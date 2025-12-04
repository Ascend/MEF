// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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
