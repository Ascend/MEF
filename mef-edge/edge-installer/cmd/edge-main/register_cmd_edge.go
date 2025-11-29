// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build !MEFEdge_SDK

// Package main for
package main

import "huawei.com/mindx/common/modulemgr/model"

func moduleExt(netType string) []model.Module {
	return nil
}
