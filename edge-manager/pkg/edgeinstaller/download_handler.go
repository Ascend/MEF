// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller download handler
package edgeinstaller

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type downloadHandler struct{}

// Handle configHandler handle entry
func (dh *downloadHandler) Handle(message *model.Message) error {
	hwlog.RunLog.Info("----------downloading software begin----------")
	hwlog.RunLog.Info("edge-installer received message from edge-connector success")
	downloadContent, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("convert to downloadContent failed")
		return errors.New("convert to downloadContent failed")
	}

	downloadSfwReq := DownloadSfwReqToSfwMgr{}
	if err := json.Unmarshal([]byte(downloadContent), &downloadSfwReq); err != nil {
		hwlog.RunLog.Errorf("parse to downloadSfwReq failed, error: %v", err)
		return fmt.Errorf("parse to downloadSfwReq failed, error: %v", err)
	}

	contentToConnector, err := downloadWithSfwMgr(message.GetNodeId(), downloadSfwReq)
	if err != nil {
		hwlog.RunLog.Errorf("download with software manager failed, error: %v", err)
		return fmt.Errorf("download with software manager failed, error: %v", err)
	}

	contentToConnectorAfterMarshal, err := json.Marshal(contentToConnector)
	if err != nil {
		hwlog.RunLog.Errorf("marshal content to edge-connector failed, error: %v", err)
		return fmt.Errorf("marshal content to edge-connector failed, error: %v", err)
	}

	if err = sendMessage(message, string(contentToConnectorAfterMarshal)); err != nil {
		hwlog.RunLog.Errorf("edge-installer send message to edge-connector for downloading failed, error: %v", err)
		return fmt.Errorf("edge-installer send message to edge-connector for downloading failed, error: %v", err)
	}

	hwlog.RunLog.Infof("edge-installer send message to edge-connector success with download url for downloading [%s]",
		contentToConnector.SoftwareName)
	hwlog.RunLog.Info("----------downloading software end----------")
	return nil
}
