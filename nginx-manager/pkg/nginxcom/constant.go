// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxcom this file is for common constant or method
package nginxcom

const (
	// KeyPrefix 前缀
	KeyPrefix = "set $"
	// ClientPipePrefix 内部https转发使用的pipe前缀
	ClientPipePrefix = "client_pipe"
	// EdgePortKey port对应在配置文件的key
	EdgePortKey = "EdgeMgrSvcPort"
	// SoftPortKey 软件仓port对应的key
	SoftPortKey = "SoftwareMgrSvcPort"
	// CertPortKey 证书服务port对应的key
	CertPortKey = "CertMgrSvcPort"
	// PortMin 端口最小值
	PortMin = 1024
	// PortMax 端口最大值
	PortMax = 65535
	// NginxDefaultConfigPath nginx配置文件模板
	NginxDefaultConfigPath = "/home/MEFCenter/conf/nginx_default.conf"
	// NginxConfigPath nginx配置文件
	NginxConfigPath = "/home/MEFCenter/conf/nginx.conf"
	// CertKeyFile 证书私钥文件
	CertKeyFile = "/home/data/config/mef-certs/nginx-manager-server.key"
	// ClientCertKeyFile 内部转发消息的证书私钥文件
	ClientCertKeyFile = "/home/data/config/mef-certs/nginx-manager.key"
	// PipePath 证书私钥管道
	PipePath = "/home/MEFCenter/pipe/apig_keyPipe"
	// ClientPipeDir 内部转发消息的证书私钥管道
	ClientPipeDir = "/home/MEFCenter/pipe/"
	// FifoPermission 证书私钥管道权限
	FifoPermission = 0600
	// NginxManagerName nginx manager模块对应收发消息的key
	NginxManagerName = "NginxManager"
	// NginxMonitorName nginx monitor模块对应收发消息的key
	NginxMonitorName = "NginxMonitor"
	// Nginx 资源名
	Nginx = "nginx"
	// Monitor 资源名
	Monitor = "monitor"
	// ReqActiveMonitor 启动监控操作
	ReqActiveMonitor = "ReqActiveMonitor"
	// ReqRestartNginx 重启Nginx操作
	ReqRestartNginx = "ReqRestartNginx"
	// RespRestartNginx 回复重启Nginx操作
	RespRestartNginx = "RespRestartNginx"
	// NginxSslPortKey nginx使用ssl的端口
	NginxSslPortKey = "NginxSslPort"
)
