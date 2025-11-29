// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package commands this file for edge control command update alarm configuration
package commands

import (
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/checker"
	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

type alarmConfigCmd struct {
	certCheckPeriod      int
	certOverdueThreshold int
	dbMgr                *config.DbMgr
}

// AlarmConfigCmd edge control command update alarm configuration
func AlarmConfigCmd() common.Command {
	return &alarmConfigCmd{}
}

// Name command name
func (cmd *alarmConfigCmd) Name() string {
	return common.AlarmConfig
}

// Description command description
func (cmd *alarmConfigCmd) Description() string {
	return common.AlarmConfigDesc
}

// BindFlag command flag binding
func (cmd *alarmConfigCmd) BindFlag() bool {
	flag.IntVar(&(cmd.certOverdueThreshold), common.CertOverdueThresholdCmd, constants.DefaultOverdueThreshold,
		"The alarm threshold of the number of days for the certificate to expire, the range is from 7 to 180")
	flag.IntVar(&(cmd.certCheckPeriod), common.CertCheckPeriodCmd, constants.DefaultCheckPeriod,
		"The number of days at which the certificate is checked, "+
			"and the range is from 1 to the certificate alarm threshold minus 3")
	return true
}

// LockFlag command lock flag
func (cmd *alarmConfigCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *alarmConfigCmd) Execute(ctx *common.Context) error {
	hwlog.RunLog.Info("start to update alarm config")
	fmt.Println("start to update alarm config...")

	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	if !util.IsFlagSet(common.CertOverdueThresholdCmd) && !util.IsFlagSet(common.CertCheckPeriodCmd) {
		hwlog.RunLog.Info("does not modify any configuration")
		fmt.Println("does not modify any configuration.")
		return nil
	}

	dbMgr, err := config.GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		return err
	}
	cmd.dbMgr = dbMgr

	if err = cmd.checkParam(); err != nil {
		return fmt.Errorf("check param failed, error: %v", err)
	}
	if err = cmd.updateConfig(); err != nil {
		return fmt.Errorf("update alarm config failed, error: %v", err)
	}

	hwlog.RunLog.Info("update alarm config success")
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *alarmConfigCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *alarmConfigCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

func (cmd *alarmConfigCmd) checkParam() error {
	if !util.IsFlagSet(common.CertCheckPeriodCmd) {
		period, err := cmd.dbMgr.GetAlarmConfig(constants.CertCheckPeriodDB)
		if err != nil {
			return err
		}
		cmd.certCheckPeriod = period
	}
	if !util.IsFlagSet(common.CertOverdueThresholdCmd) {
		threshold, err := cmd.dbMgr.GetAlarmConfig(constants.CertOverdueThresholdDB)
		if err != nil {
			return err
		}
		cmd.certOverdueThreshold = threshold
	}

	var checkFuncs = []func() error{
		cmd.checkThreshold,
		cmd.checkPeriod,
	}
	for _, checkFunc := range checkFuncs {
		if err := checkFunc(); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *alarmConfigCmd) checkThreshold() error {
	minThreshold := utils.MaxInt(constants.MinOverdueThreshold, cmd.certCheckPeriod+common.DiffTime)
	if !checker.IntChecker(cmd.certOverdueThreshold, minThreshold, constants.MaxOverdueThreshold) {
		errInfo := fmt.Sprintf("param %s error, should be within [%d, %d], and "+
			"at least the certificate check period plus %d", common.CertOverdueThresholdCmd,
			constants.MinOverdueThreshold, constants.MaxOverdueThreshold, common.DiffTime)
		hwlog.RunLog.Error(errInfo)
		fmt.Println(errInfo)
		return fmt.Errorf("param %s is invalid", common.CertOverdueThresholdCmd)
	}
	return nil
}

func (cmd *alarmConfigCmd) checkPeriod() error {
	if !checker.IntChecker(cmd.certCheckPeriod, constants.MinCheckPeriod, cmd.certOverdueThreshold-common.DiffTime) {
		errInfo := fmt.Sprintf("param %s error, should be from %d to the certificate alarm threshold minus %d",
			common.CertCheckPeriodCmd, constants.MinCheckPeriod, common.DiffTime)
		hwlog.RunLog.Error(errInfo)
		fmt.Println(errInfo)
		return fmt.Errorf("param %s is invalid", common.CertCheckPeriodCmd)
	}
	return nil
}

func (cmd *alarmConfigCmd) updateConfig() error {
	var alarmCfgMap = make(map[string]int)
	if util.IsFlagSet(common.CertOverdueThresholdCmd) {
		alarmCfgMap[constants.CertOverdueThresholdDB] = cmd.certOverdueThreshold
	}
	if util.IsFlagSet(common.CertCheckPeriodCmd) {
		alarmCfgMap[constants.CertCheckPeriodDB] = cmd.certCheckPeriod
	}

	for name, value := range alarmCfgMap {
		cfg := &config.AlarmConfig{
			ConfigName:  name,
			ConfigValue: value,
			HasModified: util.GetBoolPointer(true),
		}
		if err := cmd.dbMgr.SetAlarmConfig(cfg); err != nil {
			hwlog.RunLog.Errorf("set alarm config %s failed, error: %v", cfg.ConfigName, err)
			return fmt.Errorf("set alarm config %s failed, error: %v", cfg.ConfigName, err)
		}
	}

	return nil
}
