// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package msgconv
package msgconv

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

var (
	handlersMap sync.Map
	initialized int32
)

const (
	syncMessageTimeout = 3 * time.Second
)

// Source defines the source of messages
type Source int

// constants for source
const (
	Edge Source = iota
	Cloud
)

// Event defines the event for dispatching messages
type Event int

// constants for events
const (
	BeforeModification Event = iota
	AfterDispatch
)

// ForwardingRegisterInfo defines the rules for forwarding messages
type ForwardingRegisterInfo struct {
	Source      Source
	Operation   string
	Resource    string
	Event       Event
	Destination string
}

// DispatchFunc dispatches messages
type DispatchFunc func(message *model.Message) error

// Proxy dispatches message
type Proxy struct {
	MessageSource Source
	DispatchFunc  DispatchFunc
}

// Init configures msgconv.Proxy
func Init(forwardingRules ...ForwardingRegisterInfo) error {
	if !atomic.CompareAndSwapInt32(&initialized, 0, 1) {
		return errors.New("can't initialize proxy again")
	}

	for idx := range msgconvHandlers {
		handler := &msgconvHandlers[idx]
		handlersMap.Store(handler.getMessageRouteType(), handler)
	}
	for _, fr := range forwardingRules {
		handler := &messageHandler{operation: fr.Operation, resource: fr.Resource}
		value, loaded := handlersMap.LoadOrStore(getMessageRouteType(fr.Source, fr.Operation, fr.Resource), handler)
		if loaded {
			h, ok := value.(*messageHandler)
			if !ok {
				return errors.New("bad message handler type")
			}
			handler = h
		}
		handler.forwardingHandler.addForwardingRule(fr)
	}
	return nil
}

// DispatchMessage dispatches message
func (p *Proxy) DispatchMessage(message *model.Message) error {
	if p.DispatchFunc == nil {
		return errors.New("cannot dispatch message because dispatch func is not defined")
	}
	handler := p.getHandler(message.KubeEdgeRouter)
	// dispatch message directly if we didn't define a handler
	if handler == nil {
		return p.DispatchFunc(message)
	}
	if err := handler.dispatchMessage(message, p.DispatchFunc); err != nil {
		hwlog.RunLog.Errorf("route: %+v, error: %v", message.KubeEdgeRouter, err)
		return err
	}
	return nil
}

func (p *Proxy) getHandler(route model.MessageRoute) *messageHandler {
	var handler *messageHandler
	handlersMap.Range(func(_, value interface{}) bool {
		h, ok := value.(*messageHandler)
		if !ok {
			return false
		}
		if h.accept(route, p.MessageSource) {
			handler = h
			return false
		}
		return true
	})
	return handler
}

type messageForwardingHandler struct {
	mu    sync.RWMutex
	rules []ForwardingRegisterInfo
}

func (h *messageForwardingHandler) addForwardingRule(rule ForwardingRegisterInfo) {
	h.mu.Lock()
	h.rules = append(h.rules, rule)
	h.mu.Unlock()
}

func (h *messageForwardingHandler) tryForward(event Event, message *model.Message) error {
	// make a local copy of rules to improve performance
	h.mu.RLock()
	rules := make([]ForwardingRegisterInfo, len(h.rules))
	copy(rules, h.rules)
	h.mu.RUnlock()

	for _, rule := range rules {
		if rule.Event != event {
			continue
		}
		forwardedMessage, err := h.createForwardedMessage(rule.Destination, message)
		if err != nil {
			return err
		}
		resp, err := modulemgr.SendSyncMessage(forwardedMessage, syncMessageTimeout)
		if err != nil {
			return err
		}
		var result string
		if err := resp.ParseContent(&result); err != nil {
			return err
		}
		if result != constants.OK {
			return fmt.Errorf("get an unsuccessful resposne: %s", result)
		}
	}
	return nil
}

func (h *messageForwardingHandler) createForwardedMessage(
	destination string, message *model.Message) (*model.Message, error) {
	const resUIDIndex = 2

	forwardedMessage, err := model.NewMessage()
	if err != nil {
		return nil, err
	}
	forwardedMessage.SetIsSync(true)

	forwardedMessage.Content = make([]byte, len(message.Content))
	copy(forwardedMessage.Content, message.Content)

	forwardedMessage.KubeEdgeRouter = message.KubeEdgeRouter

	tokens := strings.Split(forwardedMessage.KubeEdgeRouter.Resource, "/")
	resourcePrefix := strings.Join(tokens[:resUIDIndex], "/")
	if len(tokens) > resUIDIndex {
		resourcePrefix += "/"
	}
	forwardedMessage.SetRouter(forwardedMessage.KubeEdgeRouter.Source, destination,
		forwardedMessage.KubeEdgeRouter.Operation, resourcePrefix)

	return forwardedMessage, nil
}

type setter func(*model.Message, interface{}) error

type messageHandler struct {
	operation         string
	resource          string
	source            Source
	contentType       interface{}
	forwardingHandler messageForwardingHandler
	setters           []setter
	handleFunc        DispatchFunc
}

func (h *messageHandler) dispatchMessage(message *model.Message, defaultDispatchFunc DispatchFunc) error {
	// try to forward message before modification
	if err := h.forwardingHandler.tryForward(BeforeModification, message); err != nil {
		return fmt.Errorf("forward message before modification failed, %v", err)
	}

	// save metadata in database if message comes from edgecore
	if h.source == Edge {
		if err := h.saveMetadata(message); err != nil {
			return fmt.Errorf("save metadata for upstream message failed, %v", err)
		}
	}

	// modify message before sending
	if err := h.modifyMessage(message); err != nil {
		return fmt.Errorf("modify message failed, %v", err)
	}

	// dispatch message directly if there is no custom handler
	dispatchFunc := defaultDispatchFunc
	if h.handleFunc != nil {
		dispatchFunc = h.handleFunc
	}
	if err := dispatchFunc(message); err != nil {
		return fmt.Errorf("dispatch message failed, %v", err)
	}

	// save metadata in database if message comes from FD/MEFCenter
	if h.source == Cloud {
		if err := h.saveMetadata(message); err != nil {
			return fmt.Errorf("save metadata for downstream message failed, %v", err)
		}
	}

	// try to forward message after dispatch
	if err := h.forwardingHandler.tryForward(AfterDispatch, message); err != nil {
		return fmt.Errorf("forward message after dispatch failed, %v", err)
	}
	return nil
}

func (h *messageHandler) modifyMessage(message *model.Message) error {
	if len(h.setters) == 0 {
		return nil
	}
	if h.contentType == nil {
		return errors.New("nil content type is unsupported")
	}
	unmarshalledContentPtr := reflect.New(reflect.TypeOf(h.contentType)).Interface()
	if err := json.Unmarshal(message.Content, unmarshalledContentPtr); err != nil {
		return errors.New("unmarshal message failed")
	}

	for _, fn := range h.setters {
		if err := fn(message, unmarshalledContentPtr); err != nil {
			return fmt.Errorf("modify message failed, %v", err)
		}
	}
	return message.FillContent(unmarshalledContentPtr)
}

func (h *messageHandler) accept(route model.MessageRoute, source Source) bool {
	return route.Operation == h.operation && strings.HasPrefix(route.Resource, h.resource) && source == h.source
}

func (h *messageHandler) getMessageRouteType() string {
	return getMessageRouteType(h.source, h.operation, h.resource)
}

func getMessageRouteType(source Source, operation, resource string) string {
	return fmt.Sprintf("%s:%s:%d", operation, resource, source)
}
