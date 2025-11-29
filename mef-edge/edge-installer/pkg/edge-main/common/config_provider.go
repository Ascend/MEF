// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common this file for config operation
package common

import (
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

// WsServerCertBytes websocket server certs data, caution: key content is plain
type WsServerCertBytes struct {
	RootCaContent []byte `json:"rootca_content"`
}

// all peer certs subdirectory,
var peerCertSubDirs = []string{
	constants.MindXOMDir,
}

// GetWsCertContent get all cert, both inner and peer root ca.
func GetWsCertContent() (*WsServerCertBytes, error) {
	innerCertDir, err := path.GetCompSpecificDir(constants.InnerCertPathName)
	if err != nil {
		return nil, err
	}
	innerRootCaPath := filepath.Join(innerCertDir, constants.RootCaName)

	rootCaContent, err := certutils.GetCertContentWithBackup(innerRootCaPath)
	if err != nil {
		return nil, err
	}
	peerCertMainDir, err := path.GetCompSpecificDir(constants.PeerCerts)
	if err != nil {
		return nil, err
	}
	for _, subDir := range peerCertSubDirs {
		peerCertPath := filepath.Join(peerCertMainDir, subDir, constants.RootCaName)
		if !fileutils.IsExist(peerCertPath) && !fileutils.IsExist(peerCertPath+backuputils.BackupSuffix) {
			continue
		}
		peerCaContent, err := certutils.GetCertContentWithBackup(peerCertPath)
		if err != nil {
			return nil, err
		}
		rootCaContent = append(rootCaContent, peerCaContent...)
	}
	return &WsServerCertBytes{
		RootCaContent: rootCaContent,
	}, nil
}
