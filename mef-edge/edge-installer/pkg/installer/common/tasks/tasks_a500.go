// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build !MEFEdge_SDK

// Package tasks for some methods that are performed only on the a500 device
package tasks

import (
	"edge-installer/pkg/common/config"
)

func (ssi *SetSystemInfoTask) initOmDb() error {
	return ssi.initOmDbWithTables(config.Configuration{})
}

func (ssi *SetSystemInfoTask) setDefaultAlarmConfig() error {
	return nil
}
