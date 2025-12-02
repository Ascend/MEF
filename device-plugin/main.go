/* Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
*/

// Package main implements initialization of the startup parameters of the device plugin.
package main

import (
	"context"
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/devmanager"
	"Ascend-device-plugin/pkg/server"
)

const (
	// socket name
	defaultLogPath = "/var/log/mindx-dl/devicePlugin/devicePlugin.log"

	// defaultListWatchPeriod is the default listening device state's period
	defaultListWatchPeriod = 5

	// maxListWatchPeriod is the max listening device state's period
	maxListWatchPeriod = 60
	// minListWatchPeriod is the min listening device state's period
	minListWatchPeriod = 3
	maxLogLineLength   = 1024
	logLevel           = 0
)

var (
	fdFlag          = flag.Bool("fdFlag", false, "Whether to use fd system to manage device (default false)")
	useAscendDocker = flag.Bool("useAscendDocker", false, "Whether to use ascend docker. "+
		"This parameter will be deprecated in future versions")
	version     = flag.Bool("version", false, "Output version information")
	edgeLogFile = flag.String("edgeLogFile", "/var/alog/AtlasEdge_log/devicePlugin.log",
		"Log file path in edge scene")
	listWatchPeriod = flag.Int("listWatchPeriod", defaultListWatchPeriod,
		"Listen and watch device state's period, unit second, range [3, 60]")
	logFile = flag.String("logFile", defaultLogPath,
		"The log file path, if the file size exceeds 20MB, will be rotate")
	shareDevCount = flag.Uint("shareDevCount", 1, "share device function, enable the func by setting "+
		"a value greater than 1, range is [1, 100], only support 310B")
)

var (
	// BuildName show app name
	BuildName string
	// BuildVersion show app version
	BuildVersion string
)

func initLogModule(ctx context.Context) error {
	var loggerPath string
	loggerPath = *logFile
	if *fdFlag {
		loggerPath = *edgeLogFile
	}
	if !common.CheckFileUserSameWithProcess(loggerPath) {
		return fmt.Errorf("check log file failed")
	}
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   loggerPath,
		LogLevel:      logLevel,
		MaxBackups:    common.MaxBackups,
		MaxAge:        common.MaxAge,
		MaxLineLength: maxLogLineLength,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %#v\n", err)
		return err
	}
	return nil
}

func checkParam() bool {
	if *listWatchPeriod < minListWatchPeriod || *listWatchPeriod > maxListWatchPeriod {
		hwlog.RunLog.Errorf("list and watch period %d out of range", *listWatchPeriod)
		return false
	}
	return checkShareDevCount()
}

func checkShareDevCount() bool {
	if *shareDevCount < 1 || *shareDevCount > common.MaxShareDevCount {
		hwlog.RunLog.Error("share device function params invalid")
		return false
	}
	return true
}

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := initLogModule(ctx); err != nil {
		return
	}
	if !checkParam() {
		return
	}
	hwlog.RunLog.Infof("ascend device plugin starting and the version is %s", BuildVersion)
	setParameters()
	hdm, err := InitFunction()
	if err != nil {
		return
	}

	go hdm.ListenDevice(ctx)
	hdm.SignCatch(cancel)
}

// InitFunction init function
func InitFunction() (*server.HwDevManager, error) {
	devM, err := devmanager.AutoInit("")
	if err != nil {
		hwlog.RunLog.Errorf("init devmanager failed, err: %#v", err)
		return nil, err
	}
	hdm := server.NewHwDevManager(devM)
	if hdm == nil {
		hwlog.RunLog.Error("init device manager failed")
		return nil, fmt.Errorf("init device manager failed")
	}
	hwlog.RunLog.Info("init device manager success")
	return hdm, nil
}

func setParameters() {
	common.ParamOption = common.Option{
		GetFdFlag:          *fdFlag,
		ListAndWatchPeriod: *listWatchPeriod,
		ShareCount:         *shareDevCount,
	}
}
