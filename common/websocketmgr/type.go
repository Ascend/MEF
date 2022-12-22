// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

// CertPathInfo cert path info
type CertPathInfo struct {
	RootCaPath  string
	SvrCertPath string
	SvrKeyPath  string
	ServerFlag  bool
}

// RegisterModuleInfo register module info
type RegisterModuleInfo struct {
	MsgOpt     string
	MsgRes     string
	ModuleName string
}
