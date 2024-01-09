// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater cert update control module
package certupdater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509/certutils"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

// common definition for cert update operation
const (
	CertTypeEdgeCa           = "EdgeCa"
	CertTypeEdgeSvc          = "EdgeSvc"
	NotRunning         int64 = 0
	InRunning          int64 = 1
	httpReqTryInterval       = time.Second * 30
	httpReqTryMaxTime        = 5
	sendAlarmTryMax          = 6
	sendAlarmInterval        = time.Second * 10
	notifyInterval           = time.Second * 3
	updateCertTimeout        = time.Minute * 10
	workingQueueSize         = common.MaxNode

	failedTaskCheckInterval = time.Minute * 15
	stopCondCheckInterval   = time.Hour
)

const (
	urlReportUpdateResult       = "/inner/v1/certificates/update-result"
	urlNgxSouthCert             = "/inner/v1/ngxmanager/cert/edge-manager"
	updateSuccessCode     int64 = 1
	updateFailedCode      int64 = 2
)

// NodeInfo node info from node-manager
type NodeInfo struct {
	Sn string `json:"sn"`
	Ip string `json:"ip"`
}

type changedNodeInfo struct {
	AddedNodeInfo   []NodeInfo
	DeletedNodeInfo []NodeInfo
}

// NodeCertUpdateResult each node update result, sent by each edge node
type NodeCertUpdateResult struct {
	CertType   string `json:"certType"`
	Sn         string `json:"sn"`
	ResultCode int64  `json:"resultCode"`
	Desc       string `json:"desc"`
}

// FinalUpdateResult final update result, sent from cert-updater to cert-manager
type FinalUpdateResult struct {
	CertType   string `json:"certType"`
	ResultCode int64  `json:"resultCode"`
	Desc       string `json:"desc"`
}

// CertUpdatePayload cert update payload from cert-manager
type CertUpdatePayload struct {
	CertType    string `json:"certType"`
	ForceUpdate bool   `json:"forceUpdate"`
	CaContent   string `json:"caContent"`
}

func reportUpdateResult(result *FinalUpdateResult) error {
	reqCertParams := &requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: constants.RootCaPath,
			CertPath:   constants.ServerCertPath,
			KeyPath:    constants.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	url := fmt.Sprintf("https://%s:%d%s", common.CertMgrDns, common.CertMgrPort, urlReportUpdateResult)
	httpsReq := httpsmgr.GetHttpsReq(url, reqCertParams.ClientTlsCert)
	jsonBody, err := json.Marshal(result)
	if err != nil {
		return err
	}
	var tryCnt int
	var resp common.RespMsg
	for tryCnt = 0; tryCnt < httpReqTryMaxTime; tryCnt++ {
		respBytes, err := httpsReq.PostJson(jsonBody)
		if err != nil {
			hwlog.RunLog.Errorf("do http post request error: %v, try request for next time", err)
			time.Sleep(httpReqTryInterval)
			continue
		}
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			hwlog.RunLog.Errorf("unmarshal http body error: %v, try request for netx time", err)
			time.Sleep(httpReqTryInterval)
			continue
		}
		break
	}
	if tryCnt == httpReqTryMaxTime {
		return fmt.Errorf("report cert update result to cert manager error, please check network connection")
	}
	if resp.Status != common.Success {
		return fmt.Errorf("report cert update result error: %v", resp.Msg)
	}
	return nil
}

func getAllNodeInfo() ([]nodemanager.NodeInfo, error) {
	router := common.Router{
		Source:      common.AppManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.NodeList,
	}
	req := types.InnerGetNodeInfoResReq{
		ModuleName: common.ConfigManagerName,
	}
	resp := common.SendSyncMessageByRestful(req, &router, common.ResponseTimeout)
	if resp.Status != common.Success {
		return nil, errors.New(resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, errors.New("marshal internal response error")
	}
	var nodes []nodemanager.NodeInfo
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nodes, errors.New("unmarshal internal response error")
	}
	return nodes, nil
}

func sendCertUpdateNotifyToNode(serialNumber string, updatePayload string) error {
	notifyMsg, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("create new message failed, error: %v", err)
	}
	notifyMsg.SetNodeId(serialNumber)
	notifyMsg.SetRouter(common.CertUpdaterName, common.CloudHubName, common.OptGet, common.CertWillExpired)
	if err = notifyMsg.FillContent(updatePayload); err != nil {
		return fmt.Errorf("fill content failed: %v", err)
	}
	if err = modulemgr.SendMessage(notifyMsg); err != nil {
		return fmt.Errorf("%s sends message to %s failed, error: %v",
			common.CertUpdaterName, common.CloudHubName, err)
	}
	return nil
}

