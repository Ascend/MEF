// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start software-manager server
package main

import (
	"context"
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker"
	"huawei.com/mindxedge/base/modulemanager"
	"software-manager/pkg/restfulservice"
	"software-manager/pkg/softwaremanager"
)

const (
	runLogFile     = "/var/log/mindx-edge/software-manager/run.log"
	operateLogFile = "/var/log/mindx-edge/software-manager/operate.log"
)

var (
	serverRunConf = &hwlog.LogConfig{}
	serverOpConf  = &hwlog.LogConfig{}
	version       bool
	buildName     string
	buildVersion  string
)

func main() {
	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", buildName, buildVersion)
		return
	}
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}
	softwaremanager.InitDatabase(softwaremanager.RepositoryFilesPath)
	if inRanage := checker.IsPortInRange(common.MinPort, common.MaxPort, softwaremanager.Port); !inRanage {
		hwlog.RunLog.Errorf("port %d is not in [%d, %d]", softwaremanager.Port, common.MinPort, common.MaxPort)
		return
	}
	if valid, err := checker.IsIpValid(softwaremanager.IP); !valid {
		hwlog.RunLog.Error(err)
		return
	}
	if err := register(); err != nil {
		hwlog.RunLog.Error("register error")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}

func init() {
	flag.BoolVar(&version, "version", false,
		"Output the program version")
	// hwOpLog configuration
	flag.IntVar(&serverOpConf.LogLevel, "operateLogLevel", 0,
		"Operation log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	flag.IntVar(&serverOpConf.MaxAge, "operateLogMaxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup operation log files, must be greater than or equal to 7 days")
	flag.StringVar(&serverOpConf.LogFileName, "operateLogFile", operateLogFile,
		"Operation log file path. If the file size exceeds 20MB, will be rotated")
	flag.IntVar(&serverOpConf.MaxBackups, "operateLogMaxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup operation logs, range (0, 30]")

	// hwRunLog configuration
	flag.IntVar(&serverRunConf.LogLevel, "runLogLevel", 0,
		"Run log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	flag.IntVar(&serverRunConf.MaxAge, "runLogMaxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup run log files, must be greater than or equal to 7 days")
	flag.StringVar(&serverRunConf.LogFileName, "runLogFile", runLogFile,
		"Run log file path. If the file size exceeds 20MB, will be rotated")
	flag.IntVar(&serverRunConf.MaxBackups, "runLogMaxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup run logs, range (0, 30]")

}

func register() error {
	modulemanager.ModuleInit()
	if err := modulemanager.Registry(restfulservice.
		NewRestfulService(true, softwaremanager.IP, softwaremanager.Port)); err != nil {
		return err
	}
	if err := modulemanager.Registry(softwaremanager.NewSoftwareManager(true)); err != nil {
		return err
	}
	modulemanager.Start()
	return nil
}
