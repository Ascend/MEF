// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

// Package configpara for
package configpara

import (
	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

// GetNetType [method] for get net type
func GetNetType() (string, error) {
	return constants.FDWithOM, nil
}

func setNetConfig(cfg *config.NetManager) {
	netCfg = *cfg
}
