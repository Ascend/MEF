// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker this file is for check parameter
package checker

import (
	"bytes"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"nginx-manager/pkg/nginxcom"
)

// confItemsTemplate nginx.conf需要替换的配置项
var confItemsTemplate = []nginxcom.NginxConfItem{
	{Key: nginxcom.EdgePortKey, From: nginxcom.KeyPrefix + nginxcom.EdgePortKey},
	{Key: nginxcom.SoftPortKey, From: nginxcom.KeyPrefix + nginxcom.SoftPortKey},
	{Key: nginxcom.CertPortKey, From: nginxcom.KeyPrefix + nginxcom.CertPortKey},
	{Key: nginxcom.UserMgrSvcPortKey, From: nginxcom.KeyPrefix + nginxcom.UserMgrSvcPortKey},
}

func checkNginxConfig(param interface{}) error {
	content, ok := param.([]byte)
	if !ok {
		hwlog.RunLog.Error("nginx config data error")
		return fmt.Errorf("nginx config data error")
	}
	for _, v := range confItemsTemplate {
		if bytes.Index(content, []byte(v.From)) == -1 {
			continue
		}
	}
	return nil
}

// GetConfigItemTemplate get the template of config replace items
func GetConfigItemTemplate() []nginxcom.NginxConfItem {
	return confItemsTemplate
}
