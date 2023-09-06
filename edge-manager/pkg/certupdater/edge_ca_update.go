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
)

type edgeCaUpdater struct {
}

// indicate update operation is stopped by normal way, not an error.
var edgeCaNormalStopFlag bool
var edgeCaCertUpdateFlag int64 = 0
var nodesChangeForCaChan = make(chan changedNodeInfo, workingQueueSize)
var updateResultForCaChan = make(chan NodeCertUpdateResult, common.MaxNode)
var forceUpdateCaCertChan = make(chan CertUpdatePayload)
var edgeCaWorkingLocker = sync.Mutex{}
var edgeCaUpdaterInstance edgeCaUpdater

// StartEdgeCaCertUpdate  entry for edge root ca cert update operation
func StartEdgeCaCertUpdate(payload *CertUpdatePayload) {
	// force update way: background updating jod gets the force signal, do force update process
	if payload.ForceUpdate {
		forceUpdateCaCertChan <- *payload
		hwlog.RunLog.Info("MEF Edge ca certs will be updated by force way")
		if atomic.LoadInt64(&edgeCaCertUpdateFlag) == InRunning {
			return
		}
	}

	// if not a force update, go normal update way
	if !atomic.CompareAndSwapInt64(&edgeCaCertUpdateFlag, NotRunning, InRunning) {
		hwlog.RunLog.Warn("MEF Edge ca certs is in updating, try it later")
		return
	}
	hwlog.RunLog.Info("Start to update MEF Edge ca certs")
	// reset exit flag to default state
	edgeCaNormalStopFlag = false
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		select {
		case <-ctx.Done():
			atomic.StoreInt64(&edgeCaCertUpdateFlag, NotRunning)
			hwlog.RunLog.Info("MEF Edge ca certs update process is finished")
		}
	}()

	// drop and create database table "edge_ca_cert_status" for each ca cert update operation
	if err := RebuildDBTable(edgeCaCertStatus{}); err != nil {
		hwlog.RunLog.Errorf("rebuild database table %v error: %v", TableEdgeCaCertStatus, err)
		cancel()
		return
	}

	// retrieve nodes info from node-manager, save them as init status data to database
	if err := edgeCaUpdaterInstance.initEdgeNodesInfo(); err != nil {
		hwlog.RunLog.Errorf("init edge node info error: %v", err)
		cancel()
		return
	}

	//  [ASYNC] listen and sync nodes info from node-manager with local database
	go edgeCaUpdaterInstance.syncNodesInfo(ctx)

	// notify nginx-manager to update it's south tls cert config
	if err := notifyCertUpdateToNginxMgr(payload); err != nil {
		hwlog.RunLog.Errorf("send cert update notify to nginx-manager error: %v", err)
		cancel()
		return
	}

	// [ASYNC] update status field in database when receive update results from edge nodes
	go edgeCaUpdaterInstance.syncUpdateResult(ctx)

	// notify all edge nodes to update them service certs
	if err := edgeCaUpdaterInstance.notifyCertUpdateToEdgeNodes(payload); err != nil {
		hwlog.RunLog.Errorf("send service certs update notify to edge nodes error: %v", err)
		cancel()
		return
	}

	// [ASYNC] process failed update operation
	go edgeCaUpdaterInstance.handleFailedTask(ctx)

	// [ASYNC] check and stop update operation when stop condition is satisfied
	go edgeCaUpdaterInstance.exitConditionCheck(ctx, cancel, payload)
}

func (ea *edgeCaUpdater) initEdgeNodesInfo() error {
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
	nodeStatusData := make([]edgeCaCertStatus, 0)
	for _, node := range allNodes {
		nodeStatusData = append(nodeStatusData, edgeCaCertStatus{
			Sn:     node.SerialNumber,
			Ip:     node.IP,
			Status: UpdateStatusInit,
		})
	}
	if err = getEdgeCaCertStatusDbModInstance().CreateMultipleRecords(nodeStatusData); err != nil {
		hwlog.RunLog.Errorf("init nodes info status to db error: %v", err)
		return fmt.Errorf("init nodes info status to db error: %v", err)
	}
	return nil
}

