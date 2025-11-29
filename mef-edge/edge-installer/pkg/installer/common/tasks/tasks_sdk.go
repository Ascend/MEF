// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
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
