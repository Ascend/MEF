// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
