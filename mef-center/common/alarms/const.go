// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package alarms, common definition and function for alarm operation
package alarms

// alarm severity, operation and other const definitions
const (
	MinorSeverity    = "MINOR"
	MajorSeverity    = "MAJOR"
	CriticalSeverity = "CRITICAL"
	OkSeverity       = "OK"
	ClearFlag        = "clear"
	AlarmFlag        = "alarm"
	EventFlag        = ""
	AlarmType        = "alarm"
	EventType        = "event"
	CenterSn         = ""
)

// alarms id
const (
	NorthCertAbnormal              = "0x01000001"
	SoftwareCertAbnormal           = "0x01000002"
	ImageCertAbnormal              = "0x01000003"
	MEFCenterCaCertAbnormal        = "0x01000004"
	MEFCenterSvcCertAbnormal       = "0x01000005"
	MEFCenterCaCertUpdateAbnormal  = "0x01000006"
	MEFCenterSvcCertUpdateAbnormal = "0x01000007"
)
