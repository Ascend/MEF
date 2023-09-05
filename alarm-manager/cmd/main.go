// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start cert-manager server
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/checker/valid"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr"

	"alarm-manager/pkg/alarmmanager"
	"alarm-manager/pkg/restful"
	"alarm-manager/pkg/websocket"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
)

var (
	serverRunConf = &hwlog.LogConfig{OnlyToFile: true, LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf  = &hwlog.LogConfig{OnlyToFile: true, LogFileName: operateLogFile, BackupDirName: backupDirName}
	port          int
	ip            string
	dbPath        string
)

const (
	defaultPort    = 8102
	runLogFile     = "/var/log/mindx-edge/alarm-manager/alarm-manager-run.log"
	operateLogFile = "/var/log/mindx-edge/alarm-manager/alarm-manager-operate.log"
	backupDirName  = "/var/log_backup/mindx-edge/alarm-manager"
	defaultKmcPath = "/home/data/public-config/kmc-config.json"
	defaultDbPath  = "/home/data/config/alarm-manager.db"
)

func main() {
	flag.Parse()
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}

	if inRange := valid.IsPortInRange(common.MinPort, common.MaxPort, port); !inRange {
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
	flag.IntVar(&port, "port", defaultPort, "The server port of the http service,range[1025-40000]")
	flag.StringVar(&dbPath, "dbPath", defaultDbPath, "sqlite database path")
	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func initResource() error {
	if err := database.InitDB(dbPath); err != nil {
		hwlog.RunLog.Error("init database failed")
		return errors.New("init database failed")
	}

	if err := alarmmanager.AlarmDbInstance().DeleteAlarmTable(); err != nil {
		hwlog.RunLog.Errorf("clear alarm info table failed: %s", err.Error())
		return errors.New("clear alarm info table failed")
	}

	if err := database.CreateTableIfNotExist(alarmmanager.AlarmInfo{}); err != nil {
		hwlog.RunLog.Error("create alarm info table failed")
		return errors.New("create alarm info table failed")
	}

	err := kmc.InitKmcCfg(defaultKmcPath)
	if err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}

	return nil
}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(restful.NewRestfulService(true, ip, port)); err != nil {
		return err
	}
	if err := modulemgr.Registry(websocket.NewAlarmWsClient(true, ctx)); err != nil {
		return err
	}
	if err := modulemgr.Registry(alarmmanager.NewAlarmManager(dbPath, true, ctx)); err != nil {
		return err
	}
	modulemgr.Start()
	return nil
}
