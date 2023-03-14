// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

// RegisterModuleInfo register module info
type RegisterModuleInfo struct {
	Src        string
	MsgOpt     string
	MsgRes     string
	ModuleName string
}

// MsgOutModuleRouter find a dest websocket connection for msg
type MsgOutModuleRouter struct {
	Src      string
	MsgOpt   string
	MsgRes   string
	ConnName string
}
