// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build !MEFEdge_A500

// Package tasks for some methods that are performed on the non a500 device
package tasks

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
)

func (swp *SetWorkPathTask) prepareCfgBackupDir() error {
	// !MEFEdge_A500 do not prepare config backup dir
	return nil
}

func (p *PostEffectProcessTask) copyResetScriptToP7() error {
	// !MEFEdge_A500 do not copy reset script to p7
	return nil
}

func (p *PostEffectProcessTask) smoothConfig() error {
	if err := p.smoothCommonConfig(); err != nil {
		return err
	}
	if err := config.SmoothAlarmConfigDB(); err != nil {
		hwlog.RunLog.Errorf("smooth edge_om alarm config to db failed, error: %v", err)
		return errors.New("smooth edge_om alarm config to db failed")
	}
	return nil
}

func (p *PostEffectProcessTask) refreshDefaultCfgDir() error {
	// !MEFEdge_A500 do not need refresh default config backup dir
	return nil
}
