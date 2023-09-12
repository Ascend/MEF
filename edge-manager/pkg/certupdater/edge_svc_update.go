// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater cert update control module
package certupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
)

type edgeSvcUpdater struct {
}

// indicate update operation is stopped by normal way, not an error.
var edgeSvcNormalStopFlag bool
var edgeSvcCertUpdateFlag int64 = 0
var nodesChangeForSvcChan = make(chan changedNodeInfo, workingQueueSize)
var updateResultForSvcChan = make(chan NodeCertUpdateResult, common.MaxNode)
var forceUpdateSvcCertChan = make(chan CertUpdatePayload)
var edgeSvcWorkingLocker = sync.Mutex{}
var edgeSvcUpdaterInstance edgeSvcUpdater

// StartEdgeSvcCertUpdate  entry for edge service cert update operation
func StartEdgeSvcCertUpdate(payload *CertUpdatePayload) {
	// force update way: background updating jod gets the force signal, do force update process
	if continueRun := sendForceUpdateSignal(payload); !continueRun {
		return
	}

	// if not a force update, go normal update way
	if !atomic.CompareAndSwapInt64(&edgeSvcCertUpdateFlag, NotRunning, InRunning) {
		hwlog.RunLog.Warnf("MEF Edge service certs is in updating, try it later")
		return
	}
	hwlog.RunLog.Info("Start to update MEF Edge service certs")
	// set an alarm when cert update process starts
	// edge service cert and south ca cert are verification pair
	if err := sendAlarm(alarms.MEFCenterCaCertAbnormal, alarms.AlarmFlag); err != nil {
		hwlog.RunLog.Errorf("send cert [%v] update alarm error: %v", payload.CertType, err)
	}

	// reset exit flag to default state
	edgeSvcNormalStopFlag = false
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		select {
		case _, _ = <-ctx.Done():
			atomic.StoreInt64(&edgeSvcCertUpdateFlag, NotRunning)
			hwlog.RunLog.Info("MEF Edge service certs update process is finished")
			// clear the alarm when cert update process is finished
			// edge service cert and south ca cert are verification pair
			if err := sendAlarm(alarms.MEFCenterCaCertAbnormal, alarms.ClearFlag); err != nil {
				hwlog.RunLog.Errorf("clear cert [%v] update alarm error: %v", payload.CertType, err)
			}
		}
	}()

	// drop and create database table "edge_svc_cert_status" for each service cert update operation
	if err := RebuildDBTable(edgeSvcCertStatus{}); err != nil {
		hwlog.RunLog.Errorf("rebuild database table %v error: %v", TableEdgeSvcCertStatus, err)
		cancel()
		return
	}

	// retrieve nodes info from node-manager, save them as init status data to database
	if err := edgeSvcUpdaterInstance.initEdgeNodesInfo(); err != nil {
		hwlog.RunLog.Errorf("init edge nodes error: %v", err)
		cancel()
		return
	}

	//  [ASYNC] listen and sync nodes info from node-manager with local database
	go edgeSvcUpdaterInstance.syncNodesInfo(ctx)

	// notify nginx-manager to update it's south tls cert config
	if err := notifyCertUpdateToNginxMgr(payload); err != nil {
		hwlog.RunLog.Errorf("send cert update notify to nginx manager error: %v", err)
		cancel()
		return
	}

	// [ASYNC] update status field in database when receive update results from edge nodes
	go edgeSvcUpdaterInstance.syncUpdateResult(ctx)

	// notify all edge nodes to update them root ca certs
	if err := edgeSvcUpdaterInstance.notifyCertUpdateToEdgeNodes(payload); err != nil {
		hwlog.RunLog.Errorf("send root ca certs update notify to edge error: %v", err)
		cancel()
		return
	}

	// [ASYNC] process failed update operation
	go edgeSvcUpdaterInstance.handleFailedTask(ctx)

	// [ASYNC] check and stop update operation when stop condition is satisfied
	go edgeSvcUpdaterInstance.exitConditionCheck(ctx, cancel, payload)
}

func (es *edgeSvcUpdater) initEdgeNodesInfo() error {
	allNodes, err := getAllNodeInfo()
	if err != nil {
		return fmt.Errorf("get all node info error: %v", err)
	}
	if len(allNodes) == 0 {
		return fmt.Errorf("no edge node info found")
	}

	if len(allNodes) > common.MaxNode {
		return fmt.Errorf("edge nodes number [%v] exceeds limit [%v]", len(allNodes), common.MaxNode)
	}
	nodeStatusData := make([]edgeSvcCertStatus, 0)
	for _, node := range allNodes {
		nodeStatusData = append(nodeStatusData, edgeSvcCertStatus{
			Sn:     node.SerialNumber,
			Ip:     node.IP,
			Status: UpdateStatusInit,
		})
	}
	if err = getEdgeSvcCertStatusModInstance().CreateMultipleRecords(nodeStatusData); err != nil {
		hwlog.RunLog.Errorf("init nodes info status to db error: %v", err)
		return fmt.Errorf("init nodes info status to db error: %v", err)
	}
	return nil
}

