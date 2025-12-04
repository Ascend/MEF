// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr for deal every handler
package handlermgr

import (
	"context"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

type handlerManger struct {
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

// NewHandlerMgrModule new config module
func NewHandlerMgrModule(enable bool) model.Module {
	module := &handlerManger{
		enable: enable,
	}
	module.ctx, module.cancel = context.WithCancel(context.Background())
	return module
}

// Name module name
func (hm *handlerManger) Name() string {
	return constants.ModEdgeOm
}

// Stop module stop
func (hm *handlerManger) Stop() bool {
	hm.cancel()
	return true
}

// Start module start running
func (hm *handlerManger) Start() {
	for {
		select {
		case <-hm.ctx.Done():
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(hm.Name())
		if err != nil {
			hwlog.RunLog.Errorf("get receive module message failed,error:%v", err)
			continue
		}
		hwlog.RunLog.Infof("receive msg option:[%s] resource:[%s]", msg.GetOption(), msg.GetResource())

		go hm.dispatchMsg(msg)
	}
}

func (hm *handlerManger) dispatchMsg(msg *model.Message) {
	mgr := GetHandlerMgr()
	err := mgr.Process(msg)
	if err != nil {
		hwlog.RunLog.Errorf("process message failed, route: %+v, error: %v", msg.Router, err)
		return
	}
}

func initConfig() error {
	podCfg, err := config.LoadPodConfig()
	if err != nil {
		return fmt.Errorf("init pod config from file failed, error:%v", err)
	}
	podConfig = *podCfg
	hwlog.RunLog.Infof("init pod config successfully, content is: %+v", podConfig)
	return nil
}
