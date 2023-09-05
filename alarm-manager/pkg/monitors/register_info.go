// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package monitors defined cert alarm register info
package monitors

import (
	"alarm-manager/pkg/utils"
	"huawei.com/mindxedge/base/common/requests"
)

// alarms id
const (
	NorthCertAbnormal    = "0x01000001"
	SoftwareCertAbnormal = "0x01000002"
	ImageCertAbnormal    = "0x01000003"
)

var allAlarms = []requests.AlarmReq{
	{
		Type:                utils.AlarmType,
		AlarmId:             NorthCertAbnormal,
		AlarmName:           "Cert Abnormal",
		PerceivedSeverity:   utils.MajorSeverity,
		DetailedInformation: "This alarm is generated when the Northbound certificate is about to expire or has expired.",
		Suggestion: "1. Check whether the certificate is about to expire or has expired." +
			"2. If the certificate has expired, import the Northbound certificate again." +
			"3. Contact Vendor technical support.",
		Reason: "The Northbound certificate is expired or about to expire.",
		Impact: "After the certificate has expired, the interconnection between MEF Center and Northbound will be affected.",
	},
	{
		Type:              utils.AlarmType,
		AlarmId:           SoftwareCertAbnormal,
		AlarmName:         "Cert Abnormal",
		PerceivedSeverity: utils.MajorSeverity,
		DetailedInformation: "This alarm is generated when the Software Repository certificate " +
			"is about to expire or has expired.",
		Suggestion: "1. Check whether the certificate is about to expire or has expired." +
			"2. If the certificate has expired, import the Software Repository certificate again." +
			"3. Contact Vendor technical support.",
		Reason: "The Software Repository certificate is expired or about to expire.",
		Impact: "After the certificate has expired, the interconnection between " +
			"MEF Center and Software Repository will be affected.",
	},
	{
		Type:              utils.AlarmType,
		AlarmId:           ImageCertAbnormal,
		AlarmName:         "Cert Abnormal",
		PerceivedSeverity: utils.MajorSeverity,
		DetailedInformation: "This alarm is generated when the Image Repository certificate " +
			"is about to expire or has expired.",
		Suggestion: "1. Check whether the certificate is about to expire or has expired." +
			"2. If the certificate has expired, import the Image Repository certificate again." +
			"3. Contact Vendor technical support.",
		Reason: "The Image Repository certificate is expired or about to expire.",
		Impact: "After the certificate has expired, the interconnection between " +
			"MEF Center and Image Repository will be affected.",
	},
}

var alarmList map[string]requests.AlarmReq

func init() {
	alarmList = make(map[string]requests.AlarmReq, len(allAlarms))
	for _, alm := range allAlarms {
		alarmList[alm.AlarmId] = alm
	}
}
