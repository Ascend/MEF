// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start edge-om
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-om/handlermgr"
	"edge-installer/pkg/edge-om/innerclient"
	"edge-installer/pkg/edge-om/logmgr"
	"edge-installer/pkg/edge-om/omjob"
	"edge-installer/pkg/edge-om/subalarm"
)

var (
	// BuildName the program name
	BuildName string
	// BuildVersion the program version
	BuildVersion string

	version bool
)

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

	ctx, cancel := context.WithCancel(context.Background())
	if err := initDb(ctx, cancel); err != nil {
		return
	}
	if err := register(ctx); err != nil {
		hwlog.RunLog.Error("register error")
		return
	}
	gracefulShutdown(ctx, cancel)
	hwlog.RunLog.Info("edge-om exit")
}

func initLog() error {
	if err := util.InitComponentLog(constants.EdgeOm); err != nil {
		return fmt.Errorf("initialize log failed, error: %s", err.Error())
	}

	cfgDir, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Error("get edge om config dir error")
		return err
	}

	kmcCfgDir := filepath.Join(cfgDir, constants.KmcCfgName)
	if err = backuputils.InitConfig(kmcCfgDir, kmc.InitKmcCfg); err != nil {
		hwlog.RunLog.Warnf("init edge om kmc config from json failed: %v, use default kmc config", err)
	}

	return nil
}

func initDb(ctx context.Context, cancel context.CancelFunc) error {
	backupCtx, err := util.StartBackupEdgeOmDb(ctx)
	if err != nil {
		return err
	}
	go func() {
		<-backupCtx.Done()
		cancel()
	}()

	cfgDir, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Error("get edge om config dir error")
		return err
	}
	dbPath := filepath.Join(cfgDir, constants.DbEdgeOmPath)
	if err = database.InitDB(dbPath); err != nil {
		hwlog.RunLog.Error("init database failed")
		return err
	}

	return nil
}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	modules := []model.Module{
		innerclient.NewEdgeClient(ctx, true),
		handlermgr.NewHandlerMgrModule(true),
		logmgr.NewLogMgr(constants.LogMgrName, ctx),
		omjob.NewOmJobModule(true, ctx),
		subalarm.NewSubAlarmModule(true),
	}
	modules = append(modules, moduleExt(ctx)...)

	for _, mod := range modules {
		if err := modulemgr.Registry(mod); err != nil {
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
