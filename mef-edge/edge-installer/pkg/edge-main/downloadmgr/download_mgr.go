// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package downloadmgr for edge-main to download
package downloadmgr

import (
	"context"
	"sync"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/configpara"
)

type downloadMgr struct {
	ctx    context.Context
	enable bool
	lock   sync.Mutex
}

// NewDownloadMgr create new module instance
func NewDownloadMgr(enable bool) model.Module {
	module := &downloadMgr{
		enable: enable,
	}
	module.ctx = context.Background()
	return module
}

// Name [method] return name of module
func (d *downloadMgr) Name() string {
	return constants.DownloadManagerName
}

// Enable [method] decide this module should be enabled
func (d *downloadMgr) Enable() bool {
	return d.enable
}

// Start [method] start this module
func (d *downloadMgr) Start() {
	hwlog.RunLog.Info("download manager start success")
	for {
		select {
		case _, _ = <-d.ctx.Done():
			hwlog.RunLog.Info("-------------------download manager exit-------------------")
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(d.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receives request failed", d.Name())
			continue
		}
		hwlog.RunLog.Infof("download manager received msg, option: [%s] resource: [%s]",
			msg.GetOption(), msg.GetResource())
		go d.process(*msg)
	}
}

func (d *downloadMgr) process(msg model.Message) {
	hwlog.OpLog.Infof("[%v@%v][%v %v][msgId: %v]", configpara.GetNetConfig().NetType,
		configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), msg.Header.Id)
	result := constants.Failed
	defer func() {
		hwlog.OpLog.Infof("[%v@%v][%v %v %v][msgId: %v]", configpara.GetNetConfig().NetType,
			configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), result, msg.Header.Id)
	}()
	if !d.lock.TryLock() {
		hwlog.RunLog.Errorf("download manager process msg [option:[%s] resource:[%s]] failed, "+
			"previous download process is no finished ", msg.GetOption(), msg.GetResource())
		return
	}
	defer d.lock.Unlock()
	if err := d.processDownloadSoftware(msg); err != nil {
		hwlog.RunLog.Errorf("download manager process msg [option:[%s] resource:[%s]] failed",
			msg.GetOption(), msg.GetResource())
		return
	}
	result = constants.Success
}
