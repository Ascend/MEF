// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"bytes"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"nginx-manager/pkg/nginxcom"
)

const (
	nginxDefaultConfigPath = "/home/MEFCenter/conf/nginx_default.conf"
)

type nginxConfUpdater struct {
	confItems []nginxcom.NginxConfItem
	confPath  string
}

// NewNginxConfUpdater create an updater to modify nginx configuration file
func NewNginxConfUpdater(confItems []nginxcom.NginxConfItem) (*nginxConfUpdater, error) {
	config := nginxConfUpdater{}
	config.confPath = nginxDefaultConfigPath
	config.confItems = confItems
	return &config, nil
}

func loadConf(path string) ([]byte, error) {
	b, err := fileutils.LoadFile(path)
	if err != nil {
		hwlog.RunLog.Errorf("failed to read file. error:%s", err.Error())
		return nil, fmt.Errorf("failed to read file. error:%s", err.Error())
	}
	return b, nil
}

func calculatePipeCount(pipeType string) (int, error) {
	content, err := loadConf(nginxcom.NginxConfigPath)
	if err != nil {
		return 0, err
	}
	return bytes.Count(content, []byte(pipeType)), nil
}

// Update do the modify nginx configuration file job
func (n *nginxConfUpdater) Update() error {
	content, err := loadConf(n.confPath)
	if err != nil {
		return err
	}
	for _, conf := range n.confItems {
		content = bytes.ReplaceAll(content, []byte(conf.From), []byte(conf.To))
	}

	err = fileutils.WriteData(nginxcom.NginxConfigPath, content)
	if err != nil {
		hwlog.RunLog.Errorf("writeFile failed. error:%s", err.Error())
		return fmt.Errorf("writeFile failed. error:%s", err.Error())
	}
	return nil
}