// track node changes from node-manager
func (es *edgeSvcUpdater) syncNodesInfo(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Infof("stop node info sync job")
			return
		case changedData := <-nodesChangeForSvcChan:
			hwlog.RunLog.Info("start to sync node info for edge service cert")
			if len(changedData.DeletedNodeInfo) == 0 && len(changedData.AddedNodeInfo) == 0 {
				hwlog.RunLog.Warn("no changes node info found. skip current sync job")
				continue
			}
			edgeSvcWorkingLocker.Lock()
			if err := es.processAddedNodesInfo(changedData.AddedNodeInfo); err != nil {
				hwlog.RunLog.Errorf("sync added node info error: %v", err)
			}
			if err := es.processDeletedNodesInfo(changedData.DeletedNodeInfo); err != nil {
				hwlog.RunLog.Errorf("sync deleted node info error: %v", err)
			}
			edgeSvcWorkingLocker.Unlock()
			hwlog.RunLog.Info("sync node info for is finished")
		}
	}
}

func (es *edgeSvcUpdater) processAddedNodesInfo(nodesInfo []NodeInfo) error {
	if len(nodesInfo) == 0 {
		return nil
	}
	addedNodeData := make([]edgeSvcCertStatus, 0)
	for _, info := range nodesInfo {
		tempNodeInfo := edgeSvcCertStatus{
			Sn:              info.Sn,
			Ip:              info.Ip,
			Status:          UpdateStatusInit,
			NotifyTimestamp: time.Now().Unix(),
		}
		if err := sendCertUpdateNotifyToNode(info.Sn, CertTypeEdgeSvc); err != nil {
			tempNodeInfo.Status = UpdateStatusFail
			hwlog.RunLog.Errorf("send cert update notify to edge [%v] error: %v", info.Sn, err)
		}
		addedNodeData = append(addedNodeData, tempNodeInfo)
	}
	return getEdgeSvcCertStatusModInstance().CreateMultipleRecords(addedNodeData)
}

func (es *edgeSvcUpdater) processDeletedNodesInfo(nodesInfo []NodeInfo) error {
	if len(nodesInfo) == 0 {
		return nil
	}
	deletedNodeIds := make([]string, 0)
	for _, info := range nodesInfo {
		deletedNodeIds = append(deletedNodeIds, info.Sn)
	}
	return getEdgeSvcCertStatusModInstance().DeleteRecordsBySns(deletedNodeIds)
}

func (es *edgeSvcUpdater) notifyCertUpdateToEdgeNodes(payload *CertUpdatePayload) error {
	initNodesInfo, err := getEdgeSvcCertStatusModInstance().QueryInitRecords()
	if err != nil {
		hwlog.RunLog.Errorf("query init state nodes info error: %v", err)
		return fmt.Errorf("query init state nodes info error")
	}
	if len(initNodesInfo) == 0 {
		hwlog.RunLog.Error("no nodes info found in database, cert update operation will be aborted")
		return fmt.Errorf("no nodes info found in database, cert update operation will be aborted")
	}
	// ignore redundancy field without define a new struct
	payloadData := map[string]string{
		"certType": payload.CertType,
	}
	updatePayload, err := json.Marshal(payloadData)
	if err != nil {
		hwlog.RunLog.Errorf("serialize cert update payload error: %v", err)
		return fmt.Errorf("serialize cert update payload error")
	}
	for _, info := range initNodesInfo {
		dbFields := make(map[string]interface{})
		if err = sendCertUpdateNotifyToNode(info.Sn, string(updatePayload)); err != nil {
			hwlog.RunLog.Errorf("send cert update notify to node [%s] error: %v", info.Sn, err)
			dbFields["status"] = UpdateStatusFail
		}
		dbFields["notify_timestamp"] = time.Now().Unix()
		if err = getEdgeSvcCertStatusModInstance().UpdateRecordsBySns([]string{info.Sn}, dbFields); err != nil {
			hwlog.RunLog.Errorf("update node [%s] info in database error: %v", info.Sn, err)
		}
		time.Sleep(notifyInterval)
	}
	return nil
}

