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

// Package main for
package main

import "edge-installer/pkg/installer/edgectl/commands"

func initCmdExt() {
	registerCmd(commands.DomainConfigCmd())
	registerCmd(commands.NetConfigCmd())
	registerCmd(commands.NewGetCertInfoCmd())
	registerCmd(commands.ImportCrlCmd())
	registerCmd(commands.AlarmConfigCmd())
	registerCmd(commands.GetAlarmCfgCmd())
	registerCmd(commands.NewGetUnusedCertInfoCmd())
	registerCmd(commands.NewDeleteCertInfoCmd())
	registerCmd(commands.NewRestoreCertInfoCmd())
}
