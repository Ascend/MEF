// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package innercommands this file for inner control command restart
package innercommands

import (
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

type exchangeCertsCmd struct {
	importPath string
	exportPath string
}

// ExchangeCertsCmd is the func to init an ExchangeCertsCmd struct
func ExchangeCertsCmd() common.Command {
	return &exchangeCertsCmd{}
}

// Name command name
func (cmd *exchangeCertsCmd) Name() string {
	return common.ExchangeCertsCmd
}

// Description command description
func (cmd *exchangeCertsCmd) Description() string {
	return common.ExchangeCertsDesc
}

// BindFlag command flag binding
func (cmd *exchangeCertsCmd) BindFlag() bool {
	flag.StringVar(&(cmd.importPath), common.ImportPathFlag, "", "path that saves ca cert to import")
	flag.StringVar(&(cmd.exportPath), common.ExportPathFlag, "", "path to export MEF ca cert")
	utils.MarkFlagRequired(common.ImportPathFlag)
	utils.MarkFlagRequired(common.ExportPathFlag)
	return true
}

// LockFlag command lock flag
func (cmd *exchangeCertsCmd) LockFlag() bool {
	return false
}

// Execute execute command
func (cmd *exchangeCertsCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed, error: %v", err)
		return errors.New("get config path manager failed")
	}

	exchangeFlow := NewExchangeCaFlow(cmd.importPath, cmd.exportPath, configPathMgr)
	if err = exchangeFlow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("execute exchange flow failed: %s", err.Error())
		return errors.New("execute exchange flow failed")
	}

	return nil
}

// PrintOpLogOk print operation success log
func (cmd *exchangeCertsCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *exchangeCertsCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

type recoveryCmd struct {
}

// RecoveryCmd is the func to init an recoveryCmd struct which is to recovery the environment
// when upgrading firmware failed
func RecoveryCmd() common.Command {
	return &recoveryCmd{}
}

// Name command name
func (cmd *recoveryCmd) Name() string {
	return common.RecoveryCmd
}

// Description command description
func (cmd *recoveryCmd) Description() string {
	return common.RecoveryDesc
}

// BindFlag command flag binding
func (cmd *recoveryCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *recoveryCmd) LockFlag() bool {
	return false
}

// Execute execute command
func (cmd *recoveryCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	workPathMgr, err := path.GetWorkPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get work path manager failed, error: %v", err)
		return errors.New("get work path manager failed")
	}

	hwlog.RunLog.Info("start to recovery environment")
	tempDir := workPathMgr.GetUpgradeTempDir()
	if !fileutils.IsExist(tempDir) {
		return nil
	}
	if err = util.UnSetImmutable(tempDir); err != nil {
		hwlog.RunLog.Warnf("unset target install path [%s] immutable find errors,maybe include link file", tempDir)
	}
	if err := fileutils.DeleteAllFileWithConfusion(tempDir); err != nil {
		hwlog.RunLog.Errorf("clean target install path failed, error: %v", err)
		return err
	}

	hwlog.RunLog.Info("recovery environment successful")
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *recoveryCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *recoveryCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

type prepareEdgecoreCmd struct {
}

// PrepareEdgecoreCmd is the func to write edgecore's password into pipe file
func PrepareEdgecoreCmd() common.Command {
	return &prepareEdgecoreCmd{}
}

// Name command name
func (cmd *prepareEdgecoreCmd) Name() string {
	return common.PrepareEdgecoreCmd
}

// Description command description
func (cmd *prepareEdgecoreCmd) Description() string {
	return common.PrepareEdgecoreDesc
}

// BindFlag command flag binding
func (cmd *prepareEdgecoreCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *prepareEdgecoreCmd) LockFlag() bool {
	return false
}

// Execute execute command
func (cmd *prepareEdgecoreCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	flow := NewPrepareEdgecore()
	if err := flow.Run(); err != nil {
		hwlog.RunLog.Errorf("prepare edgecore failed: %s", err.Error())
		return errors.New("prepare edgecore pipe file failed")
	}

	return nil
}

// PrintOpLogOk print operation success log
func (cmd *prepareEdgecoreCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *prepareEdgecoreCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

type recoverLogCmd struct {
}

// NewRecoverLogCmd is the func to sync logs between disk and memory
func NewRecoverLogCmd() common.Command {
	return &recoverLogCmd{}
}

// Name command name
func (cmd *recoverLogCmd) Name() string {
	return common.RecoverLogCmd
}

// Description command description
func (cmd *recoverLogCmd) Description() string {
	return common.RecoverLogDesc
}

// BindFlag command flag binding
func (cmd *recoverLogCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *recoverLogCmd) LockFlag() bool {
	return false
}

// Execute execute command
func (cmd *recoverLogCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	logSyncMgr := util.NewLogSyncMgr()
	if err := logSyncMgr.RecoverLogs(); err != nil {
		fmt.Printf("recover log failed: %v\n", err)
		hwlog.RunLog.Errorf("recover log failed: %v", err)
		return errors.New("recover log failed")
	}

	fmt.Println("recover log success")
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *recoverLogCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *recoverLogCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}
