// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

package commands

import (
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

type certBackupPair struct {
	CertPath   string
	BackupPath string
}

type restoreCertInfoCmd struct {
	certName string
}

// NewRestoreCertInfoCmd restore backup certificate
func NewRestoreCertInfoCmd() common.Command {
	return &restoreCertInfoCmd{}
}

// Name command name
func (cmd *restoreCertInfoCmd) Name() string {
	return common.RestoreCert
}

// Description command description
func (cmd *restoreCertInfoCmd) Description() string {
	return common.RestoreCertDesc
}

// BindFlag command flag binding
func (cmd *restoreCertInfoCmd) BindFlag() bool {
	flag.StringVar(&cmd.certName, "name", "",
		"the name of backup certificate to restore. Currently, only [cloud_root] is supported.")
	utils.MarkFlagRequired("name")
	return true
}

// LockFlag command lock flag
func (cmd *restoreCertInfoCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *restoreCertInfoCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		return errors.New("parameter ctx is invalid")
	}
	certBackupPathMap := map[string]certBackupPair{
		"cloud_root": {
			CertPath:   ctx.ConfigPathMgr.GetHubSvrRootCertPath(),
			BackupPath: ctx.ConfigPathMgr.GetHubSvrRootCertPrevBackupPath(),
		},
	}
	pair, found := certBackupPathMap[cmd.certName]
	if !found {
		fmt.Println("invalid certificate name, please check")
		return errors.New("invalid certificate name, please check")
	}
	operatorIdMgr := util.NewEdgeUGidMgr()
	if err := operatorIdMgr.SetEUGidToEdge(); err != nil {
		hwlog.RunLog.Errorf("set euid/egid to mef-edge failed: %v", err)
		return errors.New("set euid/egid to mef-edge failed")
	}
	defer func() {
		if err := operatorIdMgr.ResetEUGid(); err != nil {
			hwlog.RunLog.Errorf("reset euid/egid failed: %v", err)
		}
	}()
	if !fileutils.IsExist(pair.BackupPath) {
		fmt.Println("previous backup certificate not found")
		return errors.New("previous backup certificate not found")
	}
	if _, err := x509.CheckCertsChainReturnContent(pair.BackupPath); err != nil {
		fmt.Println("check previous backup certificate failed, restore operation aborted")
		return fmt.Errorf("check previous backup certificate failed, error: %v", err)
	}
	if err := restoreCert(pair); err != nil {
		fmt.Println("restore previous backup certificate failed")
		return fmt.Errorf("restore previous backup certificate failed: %v", err)
	}
	fmt.Printf("restore certificate [%v] successfully\n", cmd.certName)
	return nil
}

func (cmd *restoreCertInfoCmd) PrintOpLogOk(user string, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

func (cmd *restoreCertInfoCmd) PrintOpLogFail(user string, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

func restoreCert(cbPair certBackupPair) error {
	if err := fileutils.DeleteFile(cbPair.CertPath); err != nil {
		return fmt.Errorf("delete old normal cert failed: %v", err)
	}
	if err := fileutils.CopyFile(cbPair.BackupPath, cbPair.CertPath); err != nil {
		return fmt.Errorf("copy backup cert to normal cert failed: %v", err)
	}
	if err := fileutils.DeleteFile(cbPair.BackupPath); err != nil {
		return fmt.Errorf("delete backup cert failed: %v", err)
	}
	if err := fileutils.SetPathPermission(cbPair.CertPath, constants.CertFileMode, false, false); err != nil {
		return fmt.Errorf("set root ca permission failed: %v", err)
	}
	return nil
}
