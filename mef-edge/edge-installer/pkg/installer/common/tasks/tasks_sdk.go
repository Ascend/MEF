// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package common for some methods that are performed only on non a500 device
package tasks

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
)

func (ssi *SetSystemInfoTask) initOmDb() error {
	return ssi.initOmDbWithTables(config.Configuration{}, config.AlarmConfig{})
}

func (ssi *SetSystemInfoTask) setDefaultAlarmConfig() error {
	if err := config.SetDefaultAlarmCfg(); err != nil {
		hwlog.RunLog.Errorf("set initial alarm config to table failed, error: %v", err)
		return errors.New("set initial alarm config to table failed")
	}
	return nil
}
