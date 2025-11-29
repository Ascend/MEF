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
