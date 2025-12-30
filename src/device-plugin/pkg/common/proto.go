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

// Package common a series of common function
package common

import (
	"github.com/fsnotify/fsnotify"
)

var (
	// ParamOption for option
	ParamOption Option
)

// DeviceHealth health status of device
type DeviceHealth struct {
	Health string
}

// NpuAllInfo all npu infos
type NpuAllInfo struct {
	AllDevTypes []string
	AllDevs     []NpuDevice
	AICoreDevs  []*NpuDevice
}

// NpuDevice npu device description
type NpuDevice struct {
	DevType    string
	DeviceName string
	Health     string
	IP         string
	LogicID    int32
	PhyID      int32
	CardID     int32
}

// DavinCiDev davinci device
type DavinCiDev struct {
	LogicID int32
	PhyID   int32
	CardID  int32
}

// Device id for Instcance
type Device struct { // Device
	DeviceID string `json:"device_id"` // device id
}

// Option option
type Option struct {
	GetFdFlag          bool     // to describe FdFlag
	ListAndWatchPeriod int      // set listening device state period
	ShareCount         uint     // share device count
	AiCoreCount        int32    // found by dcmi interface
	ProductTypes       []string // all product types
	RealCardType       string   // real card type
}

// FileWatch is used to watch sock file
type FileWatch struct {
	FileWatcher *fsnotify.Watcher
}

// Get310PProductType get 310P product type
func Get310PProductType() map[string]string {
	return map[string]string{
		"Atlas 300V Pro": Ascend310PVPro,
		"Atlas 300V":     Ascend310PV,
		"Atlas 300I Pro": Ascend310PIPro,
	}
}