func notifyCertUpdateToNginxMgr(payload *CertUpdatePayload) error {
	reqCertParams := &requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: constants.RootCaPath,
			CertPath:   constants.ServerCertPath,
			KeyPath:    constants.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	url := fmt.Sprintf("https://%s:%d%s", common.NginxMgrDns, common.NginxMgrPort, urlNgxSouthCert)
	httpsReq := httpsmgr.GetHttpsReq(url, reqCertParams.ClientTlsCert)
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var tryCnt int
	var resp common.RespMsg
	for tryCnt = 0; tryCnt < httpReqTryMaxTime; tryCnt++ {
		respBytes, err := httpsReq.PostJson(jsonBody)
		if err != nil {
			hwlog.RunLog.Errorf("do http post request error: %v, try request for next time", err)
			time.Sleep(httpReqTryInterval)
			continue
		}
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			hwlog.RunLog.Errorf("unmarshal http body error: %v, try request for netx time", err)
			time.Sleep(httpReqTryInterval)
			continue
		}
		break
	}
	if tryCnt == httpReqTryMaxTime {
		return fmt.Errorf("send cert update notify to nginx manager error, please check network connection")
	}
	if resp.Status != common.Success {
		return fmt.Errorf("report cert update result error: %v", resp.Msg)
	}
	return nil
}

// do some cleanup jobs when edge service cert update operation is finished
func edgeSvcUpdatePostProcess(payload *CertUpdatePayload, cancel context.CancelFunc) {
	defer cancel()
	if payload.CertType != CertTypeEdgeSvc {
		hwlog.RunLog.Errorf("payload cert type error: %v expect: %v", payload.CertType, CertTypeEdgeSvc)
		return
	}
	if !edgeSvcNormalStopFlag {
		hwlog.RunLog.Warnf("cert [%v] update operation is stopped by error, skip post process", payload.CertType)
		return
	}
	// 1. notify nginx-manager to update it's south ca root cert
	payload.ForceUpdate = true
	if err := notifyCertUpdateToNginxMgr(payload); err != nil {
		hwlog.RunLog.Errorf("send cert [%v] force update notify to nginx manager error: %v", payload.CertType, err)
		return
	}
	hwlog.RunLog.Infof("send cert [%v] force update notify to nginx manager success", payload.CertType)

	// 2. send an alarm if there are some edge nodes not successfully update theirs certs
	if err := sendUpdateFailedAlarm(payload); err != nil {
		// if there is an error, don't return, keep running step 3
		hwlog.RunLog.Errorf("send cert [%v] update abnormal alarm failed: %v", payload.CertType, err)
	} else {
		hwlog.RunLog.Infof("send cert [%v] update abnormal alarm success", payload.CertType)
	}

	// 3. delete temporary db table after use.
	if err := DeleteDBTable(edgeSvcCertStatus{}); err != nil {
		hwlog.RunLog.Errorf("cleanup database table [%v] error: %v", TableEdgeSvcCertStatus, err)
		return
	}
	hwlog.RunLog.Infof("cleanup database table [%v] success", TableEdgeSvcCertStatus)

}

// do some cleanup jobs when edge root ca cert update operation is finished
func edgeCaUpdatePostProcess(payload *CertUpdatePayload, cancel context.CancelFunc) {
	defer cancel()
	if payload.CertType != CertTypeEdgeCa {
		hwlog.RunLog.Errorf("payload cert type error: %v expect: %v", payload.CertType, CertTypeEdgeCa)
		return
	}
	if !edgeCaNormalStopFlag {
		hwlog.RunLog.Warnf("cert [%v] update operation is stopped by error, skip post process", payload.CertType)
		return
	}
	// 1. notify nginx-manager to update it's south ca root cert
	payload.ForceUpdate = true
	if err := notifyCertUpdateToNginxMgr(payload); err != nil {
		hwlog.RunLog.Errorf("send cert [%v] force update notify to nginx manager error: %v", payload.CertType, err)
		return
	}
	hwlog.RunLog.Infof("send cert [%v] force update notify to nginx manager success", payload.CertType)

	// 2. send an alarm if there are some edge nodes not successfully update theirs certs
	if err := sendUpdateFailedAlarm(payload); err != nil {
		// if there is an error, don't return, keep running step 3
		hwlog.RunLog.Errorf("send cert [%v] update abnormal alarm failed: %v", payload.CertType, err)
	} else {
		hwlog.RunLog.Infof("send cert [%v] update abnormal alarm success", payload.CertType)
	}

	// 3. delete temporary db table after use.
	if err := DeleteDBTable(edgeCaCertStatus{}); err != nil {
		hwlog.RunLog.Errorf("cleanup database table [%v] error: %v", TableEdgeCaCertStatus, err)
		return
	}
	hwlog.RunLog.Infof("cleanup database table [%v] success", TableEdgeCaCertStatus)
}

