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

// Package device a series of device function
package device

import (
	"fmt"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/devmanager"
	npuCommon "Ascend-device-plugin/pkg/devmanager/common"
)

// AscendTools struct definition
type AscendTools struct {
	dmgr         devmanager.DeviceInterface
	name         string
	unHealthyKey string
	devCount     int32
	healthDevice sets.String
}

// DevManager interface for manager device
type DevManager interface {
	GetNPUs() (common.NpuAllInfo, error)
	SetDmgr(devmanager.DeviceInterface)
	GetDmgr() devmanager.DeviceInterface
	GetChipAICore() int32
	GetName() string
	IsDeviceStatusChange(map[string][]*common.NpuDevice, []*common.NpuDevice, string) map[string]bool
}

// SetDmgr set devmanager
func (tool *AscendTools) SetDmgr(dmgr devmanager.DeviceInterface) {
	tool.dmgr = dmgr
}

// GetDmgr get devmanager
func (tool *AscendTools) GetDmgr() devmanager.DeviceInterface {
	return tool.dmgr
}

// GetChipAICore get ai core
func (tool *AscendTools) GetChipAICore() int32 {
	return common.ParamOption.AiCoreCount
}

// GetName get chip name
func (tool *AscendTools) GetName() string {
	return tool.name
}

func (tool *AscendTools) assembleNpuDeviceStruct(deviType, deviceName string,
	davinCiDev common.DavinCiDev) common.NpuDevice {
	hwlog.RunLog.Debugf("Found Huawei Ascend, deviceType: %s, deviceName: %s", deviType, deviceName)
	return common.NpuDevice{
		DevType:    deviType,
		DeviceName: deviceName,
		Health:     v1beta1.Healthy,
		LogicID:    davinCiDev.LogicID,
		PhyID:      davinCiDev.PhyID,
		CardID:     davinCiDev.CardID,
	}
}

func (tool *AscendTools) assemblePhyDevices(davinCiDev common.DavinCiDev, devices *[]common.NpuDevice,
	deviceTypes *[]string) {
	deviceName := fmt.Sprintf("%s-%d", tool.name, davinCiDev.PhyID)
	device := tool.assembleNpuDeviceStruct(tool.name, deviceName, davinCiDev)
	*deviceTypes = append(*deviceTypes, tool.name)
	*devices = append(*devices, device)
}

func (tool *AscendTools) assembleVirtualDevices(davinCiDev common.DavinCiDev, vDevInfos npuCommon.VirtualDevInfo,
	devices *[]common.NpuDevice, vDeviceTypes *[]string) {
	for _, subVDevInfo := range vDevInfos.VDevInfo {
		vDeviType, deviceName, err := tool.assembleSpecVirtualDevice(davinCiDev.PhyID, subVDevInfo)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		device := tool.assembleNpuDeviceStruct(vDeviType, deviceName, davinCiDev)
		*devices = append(*devices, device)
		*vDeviceTypes = append(*vDeviceTypes, vDeviType)
	}
}

func (tool *AscendTools) assembleSpecVirtualDevice(phyID int32, vDevInfo npuCommon.CgoVDevQueryStru) (string,
	string, error) {
	coreNum := int32(vDevInfo.QueryInfo.Computing.Aic)
	if coreNum <= 0 {
		return "", "", fmt.Errorf("invalid vdev info, ai core is 0")
	}
	vDeviType, exist := common.GetTemplateName2DeviceTypeMap()[vDevInfo.QueryInfo.Name]
	if !exist {
		return "", "", fmt.Errorf("check templatename failed, templatename is %s", vDevInfo.QueryInfo.Name)
	}
	vDeviType = fmt.Sprintf("%s-%s", tool.name, vDeviType)
	devID := fmt.Sprintf("%s-%d-%d", vDeviType, vDevInfo.VDevID, phyID)
	return vDeviType, devID, nil
}

func (tool *AscendTools) removeDuplicate(allDeviceTypes *[]string) []string {
	deviceTypesMap := make(map[string]string, len(*allDeviceTypes))
	var rmDupDeviceTypes []string
	for _, deviType := range *allDeviceTypes {
		deviceTypesMap[deviType] = deviType
	}
	for _, deviType := range deviceTypesMap {
		rmDupDeviceTypes = append(rmDupDeviceTypes, deviType)
	}
	return rmDupDeviceTypes
}

func (tool *AscendTools) getDavinCiDev(logicID int32) (common.DavinCiDev, error) {
	phyID, err := tool.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return common.DavinCiDev{}, err
	}
	cardID, _, err := tool.dmgr.GetCardIDDeviceID(logicID)
	if err != nil {
		return common.DavinCiDev{}, err
	}
	return common.DavinCiDev{
		LogicID: logicID,
		PhyID:   phyID,
		CardID:  cardID,
	}, nil
}

func (tool *AscendTools) getVirtualDevice(logicID int32) (npuCommon.VirtualDevInfo, error) {
	virtualDevInfos, err := tool.dmgr.GetVirtualDeviceInfo(logicID)
	if err != nil {
		return npuCommon.VirtualDevInfo{}, fmt.Errorf("query virtual device info failure: %s", err)
	}
	return virtualDevInfos, nil
}

