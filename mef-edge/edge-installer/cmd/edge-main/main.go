// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start edge-main
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/apimachinery/pkg/util/wait"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/alarm"
	"edge-installer/pkg/edge-main/cfgrestore"
	"edge-installer/pkg/edge-main/common/configpara"
	edgeMainDb "edge-installer/pkg/edge-main/common/database"
	"edge-installer/pkg/edge-main/common/resourcegc"
	"edge-installer/pkg/edge-main/edgeproxy"
	"edge-installer/pkg/edge-main/handlermgr"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
	"edge-installer/pkg/edge-main/subalarm"
)

var (
	// BuildName the program name
	BuildName string
	// BuildVersion the program version
	BuildVersion string

	version bool
)

const cpuStatusSyncInterval = 10 * time.Second

func init() {
	flag.BoolVar(&version, "version", false, "Output the program version")
}

func main() {
	if len(os.Args) < constants.MinArgsLen {
		fmt.Println("the required parameter is missing")
		os.Exit(constants.ProcessFailed)
	}

	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}

	if err := initLog(); err != nil {
		fmt.Println(err)
		return
	}

	if err := initResource(); err != nil {
		return
	}

	if err := edgeproxy.StartEdgeProxy(); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := register(ctx); err != nil {
		hwlog.RunLog.Error("register error")
		return
	}
	gracefulShutdown(ctx, cancel)
	hwlog.RunLog.Info("edge-main exit")
}

func initLog() error {
	if err := util.InitComponentLog(constants.EdgeMain); err != nil {
		return fmt.Errorf("initialize log failed, error: %v", err)
	}
	return nil
}

func initResource() error {
	cfgDir, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Error("get edge main config dir error")
		return err
	}
	dbPath := filepath.Join(cfgDir, constants.DbEdgeMainPath)
	if err = database.InitDB(dbPath); err != nil {
		hwlog.RunLog.Error("init database failed")
		return err
	}

	if err = edgeMainDb.InitMetaRepository(); err != nil {
		hwlog.RunLog.Error("init table failed")
		return err
	}

	if err = statusmanager.DeleteNodeStatus(); err != nil {
		hwlog.RunLog.Warnf("clear node status in database failed, error: %v", err)
	}

	kmcCfgDir := filepath.Join(cfgDir, constants.KmcCfgName)
	if err = backuputils.InitConfig(kmcCfgDir, kmc.InitKmcCfg); err != nil {
		hwlog.RunLog.Warnf("init edge main kmc config from json failed: %v, use default kmc config", err)
	}

	go wait.Until(util.WatchAndUpdateCPUTransientUsage, cpuStatusSyncInterval, wait.NeverStop)
	return nil
}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	netTypeStr, err := configpara.GetNetType()
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("current net type: %v", netTypeStr)
	modules := []model.Module{
		edgeproxy.NewEdgeOmProxy(true),
		edgeproxy.NewEdgeCoreProxy(true),
		cfgrestore.NewCfgRestore(ctx, true),
		alarm.NewAlarmManager(ctx, true),
		handlermgr.NewHandlerManager(ctx, true),
		subalarm.NewSubAlarmModule(ctx, true),
	}
	if netTypeStr == constants.FDWithOM {
		modules = append(modules, edgeproxy.NewDeviceOmProxy(true))
		go config.GetCapabilityCache().StartReportJob(ctx)
		go modeltask.GetModelReporter().StartReportJob(ctx)
		go resourcegc.NewResourceGCManager().StartGcJob(ctx)
	}

	modules = append(modules, moduleExt(netTypeStr)...)

	for _, mod := range modules {
		if err = modulemgr.Registry(mod); err != nil {
			return fmt.Errorf("registry %s error: %v", mod.Name(), err)
		}
	}

	modulemgr.Start()
	return nil
}

func gracefulShutdown(ctx context.Context, cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	select {
	case <-ctx.Done():
		hwlog.RunLog.Info("catch stop context is done")
	case _, ok := <-signalChan:
		if !ok {
			hwlog.RunLog.Info("catch stop signal channel is closed")
		}
	}
	cancelFunc()
}
