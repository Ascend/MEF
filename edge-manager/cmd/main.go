// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start edge-manager server
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/utils"

	"edge-manager/pkg/appmanager"
	"edge-manager/pkg/cloudhub"
	"edge-manager/pkg/config"
	"edge-manager/pkg/configmanager"
	"edge-manager/pkg/edgemsgmanager"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/restfulservice"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
)

const (
	defaultPort           = 8101
	defaultWsPort         = 10000
	defaultAuthPort       = 10001
	defaultMaxClientNum   = 1024
	defaultRunLogFile     = "/var/log/mindx-edge/edge-manager/edge-manager-run.log"
	defaultOperateLogFile = "/var/log/mindx-edge/edge-manager/edge-manager-operate.log"
	defaultBackupDirName  = "/var/log_backup/mindx-edge/edge-manager"
	defaultDbPath         = "/home/data/config/edge-manager.db"
	defaultPodConfigPath  = "/home/data/config/pod-config.json"
	defaultAuthConfigPath = "/home/data/config/auth-config.json"
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
	authPort     int
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

	if err = initResource(); err != nil {
		hwlog.RunLog.Errorf("initialize resource failed, %s", err.Error())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err = register(ctx); err != nil {
		hwlog.RunLog.Error("register error")
		return
	}
	common.GracefulShutDown(cancel)
}

func init() {
	flag.BoolVar(&version, "version", false, "Output the program version")
	flag.StringVar(&dbPath, "dbPath", defaultDbPath, "sqlite database path")
	flag.IntVar(&port, "port", defaultPort,
		"The server port of the http service,range[1025-65535]")
	flag.IntVar(&wsPort, "wsPort", defaultWsPort,
		"The server port of the websocket service,range[1025-65535]")
	flag.IntVar(&authPort, "authPort", defaultAuthPort,
		"The server port of the edge auth service,range[1025-65535]")
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
	if !checker.IsPortInRange(common.MinPort, common.MaxPort, authPort) {
		return fmt.Errorf("authPort %d is not in [%d, %d]", authPort, common.MinPort, common.MaxPort)
	}
	if authPort == wsPort {
		return fmt.Errorf("authPort can not equals to wsPort")
	}
	if _, err := utils.CheckPath(dbPath); err != nil {
		return err
	}
	return nil
}

func initPodConfig() error {
	date, err := utils.LoadFile(defaultPodConfigPath)
	if err != nil {
		return fmt.Errorf("load pod config file failed, %s", err.Error())
	}
	podConfig := config.PodConfig
	if err = json.Unmarshal(date, &podConfig); err != nil {
		return errors.New("unmarshal pod config failed")
	}

	config.PodConfig.HostPath = config.CheckAndModifyHostPath(podConfig.HostPath)
	return nil
}

func initAuthConfig() error {
	date, err := utils.LoadFile(defaultAuthConfigPath)
	if err != nil {
		return fmt.Errorf("load auth config file failed, %s", err.Error())
	}
	var authConfig config.AuthInfo
	if err = json.Unmarshal(date, &authConfig); err != nil {
		return errors.New("unmarshal auth config failed")
	}
	if err := config.CheckAuthConfig(authConfig); err != nil {
		return err
	}

	config.SetConfig(authConfig)
	return nil
}

func initResource() error {
	if err := database.InitDB(dbPath); err != nil {
		hwlog.RunLog.Error("init database failed")
		return err
	}
	if _, err := kubeclient.NewClientK8s(); err != nil {
		hwlog.RunLog.Errorf("init k8s failed: %v", err)
		return err
	}
	if err := initPodConfig(); err != nil {
		hwlog.RunLog.Errorf("init pod config failed")
		return err
	}
	if err := initAuthConfig(); err != nil {
		hwlog.RunLog.Errorf("init auth config error %v", err)
		return err
	}
	if err := kmc.InitKmcCfg(defaultKmcPath); err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}
	return nil

}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(restfulservice.NewRestfulService(true, ip, port)); err != nil {
		return err
	}
	if err := modulemgr.Registry(nodemanager.NewNodeManager(true, ctx)); err != nil {
		return err
	}
	if err := modulemgr.Registry(appmanager.NewAppManager(true)); err != nil {
		return err
	}
	if err := modulemgr.Registry(cloudhub.NewCloudServer(true, wsPort, authPort, maxClientNum)); err != nil {
		return err
	}
	if err := modulemgr.Registry(edgemsgmanager.NewNodeMsgManager(true)); err != nil {
		return err
	}
	if err := modulemgr.Registry(configmanager.NewConfigManager(true)); err != nil {
		return err
	}

	modulemgr.Start()
	return nil
}
