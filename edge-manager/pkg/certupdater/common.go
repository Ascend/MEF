// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater cert update control module
package certupdater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"huawei.com/mindxedge/base/common/requests"
)

// common definition for cert update operation
const (
	CertTypeEdgeCa          = "EdgeCa"
	CertTypeEdgeSvc         = "EdgeSvc"
	NotRunning        int64 = 0
	InRunning         int64 = 1
	notifyInterval          = time.Second * 3
	updateCertTimeout       = time.Minute * 10
	workingQueueSize        = common.MaxNode
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
	respBytes, err := httpsReq.PostJson(jsonBody)
	if err != nil {
		return err
	}
	var resp common.RespMsg
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return err
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
	notifyMsg.FillContent(updatePayload)
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
	respBytes, err := httpsReq.PostJson(jsonBody)
	if err != nil {
		return err
	}
	var resp common.RespMsg
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return err
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
	// 2. delete temporary db table after use.
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
	// 2. delete temporary db table after use.
	if err := DeleteDBTable(edgeCaCertStatus{}); err != nil {
		hwlog.RunLog.Errorf("cleanup database table [%v] error: %v", TableEdgeCaCertStatus, err)
		return
	}
	hwlog.RunLog.Infof("cleanup database table [%v] success", TableEdgeCaCertStatus)
}
