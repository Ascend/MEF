// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater cert update control module
package certupdater

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

type messageHandler func(*model.Message) error

var messageHandlerMap = make(map[string]messageHandler)

func getMsgHandler(msg *model.Message) (messageHandler, bool) {
	handlerKey := msg.GetOption() + msg.GetResource()
	handler, ok := messageHandlerMap[handlerKey]
	return handler, ok
}
func initMsgHandler() {
	messageHandlerMap[common.OptPost+common.ResCertUpdate] = handleCertUpdate
	messageHandlerMap[common.OptPost+common.ResNodeChanged] = handleNodeChange
	messageHandlerMap[common.OptResp+common.CertWillExpired] = handleUpdateResult
}

// EdgeCertUpdater dynamic update tls certs
type EdgeCertUpdater struct {
	ctx    context.Context
	enable bool
}

// NewEdgeCertUpdater new cert updater instance
func NewEdgeCertUpdater(enable bool) *EdgeCertUpdater {
	return &EdgeCertUpdater{
		ctx:    context.Background(),
		enable: enable,
	}
}

// Name  the unique name of this module
func (updater *EdgeCertUpdater) Name() string {
	return common.CertUpdaterName
}

// Enable indicates whether this module is enabled
func (updater *EdgeCertUpdater) Enable() bool {
	return updater.enable
}

// Start initializes the websocket server
func (updater *EdgeCertUpdater) Start() {
	initMsgHandler()
	for {
		select {
		case _, ok := <-updater.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("cert updater is stopped")
			return
		default:
		}

		message, err := modulemgr.ReceiveMessage(updater.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message error: %v", updater.Name(), err)
			continue
		}
		go updater.dispatch(message)
	}
}

func (updater *EdgeCertUpdater) dispatch(msg *model.Message) {
	msgHandler, ok := getMsgHandler(msg)
	if !ok {
		hwlog.RunLog.Errorf("no handler found for message:%v %v", msg.GetOption(), msg.GetResource())
		return
	}
	if err := msgHandler(msg); err != nil {
		hwlog.RunLog.Errorf("handle message:%v %v error: %v", msg.GetOption(), msg.GetResource(), err)
	}
}

func handleCertUpdate(msg *model.Message) error {
	var updateErr error
	defer func() {
		respMsg, err := msg.NewResponse()
		if err != nil {
			hwlog.RunLog.Errorf("create response message error:%v", err)
			return
		}
		respContent := common.RespMsg{
			Status: common.Success,
		}
		if updateErr != nil {
			respContent.Status = common.ErrorCertTypeError
			respContent.Msg = updateErr.Error()
		}
		respMsg.FillContent(respContent)
		if err = modulemgr.SendMessage(respMsg); err != nil {
			hwlog.RunLog.Errorf("send response message error: %v", err)
		}
	}()
	rawContent, ok := msg.GetContent().(string)
	if !ok {
		updateErr = fmt.Errorf("message content type error")
		return updateErr
	}
	var payload CertUpdatePayload
	if err := json.Unmarshal([]byte(rawContent), &payload); err != nil {
		updateErr = fmt.Errorf("parse message error: %v", err)
		return updateErr
	}
	switch payload.CertType {
	case CertTypeEdgeSvc:
		go StartEdgeSvcCertUpdate(&payload)
	case CertTypeEdgeCa:
		go StartEdgeCaCertUpdate(&payload)
	default:
		updateErr = fmt.Errorf("cert [%v] update is not supported", payload.CertType)
		return updateErr
	}
	return nil
}

func handleNodeChange(msg *model.Message) error {
	rawContent, ok := msg.GetContent().(string)
	if !ok {
		return fmt.Errorf("message content type error")
	}
	var msgData changedNodeInfo
	if err := json.Unmarshal([]byte(rawContent), &msgData); err != nil {
		return fmt.Errorf("parse message error: %v", err)
	}
	// don't do anything if cert update is not running
	if atomic.LoadInt64(&edgeSvcCertUpdateFlag) == InRunning {
		if nodesChangeForSvcChan == nil {
			nodesChangeForSvcChan = make(chan changedNodeInfo, workingQueueSize)
		}
		nodesChangeForSvcChan <- msgData
	}
	if atomic.LoadInt64(&edgeCaCertUpdateFlag) == InRunning {
		if nodesChangeForCaChan == nil {
			nodesChangeForCaChan = make(chan changedNodeInfo, workingQueueSize)
		}
		nodesChangeForCaChan <- msgData
	}
	return nil
}

func handleUpdateResult(msg *model.Message) error {
	base64msgStr, ok := msg.GetContent().(string)
	if !ok {
		return fmt.Errorf("message content error")
	}
	rawMessageContent, err := base64.StdEncoding.DecodeString(base64msgStr)
	if err != nil {
		return fmt.Errorf("decode base64 message content error: %v", err)
	}
	var msgData NodeCertUpdateResult
	if err = json.Unmarshal(rawMessageContent, &msgData); err != nil {
		return fmt.Errorf("deserialize message content error: %v", err)
	}
	if msgData.CertType != CertTypeEdgeCa && msgData.CertType != CertTypeEdgeSvc {
		return fmt.Errorf("cert type error: %v", msgData.CertType)
	}
	switch msgData.CertType {
	case CertTypeEdgeSvc:
		if atomic.LoadInt64(&edgeSvcCertUpdateFlag) != InRunning {
			hwlog.RunLog.Warnf("cert %v is not in updating state, skip process", msgData.CertType)
			return nil
		}
		if updateResultForSvcChan == nil {
			updateResultForSvcChan = make(chan NodeCertUpdateResult, common.MaxNode)
		}
		updateResultForSvcChan <- msgData
	case CertTypeEdgeCa:
		if atomic.LoadInt64(&edgeCaCertUpdateFlag) != InRunning {
			hwlog.RunLog.Warnf("cert %v is not in updating state, skip process", msgData.CertType)
			return nil
		}
		if updateResultForCaChan == nil {
			updateResultForCaChan = make(chan NodeCertUpdateResult, common.MaxNode)
		}
		updateResultForCaChan <- msgData
	default:
		return fmt.Errorf("cert type error: %v", msgData.CertType)
	}
	return nil
}
