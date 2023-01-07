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

	"edge-manager/pkg/appmanager"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/restfulservice"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker"
	"huawei.com/mindxedge/base/modulemanager"

	"edge-manager/pkg/config"
	"edge-manager/pkg/database"
	"edge-manager/pkg/edgeconnector"
	"edge-manager/pkg/edgeinstaller"
	"edge-manager/pkg/kubeclient"
)

const (
	defaultPort           = 8101
	defaultWsPort         = 10000
	defaultRunLogFile     = "/var/log/mindx-edge/edge-manager/run.log"
	defaultOperateLogFile = "/var/log/mindx-edge/edge-manager/operate.log"
	defaultDbPath         = "/etc/mindx-edge/edge-manager/edge-manager.db"
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: defaultRunLogFile}
	serverOpConf  = &hwlog.LogConfig{LogFileName: defaultOperateLogFile}
	port          int
	wsPort        int
	ip            string
	version       bool
	// kubeConfig Kube config path
	kubeConfig string
	dbPath     string
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
	flag.StringVar(&kubeConfig, "kubeconfig", "", "The k8s master config file")
	flag.StringVar(&ip, "ip", "",
		"The listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.IntVar(&port, "port", defaultPort,
		"The server port of the http service,range[1025-65535]")
	flag.IntVar(&wsPort, "wsPort", defaultWsPort,
		"The server port of the websocket service,range[1025-65535]")

	// hwOpLog configuration
	flag.IntVar(&serverOpConf.LogLevel, "operateLogLevel", 0,
		"Operation log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	flag.IntVar(&serverOpConf.MaxAge, "operateLogMaxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup operation log files, must be greater than or equal to 7 days")
	flag.StringVar(&serverOpConf.LogFileName, "operateLogFile", defaultOperateLogFile,
		"Operation log file path. If the file size exceeds 20MB, will be rotated")
	flag.IntVar(&serverOpConf.MaxBackups, "operateLogMaxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup operation logs, range (0, 30]")

	// hwRunLog configuration
	flag.IntVar(&serverRunConf.LogLevel, "runLogLevel", 0,
		"Run log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	flag.IntVar(&serverRunConf.MaxAge, "runLogMaxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup run log files, must be greater than or equal to 7 days")
	flag.StringVar(&serverRunConf.LogFileName, "runLogFile", defaultRunLogFile,
		"Run log file path. If the file size exceeds 20MB, will be rotated")
	flag.IntVar(&serverRunConf.MaxBackups, "runLogMaxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup run logs, range (0, 30]")
}

func validateFlags() error {
	if !checker.IsPortInRange(common.MinPort, common.MaxPort, port) {
		return fmt.Errorf("port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
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
	if _, err := kubeclient.NewClientK8s(kubeConfig); err != nil {
		hwlog.RunLog.Error("init k8s failed")
		return err
	}
	return nil

}

func register(ctx context.Context) error {
	modulemanager.ModuleInit()
	if err := modulemanager.Registry(restfulservice.NewRestfulService(true, port)); err != nil {
		return err
	}
	if err := modulemanager.Registry(nodemanager.NewNodeManager(true, ctx)); err != nil {
		return err
	}
	if err := modulemanager.Registry(appmanager.NewAppManager(true)); err != nil {
		return err
	}
	if err := modulemanager.Registry(edgeconnector.NewConnector(true, wsPort)); err != nil {
		return err
	}
	if err := modulemanager.Registry(edgeinstaller.NewInstaller(true)); err != nil {
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
	case <-signalChan:
	}
	cancelFunc()
}
