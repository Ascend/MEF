// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package context to start module_manager context
package context

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
)

const (
	defaultMsgTimeout = 30 * time.Second
)

type channelContext struct {
	channels sync.Map

	anonChannels sync.Map
}

func (context *channelContext) findChannel(moduleName string) (chan model.Message, error) {
	var channel interface{}
	var ok bool

	if channel, ok = context.channels.Load(moduleName); !ok {
		return nil, fmt.Errorf("can not find channel by name %s", moduleName)
	}

	rChannel, ok := channel.(chan model.Message)
	if !ok {
		return nil, fmt.Errorf("is not model message channel %s", moduleName)
	}
	return rChannel, nil
}

func (context *channelContext) sendMsgByChannel(
	channel chan model.Message, msg *model.Message, timeout time.Duration) error {
	sender := &channelSender{
		channel: channel,
		msg:     msg,
		timeout: timeout,
	}
	sender.send()
	return sender.lastErr
}

func (context *channelContext) addAnonChannel(id string, channel chan model.Message) {
	if id == "" || channel == nil {
		hwlog.RunLog.Debug("add anon channel failed, id or channel is nil")
		return
	}

	context.anonChannels.Store(id, channel)

	return
}

func (context *channelContext) getAnonChannel(id string) (chan model.Message, error) {
	if id == "" {
		return nil, errors.New("get anon channel fail: id is empty")
	}

	var channel interface{}
	var ok bool

	if channel, ok = context.anonChannels.Load(id); !ok {
		return nil, errors.New("can not find anon channel")
	}

	rChannel, ok := channel.(chan model.Message)
	if !ok {
		return nil, fmt.Errorf("is not model message channel %s", id)
	}
	return rChannel, nil
}

func (context *channelContext) deleteAnonChannel(id string) {
	var channel interface{}
	var ok bool

	defer func() {
		if exception := recover(); exception != nil {
			hwlog.RunLog.Errorf("recover when delete anno channel, exception: %#v", exception)
		}
	}()
	if channel, ok = context.anonChannels.Load(id); !ok {
		return
	}
	context.anonChannels.Delete(id)
	rChannel, ok := channel.(chan model.Message)
	if !ok {
		return
	}
	close(rChannel)
}

func (context *channelContext) Send(msg *model.Message) error {
	if msg == nil {
		return fmt.Errorf("msg is nil for channel context send")
	}

	var channel chan model.Message
	var err error

	if channel, err = context.findChannel(msg.GetDestination()); err != nil {
		return err
	}
	return context.sendMsgByChannel(channel, msg, defaultMsgTimeout)
}

func (context *channelContext) Receive(moduleName string) (*model.Message, error) {
	var channel chan model.Message
	var err error

	if channel, err = context.findChannel(moduleName); err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, errors.New("channel is nil")
	}
	msg, ok := <-channel
	if !ok {
		return nil, errors.New("channel is close")
	}
	return &msg, nil
}

func (context *channelContext) SendSync(msg *model.Message, duration time.Duration) (*model.Message, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg is nil for channel context send sync")
	}

	timeout := duration
	if duration <= 0 {
		timeout = defaultMsgTimeout
	}

	var reqChannel chan model.Message
	var err error

	if reqChannel, err = context.findChannel(msg.GetDestination()); err != nil {
		return nil, err
	}

	respChannel := make(chan model.Message)
	context.addAnonChannel(msg.GetId(), respChannel)

	defer context.deleteAnonChannel(msg.GetId())

	if err = context.sendMsgByChannel(reqChannel, msg, defaultMsgTimeout); err != nil {
		return nil, err
	}

	var resp model.Message

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case resp = <-respChannel:
		hwlog.RunLog.Debug("receive resp for send sync")
	case <-timer.C:
		hwlog.RunLog.Error("receive resp timeout for send sync")
		return nil, fmt.Errorf("receive resp timeout for send sync")
	}

	return &resp, nil
}

func (context *channelContext) SendResp(msg *model.Message) error {
	if msg == nil {
		return fmt.Errorf("input is invalid in send resp")
	}

	annoChannel, err := context.getAnonChannel(msg.GetParentId())
	if err != nil {
		return fmt.Errorf("send msg to %s failed: %s", msg.GetDestination(), err.Error())
	}

	return context.sendMsgByChannel(annoChannel, msg, 0)
}

func (context *channelContext) Registry(moduleName string) error {
	if _, existed := context.channels.Load(moduleName); existed {
		return fmt.Errorf("channel by module %s existed", moduleName)
	}
	context.channels.Store(moduleName, make(chan model.Message))
	return nil
}

func (context *channelContext) Unregistry(moduleName string) error {
	var channel interface{}
	var existed bool

	defer func() {
		if exception := recover(); exception != nil {
			hwlog.RunLog.Errorf("recover when unregistry, exception: %#v", exception)
		}
	}()

	if channel, existed = context.channels.Load(moduleName); !existed {
		return fmt.Errorf("delete %s channel not existed", moduleName)
	}

	context.channels.Delete(moduleName)
	rChannel, ok := channel.(chan model.Message)
	if !ok {
		return fmt.Errorf("delete %s channel is not model message channel", moduleName)
	}
	close(rChannel)
	return nil
}

// NewChannelContext new channel context
func NewChannelContext() *channelContext {
	c := channelContext{}
	return &c
}

type channelSender struct {
	msg     *model.Message
	channel chan model.Message
	timeout time.Duration
	lastErr error
}

func (s *channelSender) send() {
	if s.msg == nil {
		s.lastErr = errors.New("model Message is nil")
		return
	}
	if s.channel == nil {
		s.lastErr = errors.New("model Message channel is nil")
		return
	}

	defer func() {
		if exception := recover(); exception != nil {
			s.lastErr = fmt.Errorf("recover when send msg, exception: %#v", exception)
		}
	}()

	if s.timeout == 0 {
		s.channel <- *s.msg
		return
	}

	timer := time.NewTimer(s.timeout)
	defer timer.Stop()
	select {
	case s.channel <- *s.msg:
	case <-timer.C:
		s.lastErr = fmt.Errorf("channel context send msg timeout")
	}
}
