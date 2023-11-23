// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start edge-manager server
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr"

	"edge-manager/pkg/alarmmanager"
	"edge-manager/pkg/appmanager"
	"edge-manager/pkg/certupdater"
	"edge-manager/pkg/cloudhub"
	"edge-manager/pkg/config"
	"edge-manager/pkg/configmanager"
	"edge-manager/pkg/edgemsgmanager"
	"edge-manager/pkg/innerserver"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/logmanager"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/restfulservice"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/common/taskschedule"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
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
	logMaxLineLength      = 1024
	maxIPConnLimit        = 128
	maxConcurrency        = 512
	defaultConnection     = 3
	defaultConcurrency    = 3
	defaultDataLimit      = 1024 * 1024
	defaultCachSize       = 1024 * 1024 * 10
)

var (
	serverRunConf = &hwlog.LogConfig{OnlyToFile: true, LogFileName: defaultRunLogFile, FileMaxSize: defaultRunLogMaxSize,
		BackupDirName: defaultBackupDirName, MaxLineLength: logMaxLineLength}
	serverOpConf = &hwlog.LogConfig{OnlyToFile: true, LogFileName: defaultOperateLogFile, FileMaxSize: defaultOpLogMaxSize,
		BackupDirName: defaultBackupDirName}
	port           int
	wsPort         int
	authPort       int
	ip             string
	version        bool
	dbPath         string
	maxClientNum   int
	limitIPReq     string
	concurrency    int
	cacheSize      int
	limitIPConn    int
	limitTotalConn int
	dataLimit      int64
)

func main() {
	if len(os.Args) < util.NoArgCount {
		fmt.Println("the required parameter is missing")
		os.Exit(util.ErrorExitCode)
	}

	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", config.BuildName, config.BuildVersion)
		return
	}
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}
	if err := validateFlags(); err != nil {
		hwlog.RunLog.Errorf("argument validation error: %s", err.Error())
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
	flag.IntVar(&cacheSize, "cacheSize", defaultCachSize, "the cacheSize for ip limit,"+
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

func validateFlags() error {
	if res := checker.GetIntChecker("", common.MinPort, common.MaxPort, true).Check(port); !res.Result {
		return fmt.Errorf("port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
	}
	if res := checker.GetIntChecker("", common.MinPort, common.MaxPort, true).Check(wsPort); !res.Result {
		return fmt.Errorf("wsPort %d is not in [%d, %d]", wsPort, common.MinPort, common.MaxPort)
	}
	if res := checker.GetIntChecker("", common.MinPort, common.MaxPort, true).Check(authPort); !res.Result {
		return fmt.Errorf("authPort %d is not in [%d, %d]", authPort, common.MinPort, common.MaxPort)
	}
	if err := common.LimitChecker(getLimitParam(), maxConcurrency, maxIPConnLimit); err != nil {
		return err
	}
	if authPort == wsPort {
		return fmt.Errorf("authPort can not equals to wsPort")
	}
	if _, err := fileutils.CheckOriginPath(dbPath); err != nil {
		return err
	}
	return nil
}

func initPodConfig(configPath string) error {
	date, err := fileutils.LoadFile(configPath)
	if err != nil {
		return fmt.Errorf("load pod config file failed, %s", err.Error())
	}
	podConfig := config.PodConfig
	if err = json.Unmarshal(date, &podConfig); err != nil {
		return errors.New("unmarshal pod config failed")
	}
	config.PodConfig.HostPath = config.CheckAndModifyHostPath(podConfig.HostPath)
	config.PodConfig.MaxPodNumberPerNode = config.CheckAndModifyMaxLimitNumber(podConfig.MaxPodNumberPerNode)
	config.PodConfig.MaxDsNumberPerNodeGroup = config.CheckAndModifyMaxLimitNumber(podConfig.MaxDsNumberPerNodeGroup)
	return nil
}

func initAuthConfig(configPath string) error {
	date, err := fileutils.LoadFile(configPath)
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

func initScheduler() error {
	const (
		maxHistoryMasterTasks = 2000
		maxActiveTasks        = 200
		allowedMaxTasksInDb   = 300000
	)
	db, err := gorm.Open(sqlite.Open(":memory:?cache=shared"))
	if err != nil {
		return err
	}
	rawDb, err := db.DB()
	if err != nil {
		return err
	}
	rawDb.SetMaxOpenConns(1)
	return taskschedule.InitDefaultScheduler(context.Background(), db, taskschedule.SchedulerSpec{
		MaxHistoryMasterTasks: maxHistoryMasterTasks,
		MaxActiveTasks:        maxActiveTasks,
		AllowedMaxTasksInDb:   allowedMaxTasksInDb,
	})
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
		return err
	}
	if _, err := kubeclient.NewClientK8s(); err != nil {
		hwlog.RunLog.Errorf("init k8s failed: %v", err)
		return err
	}
	if err := backuputils.InitConfig(defaultPodConfigPath, initPodConfig); err != nil {
		hwlog.RunLog.Errorf("init pod config failed: %v", err)
		return err
	}
	if err := backuputils.InitConfig(defaultAuthConfigPath, initAuthConfig); err != nil {
		hwlog.RunLog.Errorf("init auth config error %v", err)
		return err
	}
	if err := backuputils.InitConfig(defaultKmcPath, kmc.InitKmcCfg); err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}
	if err := initScheduler(); err != nil {
		hwlog.RunLog.Errorf("init scheduler failed, %v", err)
		return err
	}
	return nil

}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(certupdater.NewEdgeCertUpdater(true)); err != nil {
		return err
	}
	if err := modulemgr.Registry(restfulservice.NewRestfulService(true,
		&httpsmgr.HttpsServer{
			IP:          ip,
			Port:        port,
			SwitchLimit: true,
			ServerParam: getLimitParam(),
		})); err != nil {
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
	if err := modulemgr.Registry(logmanager.NewLogManager(ctx, true)); err != nil {
		return err
	}
	if err := modulemgr.Registry(alarmmanager.NewAlarmManager(true)); err != nil {
		return err
	}
	if err := modulemgr.Registry(innerserver.NewInnerServer(true, common.EdgeManagerInnerWsPort)); err != nil {
		return err
	}
	modulemgr.Start()
	return nil
}

func getLimitParam() httpsmgr.ServerParam {
	return httpsmgr.ServerParam{
		Concurrency:    concurrency,
		BodySizeLimit:  dataLimit,
		LimitIPReq:     limitIPReq,
		LimitIPConn:    limitIPConn,
		LimitTotalConn: limitTotalConn,
		CacheSize:      cacheSize,
	}
}
