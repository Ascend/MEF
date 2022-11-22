// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start software-manager server
package main

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

const (
	runLogFile     = "/var/log/mindx-edge/software-manager/run.log"
	operateLogFile = "/var/log/mindx-edge/software-manager/operate.log"
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: runLogFile}
	serverOpConf  = &hwlog.LogConfig{LogFileName: operateLogFile}
)

func main() {
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}
	hwlog.RunLog.Info("start software manager")
}
