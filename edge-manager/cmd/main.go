// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start edge-manager server
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-manager/pkg/logmanager"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/modulemanager"

	"edge-manager/pkg/appmanager"
	"edge-manager/pkg/cloudhub"
	"edge-manager/pkg/config"
	"edge-manager/pkg/configmanager"
	"edge-manager/pkg/database"
	"edge-manager/pkg/edgemsgmanager"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/restfulservice"
)

const (
	defaultPort           = 8101
	defaultWsPort         = 10000
	defaultMaxClientNum   = 1024
	defaultRunLogFile     = "/var/log/mindx-edge/edge-manager/run.log"
	defaultOperateLogFile = "/var/log/mindx-edge/edge-manager/operate.log"
	defaultBackupDirName  = "/var/log_backup/mindx-edge/edge-manager"
	defaultDbPath         = "/home/data/config/edge-manager.db"
	defaultKmcPath        = "/home/data/public-config/kmc-config.json"
	defaultOpLogMaxSize   = 200
	defaultRunLogMaxSize  = 400
	logMaxLineLength      = 512
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: defaultRunLogFile, FileMaxSize: defaultRunLogMaxSize,
		BackupDirName: defaultBackupDirName, MaxLineLength: logMaxLineLength}
	serverOpConf = &hwlog.LogConfig{LogFileName: defaultOperateLogFile, FileMaxSize: defaultOpLogMaxSize,
		BackupDirName: defaultBackupDirName}
	port         int
	wsPort       int
	ip           string
	version      bool
	dbPath       string
	maxClientNum int
)

func main() {
	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", config.BuildName, config.BuildVersion)
		return
	}
	if err := validateFlags(); err != nil {
		fmt.Printf("argument validation error: %s\n", err.Error())
		return
	}
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}
	podIp, err := common.GetPodIP()
	if err != nil {
		hwlog.RunLog.Errorf("get edge manager pod ip failed: %s", err.Error())
		return
	}
	ip = podIp

	if err := initResource(); err != nil {
		hwlog.RunLog.Errorf("initialize resource failed, %s", err.Error())
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
	flag.BoolVar(&version, "version", false, "Output the program version")
	flag.StringVar(&dbPath, "dbPath", defaultDbPath, "sqlite database path")
	flag.IntVar(&port, "port", defaultPort,
		"The server port of the http service,range[1025-65535]")
	flag.IntVar(&wsPort, "wsPort", defaultWsPort,
		"The server port of the websocket service,range[1025-65535]")
	flag.IntVar(&maxClientNum, "maxClientNum", defaultMaxClientNum,
		"The max number of connected edge client")
	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func validateFlags() error {
	if !checker.IsPortInRange(common.MinPort, common.MaxPort, port) {
		return fmt.Errorf("port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
	}
	if !checker.IsPortInRange(common.MinPort, common.MaxPort, wsPort) {
		return fmt.Errorf("wsPort %d is not in [%d, %d]", wsPort, common.MinPort, common.MaxPort)
	}
	if _, err := utils.CheckPath(dbPath); err != nil {
		return err
	}
	return nil
}

func initResource() error {
	if err := database.InitDB(dbPath); err != nil {
		hwlog.RunLog.Error("init database failed")
		return err
	}
	if _, err := kubeclient.NewClientK8s(""); err != nil {
		hwlog.RunLog.Error("init k8s failed")
		return err
	}
	err := common.InitKmcCfg(defaultKmcPath)
	if err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}
	return nil

}

func register(ctx context.Context) error {
	modulemanager.ModuleInit()
	if err := modulemanager.Registry(restfulservice.NewRestfulService(true, ip, port)); err != nil {
		return err
	}
	if err := modulemanager.Registry(nodemanager.NewNodeManager(true, ctx)); err != nil {
		return err
	}
	if err := modulemanager.Registry(appmanager.NewAppManager(true)); err != nil {
		return err
	}
	if err := modulemanager.Registry(cloudhub.NewCloudServer(true, wsPort, maxClientNum)); err != nil {
		return err
	}
	if err := modulemanager.Registry(edgemsgmanager.NewNodeMsgManager(true)); err != nil {
		return err
	}
	if err := modulemanager.Registry(configmanager.NewConfigManager(true)); err != nil {
		return err
	}
	if err := modulemanager.Registry(logmanager.NewLogManager(ctx, true)); err != nil {
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
