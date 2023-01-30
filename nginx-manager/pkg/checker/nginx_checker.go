// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker this file is for check parameter
package checker

import (
	"bytes"
	"fmt"

	"nginx-manager/pkg/nginxcom"
)

// confItemsTemplate nginx.conf需要替换的配置项
var confItemsTemplate = []nginxcom.NginxConfItem{
	{Key: nginxcom.EdgeUrlKey, From: nginxcom.KeyPrefix + nginxcom.EdgeUrlKey},
	{Key: nginxcom.EdgePortKey, From: nginxcom.KeyPrefix + nginxcom.EdgePortKey},
	{Key: nginxcom.SoftUrlKey, From: nginxcom.KeyPrefix + nginxcom.SoftUrlKey},
	{Key: nginxcom.SoftPortKey, From: nginxcom.KeyPrefix + nginxcom.SoftPortKey},
	{Key: nginxcom.CertUrlKey, From: nginxcom.KeyPrefix + nginxcom.CertUrlKey},
	{Key: nginxcom.CertPortKey, From: nginxcom.KeyPrefix + nginxcom.CertPortKey},
}

func checkNginxConfig(param interface{}) error {
	content, ok := param.([]byte)
	if !ok {
		return fmt.Errorf("nginx config data error")
	}
	for _, v := range confItemsTemplate {
		if bytes.Index(content, []byte(v.From)) == -1 {
			return fmt.Errorf("cannot find property %s in nginx conf file", v.From)
		}
	}
	return nil
}

// GetConfigItemTemplate get the template of config replace items
func GetConfigItemTemplate() []nginxcom.NginxConfItem {
	return confItemsTemplate
}
