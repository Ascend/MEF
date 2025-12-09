// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package devmanager this for device driver manager
package devmanager

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"Ascend-device-plugin/pkg/devmanager/common"
	"Ascend-device-plugin/pkg/devmanager/dcmi"
)

// DeviceInterface for common device interface
type DeviceInterface interface {
	Init() error
	ShutDown() error
	GetDeviceCount() (int32, error)
	GetCardList() (int32, []int32, error)
	GetDeviceNumInCard(cardID int32) (int32, error)
	GetDeviceList() (int32, []int32, error)
	GetDeviceHealth(logicID int32) (uint32, error)
	GetDeviceErrorCode(logicID int32) (int32, int64, error)
	GetChipInfo(logicID int32) (*common.ChipInfo, error)
	GetPhysicIDFromLogicID(logicID int32) (int32, error)
	GetDeviceLogicID(cardID, deviceID int32) (int32, error)
	GetCardIDDeviceID(logicID int32) (int32, int32, error)
	GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error)
	GetDevType() string
	GetProductType(cardID, deviceID int32) (string, error)
	GetAllProductType() ([]string, error)
}

// DeviceManager common device manager for Ascend910/310P/310
type DeviceManager struct {
	// DcMgr for common dev manager
	DcMgr dcmi.DcDriverInterface
	// DevType the value is the same as the device type corresponding to the DcMgr variable.
	// Options: common.Ascend310,common.Ascend310P,common.Ascend910
	DevType string
}

// GetDevType return dev type
func (d *DeviceManager) GetDevType() string {
	return d.DevType
}

// AutoInit auto detect npu chip type and return the corresponding processing object
func AutoInit(dType string) (*DeviceManager, error) {
	chipInfo, err := getChipInfoForInit()
	if err != nil {
		return nil, fmt.Errorf("auto init failed, err: %s", err)
	}
	devManager := &DeviceManager{}
	devType := common.GetDeviceTypeByChipName(chipInfo.Name)
	switch devType {
	case common.Ascend310P:
		devManager.DcMgr = &A310PManager{}
	case common.Ascend310, common.Ascend310B:
		devManager.DcMgr = &A310Manager{}
	default:
		return nil, fmt.Errorf("unsupport device type (%s)", devType)
	}
	if dType != "" && devType != dType {
		return nil, fmt.Errorf("the value of dType(%s) is inconsistent with the actual chip type(%s)",
			dType, devType)
	}
	devManager.DevType = devType
	if err = devManager.Init(); err != nil {
		return nil, fmt.Errorf("deviceManager init failed, err: %#v", err)
	}
	return devManager, nil
}

func getChipInfoForInit() (common.ChipInfo, error) {
	dcMgr := dcmi.DcManager{}
	if err := dcMgr.DcInit(); err != nil {
		return common.ChipInfo{}, fmt.Errorf("dc init failed, err: %#v", err)
	}
	defer func() {
		if err := dcMgr.DcShutDown(); err != nil {
			hwlog.RunLog.Error(err)
		}
	}()
	// get card list
	carNum, cardList, err := dcMgr.DcGetCardList()
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.ChipInfo{}, fmt.Errorf("get card list failed for init")
	}
	if carNum == 0 {
		return common.ChipInfo{}, fmt.Errorf("get chip info failed, no card found")
	}
	// get device in card, then get chip info by cardID and deviceID
	for _, cardID := range cardList {
		devNum, err := dcMgr.DcGetDeviceNumInCard(cardID)
		if err != nil || devNum == 0 {
			hwlog.RunLog.Debugf("get device num by cardID(%d) failed, error: %#v", cardID, err)
			continue
		}
		for devID := int32(0); devID < devNum; devID++ {
			chipInfo, err := dcMgr.DcGetChipInfo(cardID, devID)
			if err != nil {
				hwlog.RunLog.Debugf("get chip info failed by cardID(%d), deviceID(%d), error: %#v", cardID, devID,
					err)
				continue
			}
			if !common.IsValidChipInfo(chipInfo) {
				hwlog.RunLog.Debugf("invalid chip info by cardID(%d), deviceID(%d), error: %#v", cardID, devID,
					err)
				continue
			}
			return *chipInfo, nil
		}
	}

	return common.ChipInfo{}, errors.New("cannot get valid chip info")
}

// Init load symbol and initialize dcmi
func (d *DeviceManager) Init() error {
	return d.DcMgr.DcInit()
}

// ShutDown clean the dynamically loaded resource
func (d *DeviceManager) ShutDown() error {
	return d.DcMgr.DcShutDown()
}

