// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
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
