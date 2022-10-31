package context

import (
	"edge-manager/module_manager/model"
	"fmt"
	"sync"
	"time"
)

const defaultMsgTimeout = 30 * time.Second

type channelContext struct {
	channels map[string]chan model.Message

	anonChannels map[string]chan model.Message
	anonChsLock  sync.RWMutex
}

func (context *channelContext) findChannel(moduleName string) (chan model.Message, error) {
	var channel chan model.Message
	var ok bool
	if channel, ok = context.channels[moduleName]; !ok {
		return nil, fmt.Errorf("can not find channel by name %s", moduleName)
	}
	return channel, nil
}

func (context *channelContext) sendMsgByChannel(channel chan model.Message, msg *model.Message) error {
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
	defer context.anonChsLock.Unlock()

	context.anonChsLock.Lock()
	context.anonChannels[id] = channel
	return
}

func (context *channelContext) getAnonChannel(id string) (chan model.Message, error) {
	if id == "" {
		return nil, fmt.Errorf("get anon channel fail(%s)", id)
	}

	var channel chan model.Message
	var ok bool
	defer context.anonChsLock.Unlock()
	context.anonChsLock.Lock()
	if channel, ok = context.anonChannels[id]; !ok {
		return nil, fmt.Errorf("can not find anon channel by id %s", id)
	}
	return channel, nil
}

func (context *channelContext) deleteAnonChannel(id string) {
	var channel chan model.Message
	var ok bool

	context.anonChsLock.Lock()
	if channel, ok = context.anonChannels[id]; !ok {
		context.anonChsLock.Unlock()
		return
	}
	delete(context.anonChannels, id)
	context.anonChsLock.Unlock()
	close(channel)
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
	msg := <-channel
	return &msg, nil
}

func (context *channelContext) SendSync(msg *model.Message, duration time.Duration) (*model.Message, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg is nil for channel context send sync")
	}

	var timeount time.Duration
	if duration <= 0 {
		timeount = defaultMsgTimeout
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
	case <-time.After(timeount):
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

	annoChannel <- *msg
	return nil
}

func (context *channelContext) Registry(moduleName string) error {
	if _, existed := context.channels[moduleName]; existed {
		return fmt.Errorf("channel by module %s existed", moduleName)
	}
	context.channels[moduleName] = make(chan model.Message)
	return nil
}

func (context *channelContext) Unregistry(moduleName string) error {
	var channel chan model.Message
	var existed bool

	if channel, existed = context.channels[moduleName]; !existed {
		return fmt.Errorf("delete %s channel not existed", moduleName)
	}

	delete(context.channels, moduleName)
	close(channel)
	return nil
}

func (context *channelContext) init() {
	context.channels = make(map[string]chan model.Message)
	context.anonChannels = make(map[string]chan model.Message)
}

func NewChannelContext() *channelContext {
	c := channelContext{}
	c.init()
	return &c
}
