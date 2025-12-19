// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nginxcom this file is for common constant or method
package nginxcom

const (
	// KeyPrefix 前缀
	KeyPrefix = "$"
	// ClientPipePrefix 内部https转发使用的pipe前缀
	ClientPipePrefix = "client_pipe"
	// EdgePortKey port对应在配置文件的key
	EdgePortKey = "EdgeMgrSvcPort"
	// AlarmPortKey port对应在配置文件的key
	AlarmPortKey = "AlarmMgrSvcPort"
	// CertPortKey 证书服务port对应的key
	CertPortKey = "CertMgrSvcPort"
	// AuthPortKey 认证接口服务port对应的key
	AuthPortKey = "AuthPort"
	// WebsocketPortKey 边云对接websocket接口服务port对应的key
	WebsocketPortKey = "WebsocketPort"
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
	// NorthernCrlFile 北向证书吊销列表文件
	NorthernCrlFile = "/home/data/config/mef-certs/northern-root.crl"
	// SouthAuthCertFile  南向认证接口服务证书
	SouthAuthCertFile = "/home/data/config/mef-certs/south-auth-server.crt"
	// SouthAuthCertKeyFile  南向认证接口证书私钥文件
	SouthAuthCertKeyFile = "/home/data/config/mef-certs/south-auth-server.key"
	// WebsocketCertFile  南向服务接口服务证书
	WebsocketCertFile = "/home/data/config/mef-certs/south-websocket-server.crt"
	// WebsocketCertKeyFile  南向服务接口证书私钥文件
	WebsocketCertKeyFile = "/home/data/config/mef-certs/south-websocket-server.key"
	// SouthernCertFile  南向证书文件
	SouthernCertFile = "/home/data/config/mef-certs/southern-root.crt"
	// ClientCertFile 内部转发消息的证书
	ClientCertFile = "/home/data/config/mef-certs/nginx-manager.crt"
	// ClientCertKeyFile 内部转发消息的证书私钥文件
	ClientCertKeyFile = "/home/data/config/mef-certs/nginx-manager.key"
	// RootCaPath 内部消息转发根证书文件
	RootCaPath = "/home/data/inner-root-ca/RootCA.crt"
	// ThirdPartyServiceCertPath used for third party services cert
	ThirdPartyServiceCertPath = "/home/data/config/mef-certs/third-party-svc.crt"
	// ThirdPartyServiceKeyPath used for third party services key
	ThirdPartyServiceKeyPath = "/home/data/config/mef-certs/third-party-svc.key"
	// ThirdPipePrefix used for third pipe
	ThirdPipePrefix = "third_pipe"
	// PipePath 证书私钥管道
	PipePath = "/home/MEFCenter/pipe/apig_keyPipe"
	// AuthPipePath 认证端口的证书私钥管段
	AuthPipePath = "/home/MEFCenter/pipe/apig_auth_keyPipe"
	// WebsocketPipePath 服务端口的证书私钥管段
	WebsocketPipePath = "/home/MEFCenter/pipe/apig_websocket_keyPipe"
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
)
