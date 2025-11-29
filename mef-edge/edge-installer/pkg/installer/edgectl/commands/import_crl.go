// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package commands

import (
	"errors"
	"flag"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/importcrl"
)

type importCrlCmd struct {
	crlPath string
	peer    string
}

// ImportCrlCmd  edge control command import crl
func ImportCrlCmd() common.Command {
	return &importCrlCmd{}
}

func (cmd *importCrlCmd) Name() string {
	return common.ImportCrl
}

func (cmd *importCrlCmd) Description() string {
	return common.ImportCrlDesc
}

// BindFlag command flag binding
func (cmd *importCrlCmd) BindFlag() bool {
	flag.StringVar(&(cmd.crlPath), constants.CrlPathSubCmd, "", "the path of the importing crl")
	flag.StringVar(&(cmd.peer), constants.PeerSubCmd, "", "the peer of the crl, now only supports mef_center")
	utils.MarkFlagRequired(constants.CrlPathSubCmd)
	utils.MarkFlagRequired(constants.PeerSubCmd)
	return true
}

// LockFlag command lock flag
func (cmd *importCrlCmd) LockFlag() bool {
	return false
}

// Execute execute command
func (cmd *importCrlCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("context is nil")
	}

	if err := cmd.checkParam(); err != nil {
		return err
	}

	importFlow := importcrl.NewCrlImportFlow(ctx.ConfigPathMgr, cmd.crlPath, cmd.peer)
	return importFlow.RunFlow()
}

// PrintOpLogOk print operation success log
func (cmd *importCrlCmd) PrintOpLogOk(user, ip string) {
	hwlog.OpLog.Infof("[%s@%s] import crl success", user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *importCrlCmd) PrintOpLogFail(user, ip string) {
	hwlog.OpLog.Errorf("[%s@%s] import crl failed", user, ip)
}

func (cmd *importCrlCmd) checkParam() error {
	if _, err := fileutils.RealFileCheck(cmd.crlPath, true, false, constants.MaxCertSize); err != nil {
		hwlog.RunLog.Errorf("check crl path failed: %s", err.Error())
		return errors.New("check crl path failed")
	}

	return nil
}
