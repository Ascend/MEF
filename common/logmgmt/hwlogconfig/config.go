// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package hwlogconfig provides utils to set up hwlog.
package hwlogconfig

import (
	"flag"
	"fmt"
	"reflect"

	"huawei.com/mindx/common/hwlog"
)

var defaultOpConf = hwlog.LogConfig{
	MaxAge:        30,
	FileMaxSize:   100,
	MaxBackups:    10,
	MaxLineLength: 256,
	ExpiredTime:   1,
	CacheSize:     10240,
	IsCompress:    true,
}

var defaultRunConf = hwlog.LogConfig{
	MaxAge:        30,
	FileMaxSize:   100,
	MaxBackups:    30,
	MaxLineLength: 256,
	ExpiredTime:   1,
	CacheSize:     10240,
	IsCompress:    true,
}

// BindFlags is wrapper of command flags
func BindFlags(serverOpConf, serverRunConf *hwlog.LogConfig) {
	opLogDefaults, runLogDefaults := defaultOpConf, defaultRunConf
	setDefaults(serverOpConf, &opLogDefaults)
	setDefaults(serverRunConf, &runLogDefaults)

	// hwOpLog configurations
	flag.IntVar(&serverOpConf.LogLevel, "operateLogLevel", opLogDefaults.LogLevel,
		fmt.Sprintf("Operation log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical (default %d)",
			opLogDefaults.LogLevel))
	flag.IntVar(&serverOpConf.MaxAge, "operateLogMaxAge", opLogDefaults.MaxAge,
		fmt.Sprintf("Maximum number of days for backup operation log files,"+
			"must be greater than or equal to %d days", opLogDefaults.MaxAge))
	flag.StringVar(&serverOpConf.LogFileName, "operateLogFile", opLogDefaults.LogFileName,
		fmt.Sprintf("Operation log file path. If the file size exceeds %dMB, will be rotated",
			opLogDefaults.FileMaxSize))
	flag.IntVar(&serverOpConf.MaxBackups, "operateLogMaxBackups", opLogDefaults.MaxBackups,
		fmt.Sprintf("Maximum number of backup operation logs, range (0, %d]", opLogDefaults.MaxBackups))

	// hwRunLog configurations
	flag.IntVar(&serverRunConf.LogLevel, "runLogLevel", runLogDefaults.LogLevel,
		fmt.Sprintf("Run log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical (default %d)",
			runLogDefaults.LogLevel))
	flag.IntVar(&serverRunConf.MaxAge, "runLogMaxAge", runLogDefaults.MaxAge,
		fmt.Sprintf("Maximum number of days for backup run log files, must be greater than or equal to %d days",
			runLogDefaults.MaxAge))
	flag.StringVar(&serverRunConf.LogFileName, "runLogFile", runLogDefaults.LogFileName,
		fmt.Sprintf("Run log file path. If the file size exceeds %dMB, will be rotated",
			runLogDefaults.FileMaxSize))
	flag.IntVar(&serverRunConf.MaxBackups, "runLogMaxBackups", runLogDefaults.MaxBackups,
		fmt.Sprintf("Maximum number of backup run logs, range (0, %d]", runLogDefaults.MaxBackups))
}

func setDefaults(confOverride, defaults *hwlog.LogConfig) {
	overrideVal := reflect.ValueOf(confOverride).Elem()
	defaultsVal := reflect.ValueOf(defaults).Elem()
	configType := reflect.TypeOf(hwlog.LogConfig{})
	for i := 0; i < configType.NumField(); i++ {
		if overrideVal.Field(i).IsZero() {
			overrideVal.Field(i).Set(defaultsVal.Field(i))
		} else {
			defaultsVal.Field(i).Set(overrideVal.Field(i))
		}
	}
}
