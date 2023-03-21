// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start software-manager server
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"software-manager/pkg/restfulservice"
	"software-manager/pkg/softwaremanager"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/modulemanager"
)

const (
	runLogFile     = "/var/log/mindx-edge/software-manager/software-manager-run.log"
	operateLogFile = "/var/log/mindx-edge/software-manager/software-manager-operate.log"
	backupDirName  = "/var/log_backup/mindx-edge/software-manager"
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf  = &hwlog.LogConfig{LogFileName: operateLogFile, BackupDirName: backupDirName}
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
	if err := softwaremanager.InitDB(); err != nil {
		hwlog.RunLog.Errorf(err.Error())
	}
	if inRanage := checker.IsPortInRange(common.MinPort, common.MaxPort, softwaremanager.Port); !inRanage {
		hwlog.RunLog.Errorf("port %d is not in [%d, %d]", softwaremanager.Port, common.MinPort, common.MaxPort)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := register(ctx); err != nil {
		hwlog.RunLog.Error("register error")
		return
	}
	gracefulShutdown(cancel)
}

func init() {
	flag.BoolVar(&version, "version", false,
		"Output the program version")

	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func register(ctx context.Context) error {
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
