// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
