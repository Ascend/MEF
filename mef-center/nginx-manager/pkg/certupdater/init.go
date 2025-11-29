// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater dynamic update cloudhub server's tls ca and service certs
package certupdater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	messageHandlerMap[common.OptPost+common.ResEdgeMgrCertUpdate] = handleCertUpdate
}

// SouthCertUpdater dynamic update tls certs
type SouthCertUpdater struct {
	ctx    context.Context
	enable bool
}

// NewSouthCertUpdater new south cert updater instance
func NewSouthCertUpdater(enable bool, ctx context.Context) *SouthCertUpdater {
	return &SouthCertUpdater{
		ctx:    ctx,
		enable: enable,
	}
}

// Name  the unique name of this module
func (updater *SouthCertUpdater) Name() string {
	return common.CertUpdaterName
}

// Enable indicates whether this module is enabled
func (updater *SouthCertUpdater) Enable() bool {
	return updater.enable
}

// Start initializes the websocket server
func (updater *SouthCertUpdater) Start() {
	initMsgHandler()
	for {
		select {
		case <-updater.ctx.Done():
			hwlog.RunLog.Info("nginx south cert updater is stopped")
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

func (updater *SouthCertUpdater) dispatch(msg *model.Message) {
	msgHandler, ok := getMsgHandler(msg)
	if !ok {
		hwlog.RunLog.Errorf("no handler found for message: [option: %v  resource: %v]",
			msg.GetOption(), msg.GetResource())
		return
	}
	if err := msgHandler(msg); err != nil {
		hwlog.RunLog.Errorf("handle message:[option: %v  resource: %v] error: %v",
			msg.GetOption(), msg.GetResource(), err)
	}
}

func handleCertUpdate(msg *model.Message) error {
	respMsg, err := msg.NewResponse()
	if err != nil {
		return fmt.Errorf("create response message error: %v", err)
	}
	var updateErr error
	defer func() {
		fillErr := respMsg.FillContent(common.OK)
		if updateErr != nil {
			fillErr = respMsg.FillContent(updateErr.Error())
		}
		if fillErr != nil {
			hwlog.RunLog.Errorf("fill content into resp failed: %v", err)
		}
		if err = modulemgr.SendMessage(respMsg); err != nil {
			hwlog.RunLog.Errorf("send response message failed: %v", err)
		}
	}()

	var rawContent []byte
	if err := msg.ParseContent(&rawContent); err != nil {
		updateErr = errors.New("message content type error, expect byte slice")
		hwlog.RunLog.Error(updateErr)
		return updateErr
	}
	var payload CertUpdatePayload
	if err = json.Unmarshal(rawContent, &payload); err != nil {
		updateErr = fmt.Errorf("deseliarize message content error: %v", err)
		hwlog.RunLog.Error(updateErr)
		return updateErr
	}
	switch payload.CertType {
	case CertTypeEdgeSvc:
		if err = updateSouthCaCert(&payload); err != nil {
			updateErr = fmt.Errorf("update nginx south ca cert error: %v", err)
			hwlog.RunLog.Error(updateErr)
			return updateErr
		}
	case CertTypeEdgeCa:
		if err = updateSouthSvcCert(&payload); err != nil {
			updateErr = fmt.Errorf("update nginx south service cert error: %v", err)
			hwlog.RunLog.Error(updateErr)
			return updateErr
		}
	default:
		updateErr = fmt.Errorf("cert [%v] update is not supported", payload.CertType)
		hwlog.RunLog.Error(updateErr)
		return updateErr
	}
	hwlog.RunLog.Infof("update cert [%v] success", payload.CertType)
	return nil
}
