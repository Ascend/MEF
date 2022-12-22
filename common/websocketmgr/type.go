// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

type CertPathInfo struct {
	RootCaPath  string
	SvrCertPath string
	SvrKeyPath  string
	ServerFlag  bool
}

type RegisterModuleInfo struct {
	MsgOpt     string
	MsgRes     string
	ModuleName string
}