func (tool *AscendTools) getDeviceListIP(devices []string, deviceType string) (map[int]string, error) {
	ascendRuntimeOptions := ""
	if common.IsVirtualDev(deviceType) {
		ascendRuntimeOptions = common.VirtualDev
	}
	_, ascendDevices, err := common.GetDeviceListID(devices, ascendRuntimeOptions)
	if err != nil {
		hwlog.RunLog.Errorf("get device list id err: %#v", err)
		return nil, err
	}
	devicesWithIP := make(map[int]string, len(devices))
	for _, id := range ascendDevices {
		if ascendRuntimeOptions == common.VirtualDev {
			devicesWithIP[id] = common.DefaultDeviceIP
			continue
		}
		if !strings.Contains(deviceType, common.Ascend910) {
			devicesWithIP[id] = ""
			continue
		}
	}
	return devicesWithIP, nil
}

// IsDeviceStatusChange is device status change
func (tool *AscendTools) IsDeviceStatusChange(groupDevice map[string][]*common.NpuDevice,
	aiCoreDevs []*common.NpuDevice, runMode string) map[string]bool {
	// get all chip by logic id
	healthStatus := make(map[int32]common.DeviceHealth, 1)
	for _, devices := range groupDevice {
		for _, device := range devices {
			healthStatus[device.LogicID] = common.DeviceHealth{Health: v1beta1.Healthy}
		}
	}
	// get all chip's health
	for logicID := range healthStatus {
		healthStatus[logicID] = common.DeviceHealth{Health: tool.getDevState(logicID)}
	}

	// update all device's health
	isStateChange := make(map[string]bool, len(groupDevice))
	for devType, devices := range groupDevice {
		for idx, device := range devices {
			if healthStatus[device.LogicID].Health != device.Health {
				isStateChange[devType] = true
				devices[idx].Health = healthStatus[device.LogicID].Health
			}
		}
	}
	tool.syncDuoCardState(groupDevice)
	return isStateChange
}

func (tool *AscendTools) syncDuoCardState(groupDevice map[string][]*common.NpuDevice) {
	if !common.IsContainAtlas300IDuo() {
		return
	}

	ascend310PDevices, ok := groupDevice[common.Ascend310P]
	if !ok {
		hwlog.RunLog.Debugf("not found 310P devices")
		return
	}
	unHealthyCards := getUnHealthyCard(ascend310PDevices)
	for devType, devices := range groupDevice {
		if devType != common.Ascend310P {
			continue
		}
		for idx, device := range devices {
			if _, ok := unHealthyCards[device.CardID]; ok {
				devices[idx].Health = v1beta1.Unhealthy
			}
		}
	}
}

func getUnHealthyCard(ascend310PDevices []*common.NpuDevice) map[int32]interface{} {
	unHealthyCards := make(map[int32]interface{}, len(ascend310PDevices))
	for _, device := range ascend310PDevices {
		if device.Health == v1beta1.Healthy {
			continue
		}
		unHealthyCards[device.CardID] = struct{}{}
	}
	return unHealthyCards
}

// ClassifyDevices classify diff type devices
func ClassifyDevices(allDevs []common.NpuDevice, devTypes []string) map[string][]*common.NpuDevice {
	var classifyMap = make(map[string][]*common.NpuDevice, len(devTypes))
	for _, suffix := range devTypes {
		classifyMap[suffix] = classifyDevByType(allDevs, suffix)
	}
	return classifyMap
}

func classifyDevByType(allDevs []common.NpuDevice, suffix string) []*common.NpuDevice {
	var classifyDev []*common.NpuDevice
	for index, device := range allDevs {
		if device.DevType == suffix {
			classifyDev = append(classifyDev, &allDevs[index])
		}
	}
	return classifyDev
}

func (tool *AscendTools) getDevState(logicID int32) string {
	healthState, err := tool.dmgr.GetDeviceHealth(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("get device healthy state failed, deviceId: %d, err: %#v", logicID, err)
		return v1beta1.Unhealthy
	}
	switch healthState {
	case common.NormalState, common.GeneralAlarm:
		return v1beta1.Healthy
	default:
		if err = tool.unhealthyState(healthState, logicID); err != nil {
			hwlog.RunLog.Errorf("UnhealthyState, err: %#v", err)
		}
		return v1beta1.Unhealthy
	}
}

// UnhealthyState state unhealthy info
func (tool *AscendTools) unhealthyState(healthyState uint32, logicID int32) error {
	phyID, err := tool.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return fmt.Errorf("get phyID failed %#v", err)
	}
	if _, _, err := tool.dmgr.GetDeviceErrorCode(logicID); err != nil {
		return fmt.Errorf("get device error code failed %#v", err)
	}
	hwlog.RunLog.Errorf("device logicID: %d, phyID: %d, state is %d", logicID, phyID, healthyState)
	return nil
}
