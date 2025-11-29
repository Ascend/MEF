// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
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
