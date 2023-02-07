// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start nginx-manager
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"nginx-manager/pkg/checker"
	"nginx-manager/pkg/database"
	"nginx-manager/pkg/nginxcom"
	"nginx-manager/pkg/nginxmgr"
	"nginx-manager/pkg/nginxmonitor"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
)

const (
	runLogFile     = "/home/MEFCenter/logs/run.log"
	operateLogFile = "/home/MEFCenter/logs/operate.log"
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: runLogFile}
	serverOpConf  = &hwlog.LogConfig{LogFileName: operateLogFile}
	restfulPort   = 8080
)

func main() {
	flag.Parse()

	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}
	if err := initResource(); err != nil {
		hwlog.RunLog.Errorf("initialize resource failed, %s", err.Error())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := register(ctx); err != nil {
		hwlog.RunLog.Errorf("register module failed, %s", err.Error())
		return
	}
	gracefulShutdown(cancel)
}

func init() {
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

func initResource() error {
	nginxcom.InitEnvs()
	if err := checker.Check(checker.Env, nginxcom.Envs); err != nil {
		return err
	}
	usrMgrPort, err := nginxcom.GetEnvAsInt(nginxcom.UserMgrSvcPortKey)
	if err != nil {
		return err
	}
	restfulPort = usrMgrPort
	if err := database.InitDB(nginxcom.DefaultDbPath); err != nil {
		hwlog.RunLog.Errorf("init database failed, error: %s", err.Error())
		return err
	}
	return nil
}

func register(ctx context.Context) error {
	modulemanager.ModuleInit()
	if err := modulemanager.Registry(nginxmgr.NewNginxManager(true, ctx)); err != nil {
		return err
	}
	if err := modulemanager.Registry(nginxmonitor.NewNginxMonitor(true, ctx)); err != nil {
		return err
	}
	modulemanager.Start()
	return nil
}

func gracefulShutdown(cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	select {
	case _, ok := <-signalChan:
		if !ok {
			hwlog.RunLog.Info("catch stop signal channel is closed")
		}
	}
	cancelFunc()
}
