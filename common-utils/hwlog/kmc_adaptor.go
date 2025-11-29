//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

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
