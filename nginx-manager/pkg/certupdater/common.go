// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater dynamic update cloudhub server's tls ca and service certs
package certupdater

import (
	"fmt"
	"sync"
	"time"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"nginx-manager/pkg/nginxcom"
	"nginx-manager/pkg/nginxmgr"
)

// cert types and backup file suffix defination
const (
	backupFileSuffix       = ".bak"
	CertTypeEdgeCa         = "EdgeCa"
	CertTypeEdgeSvc        = "EdgeSvc"
	nginxReloadConfTimeout = time.Second * 20
)

// CertUpdatePayload cert update payload from edge-manager
type CertUpdatePayload struct {
	CertType    string `json:"certType"`
	ForceUpdate bool   `json:"forceUpdate"`
	CaContent   string `json:"caContent"`
}

var nginxReloadLocker sync.Mutex

func reloadNginxConf() error {
	nginxReloadLocker.Lock()
	defer nginxReloadLocker.Unlock()
	// create cert key pipes
	if err := nginxmgr.CreateKeyPipes(); err != nil {
		hwlog.RunLog.Errorf("create key pipes error: %v", err)
		return fmt.Errorf("create keys pipes failed: %v", err)
	}
	//  reload nginx config, nginx will read conf file twice, 1st for testing, 2nd for making effect
	if _, err := envutils.RunResidentCmd("./nginx", "-s", "reload", "-c",
		nginxcom.NginxConfigPath); err != nil {
		hwlog.RunLog.Errorf("reload nginx config failed: %v", err)
		return fmt.Errorf("reload nginx config failed: %v", err)
	}
	// write cert key data to pipe for the first time, DON'T delete pipes
	if err := nginxmgr.LoadKeysDataToPipes(false); err != nil {
		hwlog.RunLog.Errorf("load cert keys to pipe failed: %v", err)
		return fmt.Errorf("load cert keys to pipe failed: %v", err)
	}
	// write cert key data to pipe for the second time, delete after use.
	if err := nginxmgr.LoadKeysDataToPipes(true); err != nil {
		hwlog.RunLog.Errorf("load cert keys to pipe failed: %v", err)
		return fmt.Errorf("load cert keys to pipe failed: %v", err)
	}
	// above operations are async, nginxReloadLocker will be useless,
	// add sleep to wait for reload is finished.
	time.Sleep(nginxReloadConfTimeout)
	return nil
}

type fileProcessor func(filePath string) error

func processFiles(filePaths []string, processor fileProcessor) error {
	for _, path := range filePaths {
		if err := processor(path); err != nil {
			return err
		}
	}
	return nil
}

// create a backup file with .tmp suffix in the same directory, in case of update operation is interrupted.
func doBackup(filePath string) error {
	if !fileutils.IsExist(filePath) {
		return fmt.Errorf("source file path [%v] not exists", filePath)
	}
	backupPath := filePath + backupFileSuffix
	if err := fileutils.DeleteFile(backupPath); err != nil {
		return fmt.Errorf("remove previously created backup file [%v] error: %v", backupPath, err)
	}

	if err := fileutils.CopyFile(filePath, backupPath); err != nil {
		return fmt.Errorf("backup source file [%v] to dest file [%v] failed: %v", filePath, backupPath, err)
	}
	return nil
}

// delete backup file when operation is finished
func removeBackup(filePath string) error {
	backupPath := filePath + backupFileSuffix
	// file list contains key files, use DeleteAllFileWithConfusion instead of DeleteFile
	if err := fileutils.DeleteAllFileWithConfusion(backupPath); err != nil {
		return fmt.Errorf("remove backup file [%v] failed: %v", backupPath, err)
	}
	return nil
}

// set file mode to 600 (rw) for writing new data to it
func setWritePermission(filePath string) error {
	return fileutils.SetPathPermission(filePath, fileutils.Mode600, false, false)
}
