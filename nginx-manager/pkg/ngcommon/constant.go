// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package ngcommon this file is for common constant or method
package ngcommon

const (
	// KeyPrefix 前缀
	KeyPrefix = "set $"
	// EdgeUrlKey url对应在配置文件的key
	EdgeUrlKey = "EdgeMgrSvcDomain"
	// EdgePortKey port对应在配置文件的key
	EdgePortKey = "EdgeMgrSvcPort"
	// SoftUrlKey 软件仓url对应的key
	SoftUrlKey = "SoftwareMgrSvcDomain"
	// SoftPortKey 软件仓port对应的key
	SoftPortKey = "SoftwareMgrSvcPort"
	// UrlPattern 校验url的正则
	UrlPattern = "^[a-z][a-z-.]{1,64}$"
	// PortMin 端口最小值
	PortMin = 1
	// PortMax 端口最大值
	PortMax = 65535
	// NginxDefaultConfigPath nginx配置文件模板
	NginxDefaultConfigPath = "/home/hwMindX/conf/nginx_default.conf"
	// NginxConfigPath nginx配置文件
	NginxConfigPath = "/home/hwMindX/conf/nginx.conf"
	// CertKeyFile 证书私钥文件
	CertKeyFile = "/home/hwMindX/certs/nginx-manager.key"
	// PipePath 证书私钥管道
	PipePath = "/home/hwMindX/conf/keyPipe"
	// FifoPermission 证书私钥管道权限
	FifoPermission = 0600
	NginxManagerName = "NginxManager"
)