// GetDeviceCount get npu device count
func (d *DeviceManager) GetDeviceCount() (int32, error) {
	return d.DcMgr.DcGetDeviceCount()
}

// GetCardList  get all card list
func (d *DeviceManager) GetCardList() (int32, []int32, error) {
	return d.DcMgr.DcGetCardList()
}

// GetDeviceNumInCard  get all device list in one card
func (d *DeviceManager) GetDeviceNumInCard(cardID int32) (int32, error) {
	return d.DcMgr.DcGetDeviceNumInCard(cardID)
}

// GetDeviceList get all device logicID list
func (d *DeviceManager) GetDeviceList() (int32, []int32, error) {
	return d.DcMgr.DcGetLogicIDList()
}

// GetDeviceHealth query npu device health status
func (d *DeviceManager) GetDeviceHealth(logicID int32) (uint32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get health code by logicID(%d)", logicID)
	}
	healthCode, err := d.DcMgr.DcGetDeviceHealth(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get health code by logicID(%d)", logicID)
	}

	return uint32(healthCode), nil
}

// GetDeviceErrorCode get npu device error code
func (d *DeviceManager) GetDeviceErrorCode(logicID int32) (int32, int64, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, common.RetError, fmt.Errorf("failed to get device error code by logicID(%d)",
			logicID)
	}
	errCount, errCode, err := d.DcMgr.DcGetDeviceErrorCode(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, common.RetError, fmt.Errorf("failed to get device error code by logicID(%d)",
			logicID)
	}

	return errCount, errCode, nil
}

// GetChipInfo get npu device error code
func (d *DeviceManager) GetChipInfo(logicID int32) (*common.ChipInfo, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get chip info code by logicID(%d)", logicID)
	}
	chipInfo, err := d.DcMgr.DcGetChipInfo(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get chip info code by logicID(%d)", logicID)
	}

	return chipInfo, nil
}

// GetPhysicIDFromLogicID get device physic id from logic id
func (d *DeviceManager) GetPhysicIDFromLogicID(logicID int32) (int32, error) {
	physicID, err := d.DcMgr.DcGetPhysicIDFromLogicID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, fmt.Errorf("failed to get physicID by logicID(%d)", logicID)
	}

	return physicID, nil
}

// GetDeviceLogicID get device logic id from card id and device id
func (d *DeviceManager) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return d.DcMgr.DcGetDeviceLogicID(cardID, deviceID)
}

// GetVirtualDeviceInfo get virtual device info
func (d *DeviceManager) GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	cgoVDevInfo, err := d.DcMgr.DcGetVDeviceInfo(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.VirtualDevInfo{}, fmt.Errorf("get virtual device info failed, error is: %#v "+
			"and vdev num is: %d", err, int32(cgoVDevInfo.TotalResource.VDevNum))
	}
	for _, vDevInfo := range cgoVDevInfo.VDevInfo {
		if !common.IsValidTemplateName(d.DevType, vDevInfo.QueryInfo.Name) {
			return common.VirtualDevInfo{}, fmt.Errorf("vdevice id %d, it's template name is invalid: %s",
				vDevInfo.VDevID, vDevInfo.QueryInfo.Name)
		}
	}
	return cgoVDevInfo, nil
}

// GetCardIDDeviceID get cardID and deviceID by logicID
func (d *DeviceManager) GetCardIDDeviceID(logicID int32) (int32, int32, error) {
	return d.DcMgr.DcGetCardIDDeviceID(logicID)
}

// GetProductType get product type by cardID and deviceID
func (d *DeviceManager) GetProductType(cardID, deviceID int32) (string, error) {
	return d.DcMgr.DcGetProductType(cardID, deviceID)
}

// GetAllProductType get all product type
func (d *DeviceManager) GetAllProductType() ([]string, error) {
	var productTypes []string
	cardNum, cardList, err := d.GetCardList()
	if cardNum == 0 || err != nil {
		hwlog.RunLog.Errorf("failed to get card list, err: %#v", err)
		return productTypes, err
	}
	for _, cardID := range cardList {
		devNum, err := d.GetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Debugf("get device num by cardID(%d) failed, error: %#v", cardID, err)
			continue
		}
		if devNum == 0 {
			hwlog.RunLog.Debugf("not found device on card %d", cardID)
			continue
		}
		for devID := int32(0); devID < devNum; devID++ {
			productType, err := d.GetProductType(cardID, devID)
			if err != nil {
				hwlog.RunLog.Debugf("get product type by card %d deviceID %d failed, err: %#v", cardID, devID, err)
				continue
			}
			productTypes = append(productTypes, productType)
			break
		}
	}
	productTypes = common.RemoveDuplicate(&productTypes)
	return productTypes, nil
}
