// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxcom this file is for common constant or method
package nginxcom

const (
	// KeyPrefix 前缀
	KeyPrefix = "$"
	// ClientPipePrefix 内部https转发使用的pipe前缀
	ClientPipePrefix = "client_pipe"
	// EdgePortKey port对应在配置文件的key
	EdgePortKey = "EdgeMgrSvcPort"
	// CertPortKey 证书服务port对应的key
	CertPortKey = "CertMgrSvcPort"
	// CrlConfigKey 证书吊销列表配置对应key
	CrlConfigKey = "SslCrlPath"
	// NginxConfigPath nginx配置文件
	NginxConfigPath = "/home/MEFCenter/conf/nginx.conf"
	// ServerCertFile nginx对外服务证书
	ServerCertFile = "/home/data/config/mef-certs/nginx-manager-server.crt"
	// ServerCertKeyFile nginx对外服务证书私钥文件
	ServerCertKeyFile = "/home/data/config/mef-certs/nginx-manager-server.key"
	// NorthernCertFile 北向证书文件
	NorthernCertFile = "/home/data/config/mef-certs/northern-root.crt"
	// NorthernCrlFile 北向证书文件
	NorthernCrlFile = "/home/data/config/mef-certs/northern-root.crl"
	// ClientCertFile 内部转发消息的证书
	ClientCertFile = "/home/data/config/mef-certs/nginx-manager.crt"
	// ClientCertKeyFile 内部转发消息的证书私钥文件
	ClientCertKeyFile = "/home/data/config/mef-certs/nginx-manager.key"
	// RootCaPath 内部消息转发根证书文件
	RootCaPath = "/home/data/inner-root-ca/RootCA.crt"
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
	// PodIpKey the key of this pod's ip
	PodIpKey = "POD_IP"
	// LockTimeKey lock time key for user and ip
	LockTimeKey = "LockTime"
	// TokenExpireTimeKey token expire time
	TokenExpireTimeKey = "TokenExpireTime"
	// EnableResolverKey enable the nginx dynamic ip resolver, true or false
	EnableResolverKey = "EnableResolver"
)
