/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
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