func sendAlarm(alarmId, notifyType string) error {
	alarm, err := alarms.CreateAlarm(alarmId, common.CertUpdaterName, notifyType)
	if err != nil {
		hwlog.RunLog.Errorf("create alarm [%v] error: %v", alarmId, err)
		return fmt.Errorf("create alarm [%v] error: %v", alarmId, err)
	}

	hostIp, err := common.GetHostIP("NODE_IP")
	if err != nil {
		hwlog.RunLog.Errorf("get host ip failed, error: %v", err)
		return errors.New("get host ip failed")
	}

	alarmReq := requests.AddAlarmReq{
		Alarms: []requests.AlarmReq{*alarm},
		Ip:     hostIp,
	}
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create new message error: %v", err)
		return fmt.Errorf("create new message error: %v", err)
	}
	if err = msg.FillContent(alarmReq, true); err != nil {
		hwlog.RunLog.Errorf("fill alarm req into content failed: %v", err)
		return errors.New("fill alarm req into content failed")
	}
	msg.SetNodeId(common.AlarmManagerClientName)
	msg.SetRouter(common.CertUpdaterName, common.InnerServerName, common.OptPost, requests.ReportAlarmRouter)
	var tryCnt int
	for tryCnt = 0; tryCnt < sendAlarmTryMax; tryCnt++ {
		if err = modulemgr.SendAsyncMessage(msg); err != nil {
			hwlog.RunLog.Errorf("send msg to alarm inner server failed: %v, try for next time", err)
			time.Sleep(sendAlarmInterval)
			continue
		}
		break
	}
	if tryCnt == sendAlarmTryMax {
		hwlog.RunLog.Errorf("send alarm [%v] failed, please check service status", alarmId)
		return fmt.Errorf("send alarm [%v] failed, please check service status", alarmId)
	}
	return nil
}

func sendUpdateFailedAlarm(payload *CertUpdatePayload) error {
	var alarmId string
	var needSendAlarm bool
	failedSns := make([]string, 0)
	switch payload.CertType {
	case CertTypeEdgeSvc:
		failedRecords, err := getEdgeSvcCertStatusModInstance().QueryUnsuccessfulRecords()
		if err != nil {
			return fmt.Errorf("query edge service cert update failed records error: %v", err)
		}
		if len(failedRecords) > 0 {
			needSendAlarm = true
			// edge service cert and south ca cert are verification pair
			alarmId = alarms.MEFCenterCaCertUpdateAbnormal
			for _, record := range failedRecords {
				failedSns = append(failedSns, record.Sn)
			}
		}
	case CertTypeEdgeCa:
		failedRecords, err := getEdgeCaCertStatusDbModInstance().QueryUnsuccessfulRecords()
		if err != nil {
			return fmt.Errorf("query edge ca cert update failed records error: %v", err)
		}
		if len(failedRecords) > 0 {
			needSendAlarm = true
			// edge ca cert and south service cert are verification pair
			alarmId = alarms.MEFCenterSvcCertUpdateAbnormal
			for _, record := range failedRecords {
				failedSns = append(failedSns, record.Sn)
			}
		}
	default:
		return fmt.Errorf("invalid cert type for alram: %v", payload.CertType)
	}
	if needSendAlarm {
		hwlog.RunLog.Warnf("the following edge nodes are not successfully update certs, "+
			"please do net-config operation on edge nodes as soon as possible. nodes sn: [%v]", failedSns)
		if err := sendAlarm(alarmId, alarms.AlarmFlag); err != nil {
			return fmt.Errorf("send cert update abnormal alarm [id: %v] error: %v", alarmId, err)
		}
	}
	return nil
}

// return bool indicate whether outer function need continue to run.
// true: continue run, false: stop and return
func sendForceUpdateSignal(payload *CertUpdatePayload) bool {
	if payload == nil {
		hwlog.RunLog.Errorf("invalid payload data")
		return false
	}
	if !payload.ForceUpdate {
		return true
	}
	switch payload.CertType {
	case CertTypeEdgeCa:
		if forceUpdateCaCertChan == nil {
			forceUpdateCaCertChan = make(chan CertUpdatePayload)
		}
		forceUpdateCaCertChan <- *payload
		if atomic.LoadInt64(&edgeCaCertUpdateFlag) == InRunning {
			hwlog.RunLog.Info("MEF Edge ca cert will be updated by force way")
			return false
		}
		return true
	case CertTypeEdgeSvc:
		if forceUpdateSvcCertChan == nil {
			forceUpdateSvcCertChan = make(chan CertUpdatePayload)
		}
		forceUpdateSvcCertChan <- *payload
		if atomic.LoadInt64(&edgeSvcCertUpdateFlag) == InRunning {
			hwlog.RunLog.Info("MEF Edge service cert will be updated by force way")
			return false
		}
		return true
	default:
		hwlog.RunLog.Errorf("invalid cert type: %v", payload.CertType)
		return false
	}
}
