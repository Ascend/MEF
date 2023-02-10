// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxcom this file is for common constant or method
package nginxcom

import "time"

const (
	// KeyPrefix 前缀
	KeyPrefix = "$"
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
	// ServerCertFile nginx对外服务证书
	ServerCertFile = "/home/data/config/mef-certs/nginx-manager-server.crt"
	// ServerCertKeyFile nginx对外服务证书私钥文件
	ServerCertKeyFile = "/home/data/config/mef-certs/nginx-manager-server.key"
	// UserCertFile 用户管理模块证书文件
	UserCertFile = "/home/data/config/mef-certs/user-manager.crt"
	// UserCertKeyFile 用户管理模块证书key
	UserCertKeyFile = "/home/data/config/mef-certs/user-manager.key"
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
	// DefaultDbPath 默认数据库地址
	DefaultDbPath = "/home/data/config/user-manager.db"
	// UserManagerName user manager模块对应收发消息的服务名
	UserManagerName = "UserManager"
	// UserRestfulServiceName 用户管理模块对应的restful服务名
	UserRestfulServiceName = "UserRestfulService"
	// DefaultUsernameKey 用户管理模块默认账号对应的key
	DefaultUsernameKey = "DefaultUsername"
	// UserMgrSvcPortKey 用户管理模块端口
	UserMgrSvcPortKey = "UserMgrSvcPort"
	// NginxSslPortKey nginx使用ssl的端口
	NginxSslPortKey = "NginxSslPort"
	// UserLockTime 用户锁定时长
	UserLockTime = time.Second * 30
	// IpLockTime Ip锁定时长
	IpLockTime = time.Second * 30
	// MaxPwdWrongTimes 密码最大错误次数
	MaxPwdWrongTimes = 5
	// HistoryPasswordSaveCount 相同密码缓存次数
	HistoryPasswordSaveCount = 5
)
