// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"fmt"
	"os"
	"syscall"

	"nginx-manager/pkg/nginxcom"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

// Load 加载北向密钥文件并写入pipe
func Load(keyPath, pipePath string) error {
	var keyContent []byte
	var err error
	if keyContent, err = loadKey(keyPath); err != nil {
		return err
	}
	err = utils.MakeSureDir(pipePath)
	if err != nil {
		return err
	}
	if utils.IsExist(pipePath) {
		return nil
	}
	err = createPipe(pipePath)
	if err != nil {
		return err
	}
	go writeKeyToPipe(pipePath, keyContent)
	return nil
}

// LoadForClient 加载用于内部转发的密钥文件并写入pipe
func LoadForClient(keyPath, pipeDir string, pipeCount int) error {
	var keyContent []byte
	var err error
	if keyContent, err = loadKey(keyPath); err != nil {
		return err
	}
	var pipePaths []string
	for i := 0; i < pipeCount; i++ {
		pipePath := fmt.Sprintf("%s%s_%d", pipeDir, nginxcom.ClientPipePrefix, i)
		pipePaths = append(pipePaths, pipePath)
	}

	for _, pipePath := range pipePaths {
		if utils.IsExist(pipePath) {
			return nil
		}
		err = createPipe(pipePath)
		if err != nil {
			return err
		}
	}
	for _, pipePath := range pipePaths {
		go writeKeyToPipe(pipePath, keyContent)
	}
	return nil
}

func loadKey(path string) ([]byte, error) {
	encryptKeyContent, err := utils.LoadFile(path)
	if encryptKeyContent == nil {
		return nil, fmt.Errorf("load key path [%s] file failed", path)
	}
	decryptKeyByte, err := common.DecryptContent(encryptKeyContent, common.GetDefKmcCfg())
	if err != nil {
		return nil, fmt.Errorf("decrypt key content failed: %s" + err.Error())
	}
	return decryptKeyByte, nil
}

func writeKeyToPipe(pipeFile string, content []byte) {
	pipe, err := os.OpenFile(pipeFile, os.O_WRONLY|os.O_SYNC, os.ModeNamedPipe)
	if err != nil {
		hwlog.RunLog.Error("open pipe failed")
		return
	}
	defer func() {
		err := pipe.Close()
		if err != nil {
			hwlog.RunLog.Error("pipe close error")
		}
		err = os.Remove(pipeFile)
		if err != nil {
			hwlog.RunLog.Error("pipe remove error")
		}
	}()
	_, err = pipe.Write(content)
	if err != nil {
		hwlog.RunLog.Errorf("pass key to pipe failed:%v", err)
		return
	}
	_, err = pipe.WriteString("\n")
	if err != nil {
		hwlog.RunLog.Errorf("pass key to pipe failed:%v", err)
		return
	}
	hwlog.RunLog.Infof("write pipe %s success", pipeFile)
}

func createPipe(pipeFile string) error {
	_, err := os.Stat(pipeFile)
	if os.IsExist(err) {
		hwlog.RunLog.Infof("pipe %v exists", pipeFile)
		return nil
	}
	err = syscall.Mkfifo(pipeFile, nginxcom.FifoPermission)
	if err != nil {
		return fmt.Errorf("make pipe %s error: %v", pipeFile, err)
	}
	hwlog.RunLog.Infof("make pipe %v success", pipeFile)
	return nil
}
