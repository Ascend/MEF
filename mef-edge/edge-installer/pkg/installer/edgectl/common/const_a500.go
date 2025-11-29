// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

// Package common for control constants
package common

// inner control commands
const (
	CopyResetScriptCmd  = "copy_reset_script"
	CopyResetScriptDesc = "to copy reset script to filesystem p7"
	RestoreCfgCmd       = "restore_config"
	RestoreCfgDesc      = "to copy restore default config"
)
