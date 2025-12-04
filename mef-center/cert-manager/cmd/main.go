// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
	"huawei.com/mindx/common/httpsmgr"
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
	maxIPConnLimit        = 100
	maxConcurrency        = 100
	defaultConnection     = 100
	defaultConcurrency    = 100
	defaultConnPerIP      = 100
	maxBurstIP            = 100
	defaultDataLimit      = 1024 * 1024
	defaultCachSize       = 1024 * 1024 * 10
)

var (
	serverRunConf = &hwlog.LogConfig{OnlyToFile: true, LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf  = &hwlog.LogConfig{OnlyToFile: true, LogFileName: operateLogFile, BackupDirName: backupDirName}
	// BuildName cert-manager's build name
	BuildName string
	// BuildVersion cert-manager's build version
	BuildVersion   string
	port           int
	ip             string
	version        bool
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
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		return
	}
	if err := checkParam(); err != nil {
		hwlog.RunLog.Errorf("parameter validation error: %v", err)
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

func checkParam() error {
	if res := checker.GetIntChecker("", common.MinPort, common.MaxPort, true).Check(port); !res.Result {
		return fmt.Errorf("port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
	}
	if err := common.LimitChecker(getLimitParam(), maxConcurrency, maxIPConnLimit); err != nil {
		return err
	}
	return nil
}

func init() {
	flag.IntVar(&port, "port", portConst,
		"The server port of the http service,range[1025-40000]")
	flag.BoolVar(&version, "version", false, "Output the program version")
	flag.IntVar(&cacheSize, "cacheSize", defaultCachSize, "the cacheSize for ip limit,"+
		"keep default normally")
	flag.IntVar(&limitIPConn, "limitIPConn", defaultConnPerIP, "the tcp connection limit for each Ip")
	flag.IntVar(&limitTotalConn, "limitTotalConn", defaultConnection, "the tcp connection limit for all request")
	flag.StringVar(&limitIPReq, "limitIPReq", "10/1",
		"the http request limit counts for each Ip, 10/1 means allow 10 request in 1 seconds")
	flag.IntVar(&concurrency, "concurrency", defaultConcurrency,
		"The max concurrency of the http server, range is [1-512]")
	flag.Int64Var(&dataLimit, "dataLimit", defaultDataLimit,
		"bytes, limit the data size of request's body, the default value is 1MB")

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

	if err := modulemgr.Registry(restful.NewRestfulService(true,
		&httpsmgr.HttpsServer{
			IP:          ip,
			Port:        port,
			SwitchLimit: true,
			ServerParam: getLimitParam(),
		})); err != nil {
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
