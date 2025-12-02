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

const (
	// Component component name
	Component = "device-plugin"
	// MaxBackups log file max backup
	MaxBackups = 30
	// MaxAge the log file last time
	MaxAge = 7

	// MaxContainerLimit max container num
	MaxContainerLimit = 300000
	// MaxDeviceNameLen max length of device name, like "Ascend310P-4c.3cpu-100-0"
	MaxDeviceNameLen = 50
	// MaxGRPCRecvMsgSize 4MB
	MaxGRPCRecvMsgSize = 4 * 1024 * 1024
	// MaxGRPCConcurrentStreams limit on the number of concurrent streams to each ServerTransport.
	MaxGRPCConcurrentStreams = 64
	// MaxConcurrentLimit limit over listener
	MaxConcurrentLimit = 64
	// MaxIPConnectionLimit limit over ip
	MaxIPConnectionLimit = 64
	// CacheSize cache for ip
	CacheSize = 128
	// MaxVirtualDeviceNum max num of virtual device
	MaxVirtualDeviceNum = 1024

	// VirtualDev Virtual device tag
	VirtualDev = "VIRTUAL"
	// PhyDeviceLen like Ascend910-0 split length is 2
	PhyDeviceLen = 2
	// VirDeviceLen like Ascend910-2c-100-1 split length is 4
	VirDeviceLen = 4
	// MaxDevicesNum max device num
	MaxDevicesNum = 100
	// MaxCardNum max card num
	MaxCardNum = 64
	// MaxDevNumInCard max device num in card
	MaxDevNumInCard = 4
	// MaxRequestVirtualDeviceNum max request device num
	MaxRequestVirtualDeviceNum = 1
	// DefaultDeviceIP device ip address
	DefaultDeviceIP = "127.0.0.1"
	// NormalState health state
	NormalState = uint32(0)
	// GeneralAlarm health state
	GeneralAlarm = uint32(1)

	// SocketChmod socket file mode
	SocketChmod = 0600

	// Interval interval time
	Interval = 1
	// Timeout time
	Timeout = 10
	// SleepTime The unit is seconds
	SleepTime = 5
)

const (
	// ResourceNamePrefix prefix
	ResourceNamePrefix = "huawei.com/"
	// Ascend310P 310p
	Ascend310P = "Ascend310P"
	// Ascend310PV 310P-V
	Ascend310PV = Ascend310P + "-V"
	// Ascend310PVPro 310P-VPro
	Ascend310PVPro = Ascend310P + "-VPro"
	// Ascend310PIPro 310P-IPro
	Ascend310PIPro = Ascend310P + "-IPro"

	// Ascend910 910
	Ascend910 = "Ascend910"

	// Ascend310 310
	Ascend310 = "Ascend310"
	// Ascend310B 310B chip
	Ascend310B = "Ascend310B"
	// HuaweiAscend310 with prefix
	HuaweiAscend310 = ResourceNamePrefix + Ascend310
	// AscendfdPrefix use in fd
	AscendfdPrefix = "davinci-mini"

	// HuaweiUnHealthAscend310P 310p unhealthy
	HuaweiUnHealthAscend310P = ResourceNamePrefix + Ascend310P + "-Unhealthy"
	// HuaweiUnHealthAscend310 310 unhealthy
	HuaweiUnHealthAscend310 = ResourceNamePrefix + Ascend310 + "-Unhealthy"

	// AiCoreResourceName resource name for virtual device
	AiCoreResourceName = "npu-core"

	// Core1 1 core
	Core1 = "1c"
	// Core2 2 core
	Core2 = "2c"
	// Core4 4 core
	Core4 = "4c"
	// Core8 8 core
	Core8 = "8c"

	// Core4Cpu3 4core 3cpu
	Core4Cpu3 = "4c.3cpu"
	// Core2Cpu1 2core 1cpu
	Core2Cpu1 = "2c.1cpu"
	// Core4Cpu4Dvpp 4core 4cpu dvpp
	Core4Cpu4Dvpp = "4c.4cpu.dvpp"
	// Core4Cpu3Ndvpp 4core 3cpu ndvpp
	Core4Cpu3Ndvpp = "4c.3cpu.ndvpp"

	// Vir01 template name vir01
	Vir01 = "vir01"
	// Vir02 template name vir02
	Vir02 = "vir02"
	// Vir04 template name vir04
	Vir04 = "vir04"
	// Vir08 template name vir08
	Vir08 = "vir08"
	// Vir04C3 template name vir04_3c
	Vir04C3 = "vir04_3c"
	// Vir02C1 template name vir02_1c
	Vir02C1 = "vir02_1c"
	// Vir04C4Dvpp template name vir04_4c_dvpp
	Vir04C4Dvpp = "vir04_4c_dvpp"
	// Vir04C3Ndvpp template name vir04_3c_ndvpp
	Vir04C3Ndvpp = "vir04_3c_ndvpp"

	// MinAICoreNum min ai core num
	MinAICoreNum = 8

	// MaxShareDevCount open share device function, max share count is 100
	MaxShareDevCount = 100
)

