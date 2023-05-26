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

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/logmgmt/logrotate"
	"huawei.com/mindx/common/modulemgr"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"

	"nginx-manager/pkg/nginxcom"
	"nginx-manager/pkg/nginxlogrotate"
	"nginx-manager/pkg/nginxmgr"
	"nginx-manager/pkg/nginxmonitor"
)

const (
	runLogFile     = "/home/MEFCenter/logs/nginx-manager-run.log"
	operateLogFile = "/home/MEFCenter/logs/nginx-manager-operate.log"
	backupDirName  = "/home/MEFCenter/logs_backup"
	defaultKmcPath = "/home/data/public-config/kmc-config.json"
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf  = &hwlog.LogConfig{LogFileName: operateLogFile, BackupDirName: backupDirName}
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
	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func initResource() error {
	if err := nginxcom.GetEnvManager().Load(); err != nil {
		return err
	}

	err := kmc.InitKmcCfg(defaultKmcPath)
	if err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}
	return nil
}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(nginxmgr.NewNginxManager(true, ctx)); err != nil {
		return err
	}
	if err := modulemgr.Registry(nginxmonitor.NewNginxMonitor(true, ctx)); err != nil {
		return err
	}
	if err := modulemgr.Registry(logrotate.Module("", ctx, nginxlogrotate.Setup)); err != nil {
		return err
	}

	modulemgr.Start()
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
