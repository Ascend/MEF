// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF Center update alarm config to db
package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type alarmCfgController struct {
	operate              string
	certCheckPeriod      int
	certOverdueThreshold int
	installParam         *util.InstallParamJsonTemplate
	dbMgr                *common.DbMgr
}

// update alarm config commands config
const (
	CertCheckPeriodCmd      = "cert_period"
	CertOverdueThresholdCmd = "cert_threshold"
	DiffTime                = 3
)

func (acc *alarmCfgController) bindFlag() bool {
	flag.IntVar(&(acc.certOverdueThreshold), CertOverdueThresholdCmd, util.DefaultOverdueThreshold,
		"The alarm threshold of the number of days for the certificate to expire, the range is from 7 to 180")
	flag.IntVar(&(acc.certCheckPeriod), CertCheckPeriodCmd, util.DefaultCheckPeriod,
		"The number of days at which the certificate is checked, "+
			"and the range is from 1 to the certificate alarm threshold minus 3")
	return true
}

func (acc *alarmCfgController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	acc.installParam = installParam
}

func (acc *alarmCfgController) doControl() error {
	if !utils.IsFlagSet(CertOverdueThresholdCmd) && !utils.IsFlagSet(CertCheckPeriodCmd) {
		hwlog.RunLog.Info("does not modify any configuration")
		fmt.Println("does not modify any configuration.")
		return nil
	}

	pathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init path mgr failed: %v", err)
		return errors.New("init path mgr failed")
	}
	defer func() {
		if err = util.ResetCfgPathPermAfterReducePriv(pathMgr); err != nil {
			hwlog.RunLog.Errorf("reset config path permission after reducing privilege failed, error: %v", err)
		}
	}()
	if err = util.SetCfgPathPermAndReducePriv(pathMgr); err != nil {
		return err
	}

	configDir := pathMgr.GetConfigPath()
	alarmDbDir := filepath.Join(configDir, util.AlarmManagerName)
	acc.dbMgr = common.NewDbMgr(alarmDbDir, common.AlarmConfigDBName)

	if err = acc.checkParam(); err != nil {
		return fmt.Errorf("check param failed, error: %v", err)
	}
	if err = acc.updateConfig(); err != nil {
		return fmt.Errorf("update alarm config failed, error: %v", err)
	}

	return nil
}

func (acc *alarmCfgController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to update alarm config-------------------")
	hwlog.OpLog.Infof("[%s@%s] start to update alarm config", user, ip)
	fmt.Println("start to update alarm config")
}

func (acc *alarmCfgController) printSuccessLog(user, ip string) {
	hwlog.RunLog.Info("-------------------update alarm config successful-------------------")
	hwlog.OpLog.Infof("[%s@%s] update alarm config successful", user, ip)
	fmt.Println("update alarm config successful")
}

func (acc *alarmCfgController) printFailedLog(user, ip string) {
	hwlog.RunLog.Error("-------------------update alarm config failed-------------------")
	hwlog.OpLog.Errorf("[%s@%s] update alarm config failed", user, ip)
	fmt.Println("update alarm config failed")
}

func (acc *alarmCfgController) getName() string {
	return acc.operate
}

func (acc *alarmCfgController) checkParam() error {
	if !utils.IsFlagSet(CertCheckPeriodCmd) {
		period, err := acc.dbMgr.GetAlarmConfig(common.CertCheckPeriodDB)
		if err != nil {
			hwlog.RunLog.Errorf("get alarm config %s failed, error: %v", common.CertCheckPeriodDB, err)
			return err
		}
		acc.certCheckPeriod = period
	}
	if !utils.IsFlagSet(CertOverdueThresholdCmd) {
		threshold, err := acc.dbMgr.GetAlarmConfig(common.CertOverdueThresholdDB)
		if err != nil {
			hwlog.RunLog.Errorf("get alarm config %s failed, error: %v", common.CertOverdueThresholdDB, err)
			return err
		}
		acc.certOverdueThreshold = threshold
	}

	var checkFuncs = []func() error{
		acc.checkThreshold,
		acc.checkPeriod,
	}
	for _, checkFunc := range checkFuncs {
		if err = checkFunc(); err != nil {
			return err
		}
	}

	return nil
}

func (acc *alarmCfgController) checkThreshold() error {
	minThreshold := utils.MaxInt(util.MinOverdueThreshold, acc.certCheckPeriod+DiffTime)
	thresholdChecker := checker.GetIntChecker("", int64(minThreshold), util.MaxOverdueThreshold, true)
	checkRes := thresholdChecker.Check(acc.certOverdueThreshold)
	if checkRes.Result {
		return nil
	}
	errInfo := fmt.Sprintf("param %s error, should be within [%d, %d], and "+
		"at least the certificate check period plus %d", CertOverdueThresholdCmd,
		util.MinOverdueThreshold, util.MaxOverdueThreshold, DiffTime)
	hwlog.RunLog.Error(errInfo)
	fmt.Println(errInfo)
	return fmt.Errorf("param %s is invalid", CertOverdueThresholdCmd)
}

func (acc *alarmCfgController) checkPeriod() error {
	periodChecker := checker.GetIntChecker("", util.MinCheckPeriod, int64(acc.certOverdueThreshold-DiffTime), true)
	checkRes := periodChecker.Check(acc.certCheckPeriod)
	if checkRes.Result {
		return nil
	}
	errInfo := fmt.Sprintf("param %s error, should be from %d to the certificate alarm threshold minus %d",
		CertCheckPeriodCmd, util.MinCheckPeriod, DiffTime)
	hwlog.RunLog.Error(errInfo)
	fmt.Println(errInfo)
	return fmt.Errorf("param %s is invalid", CertCheckPeriodCmd)
}

func (acc *alarmCfgController) updateConfig() error {
	var alarmCfgMap = make(map[string]int)
	if utils.IsFlagSet(CertOverdueThresholdCmd) {
		alarmCfgMap[common.CertOverdueThresholdDB] = acc.certOverdueThreshold
	}
	if utils.IsFlagSet(CertCheckPeriodCmd) {
		alarmCfgMap[common.CertCheckPeriodDB] = acc.certCheckPeriod
	}

	for name, value := range alarmCfgMap {
		cfg := &common.AlarmConfig{
			ConfigName:  name,
			ConfigValue: value,
			HasModified: util.GetBoolPointer(true),
		}
		if err = acc.dbMgr.SetAlarmConfig(cfg); err != nil {
			hwlog.RunLog.Errorf("set alarm config %s failed, error: %v", cfg.ConfigName, err)
			return fmt.Errorf("set alarm config %s failed, error: %v", cfg.ConfigName, err)
		}
	}

	return nil
}
