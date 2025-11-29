// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

//go:build MEFEdge_SDK

// Package handlermgr
package handlermgr

func init() {
	initFuncList = append(initFuncList, initLogDumpDirs)
}
