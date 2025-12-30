// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"context"
	"errors"
)

// RunLog run logger
var RunLog *logger

// OpLog operate logger
var OpLog *logger

// SecLog security logger
var SecLog *logger

// UserLog user logger
var UserLog *logger

// DebugLog debug logger
var DebugLog *logger

// InitRunLogger initialize run logger
func InitRunLogger(config *LogConfig, ctx context.Context) error {
	if config == nil {
		return errors.New("run logger config is nil")
	}
	if RunLog != nil && RunLog.isInit() {
		RunLog.Warn("run logger is been initialized.")
		return nil
	}
	RunLog = new(logger)
	if RunLog == nil {
		return errors.New("malloc new logger flied")
	}
	if err := RunLog.setLogger(config); err != nil {
		return err
	}
	if !RunLog.isInit() {
		return errors.New("run logger init failed")
	}
	return nil
}

// InitOperateLogger initialize operate logger
func InitOperateLogger(config *LogConfig, ctx context.Context) error {
	if config == nil {
		return errors.New("operate logger config is nil")
	}
	if OpLog != nil && OpLog.isInit() {
		OpLog.Warn("operate logger is been initialized.")
		return nil
	}
	OpLog = new(logger)
	if OpLog == nil {
		return errors.New("malloc new logger flied")
	}
	if err := OpLog.setLogger(config); err != nil {
		return err
	}
	if !OpLog.isInit() {
		return errors.New("operate logger init failed")
	}
	return nil
}

// InitSecurityLogger initialize security logger
func InitSecurityLogger(config *LogConfig, ctx context.Context) error {
	if config == nil {
		return errors.New("security logger config is nil")
	}
	if SecLog != nil && SecLog.isInit() {
		SecLog.Warn("security logger is been initialized.")
		return nil
	}
	SecLog = new(logger)
	if SecLog == nil {
		return errors.New("malloc new logger flied")
	}
	if err := SecLog.setLogger(config); err != nil {
		return err
	}
	if !SecLog.isInit() {
		return errors.New("security logger init failed")
	}
	return nil
}

// InitUserLogger initialize user logger
func InitUserLogger(config *LogConfig, ctx context.Context) error {
	if config == nil {
		return errors.New("user logger config is nil")
	}
	if UserLog != nil && UserLog.isInit() {
		UserLog.Warn("user logger is been initialized.")
		return nil
	}
	UserLog = new(logger)
	if UserLog == nil {
		return errors.New("malloc new logger flied")
	}
	if err := UserLog.setLogger(config); err != nil {
		return err
	}
	if !UserLog.isInit() {
		return errors.New("user logger init failed")
	}
	return nil
}

// InitDebugLogger initialize debug logger
func InitDebugLogger(config *LogConfig, ctx context.Context) error {
	if config == nil {
		return errors.New("debug logger config is nil")
	}
	if DebugLog != nil && DebugLog.isInit() {
		DebugLog.Warn("debug logger is been initialized.")
		return nil
	}
	DebugLog = new(logger)
	if DebugLog == nil {
		return errors.New("malloc new logger flied")
	}
	if err := DebugLog.setLogger(config); err != nil {
		return err
	}
	if !DebugLog.isInit() {
		return errors.New("debug logger init failed")
	}
	return nil
}

// InitHwLogger initialize run and operate logger
func InitHwLogger(ServerRunConf, ServerOpConf *LogConfig) error {
	if err := InitRunLogger(ServerRunConf, context.Background()); err != nil {
		return err
	}
	if err := InitOperateLogger(ServerOpConf, context.Background()); err != nil {
		return err
	}
	return nil
}
