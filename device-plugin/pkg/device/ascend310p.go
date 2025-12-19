/* Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MEF is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
*/

// Package device a series of device function
package device

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"Ascend-device-plugin/pkg/common"
)

// HwAscend310PManager manages huawei Ascend310P devices.
type HwAscend310PManager struct {
	AscendTools
}

// NewHwAscend310PManager used to create ascend 310P manager
func NewHwAscend310PManager() *HwAscend310PManager {
	return &HwAscend310PManager{
		AscendTools: AscendTools{
			name:         common.Ascend310P,
			unHealthyKey: common.HuaweiUnHealthAscend310P,
			devCount:     common.MaxDevicesNum,
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend310P devices by call devmanager interface
func (hnm *HwAscend310PManager) GetNPUs() (common.NpuAllInfo, error) {
	_, devList, err := hnm.dmgr.GetDeviceList()
	if err != nil {
		return common.NpuAllInfo{}, err
	}
	if int32(len(devList)) > hnm.devCount {
		return common.NpuAllInfo{}, fmt.Errorf("invalid device num: %d", len(devList))
	}
	var allDevices []common.NpuDevice
	var aiCoreDevices []*common.NpuDevice
	var allDeviceTypes []string
	for _, dev := range devList {
		davinCiDev, err := hnm.getDavinCiDev(dev)
		if err != nil {
			return common.NpuAllInfo{}, err
		}

		vDevInfos, err := hnm.getVirtualDevice(dev)
		if err != nil {
			hwlog.RunLog.Errorf("The virtual device is considered not exist, please check the error: %#v", err)
		}
		if vDevInfos.TotalResource.VDevNum > common.MaxVirtualDeviceNum {
			return common.NpuAllInfo{}, fmt.Errorf("invalid virtual device count")
		}
		if vDevInfos.TotalResource.VDevNum == 0 {
			hnm.assemblePhyDevices(davinCiDev, &allDevices, &allDeviceTypes)
			continue
		}
		hnm.assembleVirtualDevices(davinCiDev, vDevInfos, &allDevices, &allDeviceTypes)
	}
	allDeviceTypes = hnm.removeDuplicate(&allDeviceTypes)
	return common.NpuAllInfo{AllDevs: allDevices, AICoreDevs: aiCoreDevices, AllDevTypes: allDeviceTypes}, nil
}
