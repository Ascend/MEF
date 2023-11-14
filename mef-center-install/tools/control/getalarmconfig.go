// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF Center get alarm config from db
package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type getAlarmCfgController struct {
	operate              string
	installParam         *util.InstallParamJsonTemplate
	certCheckPeriod      int
	certOverdueThreshold int
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

	period, err := dbMgr.GetAlarmConfig(common.CertCheckPeriodDB)
	if err != nil {
		hwlog.RunLog.Errorf("get alarm config %s failed, error: %v", common.CertCheckPeriodDB, err)
		return err
	}
	gcc.certCheckPeriod = period

	threshold, err := dbMgr.GetAlarmConfig(common.CertOverdueThresholdDB)
	if err != nil {
		hwlog.RunLog.Errorf("get alarm config %s failed, error: %v", common.CertOverdueThresholdDB, err)
		return err
	}
	gcc.certOverdueThreshold = threshold

	var checkFuncs = []func() error{
		gcc.checkThreshold,
		gcc.checkPeriod,
	}
	for _, checkFunc := range checkFuncs {
		if err = checkFunc(); err != nil {
			return err
		}
	}
	gcc.printAlarmConfig()
	return nil
}

func (gcc *getAlarmCfgController) checkPeriod() error {
	periodChecker := checker.GetIntChecker("", util.MinCheckPeriod, int64(gcc.certOverdueThreshold-DiffTime), true)
	checkRes := periodChecker.Check(gcc.certCheckPeriod)
	if checkRes.Result {
		return nil
	}

	errInfo := fmt.Sprintf("%s error, should be from %d to the certificate alarm threshold minus %d",
		CertCheckPeriodCmd, util.MinCheckPeriod, DiffTime)
	hwlog.RunLog.Error(errInfo)
	fmt.Println(errInfo)
	return fmt.Errorf("%s is invalid", CertCheckPeriodCmd)
}

func (gcc *getAlarmCfgController) checkThreshold() error {
	thresholdChecker := checker.GetIntChecker("", util.MinOverdueThreshold, util.MaxOverdueThreshold, true)
	checkRes := thresholdChecker.Check(gcc.certOverdueThreshold)
	if checkRes.Result {
		return nil
	}
	errInfo := fmt.Sprintf("%s error, should be within [%d, %d]",
		CertOverdueThresholdCmd, util.MinOverdueThreshold, util.MaxOverdueThreshold)
	hwlog.RunLog.Error(errInfo)
	fmt.Println(errInfo)
	return fmt.Errorf("%s is invalid", CertOverdueThresholdCmd)
}

func (gcc *getAlarmCfgController) printAlarmConfig() {
	alarmCfgs := []struct {
		cfgInDb int
		cfgCmd  string
		unit    string
	}{
		{gcc.certCheckPeriod, CertCheckPeriodCmd, UnitDay},
		{gcc.certOverdueThreshold, CertOverdueThresholdCmd, UnitDay},
	}
	for _, alarmCfg := range alarmCfgs {
		fmt.Printf("%s is [%v], the unit is %s\n", alarmCfg.cfgCmd, alarmCfg.cfgInDb, alarmCfg.unit)
	}
	return
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
