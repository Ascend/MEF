// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater dynamic update cloudhub server's tls ca and service certs
package certupdater

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"nginx-manager/pkg/nginxcom"
	"nginx-manager/pkg/nginxmgr"
)

var updateFilePaths = []string{
	nginxcom.SouthAuthCertFile,
	nginxcom.SouthAuthCertKeyFile,
	nginxcom.WebsocketCertFile,
	nginxcom.WebsocketCertKeyFile}

func updateSouthSvcCert(payload *CertUpdatePayload) error {
	var optErr error
	if payload.ForceUpdate {
		if err := processFiles(updateFilePaths, removeBackup); err != nil {
			optErr = fmt.Errorf("remove backup file error: %v", err)
			hwlog.RunLog.Error(optErr)
			return optErr
		}
	} else {
		if err := processFiles(updateFilePaths, doBackup); err != nil {
			optErr = fmt.Errorf("backup file error: %v", err)
			hwlog.RunLog.Error(optErr)
			return optErr
		}
	}
	if err := processFiles(updateFilePaths, setWritePermission); err != nil {
		optErr = fmt.Errorf("set file write permission error: %v", err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	defer func() {
		if err := processFiles(updateFilePaths, clearWritePermission); err != nil {
			hwlog.RunLog.Errorf("clear file write permission error: %v", err)
		}
	}()
	if err := nginxmgr.PrepareServiceCert(nginxcom.SouthAuthCertKeyFile, nginxcom.SouthAuthCertFile,
		common.WsSerName, true, &nginxReloadLocker); err != nil {
		optErr = fmt.Errorf("prepare service cert [%v] error: %v", nginxcom.SouthAuthCertFile, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	hwlog.RunLog.Info("update south auth certs success")
	if err := nginxmgr.PrepareServiceCert(nginxcom.WebsocketCertKeyFile, nginxcom.WebsocketCertFile,
		common.WsSerName, true, &nginxReloadLocker); err != nil {
		optErr = fmt.Errorf("prepare service cert [%v] error: %v", nginxcom.WebsocketCertFile, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	hwlog.RunLog.Info("update south auth and websocket service certs success")
	if err := reloadNginxConf(); err != nil {
		optErr = fmt.Errorf("reload nginx configuration error: %v", err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	hwlog.RunLog.Info("reload nginx configuration success")
	return nil
}
