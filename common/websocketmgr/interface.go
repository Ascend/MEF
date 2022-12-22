// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

type NetProxyIntf interface {
	Start() error
	Send(msg interface{}) error
	Stop() error
	GetName() string
}

type HandleMsgIntf interface {
	handleMsg(msg []byte)
}
