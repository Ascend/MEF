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

// CryptoLogger  kmc log call back interface
type CryptoLogger interface {
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
	Trace(msg string)
	Log(msg string)
}

// LoggerAdaptor kmc log adaptor
type LoggerAdaptor struct {
}

// Error print error log
func (kla *LoggerAdaptor) Error(msg string) {
	RunLog.Error(msg)
}

// Warn print warning log
func (kla *LoggerAdaptor) Warn(msg string) {
	RunLog.Warn(msg)
}

// Info print info log
func (kla *LoggerAdaptor) Info(msg string) {
	RunLog.Info(msg)
}

// Debug print debug log
func (kla *LoggerAdaptor) Debug(msg string) {
	RunLog.Debug(msg)
}

// Trace print trace log
func (kla *LoggerAdaptor) Trace(msg string) {
	RunLog.Debug(msg)
}

// Log print log
func (kla *LoggerAdaptor) Log(msg string) {
	RunLog.Info(msg)
}
