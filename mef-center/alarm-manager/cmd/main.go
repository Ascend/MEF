// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package main to start alarm-manager
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/checker/valid"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr"

	"alarm-manager/pkg/alarmmanager"
	"alarm-manager/pkg/restful"
	"alarm-manager/pkg/websocket"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

var (
	serverRunConf  = &hwlog.LogConfig{OnlyToFile: true, LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf   = &hwlog.LogConfig{OnlyToFile: true, LogFileName: operateLogFile, BackupDirName: backupDirName}
	port           int
	ip             string
	dbPath         string
	limitIPReq     string
	concurrency    int
	cacheSize      int
	limitIPConn    int
	limitTotalConn int
	dataLimit      int64
)

const (
	defaultPort    = 8102
	runLogFile     = "/var/log/mindx-edge/alarm-manager/alarm-manager-run.log"
	operateLogFile = "/var/log/mindx-edge/alarm-manager/alarm-manager-operate.log"
	backupDirName  = "/var/log_backup/mindx-edge/alarm-manager"
	defaultKmcPath = "/home/data/public-config/kmc-config.json"
	defaultDbPath  = "/home/data/config/alarm-manager.db"

	maxIPConnLimit     = 100
	maxConcurrency     = 100
	defaultConnection  = 100
	defaultConcurrency = 100
	maxBurstIP         = 100
	defaultDataLimit   = 1024 * 1024
	defaultCacheSize   = 1024 * 1024 * 10
)

func main() {
	if len(os.Args) < util.NoArgCount {
		fmt.Println("the required parameter is missing")
		os.Exit(util.ErrorExitCode)
	}

	flag.Parse()
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}

	if err := checker(); err != nil {
		hwlog.RunLog.Errorf("parameter check error: %v", err)
		return
	}

	podIp, err := common.GetPodIP()
	if err != nil {
		hwlog.RunLog.Errorf("get alarm manager pod ip failed: %s", err.Error())
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
	flag.IntVar(&port, "port", defaultPort, "The server port of the http service,range[1025-65535]")
	flag.StringVar(&dbPath, "dbPath", defaultDbPath, "sqlite database path")
	flag.IntVar(&cacheSize, "cacheSize", defaultCacheSize, "the cacheSize for ip limit,"+
		"keep default normally")
	flag.IntVar(&limitIPConn, "limitIPConn", defaultConcurrency, "the tcp connection limit for each Ip")
	flag.IntVar(&limitTotalConn, "limitTotalConn", defaultConnection, "the tcp connection limit for all request")
	flag.StringVar(&limitIPReq, "limitIPReq", "2/1",
		"the http request limit counts for each Ip,2/1 means allow 2 request in 1 seconds")
	flag.IntVar(&concurrency, "concurrency", defaultConcurrency,
		"The max concurrency of the http server, range is [1-512]")
	flag.Int64Var(&dataLimit, "dataLimit", defaultDataLimit,
		"bytes, limit the data size of request's body, the default value is 1MB")
	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func initResource() error {
	opts := database.Options{
		EnableBackup:      true,
		BackupDbPath:      dbPath + common.BackupDbSuffix,
		TestInterval:      common.DbTestInterval,
		EnableAutoRecover: true,
	}
	if err := database.InitDB(dbPath, opts); err != nil {
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

	if err := backuputils.InitConfig(defaultKmcPath, kmc.InitKmcCfg); err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}

	return nil
}

func checker() error {
	if inRange := valid.IsPortInRange(common.MinPort, common.MaxPort, port); !inRange {
		return fmt.Errorf("port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
	}
	if err := common.LimitChecker(getLimitParam(), maxConcurrency, maxIPConnLimit); err != nil {
		return err
	}
	return nil
}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(restful.NewRestfulService(true,
		&httpsmgr.HttpsServer{
			IP:          ip,
			Port:        port,
			SwitchLimit: true,
			ServerParam: getLimitParam(),
		})); err != nil {
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

func getLimitParam() httpsmgr.ServerParam {
	return httpsmgr.ServerParam{
		BurstIPReq:     maxBurstIP,
		Concurrency:    concurrency,
		BodySizeLimit:  dataLimit,
		LimitIPReq:     limitIPReq,
		LimitIPConn:    limitIPConn,
		LimitTotalConn: limitTotalConn,
		CacheSize:      cacheSize,
	}
}