// track node changes from node-manager
func (ea *edgeCaUpdater) syncNodesInfo(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("stop node info sync job")
			return
		case changedData := <-nodesChangeForCaChan:
			if len(changedData.DeletedNodeInfo) == 0 && len(changedData.AddedNodeInfo) == 0 {
				hwlog.RunLog.Warn("no changed node info found. skip current sync job")
				continue
			}
			hwlog.RunLog.Info("start to sync node info")
			edgeCaWorkingLocker.Lock()
			if err := ea.processAddedNodesInfo(changedData.AddedNodeInfo); err != nil {
				hwlog.RunLog.Errorf("sync added node info error: %v", err)
			}
			if err := ea.processDeletedNodesInfo(changedData.DeletedNodeInfo); err != nil {
				hwlog.RunLog.Errorf("sync deleted node info error: %v", err)
			}
			edgeCaWorkingLocker.Unlock()
			hwlog.RunLog.Info("sync node info is finished")
		}
	}
}

func (ea *edgeCaUpdater) processAddedNodesInfo(nodesInfo []NodeInfo) error {
	if len(nodesInfo) == 0 {
		return nil
	}
	addedNodeData := make([]edgeCaCertStatus, 0)
	for _, info := range nodesInfo {
		tempNodeInfo := edgeCaCertStatus{
			Sn:              info.Sn,
			Ip:              info.Ip,
			Status:          UpdateStatusInit,
			NotifyTimestamp: time.Now().Unix(),
		}
		if err := sendCertUpdateNotifyToNode(info.Sn, CertTypeEdgeCa); err != nil {
			tempNodeInfo.Status = UpdateStatusFail
			hwlog.RunLog.Errorf("send cert update notify to edge [%v] error: %v", info.Sn, err)
		}
		addedNodeData = append(addedNodeData, tempNodeInfo)
	}
	return getEdgeCaCertStatusDbModInstance().CreateMultipleRecords(addedNodeData)
}

func (ea *edgeCaUpdater) processDeletedNodesInfo(nodesInfo []NodeInfo) error {
	if len(nodesInfo) == 0 {
		return nil
	}
	deletedNodeIds := make([]string, 0)
	for _, info := range nodesInfo {
		deletedNodeIds = append(deletedNodeIds, info.Sn)
	}
	return getEdgeCaCertStatusDbModInstance().DeleteRecordsBySns(deletedNodeIds)
}

func (ea *edgeCaUpdater) notifyCertUpdateToEdgeNodes(payload *CertUpdatePayload) error {
	nodesInfo, err := getEdgeCaCertStatusDbModInstance().QueryInitRecords()
	if err != nil {
		hwlog.RunLog.Errorf("query init state nodes info error: %v", err)
		return fmt.Errorf("query init state nodes info error")
	}
	if len(nodesInfo) == 0 {
		hwlog.RunLog.Error("no nodes info found in database, cert update operation will be aborted")
		return fmt.Errorf("no nodes info found in database, cert update operation will be aborted")
	}
	updatePayload, err := json.Marshal(payload)
	if err != nil {
		hwlog.RunLog.Errorf("serialize cert update payload error: %v", err)
		return fmt.Errorf("serialize cert update payload error")
	}
	for _, info := range nodesInfo {
		dbFields := map[string]interface{}{}
		if err = sendCertUpdateNotifyToNode(info.Sn, string(updatePayload)); err != nil {
			hwlog.RunLog.Errorf("send cert update notify to node [%s] error: %v", info.Sn, err)
			dbFields["status"] = UpdateStatusFail
		}
		dbFields["notify_timestamp"] = time.Now().Unix()
		if err = getEdgeCaCertStatusDbModInstance().UpdateRecordsBySns([]string{info.Sn}, dbFields); err != nil {
			hwlog.RunLog.Errorf("update node [%s] info in database error: %v", info.Sn, err)
		}
		time.Sleep(notifyInterval)
	}
	return nil
}

