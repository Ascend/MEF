// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

// Package main for register command
package main

import "edge-installer/pkg/installer/edgectl/innercommands"

func initCmdExt() {
	registerCmd(innercommands.ExchangeCertsCmd())
	registerCmd(innercommands.RecoveryCmd())
	registerCmd(innercommands.CopyResetScriptCmd())
	registerCmd(innercommands.RestoreCfgCmd())
}
