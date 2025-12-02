/* Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
*/

// Package device a series of device function.
package device

import (
	"fmt"

	"Ascend-device-plugin/pkg/common"
)

// HwAscend310Manager manages huawei Ascend310 devices.
type HwAscend310Manager struct {
	AscendTools
}

// NewHwAscend310Manager used to create ascend 310 manager
func NewHwAscend310Manager() *HwAscend310Manager {
	name := common.Ascend310
	if common.ParamOption.GetFdFlag {
		name = common.AscendfdPrefix
	}
	return &HwAscend310Manager{
		AscendTools: AscendTools{
			name:         name,
			unHealthyKey: common.HuaweiUnHealthAscend310,
			devCount:     common.MaxCardNum * common.MaxDevNumInCard,
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend310 devices by call devmanager interface
func (hnm *HwAscend310Manager) GetNPUs() (common.NpuAllInfo, error) {
	_, devList, err := hnm.dmgr.GetDeviceList()
	if err != nil {
		return common.NpuAllInfo{}, err
	}
	if int32(len(devList)) > hnm.devCount {
		return common.NpuAllInfo{}, fmt.Errorf("invalid device num: %d", len(devList))
	}
	var allDevices []common.NpuDevice
	for _, dev := range devList {
		davinCiDev, err := hnm.getDavinCiDev(dev)
		if err != nil {
			return common.NpuAllInfo{}, err
		}
		normalDevices := hnm.getNPUsByNormalMode(davinCiDev)
		if common.ShareDev() {
			normalDevices = hnm.getNPUsByShareMode(davinCiDev)
		}
		allDevices = append(allDevices, normalDevices...)
	}
	return common.NpuAllInfo{AllDevs: allDevices, AllDevTypes: []string{hnm.name}}, err
}

func (hnm *HwAscend310Manager) getNPUsByNormalMode(davinCiDev common.DavinCiDev) []common.NpuDevice {
	deviceName := fmt.Sprintf("%s-%d", hnm.name, davinCiDev.PhyID)
	return []common.NpuDevice{hnm.assembleNpuDeviceStruct(hnm.name, deviceName, davinCiDev)}
}

func (hnm *HwAscend310Manager) getNPUsByShareMode(davinCiDev common.DavinCiDev) []common.NpuDevice {
	shareDevices := make([]common.NpuDevice, 0, common.ParamOption.ShareCount)
	for id := uint(davinCiDev.LogicID) * common.ParamOption.ShareCount; id < uint(davinCiDev.LogicID+1)*
		common.ParamOption.ShareCount; id++ {
		deviceName := fmt.Sprintf("%s-%d", hnm.name, id)
		device := hnm.assembleNpuDeviceStruct(hnm.name, deviceName, davinCiDev)
		shareDevices = append(shareDevices, device)
	}
	return shareDevices
}
