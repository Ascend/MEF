// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package handlermgr for sdk get net config
package handlermgr

import (
	"encoding/json"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

func getNetConfig(dbMgr *config.DbMgr) string {
	netConfig, err := config.GetNetManager(dbMgr)
	if err != nil {
		hwlog.RunLog.Errorf("get net manager config failed: %v", err)
		return constants.Failed
	}
	config.SetNetManagerCache(*netConfig)

	if err = decryptToken(netConfig); err != nil {
		hwlog.RunLog.Errorf("decrypt token failed %v", err)
		return constants.Failed
	}
	defer utils.ClearSliceByteMemory(netConfig.Token)

	bytes, err := json.Marshal(netConfig)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data failed: %v", err)
		return constants.Failed
	}

	return string(bytes)
}

func decryptToken(netConfig *config.NetManager) error {
	if netConfig.NetType != constants.MEF {
		return nil
	}
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		return err
	}
	kmcDir := configPathMgr.GetCompKmcDir(constants.EdgeOm)
	kmcCfg, err := util.GetKmcConfig(kmcDir)
	if err != nil {
		hwlog.RunLog.Error("get kmc config error when decrypt token")
		return err
	}
	token, err := kmc.DecryptContent(netConfig.Token, kmcCfg)
	if err != nil {
		hwlog.RunLog.Error("encrypt token failed")
		return err
	}
	netConfig.Token = token
	return nil
}
