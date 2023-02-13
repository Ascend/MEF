// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker this file is for check parameter
package checker

import (
	"bytes"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"nginx-manager/pkg/nginxcom"
)

// confItemsTemplate the items needed to replace into nginx.conf
var confItemsTemplate = []nginxcom.NginxConfItem{
	{Key: nginxcom.NginxSslPortKey, From: nginxcom.KeyPrefix + nginxcom.NginxSslPortKey},
	{Key: nginxcom.UserMgrSvcPortKey, From: nginxcom.KeyPrefix + nginxcom.UserMgrSvcPortKey},
	{Key: nginxcom.PodIpKey, From: nginxcom.KeyPrefix + nginxcom.PodIpKey},
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
