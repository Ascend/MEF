// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logrotate provides log rotation function for third-party software
package logrotate

import (
	"context"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/modulemanager/model"
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
