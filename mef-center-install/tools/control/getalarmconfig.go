// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF Center get alarm config from db
package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type getAlarmCfgController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
}

// UnitDay unit day
const UnitDay = "day"

func (gcc *getAlarmCfgController) bindFlag() bool {
	return false
}

func (gcc *getAlarmCfgController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	gcc.installParam = installParam
}

func (gcc *getAlarmCfgController) doControl() (err error) {
	pathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init path mgr failed: %v", err)
		return errors.New("init path mgr failed")
	}
	defer func() {
		if resetErr := util.ResetPriv(); resetErr != nil {
			err = resetErr
			hwlog.RunLog.Errorf("reset euid/gid back to root failed: %v", err)
		}
	}()
	if err = util.ReducePriv(); err != nil {
		return err
	}

	configDir := pathMgr.GetConfigPath()
	alarmDbDir := filepath.Join(configDir, util.AlarmManagerName)
	dbMgr := common.NewDbMgr(alarmDbDir, common.AlarmConfigDBName)
	alarmCfgs := []struct {
		cfgInDb string
		cfgCmd  string
		unit    string
	}{
		{common.CertCheckPeriodDB, CertCheckPeriodCmd, UnitDay},
		{common.CertOverdueThresholdDB, CertOverdueThresholdCmd, UnitDay},
	}
	for _, alarmCfg := range alarmCfgs {
		cfg, err := dbMgr.GetAlarmConfig(alarmCfg.cfgInDb)
		if err != nil {
			return err
		}
		fmt.Printf("%s is [%v], the unit is %s\n", alarmCfg.cfgCmd, cfg, alarmCfg.unit)
	}

	return nil
}

func (gcc *getAlarmCfgController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to get alarm config-------------------")
	hwlog.OpLog.Infof("[%s@%s] start to get alarm config", user, ip)
	fmt.Println("start to get alarm config")
}

func (gcc *getAlarmCfgController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------get alarm config successful-------------------")
	hwlog.OpLog.Infof("[%s@%s] get alarm config successful", user, ip)
	fmt.Println("get alarm config successful")
}

func (gcc *getAlarmCfgController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------get alarm config failed-------------------")
	hwlog.OpLog.Errorf("[%s@%s] get alarm config failed", user, ip)
	fmt.Println("get alarm config failed")
}

func (gcc *getAlarmCfgController) getName() string {
	return gcc.operate
}
