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

	"nginx-manager/pkg/database"
	"nginx-manager/pkg/nginxcom"
	"nginx-manager/pkg/nginxmgr"
	"nginx-manager/pkg/nginxmonitor"
	"nginx-manager/pkg/usermgr"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/common/logmgmt/logrotate"
	"huawei.com/mindxedge/base/modulemanager"
	"nginx-manager/pkg/nginxlogrotate"
)

const (
	runLogFile     = "/home/MEFCenter/logs/run.log"
	operateLogFile = "/home/MEFCenter/logs/operate.log"
	backupDirName  = "/home/MEFCenter/logs_backup"
	defaultKmcPath = "/home/data/public-config/kmc-config.json"
	defaultPort    = 8080
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf  = &hwlog.LogConfig{LogFileName: operateLogFile, BackupDirName: backupDirName}
	restfulPort   = defaultPort
	ip            string
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
	usrMgrPort, err := nginxcom.GetEnvManager().GetInt(nginxcom.UserMgrSvcPortKey)
	if err != nil {
		return err
	}
	restfulPort = usrMgrPort
	if err = database.InitDB(nginxcom.DefaultDbPath); err != nil {
		hwlog.RunLog.Errorf("init database failed, error: %s", err.Error())
		return err
	}
	podIp, err := common.GetPodIP()
	if err != nil {
		hwlog.RunLog.Errorf("get nginx manager pod ip failed: %s", err.Error())
		return err
	}
	ip = podIp
	err = common.InitKmcCfg(defaultKmcPath)
	if err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
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
	if err := modulemanager.Registry(usermgr.NewRestfulService(true, ip, restfulPort)); err != nil {
		return err
	}
	if err := modulemanager.Registry(usermgr.NewUserManager(true, ctx)); err != nil {
		return err
	}

	if err := modulemanager.Registry(logrotate.Module("", ctx, nginxlogrotate.Setup)); err != nil {
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
