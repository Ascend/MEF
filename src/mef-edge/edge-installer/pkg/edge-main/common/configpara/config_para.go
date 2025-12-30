// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package configpara for config para
package configpara

import (
	"edge-installer/pkg/common/config"
)

var podConfig config.PodConfig
var netCfg config.NetManager
var installerCfg config.InstallerConfig
var podFlag bool
var netFlag bool
var installFlag bool
var npuSharingFlag bool

// SetCfgPara [method] set config para
func SetCfgPara(value interface{}) {
	switch value.(type) {
	case *config.PodConfig:
		podConfig = *(value.(*config.PodConfig))
		podFlag = true
	case *config.NetManager:
		setNetConfig(value.(*config.NetManager))
		netFlag = true
	case *config.InstallerConfig:
		installerCfg = *(value.(*config.InstallerConfig))
		installFlag = true
	case *config.StaticInfo:
		config.GetCapabilityCache().SetEdgeOmCaps(*(value.(*config.StaticInfo)))
		npuSharingFlag = true
	default:
		return
	}

}

// CheckCfgIsReady [method] for check cfg from edge om is ready
func CheckCfgIsReady() bool {
	return installFlag && netFlag && podFlag && npuSharingFlag
}

// GetInstallerConfig [method] for get installer config in mem
func GetInstallerConfig() config.InstallerConfig {
	return installerCfg
}

// GetNetConfig [method] for get net manager config in mem
func GetNetConfig() config.NetManager {
	return netCfg
}

// GetPodConfig [method] for get pod config in mem
func GetPodConfig() config.PodConfig {
	return podConfig
}
