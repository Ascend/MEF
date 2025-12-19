// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

type deleteCertInfoCmd struct {
	certName string
}

// user operation choices
const (
	Accept = "yes"
	Deny   = "no"
)

// NewDeleteCertInfoCmd delete unused backup cert
func NewDeleteCertInfoCmd() common.Command {
	return &deleteCertInfoCmd{}
}

// Name command name
func (cmd *deleteCertInfoCmd) Name() string {
	return common.DeleteCert
}

// Description command description
func (cmd *deleteCertInfoCmd) Description() string {
	return common.DeleteUnusedCertDesc
}

// BindFlag command flag binding
func (cmd *deleteCertInfoCmd) BindFlag() bool {
	flag.StringVar(&cmd.certName, "name", "",
		"the name of unused certificate. Currently, only [cloud_root] is supported.")
	utils.MarkFlagRequired("name")
	return true
}

// LockFlag command lock flag
func (cmd *deleteCertInfoCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *deleteCertInfoCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		return errors.New("parameter ctx is invalid")
	}
	unusedCertPathMap := map[string]string{
		"cloud_root": ctx.ConfigPathMgr.GetHubSvrRootCertPrevBackupPath(),
	}
	certPath, found := unusedCertPathMap[cmd.certName]
	if !found {
		fmt.Println("invalid certificate name, please check")
		return errors.New("invalid certificate name, please check")
	}
	if !fileutils.IsExist(certPath) {
		fmt.Println("backup certificate not found")
		return errors.New("backup certificate not found")
	}
	promptMsg := `the following cert will be deleted, are you sure ? [yes | no]
` + certPath

	if err := confirmDelete(promptMsg); err != nil {
		fmt.Printf("delete operation is aborted: %v\n", err)
		return nil
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
	if err := fileutils.DeleteFile(certPath); err != nil {
		fmt.Printf("delete certificate [%v] failed\n", cmd.certName)
		return fmt.Errorf("delete certificate [%v] failed: %v", cmd.certName, err)
	}
	fmt.Printf("delete certificate [%v] successfully\n", cmd.certName)
	return nil
}

func (cmd *deleteCertInfoCmd) PrintOpLogOk(user string, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

func (cmd *deleteCertInfoCmd) PrintOpLogFail(user string, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

func confirmDelete(prompt string) error {
	fmt.Println(prompt)
	const inputBuffLimit = 8
	buf := make([]byte, inputBuffLimit)
	inputLen, err := os.Stdin.Read(buf)
	if err != nil {
		hwlog.RunLog.Errorf("read user input failed: %v", err)
		return errors.New("read user input failed")
	}
	choice := bytes.TrimSpace(buf[:inputLen])
	switch string(choice) {
	case Accept:
		return nil
	case Deny:
		hwlog.RunLog.Infof("user cancel the delete operation")
		return errors.New("user cancel the delete operation")
	default:
		hwlog.RunLog.Errorf("invalid operation: %v", string(choice))
		return fmt.Errorf("invalid operation: %v", string(choice))
	}
}
