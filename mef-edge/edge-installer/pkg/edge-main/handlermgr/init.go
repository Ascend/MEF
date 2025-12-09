// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr
package handlermgr

import (
	"context"
	"errors"
	"sync/atomic"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/handlermgr/handlers"
)

const (
	maxGoRoutine  = 1024
	maxRetryCount = 20
)

var goRoutineCount int32

type handlerManager struct {
	ctx    context.Context
	enable bool
}

// NewHandlerManager create the module of handler manager
func NewHandlerManager(ctx context.Context, enable bool) model.Module {
	return &handlerManager{
		enable: enable,
		ctx:    ctx,
	}
}

// Name module name
func (h *handlerManager) Name() string {
	return constants.ModHandlerMgr
}

// Enable module enable
func (h *handlerManager) Enable() bool {
	return h.enable
}

// Start module start running
func (h *handlerManager) Start() {
	for {
		select {
		case <-h.ctx.Done():
			hwlog.RunLog.Info("----------------handler manager module exit-------------------")
			return
		default:
		}
		message, err := modulemgr.ReceiveMessage(h.Name())
		if err != nil {
			hwlog.RunLog.Errorf("failed to receive message, error: %v", err)
			continue
		}

		if err = h.dispatchMessage(message); err != nil {
			hwlog.RunLog.Errorf("failed to dispatch message, error: %v", err)
		}
	}
}

func (h *handlerManager) dispatchMessage(message *model.Message) error {
	for i := 1; i <= maxRetryCount; i++ {
		count := atomic.LoadInt32(&goRoutineCount)
		if count > maxGoRoutine {
			hwlog.RunLog.Errorf("number of go routine over limit, discard message, router: %+v, route: %+v",
				message.Router, message.KubeEdgeRouter)
			return errors.New("number of go routine over limit")
		}
		if atomic.CompareAndSwapInt32(&goRoutineCount, count, count+1) {
			go h.processMessage(message)
			return nil
		}
	}
	return errors.New("check go routine count failed")
}

func (h *handlerManager) processMessage(message *model.Message) {
	defer atomic.AddInt32(&goRoutineCount, -1)
	handler := handlers.GetHandler()
	if err := handler.Process(message); err != nil {
		hwlog.RunLog.Errorf("handle message failed, route: %+v, error: %v", message.KubeEdgeRouter, err)
		return
	}
}
