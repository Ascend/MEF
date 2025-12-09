// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common
package common

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/database"
)

// LoadNpuFromDb get npu name
func LoadNpuFromDb() (string, bool) {
	metas, err := database.GetMetaRepository().GetByType(constants.ResourceTypeNode)
	if err != nil {
		hwlog.RunLog.Errorf("get node from db err: %v", err)
		return "", false
	}
	if len(metas) != 1 {
		hwlog.RunLog.Errorf("%s meta count not correct", constants.ResourceTypeNode)
		return "", false
	}

	nodeMeta := metas[0].Value
	mapContent, err := util.GetContentMap(nodeMeta)
	if err != nil {
		hwlog.RunLog.Errorf("convert data fail when load npu from db : %s", err.Error())
		return "", false
	}
	contentWrapper := util.NewWrapper(mapContent)
	capacityObj := contentWrapper.GetObject("status").GetObject("capacity").GetData()
	return util.FindMostQualifiedNpu(capacityObj)
}
