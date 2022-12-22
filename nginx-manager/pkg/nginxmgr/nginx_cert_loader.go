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

type nginxCertLoader struct {
	keyPath    string
	pipePath   string
	keyContent []byte
}

// NewNginxCertLoader 创建一个nginx证书加载器
func NewNginxCertLoader(keyPath, pipePath string) *nginxCertLoader {
	loader := &nginxCertLoader{
		keyPath:  keyPath,
		pipePath: pipePath,
	}
	return loader
}

func (n *nginxCertLoader) Load() error {
	if err := n.loadKey(n.keyPath); err != nil {
		return err
	}
	if utils.IsExist(n.pipePath) {
		return nil
	}
	err := n.createPipe(n.pipePath)
	if err != nil {
		return err
	}
	go n.writeKeyToPipe(n.pipePath, n.keyContent)
	return nil
}

func (n *nginxCertLoader) loadKey(path string) error {
	encryptKeyContent, err := utils.LoadFile(path)
	if err != nil {
		return fmt.Errorf("load key file failed: %s" + err.Error())
	}
	decryptKeyByte, err := common.DecryptContent(encryptKeyContent, common.GetDefKmcCfg())
	if err != nil {
		return fmt.Errorf("decrypt key content failed: %s" + err.Error())
	}
	n.keyContent = decryptKeyByte
	return nil
}

func (n *nginxCertLoader) writeKeyToPipe(pipeFile string, content []byte) {
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

func (n *nginxCertLoader) createPipe(pipeFile string) error {
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
