// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package main to start nginx-manager
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/logmgmt/logrotate"
	"huawei.com/mindx/common/modulemgr"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"

	"nginx-manager/pkg/certupdater"
	"nginx-manager/pkg/nginxcom"
	"nginx-manager/pkg/nginxlogrotate"
	"nginx-manager/pkg/nginxmgr"
	"nginx-manager/pkg/nginxmonitor"
	"nginx-manager/pkg/restfulservice"
)

const (
	runLogFile     = "/home/MEFCenter/logs/nginx-manager-run.log"
	operateLogFile = "/home/MEFCenter/logs/nginx-manager-operate.log"
	backupDirName  = "/home/MEFCenter/logs_backup"
	defaultKmcPath = "/home/data/public-config/kmc-config.json"
)

var (
	ip            string
	serverRunConf = &hwlog.LogConfig{OnlyToFile: true, LogFileName: runLogFile, BackupDirName: backupDirName}
	serverOpConf  = &hwlog.LogConfig{OnlyToFile: true, LogFileName: operateLogFile, BackupDirName: backupDirName}
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
	if err := initResource(); err != nil {
		hwlog.RunLog.Errorf("initialize resource failed, %s", err.Error())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := register(ctx); err != nil {
		hwlog.RunLog.Errorf("register module failed, %s", err.Error())
		return
	}
	common.GracefulShutDown(cancel)
}

func init() {
	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func initResource() error {
	var err error
	if err = nginxcom.GetEnvManager().Load(); err != nil {
		return err
	}

	if err = backuputils.InitConfig(defaultKmcPath, kmc.InitKmcCfg); err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, use default kmc config", err)
	}
	ip, err = common.GetPodIP()
	if err != nil {
		return fmt.Errorf("get nginx manager pod ip failed: %v", err)
	}
	return nil
}

func register(ctx context.Context) error {
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(restfulservice.NewNgxMgrServer(true, ip, common.NginxMgrPort)); err != nil {
		return err
	}
	if err := modulemgr.Registry(certupdater.NewSouthCertUpdater(true, ctx)); err != nil {
		return err
	}
	if err := modulemgr.Registry(nginxmgr.NewNginxManager(true, ctx)); err != nil {
		return err
	}
	if err := modulemgr.Registry(nginxmonitor.NewNginxMonitor(true, ctx)); err != nil {
		return err
	}
	if err := modulemgr.Registry(logrotate.Module("", ctx, nginxlogrotate.Setup)); err != nil {
		return err
	}

	modulemgr.Start()
	return nil
}
