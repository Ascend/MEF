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
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

const (
	logDebugLv = iota - 1
	logInfoLv
	logWarnLv
	logErrorLv
	logCriticalLv
)

type logger struct {
	lgDebug      *log.Logger
	lgInfo       *log.Logger
	lgWarn       *log.Logger
	lgError      *log.Logger
	lgCritical   *log.Logger
	lgCtrl       *LogLimiter
	lgEscapeHtml bool
	lgLevel      int
	lgMaxLine    int
}

func (lg *logger) initLogWriter(w io.Writer) {
	lg.lgDebug = log.New(w, "[DEBUG]    ", log.Ldate|log.Lmicroseconds)
	lg.lgInfo = log.New(w, "[INFO]     ", log.Ldate|log.Lmicroseconds)
	lg.lgWarn = log.New(w, "[WARN]     ", log.Ldate|log.Lmicroseconds)
	lg.lgError = log.New(w, "[ERROR]    ", log.Ldate|log.Lmicroseconds)
	lg.lgCritical = log.New(w, "[Critical] ", log.Ldate|log.Lmicroseconds)
}

func (lg *logger) setLoggerLevel(lv int) {
	if lv < minLogLevel || lv > maxLogLevel {
		lg.lgLevel = 0
		return
	}
	lg.lgLevel = lv
}

func (lg *logger) setLoggerMaxLine(lml int) {
	if lml <= 0 || lml > maxEachLineLen {
		lg.lgMaxLine = defaultMaxEachLineLen
		return
	}
	lg.lgMaxLine = lml
}

func (lg *logger) setLoggerEscapeHtml(escapeHtml bool) {
	lg.lgEscapeHtml = escapeHtml
}

func (lg *logger) setLoggerWriter(config *LogConfig) {
	rollLogger := &Logs{
		FileName:                    config.LogFileName,
		backupDir:                   config.BackupDirName,
		Capacity:                    config.FileMaxSize, // megabytes
		SaveVolume:                  config.MaxBackups,
		SaveTime:                    config.MaxAge, // days
		isCompress:                  config.IsCompress,
		disableRotationIfSwitchUser: config.DisableRotationIfSwitchUser,
	}
	logWriter := &LogLimiter{
		Logs:        rollLogger,
		ExpiredTime: config.ExpiredTime, // seconds
		CacheSize:   config.CacheSize,
	}
	if config.OnlyToStdout {
		lg.initLogWriter(os.Stdout)
		return
	}
	if config.OnlyToFile {
		lg.initLogWriter(logWriter)
		return
	}
	writer := io.MultiWriter(os.Stdout, logWriter)
	lg.initLogWriter(writer)
	lg.lgCtrl = logWriter
}

func (lg *logger) setLogger(config *LogConfig) error {
	if err := validateLogConfigFiled(config); err != nil {
		return err
	}
	lg.setLoggerWriter(config)
	lg.setLoggerLevel(config.LogLevel)
	lg.setLoggerMaxLine(config.MaxLineLength)
	lg.setLoggerEscapeHtml(config.EscapeHtml)
	msg := fmt.Sprintf("%s's logger init success", path.Base(config.LogFileName))
	// skip change file mode and fs notify
	if config.OnlyToStdout {
		msg = fmt.Sprintf("%s, only to stdout", msg)
		return nil
	}
	lg.Info(msg)
	if err := os.Chmod(config.LogFileName, LogFileMode); err != nil {
		lg.Errorf("change file mode failed: %v", err)
		return fmt.Errorf("set log file mode failed")
	}
	return nil
}

func (lg *logger) isInit() bool {
	return lg.lgDebug != nil && lg.lgInfo != nil && lg.lgWarn != nil && lg.lgError != nil && lg.lgCritical != nil
}

// Debug record debug not format
func (lg *logger) Debug(args ...interface{}) {
	lg.DebugWithCtx(nil, args...)
}

// Debugf record debug
func (lg *logger) Debugf(format string, args ...interface{}) {
	lg.DebugfWithCtx(nil, format, args...)
}

// DebugWithCtx record Debug not format
func (lg *logger) DebugWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logDebugLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgDebug, fmt.Sprint(args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// DebugfWithCtx record Debug  format
func (lg *logger) DebugfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logDebugLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgDebug, fmt.Sprintf(format, args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// Info record info not format
func (lg *logger) Info(args ...interface{}) {
	lg.InfoWithCtx(nil, args...)
}

// Infof record info
func (lg *logger) Infof(format string, args ...interface{}) {
	lg.InfofWithCtx(nil, format, args...)
}

// InfoWithCtx record Info not format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) InfoWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logInfoLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgInfo, fmt.Sprint(args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// InfofWithCtx record Info  format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) InfofWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logInfoLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgInfo, fmt.Sprintf(format, args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// Warn record warn not format
func (lg *logger) Warn(args ...interface{}) {
	lg.WarnWithCtx(nil, args...)
}

// Warnf record warn
func (lg *logger) Warnf(format string, args ...interface{}) {
	lg.WarnfWithCtx(nil, format, args...)
}

// WarnWithCtx record Warn not format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) WarnWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logWarnLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgWarn, fmt.Sprint(args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// WarnfWithCtx record Warn  format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) WarnfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logWarnLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgWarn, fmt.Sprintf(format, args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// Error record error not format
func (lg *logger) Error(args ...interface{}) {
	lg.ErrorWithCtx(nil, args...)
}

// Errorf record error
func (lg *logger) Errorf(format string, args ...interface{}) {
	lg.ErrorfWithCtx(nil, format, args...)
}

// ErrorWithCtx record Error not format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) ErrorWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logErrorLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgError, fmt.Sprint(args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// ErrorfWithCtx record Error  format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) ErrorfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logErrorLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgError, fmt.Sprintf(format, args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// Critical record critical not format
func (lg *logger) Critical(args ...interface{}) {
	lg.CriticalWithCtx(nil, args...)
}

// Criticalf record Critical log format
func (lg *logger) Criticalf(format string, args ...interface{}) {
	lg.CriticalfWithCtx(nil, format, args...)
}

// CriticalWithCtx record Critical not format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) CriticalWithCtx(ctx context.Context, args ...interface{}) {
	if lg.lgLevel > logCriticalLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgCritical, fmt.Sprint(args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

// CriticalfWithCtx record Critical format with context, if you have no ctx, please use the method with not ctx
func (lg *logger) CriticalfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.lgLevel > logCriticalLv {
		return
	}
	if lg.validate() {
		printHelper(lg.lgCritical, fmt.Sprintf(format, args...), lg.lgMaxLine, lg.lgEscapeHtml, ctx)
	}
}

func (lg *logger) validate() bool {
	if lg == nil || !lg.isInit() {
		fmt.Println("Fatal function's logger is nil")
		return false
	}
	return true
}

// FlushMem writes the contents of the memory to the disk
func (lg *logger) FlushMem() error {
	return lg.lgCtrl.Flush()
}
