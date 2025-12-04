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

// Package common a series of common function
package common

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
)

// GetDeviceID get device physical id and virtual by device name
func GetDeviceID(deviceName string, ascendRuntimeOptions string) (int, int, error) {
	// hiAIAscend310Prefix: davinci-mini
	// vnpu: davinci-coreNum-vid-devID, like Ascend910-2c-111-0
	// ascend310:  davinci-mini0
	idSplit := strings.Split(deviceName, MiddelLine)

	if len(idSplit) < PhyDeviceLen {
		return 0, 0, fmt.Errorf("id: %s is invalid", deviceName)
	}
	phyIDStr := idSplit[len(idSplit)-1]
	// for virtual device, index 2 data means it's id
	var virID int
	if ascendRuntimeOptions == VirtualDev && len(idSplit) == VirDeviceLen {
		var err error
		virID, err = strconv.Atoi(idSplit[PhyDeviceLen])
		if err != nil {
			return 0, 0, fmt.Errorf("convert vnpu id %s failed, erros is %v", idSplit[PhyDeviceLen], err)
		}
	}
	phyID, err := strconv.Atoi(phyIDStr)
	if err != nil {
		return 0, 0, fmt.Errorf("convert physical id %s failed, erros is %v", phyIDStr, err)
	}
	return phyID, virID, nil
}

// GetDeviceListID get device id by input device name
func GetDeviceListID(devices []string, ascendRuntimeOptions string) (map[int]int, []int, error) {
	if len(devices) > MaxDevicesNum {
		return nil, nil, fmt.Errorf("device num excceed max num, when get device list id")
	}
	var ascendVisibleDevices []int
	phyDevMapVirtualDev := make(map[int]int, MaxDevicesNum)
	for _, id := range devices {
		deviceID, virID, err := GetDeviceID(id, ascendRuntimeOptions)
		if err != nil {
			hwlog.RunLog.Errorf("get device ID err: %#v", err)
			return nil, nil, err
		}
		if ascendRuntimeOptions == VirtualDev {
			ascendVisibleDevices = append(ascendVisibleDevices, virID)
			phyDevMapVirtualDev[virID] = deviceID
			continue
		}
		if ShareDev() {
			deviceID = deviceID / int(ParamOption.ShareCount)
		}
		ascendVisibleDevices = append(ascendVisibleDevices, deviceID)
	}
	return phyDevMapVirtualDev, ascendVisibleDevices, nil
}

// ShareDev open the share dev function
func ShareDev() bool {
	return ParamOption.ShareCount > 1 && ParamOption.RealCardType == Ascend310B
}

// IsVirtualDev used to judge whether a physical device or a virtual device
func IsVirtualDev(devType string) bool {
	patternMap := GetPattern()
	reg310P := regexp.MustCompile(patternMap["vir310p"])
	return reg310P.MatchString(devType)
}

// ToString convert input data to string
func ToString(devices sets.String, sepType string) string {
	return strings.Join(devices.List(), sepType)
}

// GetTemplateName2DeviceTypeMap get virtual device type by template
func GetTemplateName2DeviceTypeMap() map[string]string {
	return map[string]string{
		Vir08:        Core8,
		Vir04:        Core4,
		Vir02:        Core2,
		Vir01:        Core1,
		Vir04C3:      Core4Cpu3,
		Vir02C1:      Core2Cpu1,
		Vir04C4Dvpp:  Core4Cpu4Dvpp,
		Vir04C3Ndvpp: Core4Cpu3Ndvpp,
	}
}
