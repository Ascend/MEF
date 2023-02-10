// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package context to start module_manager context
package context

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/modulemanager/model"
)

const defaultMsgTimeout = 30 * time.Second

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

func (context *channelContext) sendMsgByChannel(channel chan model.Message, msg *model.Message) error {
	if channel == nil {
		return errors.New("model.Message channel is nil")
	}
	select {
	case channel <- *msg:
	case <-time.After(defaultMsgTimeout):
		return fmt.Errorf("channel context send msg timeout")
	}
	return nil
}

func (context *channelContext) addAnonChannel(id string, channel chan model.Message) {
	if id == "" || channel == nil {
		return
	}

	context.anonChannels.Store(id, channel)

	return
}

func (context *channelContext) getAnonChannel(id string) (chan model.Message, error) {
	if id == "" {
		return nil, fmt.Errorf("get anon channel fail(%s)", id)
	}

	var channel interface{}
	var ok bool

	if channel, ok = context.anonChannels.Load(id); !ok {
		return nil, fmt.Errorf("can not find anon channel by id %s", id)
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
	return context.sendMsgByChannel(channel, msg)
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

	if err = context.sendMsgByChannel(reqChannel, msg); err != nil {
		return nil, err
	}

	var resp model.Message

	select {
	case resp = <-respChannel:
		hwlog.RunLog.Debug("receive resp for send sync")
	case <-time.After(timeout):
		hwlog.RunLog.Error("receive resp timeount for send sync")
		return nil, fmt.Errorf("receive resp timeount for send sync")
	}

	return &resp, nil
}

func (context *channelContext) SendResp(msg *model.Message) error {
	if msg == nil {
		return fmt.Errorf("input is invalid in sedn resp")
	}

	annoChannel, err := context.getAnonChannel(msg.GetParentId())
	if err != nil {
		return err
	}
	if annoChannel == nil {
		return errors.New("annoChannel is nil")
	}
	annoChannel <- *msg
	return nil
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
