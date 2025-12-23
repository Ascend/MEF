// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
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
