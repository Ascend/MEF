// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package ngcommon this file is for common constant or method
package ngcommon

// Endpoint 路径端口信息
type Endpoint struct {
	UrlKey  string
	PortKey string
	UrlVal  string
	PortVal string
	Regexp  string
}

// NginxConfItem nginx配置替换项
type NginxConfItem struct {
	Key  string
	From string
	To   string
}
