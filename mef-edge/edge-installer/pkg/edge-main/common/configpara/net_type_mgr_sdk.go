// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package configpara for
package configpara

import (
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

// GetNetType [method] for get net type
func GetNetType() (string, error) {
	return constants.MEF, nil
}

func setNetConfig(cfg *config.NetManager) {
	utils.ClearSliceByteMemory(netCfg.Token)
	netCfg = *cfg
}
