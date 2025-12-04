// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package omjob this file for report manager module register
package omjob

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-om/omjob/handlers"
)

// omJobMgr report module
type omJobMgr struct {
	ctx    context.Context
	enable bool
}

// NewOmJobModule for om job module
func NewOmJobModule(enable bool, ctx context.Context) model.Module {
	module := &omJobMgr{
		enable: enable,
		ctx:    ctx,
	}
	return module
}

// Name module name
func (m *omJobMgr) Name() string {
	return constants.OmJobManager
}

// Enable module enable
func (m *omJobMgr) Enable() bool {
	return m.enable
}

// Start module start running
func (m *omJobMgr) Start() {
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(m.Name())
		if err != nil {
			hwlog.RunLog.Errorf("get receive module message failed,error:%v", err)
			continue
		}
		hwlog.RunLog.Infof("receive msg option:[%s] resource:[%s]", msg.GetOption(), msg.GetResource())

		go m.dispatchMsg(msg)
	}
}

func (m *omJobMgr) dispatchMsg(msg *model.Message) {
	handlerMgr := handlers.GetHandlerMgr()
	err := handlerMgr.Process(msg)
	if err != nil {
		hwlog.RunLog.Errorf("process msg failed: %v", err)
		return
	}
}
