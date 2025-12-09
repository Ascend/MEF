// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package logrotate provides log rotation function for third-party software
package logrotate

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
)

const (
	// DefaultModuleName default module name
	DefaultModuleName = "logrotate"
)

// Initializer init function for logRotator module
type Initializer func() (Configs, error)

type logRotatorModule struct {
	name        string
	ctx         context.Context
	rotator     *LogRotator
	initializer Initializer
}

// Module creates a logRotator module
func Module(name string, ctx context.Context, initializer Initializer) model.Module {
	return &logRotatorModule{
		name:        name,
		ctx:         ctx,
		rotator:     New(Configs{}),
		initializer: initializer,
	}
}

// Name returns module's name
func (l *logRotatorModule) Name() string {
	if l.name == "" {
		return DefaultModuleName
	}
	return l.name
}

// Enable setups environment for logRotator
func (l *logRotatorModule) Enable() bool {
	configs, err := l.initializer()
	if err != nil {
		hwlog.RunLog.Errorf("unable to enable module %s, %s", l.Name(), err.Error())
		return false
	}
	l.rotator.configs = configs
	return true
}

// Start starts module
func (l *logRotatorModule) Start() {
	l.rotator.Start(l.ctx)
}
