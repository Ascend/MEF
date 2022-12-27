// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"bytes"
	"fmt"

	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"

	"nginx-manager/pkg/checker"
	"nginx-manager/pkg/nginxcom"
)

type nginxConfUpdater struct {
	confItems []nginxcom.NginxConfItem
	confPath  string
}

// NewNginxConfUpdater 创建一个nginx配置文件修改器
func NewNginxConfUpdater(confItems []nginxcom.NginxConfItem, confPath string) (*nginxConfUpdater, error) {
	return &nginxConfUpdater{
		confItems: confItems,
		confPath:  confPath,
	}, nil
}

func loadConf(path string) ([]byte, error) {
	b, err := utils.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file. error:%s", err.Error())
	}
	return b, nil
}

func (n *nginxConfUpdater) calculatePipeCount() (int, error) {
	content, err := loadConf(n.confPath)
	if err != nil {
		return 0, err
	}
	return bytes.Count(content, []byte(nginxcom.ClientPipePrefix)), nil
}

// Update 修改配置文件
func (n *nginxConfUpdater) Update() error {
	content, err := loadConf(n.confPath)
	if err != nil {
		return err
	}
	err = checker.Check(checker.NginxConfig, content)
	if err != nil {
		return err
	}

	return n.updateUrl(content)
}

func (n *nginxConfUpdater) updateUrl(content []byte) error {
	for _, conf := range n.confItems {
		content = bytes.ReplaceAll(content, []byte(conf.From), []byte(conf.To))
	}
	err := common.WriteData(nginxcom.NginxConfigPath, content)
	if err != nil {
		return fmt.Errorf("writeFile failed. error:%s", err.Error())
	}
	return nil
}
