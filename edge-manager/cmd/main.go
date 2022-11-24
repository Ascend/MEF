// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start edge-manager server
package main

import (
	"context"
	"flag"
	"fmt"

	"edge-manager/module_manager"
	"edge-manager/pkg/appmanager"
	"edge-manager/pkg/common"
	"edge-manager/pkg/common/checker"
	"edge-manager/pkg/database"
	"edge-manager/pkg/edgeconnector"
	"edge-manager/pkg/edgeinstaller"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/restfulservice"

	"huawei.com/mindx/common/hwlog"
)

const (
	portConst      = 8101
	runLogFile     = "/var/log/mindx-edge/edge-manager/run.log"
	operateLogFile = "/var/log/mindx-edge/edge-manager/operate.log"
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: runLogFile}
	serverOpConf  = &hwlog.LogConfig{LogFileName: operateLogFile}
	buildName     string
	buildVersion  string
	port          int
	ip            string
	version       bool
	// Kubeconfig Kube config path
	Kubeconfig string
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
	if inRanage := checker.IsPortInRange(common.MinPort, common.MaxPort, port); !inRanage {
		hwlog.RunLog.Errorf("port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
		return
	}
	if valid, err := checker.IsIpValid(ip); !valid {
		hwlog.RunLog.Error(err)
		return
	}
	if err := initResource(); err != nil {
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
	flag.IntVar(&port, "port", portConst,
		"The server port of the http service,range[1025-40000]")
	flag.StringVar(&ip, "ip", "",
		"The listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.BoolVar(&version, "version", false, "Output the program version")

	flag.StringVar(&Kubeconfig, "kubeconfig", "",
		"The k8s master config file")
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
	restfulservice.BuildNameStr = buildName
	restfulservice.BuildVersionStr = buildVersion
	if err := database.InitDB(); err != nil {
		hwlog.RunLog.Error("init database failed")
		return err
	}
	if _, err := kubeclient.NewClientK8s(); err != nil {
		hwlog.RunLog.Error("init k8s failed")
		return err
	}
	return nil

}

func register() error {
	module_manager.ModuleManagerInit()
	if err := module_manager.Registry(restfulservice.NewRestfulService(true, ip, port)); err != nil {
		return err
	}
	if err := module_manager.Registry(nodemanager.NewNodeManager(true)); err != nil {
		return err
	}
	if err := module_manager.Registry(appmanager.NewAppManager(true)); err != nil {
		return err
	}
	if err := module_manager.Registry(edgeconnector.NewSocket(true)); err != nil {
		return err
	}
	if err := module_manager.Registry(edgeinstaller.NewInstaller(true)); err != nil {
		return err
	}
	module_manager.Start()
	return nil
}
