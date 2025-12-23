// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build !MEFEdge_SDK || MEFEdge_A500

// Package handlermgr for deal every handler
package handlermgr

import (
	"encoding/json"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

func getNetConfig(dbMgr *config.DbMgr) string {
	netConfig, err := config.GetNetManager(dbMgr)
	if err != nil {
		hwlog.RunLog.Errorf("get net manager config failed: %v", err)
		return constants.Failed
	}
	config.SetNetManagerCache(*netConfig)
	bytes, err := json.Marshal(netConfig)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data failed: %v", err)
		return constants.Failed
	}

	return string(bytes)
}
