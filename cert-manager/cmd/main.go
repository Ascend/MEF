// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start cert-manager server
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"

	"cert-manager/pkg/certmanager"
	"cert-manager/pkg/config"
	"cert-manager/pkg/restful"
)

const (
	portConst             = 8103
	runLogFile            = "/var/log/mindx-edge/cert-manager/cert-manager-run.log"
	operateLogFile        = "/var/log/mindx-edge/cert-manager/cert-manager-operate.log"
	backupDirName         = "/var/log_backup/mindx-edge/cert-manager"
	defaultKmcPath        = "/home/data/public-config/kmc-config.json"
	defaultCertConfigPath = "/home/data/config/cert-config.json"
)

var (
	serverRunConf = &hwlog.LogConfig{OnlyToFile: true, LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf  = &hwlog.LogConfig{OnlyToFile: true, LogFileName: operateLogFile, BackupDirName: backupDirName}
	// BuildName cert-manager's build name
	BuildName string
	// BuildVersion cert-manager's build version
	BuildVersion string
	port         int
	ip           string
	version      bool
)

func main() {
	if len(os.Args) <= util.NoArgCount {
		fmt.Println("the required parameter is missing")
		os.Exit(util.ErrorExitCode)
	}

	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}
	if res := checker.GetIntChecker("", common.MinPort, common.MaxPort, true).Check(port); !res.Result {
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
	common.GracefulShutDown(cancel)
}

func init() {
	flag.IntVar(&port, "port", portConst,
		"The server port of the http service,range[1025-40000]")
	flag.BoolVar(&version, "version", false, "Output the program version")

	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func initResource() error {
	if err := backuputils.InitConfig(defaultKmcPath, kmc.InitKmcCfg); err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}
	if err := backuputils.InitConfig(defaultCertConfigPath, initCertConfig); err != nil {
		hwlog.RunLog.Errorf("init auth config error %v", err)
		return err
	}
	return nil
}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(restful.NewRestfulService(true, ip, port)); err != nil {
		return err
	}
	if err := modulemgr.Registry(certmanager.NewCertManager(true)); err != nil {
		return err
	}
	modulemgr.Start()
	return nil
}

func initCertConfig(configPath string) error {
	data, err := fileutils.LoadFile(configPath)
	if err != nil {
		return fmt.Errorf("load auth config file failed, %s", err.Error())
	}
	var certConfig config.CertConfigInfo
	if err = json.Unmarshal(data, &certConfig); err != nil {
		return errors.New("unmarshal auth config failed")
	}
	if err := config.CheckCertConfig(certConfig); err != nil {
		return err
	}

	config.SetConfig(certConfig)
	return nil
}
