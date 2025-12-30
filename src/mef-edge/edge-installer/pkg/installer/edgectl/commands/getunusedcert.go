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
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/installer/edgectl/common"
)

type getUnusedCertInfoCmd struct {
	certName string
}

// NewGetUnusedCertInfoCmd edge control command get cert
func NewGetUnusedCertInfoCmd() common.Command {
	return &getUnusedCertInfoCmd{}
}

// Name command name
func (cmd *getUnusedCertInfoCmd) Name() string {
	return common.GetUnusedCertInfo
}

// Description command description
func (cmd *getUnusedCertInfoCmd) Description() string {
	return common.GetUnusedCertInfoDesc
}

// BindFlag command flag binding
func (cmd *getUnusedCertInfoCmd) BindFlag() bool {
	flag.StringVar(&cmd.certName, "name", "",
		"the name of unused certificate. Currently, only [cloud_root] is supported.")
	utils.MarkFlagRequired("name")
	return true
}

// LockFlag command lock flag
func (cmd *getUnusedCertInfoCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *getUnusedCertInfoCmd) Execute(ctx *common.Context) error {
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
		fmt.Println("no backup certificate found")
		return nil
	}
	fmt.Println(certPath)
	return nil
}

func (cmd *getUnusedCertInfoCmd) PrintOpLogOk(user string, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

func (cmd *getUnusedCertInfoCmd) PrintOpLogFail(user string, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}
