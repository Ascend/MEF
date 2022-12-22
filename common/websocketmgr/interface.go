// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

// NetProxyIntf proxy interface
type NetProxyIntf interface {
	Start() error
	Send(msg interface{}) error
	Stop() error
	GetName() string
}

// HandleMsgIntf  handler message interface
type HandleMsgIntf interface {
	handleMsg(msg []byte)
}
