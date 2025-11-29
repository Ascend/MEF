// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package commands this file for
package commands

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/preupgrade/flows"
)

type upgradeCmd struct {
	tarPath     string
	cmsPath     string
	crlPath     string
	delayEffect bool
}

// UpgradeCmd edge control command start
func UpgradeCmd() common.Command {
	return &upgradeCmd{}
}

// Name command name
func (cmd *upgradeCmd) Name() string {
	return common.Upgrade
}

// Description command description
func (cmd *upgradeCmd) Description() string {
	return common.UpgradeDesc
}

// LockFlag command lock flag
func (cmd *upgradeCmd) LockFlag() bool {
	return true
}

// BindFlag command flag binding
func (cmd *upgradeCmd) BindFlag() bool {
	flag.StringVar(&(cmd.tarPath), constants.TarFlag, "", "path of the software upgrade tar.gz file")
	flag.StringVar(&(cmd.cmsPath), constants.CmsFlag, "", "path of the software upgrade tar.gz.cms file")
	flag.StringVar(&(cmd.crlPath), constants.CrlFlag, "", "path of the software upgrade tar.gz.crl file")
	flag.BoolVar(&(cmd.delayEffect), "delay", false,
		"to determine if the software should be effected after upgrade")
	utils.MarkFlagRequired(constants.TarFlag)
	utils.MarkFlagRequired(constants.CmsFlag)
	utils.MarkFlagRequired(constants.CrlFlag)
	return true
}

// PrintOpLogOk print operation success log
func (cmd *upgradeCmd) PrintOpLogOk(user, ip string) {
	hwlog.OpLog.Infof("[%s@%s] command: %s, result: success", user, ip, cmd.Name())
}

// PrintOpLogFail print operation fail log
func (cmd *upgradeCmd) PrintOpLogFail(user, ip string) {
	hwlog.OpLog.Errorf("[%s@%s] command: %s, result: failed", user, ip, cmd.Name())
}

func (cmd *upgradeCmd) checkSinglePath(path string) (string, error) {
	validPath, err := fileutils.RealFileCheck(path, true, false, constants.InstallerTarGzSizeMaxInMB)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(validPath, constants.UnpackPath) {
		return "", errors.New("the path cannot be in the decompression path")
	}

	return validPath, nil
}

// Execute execute command
func (cmd *upgradeCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}
	if err := cmd.checkParam(); err != nil {
		fmt.Printf("check upgrade param failed: %s\n", err.Error())
		return fmt.Errorf("check param failed, %v", err)
	}

	param := flows.OfflineUpgradeInstallerParam{TarPath: cmd.tarPath, CmsPath: cmd.cmsPath,
		CrlPath: cmd.crlPath, EdgeDir: ctx.WorkPathMgr.GetMefEdgeDir(), DelayEffect: cmd.delayEffect}
	flow := flows.OfflineUpgradeInstaller(param)
	if err := flow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("offline upgrade edge-installer failed, error: %v", err)
		return err
	}
	hwlog.RunLog.Info("offline upgrade edge-installer success")
	return nil
}

func (cmd *upgradeCmd) checkParam() error {
	if cmd.tarPath == "" || cmd.cmsPath == "" || cmd.crlPath == "" {
		return errors.New("tar or cms or crl file not input")
	}

	validPath, err := cmd.checkSinglePath(cmd.tarPath)
	if err != nil {
		hwlog.RunLog.Errorf("check tar path failed: %s", err.Error())
		return fmt.Errorf("check tar path failed: %s", err.Error())
	}
	cmd.tarPath = validPath

	validPath, err = cmd.checkSinglePath(cmd.cmsPath)
	if err != nil {
		hwlog.RunLog.Errorf("check cms path failed: %s", err.Error())
		return fmt.Errorf("check cms path failed: %s", err.Error())
	}
	cmd.cmsPath = validPath

	validPath, err = cmd.checkSinglePath(cmd.crlPath)
	if err != nil {
		hwlog.RunLog.Errorf("check crl path failed: %s", err.Error())
		return fmt.Errorf("check crl path failed: %s", err.Error())
	}
	cmd.crlPath = validPath
	return nil
}