func (ea *edgeCaUpdater) handleFailedTask(ctx context.Context) {
	// wait for all nodes are notified at least one time
	initStateNodes, err := getEdgeCaCertStatusDbModInstance().QueryInitRecords()
	if err != nil {
		hwlog.RunLog.Errorf("query init state edge nodes error: %v", err)
	}
	waitDuration := time.Duration(len(initStateNodes)) * notifyInterval
	time.Sleep(waitDuration)

	if err = ea.processFailedRecords(); err != nil {
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
			if err = ea.processFailedRecords(); err != nil {
				hwlog.RunLog.Errorf("process failed operation nodes error: %v", err)
			}
		}
	}
}

// processFailedRecords:  re-send update notify, update database records
func (ea *edgeCaUpdater) processFailedRecords() error {
	failedRecords, err := getEdgeCaCertStatusDbModInstance().QueryUnsuccessfulRecords()
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
		if err = sendCertUpdateNotifyToNode(info.Sn, CertTypeEdgeCa); err != nil {
			hwlog.RunLog.Errorf("re-send update notify to node [%s] error: %v", info.Sn, err)
		}
		dbFields["notify_timestamp"] = time.Now().Unix()
		if err = getEdgeCaCertStatusDbModInstance().UpdateRecordsBySns([]string{info.Sn}, dbFields); err != nil {
			hwlog.RunLog.Errorf("update node [%s] notify time error: %v", info.Sn, err)
		}
	}
	return nil
}

func (ea *edgeCaUpdater) exitConditionCheck(ctx context.Context, cf context.CancelFunc, payload *CertUpdatePayload) {
	updateResult := &FinalUpdateResult{
		CertType:   CertTypeEdgeCa,
		ResultCode: updateSuccessCode,
	}
	finalCertPayload := payload
	ticker := time.NewTicker(common.HalfDay)
	defer func() {
		ticker.Stop()
		edgeCaUpdatePostProcess(finalCertPayload, cf)
	}()
	for {
		select {
		// only exitConditionCheck function can trigger exit condition and call cf function.
		// if cf function is called on other place, it will be treated as an error.
		case <-ctx.Done():
			hwlog.RunLog.Error("cert update operation exit condition check job is aborted")
			return
		case _, ok := <-ticker.C:
			if !ok {
				hwlog.RunLog.Error("cert update operation exit condition check job timer is stopped")
				return
			}
			failedRecords, err := getEdgeCaCertStatusDbModInstance().QueryUnsuccessfulRecords()
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
			edgeCaNormalStopFlag = true
			return
		case data := <-forceUpdateCaCertChan:
			// force update chan is filled with data, cert update operation is finished by force way.
			hwlog.RunLog.Warn("root cert [hub_svr] will be expired soon, do force update process now")
			finalCertPayload = &data
			edgeCaNormalStopFlag = true
			return
		}
	}
}

func (ea *edgeCaUpdater) syncUpdateResult(ctx context.Context) {
	dbFields := make(map[string]interface{})
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Infof("stop update result sync job")
			return
		case result := <-updateResultForCaChan:
			dbFields["status"] = UpdateStatusSuccess
			if result.ResultCode != UpdateStatusSuccess {
				dbFields["status"] = UpdateStatusFail
				hwlog.RunLog.Errorf("node [%v] reports failed result: %v", result.Sn, result.Desc)
			}
			if err := getEdgeCaCertStatusDbModInstance().UpdateRecordsBySns([]string{result.Sn}, dbFields); err != nil {
				hwlog.RunLog.Errorf("update node [%v] status in database error: %v", result.Sn, err)
			}
		}
	}
}