const (
	// HiAIHDCDevice hisi_hdc
	HiAIHDCDevice = "/dev/hisi_hdc"
	// HiAIManagerDevice davinci_manager
	HiAIManagerDevice = "/dev/davinci_manager"
	// HiAIManagerDeviceDocker davinci_manager for docker
	HiAIManagerDeviceDocker = "/dev/davinci_manager_docker"
	// HiAISVMDevice devmm_svm
	HiAISVMDevice = "/dev/devmm_svm"
	// HiAi200RCSVM0 svm0
	HiAi200RCSVM0 = "/dev/svm0"
	// HiAi200RCLog log_drv
	HiAi200RCLog = "/dev/log_drv"
	// HiAi200RCEventSched event_sched
	HiAi200RCEventSched = "/dev/event_sched"
	// HiAi200RCUpgrade upgrade
	HiAi200RCUpgrade = "/dev/upgrade"
	// HiAi200RCHiDvpp hi_dvpp
	HiAi200RCHiDvpp = "/dev/hi_dvpp"
	// HiAi200RCMemoryBandwidth memory_bandwidth
	HiAi200RCMemoryBandwidth = "/dev/memory_bandwidth"
	// HiAi200RCTsAisle ts_aisle
	HiAi200RCTsAisle = "/dev/ts_aisle"
)

const (
	// Atlas200ISoc 200 soc env
	Atlas200ISoc = "Atlas 200I SoC A1"
	// Atlas200ISocXSMEM is xsmem_dev
	Atlas200ISocXSMEM = "/dev/xsmem_dev"
	// Atlas200ISocSYS is sys
	Atlas200ISocSYS = "/dev/sys"
	// Atlas200ISocVDEC is vdec
	Atlas200ISocVDEC = "/dev/vdec"
	// Atlas200ISocVPC is vpc
	Atlas200ISocVPC = "/dev/vpc"
	// Atlas200ISocSpiSmbus is spi_smbus
	Atlas200ISocSpiSmbus = "/dev/spi_smbus"
	// Atlas200ISocUserConfig is user_config
	Atlas200ISocUserConfig = "/dev/user_config"
)

const (
	// Atlas310BDvppCmdlist is dvpp_cmdlist
	Atlas310BDvppCmdlist = "/dev/dvpp_cmdlist"
	// Atlas310BPngd is pngd
	Atlas310BPngd = "/dev/pngd"
	// Atlas310BVenc is venc
	Atlas310BVenc = "/dev/venc"
)

// Audio and video dependent device for Atlas310B
const (
	Atlas310BAcodec = "/dev/acodec"
	Atlas310BAi     = "/dev/ai"
	Atlas310BAo     = "/dev/ao"
	Atlas310BVo     = "/dev/vo"
	Atlas310BHdmi   = "/dev/hdmi"
)

const (
	// RootUID is root user id
	RootUID = 0
	// RootGID is root group id
	RootGID = 0

	// MiddelLine if the separator between devices for split id
	MiddelLine = "-"
)

const (
	// Atlas300IDuo for hot reset function, sync chip healthy state
	Atlas300IDuo = "Atlas 300I Duo"
)
