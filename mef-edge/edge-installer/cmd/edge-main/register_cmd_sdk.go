// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package main for
package main

import (
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/cloudcoreproxy"
	"edge-installer/pkg/edge-main/downloadmgr"
	"edge-installer/pkg/edge-main/edgehub"
)

func moduleExt(netType string) []model.Module {
	if netType != constants.MEF {
		return nil
	}

	return []model.Module{
		edgehub.NewEdgeHub(true),
		cloudcoreproxy.NewCloudCoreProxy(true),
		downloadmgr.NewDownloadMgr(true),
	}
}
