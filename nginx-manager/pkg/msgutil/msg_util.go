// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package msgutil this file contains methods used to deal messages
package msgutil

import (
	"fmt"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
)

const messageIdTimestampMask = 1e6

func newMsg(src, dst, opt, res string) *model.Message {
	msg := model.Message{}
	msg.Header.Timestamp = time.Now().UnixNano() / messageIdTimestampMask
	msg.Header.Version = "1.0"
	msg.SetRouter(src, dst, opt, res)
	return &msg
}

// SendVoidMsg 发送一条不含content的消息
func SendVoidMsg(src, dst, opt, res string) {
	msg := newMsg(src, dst, opt, res)
	if err := modulemgr.SendMessage(msg); err != nil {
		hwlog.RunLog.Errorf("send message failed, error: %s", err.Error())
	}
}

// Handle 处理分发消息
func Handle(req *model.Message) {
	method, exist := handlers.Load(Combine(req.GetOption(), req.GetResource()))
	if !exist {
		return
	}
	methodFunc, ok := method.(handlerFunc)
	if !ok {
		return
	}
	methodFunc(req)
}

type handlerFunc func(req *model.Message)

var handlers sync.Map

// RegisterHandlers 注册消息处理函数
func RegisterHandlers(key string, handler handlerFunc) {
	handlers.Store(key, handler)
}

// Combine 生成消息路由key
func Combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
