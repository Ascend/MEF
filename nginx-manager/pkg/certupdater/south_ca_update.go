// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater dynamic update cloudhub server's tls ca and service certs
package certupdater

import (
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"nginx-manager/pkg/nginxcom"
)

func updateSouthCaCert(payload *CertUpdatePayload) error {
	var optErr error
	newCaCert := payload.CaContent
	if newCaCert == "" {
		optErr = fmt.Errorf("no invalid ca cert content")
		hwlog.RunLog.Error(optErr)
		return optErr
	}

	certData, err := fileutils.LoadFile(nginxcom.SouthernCertFile)
	if err != nil {
		optErr = fmt.Errorf("load south ca cert error: %v", err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	// if force flag is true (force update or finished), use new ca cert only, otherwise use both new and old cert
	if payload.ForceUpdate {
		certData = []byte(newCaCert)
		hwlog.RunLog.Info("write final south root ca cert data")
	} else {
		certData = append(certData, newCaCert...)
		hwlog.RunLog.Info("append temporary south root ca cert data")
	}

	if err = fileutils.WriteData(nginxcom.SouthernCertFile, certData); err != nil {
		optErr = fmt.Errorf("write new south ca cert error: %v", err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}

	if err = reloadNginxConf(); err != nil {
		optErr = fmt.Errorf("reload nginx configuration error: %v", err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	hwlog.RunLog.Info("reload nginx configuration success")
	return nil
}