func (es *edgeSvcUpdater) handleFailedTask(ctx context.Context) {
	// wait for all nodes are notified at least one time
	initStateNodes, err := getEdgeSvcCertStatusModInstance().QueryInitRecords()
	if err != nil {
		hwlog.RunLog.Errorf("query init state edge nodes error: %v", err)
	}
	waitDuration := time.Duration(len(initStateNodes)) * notifyInterval
	time.Sleep(waitDuration)

	if err = es.processFailedRecords(); err != nil {
		hwlog.RunLog.Errorf("process failed operation nodes error: %v", err)
	}

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warnf("stop check failed records")
			return
		case _, ok := <-ticker.C:
			if !ok {
				hwlog.RunLog.Warnf("failed-records check job timer is stopped")
				return
			}
			if err = es.processFailedRecords(); err != nil {
				hwlog.RunLog.Errorf("process failed operation nodes error: %v", err)
			}
		}
	}
}

// processFailedRecords:  re-send update notify, update database records
func (es *edgeSvcUpdater) processFailedRecords() error {
	failedRecords, err := getEdgeSvcCertStatusModInstance().QueryUnsuccessfulRecords()
	if err != nil {
		hwlog.RunLog.Errorf("query failed cert update records error: %v", err)
		return fmt.Errorf("query failed cert update records error: %v", err)
	}
	if len(failedRecords) == 0 {
		return nil
	}
	dbFields := map[string]interface{}{
		"status": UpdateStatusFail,
	}
	for _, info := range failedRecords {
		checkTime := time.Now().Unix()
		duration := checkTime - info.NotifyTimestamp
		if duration < 0 {
			hwlog.RunLog.Errorf("node [%v] timeout check exception. "+
				"check time: %v, notify time: %v", info.Sn, checkTime, info.NotifyTimestamp)
			continue
		}
		if time.Duration(duration)*time.Second < updateCertTimeout {
			continue
		}
		// if duration is >= timeout( 10min), then re-send update notify to edge, then update notify timestamp
		if err = sendCertUpdateNotifyToNode(info.Sn, CertTypeEdgeSvc); err != nil {
			hwlog.RunLog.Errorf("re-send update notify to node [%s] error: %v", info.Sn, err)
		}
		dbFields["notify_timestamp"] = time.Now().Unix()
		if err = getEdgeSvcCertStatusModInstance().UpdateRecordsBySns([]string{info.Sn}, dbFields); err != nil {
			hwlog.RunLog.Errorf("update node [%s] notify time error: %v", info.Sn, err)
		}
	}
	return nil
}

func (es *edgeSvcUpdater) exitConditionCheck(ctx context.Context, cf context.CancelFunc, payload *CertUpdatePayload) {
	updateResult := &FinalUpdateResult{
		CertType:   CertTypeEdgeSvc,
		ResultCode: updateSuccessCode,
	}
	finalCertPayload := payload
	ticker := time.NewTicker(common.HalfDay)
	defer func() {
		ticker.Stop()
		edgeSvcUpdatePostProcess(finalCertPayload, cf)
	}()
	for {
		select {
		// only exitConditionCheck function can trigger exit condition and call cancel function.
		// if cancel function is called on other place, it will be treated as an error.
		case <-ctx.Done():
			hwlog.RunLog.Error("cert update operation exit condition check job is aborted")
			return
		case _, ok := <-ticker.C:
			if !ok {
				hwlog.RunLog.Error("cert update operation exit condition check job timer is stopped")
				return
			}
			failedRecords, err := getEdgeSvcCertStatusModInstance().QueryUnsuccessfulRecords()
			if err != nil {
				hwlog.RunLog.Errorf("query cert update unsuccessful records error: %v", err)
				continue
			}
			if len(failedRecords) > 0 {
				continue
			}
			// all records are successful status, cert update operation is finished by normal way.
			// report update result to cert-manger
			if err = reportUpdateResult(updateResult); err != nil {
				hwlog.RunLog.Errorf("report cert update result to cert-manager error: %v", err)
			}
			edgeSvcNormalStopFlag = true
			return
		case data := <-forceUpdateSvcCertChan:
			// force update chan is filled with data, cert update operation is finished by force way.
			hwlog.RunLog.Warn("root cert [hub_client] will be expired soon, do force update process now")
			finalCertPayload = &data
			edgeSvcNormalStopFlag = true
			return
		}
	}
}

func (es *edgeSvcUpdater) syncUpdateResult(ctx context.Context) {
	dbFields := make(map[string]interface{})
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Infof("stop update result sync job")
			return
		case result := <-updateResultForSvcChan:
			dbFields["status"] = UpdateStatusSuccess
			if result.ResultCode != UpdateStatusSuccess {
				dbFields["status"] = UpdateStatusFail
				hwlog.RunLog.Errorf("node [%v] reports failed result: %v", result.Sn, result.Desc)
			}
			if err := getEdgeSvcCertStatusModInstance().UpdateRecordsBySns([]string{result.Sn}, dbFields); err != nil {
				hwlog.RunLog.Errorf("update node [%v] status in database error: %v", result.Sn, err)
			}
		}
	}
}
