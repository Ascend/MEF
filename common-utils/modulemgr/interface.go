// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package modulemgr, abstract interface definition for message handler
package modulemgr

import (
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr/model"
)

// MessageHandlerIntf abstract definition for message handler
type MessageHandlerIntf interface {
	GetMessageKey() string
	GetMessageHandler() string
	GetRpsLimiter() limiter.IndependentLimiter
}

// HandleMessageIntf abstract definition for registration message
type HandleMessageIntf interface {
	Register(MessageHandlerIntf)
	HandleMsg([]byte, model.MsgPeerInfo) []byte
}
