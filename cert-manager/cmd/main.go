// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start cert-manager server
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/modulemanager"

	"cert-manager/pkg/certmanager"
	"cert-manager/pkg/restful"
)

const (
	portConst      = 8103
	runLogFile     = "/var/log/mindx-edge/cert-manager/cert-manager-run.log"
	operateLogFile = "/var/log/mindx-edge/cert-manager/cert-manager-operate.log"
	backupDirName  = "/var/log_backup/mindx-edge/cert-manager"
	defaultKmcPath = "/home/data/public-config/kmc-config.json"
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf  = &hwlog.LogConfig{LogFileName: operateLogFile, BackupDirName: backupDirName}
	// BuildName cert-manager's build name
	BuildName string
	// BuildVersion cert-manager's build version
	BuildVersion string
	port         int
	ip           string
	version      bool
)

func main() {
	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}
	if inRange := checker.IsPortInRange(common.MinPort, common.MaxPort, port); !inRange {
		hwlog.RunLog.Errorf("port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
		return
	}
	podIp, err := common.GetPodIP()
	if err != nil {
		hwlog.RunLog.Errorf("get edge manager pod ip failed: %s", err.Error())
		return
	}
	ip = podIp

	if err = initResource(); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err = register(ctx); err != nil {
		hwlog.RunLog.Error("register error")
		return
	}
	gracefulShutdown(cancel)
}

func init() {
	flag.IntVar(&port, "port", portConst,
		"The server port of the http service,range[1025-40000]")
	flag.BoolVar(&version, "version", false, "Output the program version")

	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func initResource() error {
	err := common.InitKmcCfg(defaultKmcPath)
	if err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}
	return nil
}

func register(ctx context.Context) error {
	modulemanager.ModuleInit()
	if err := modulemanager.Registry(restful.NewRestfulService(true, ip, port)); err != nil {
		return err
	}
	if err := modulemanager.Registry(certmanager.NewCertManager(true)); err != nil {
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
