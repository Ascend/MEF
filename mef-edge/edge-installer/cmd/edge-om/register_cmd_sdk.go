// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package main

import (
	"context"

	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/edge-om/upgrade"
)

func moduleExt(ctx context.Context) []model.Module {
	modules := []model.Module{
		upgrade.NewUpgradeMgr(ctx, true),
	}

	return modules
}
