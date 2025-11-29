// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarms, common definition and function for alarm operation
package alarms

import (
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common/requests"
)

var allAlarms = []requests.AlarmReq{
	{
		Type:              AlarmType,
		AlarmId:           NorthCertAbnormal,
		AlarmName:         "Third Management Platform Cert Abnormal",
		PerceivedSeverity: MajorSeverity,
		DetailedInformation: "This alarm is generated when the Third management platform certificate is about to " +
			"expire or has expired.",
		Suggestion: "1. Check whether the certificate is about to expire or has expired." +
			"2. If the certificate has expired, import the Third management platform certificate again." +
			"3. Contact Vendor technical support.",
		Reason: "The Third management platform certificate is expired or about to expire.",
		Impact: "After the certificate has expired, the interconnection between MEF Center and Third management " +
			"platform will be affected.",
	},
	{
		Type:              AlarmType,
		AlarmId:           SoftwareCertAbnormal,
		AlarmName:         "Software Repository Cert Abnormal",
		PerceivedSeverity: MajorSeverity,
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
		Type:              AlarmType,
		AlarmId:           ImageCertAbnormal,
		AlarmName:         "Image Repository Cert Abnormal",
		PerceivedSeverity: MajorSeverity,
		DetailedInformation: "This alarm is generated when the Image Repository certificate " +
			"is about to expire or has expired.",
		Suggestion: "1. Check whether the certificate is about to expire or has expired." +
			"2. If the certificate has expired, import the Image Repository certificate again." +
			"3. Contact Vendor technical support.",
		Reason: "The Image Repository certificate is expired or about to expire.",
		Impact: "After the certificate has expired, the interconnection between " +
			"MEF Center and Image Repository will be affected.",
	},
	{
		Type:              AlarmType,
		AlarmId:           MEFCenterCaCertAbnormal,
		AlarmName:         "MEF Center Ca Cert Need Update",
		PerceivedSeverity: MinorSeverity,
		DetailedInformation: "This alarm is generated when MEF south root ca cert will be expired and auto " +
			"update process begins. South root ca will be updated, edge new service certs will be issued " +
			"by this new south root ca cert.",
		Suggestion: "1. Keep the network stable between MEF Center and Edge, if edge nodes connect with" +
			" Center then disconnect it, the cert update process may take a long duration",
		Reason: "MEF south root cert will be expired in 15 days",
		Impact: "MEF Edge nodes will be unable to connect with MEF Center if the cert is expired",
	},
	{
		Type:              AlarmType,
		AlarmId:           MEFCenterSvcCertAbnormal,
		AlarmName:         "MEF Center Service Cert Need Update",
		PerceivedSeverity: MinorSeverity,
		DetailedInformation: "This alarm is generated when the root ca for issuing MEF south service cert will be " +
			"expired and auto update process begins. New root ca will be generated, issue new south service" +
			"cert, then sent to all edge nodes.",
		Suggestion: "1. Keep the network stable between MEF Center and Edge, if edge nodes connect with" +
			" Center then disconnect it, the cert update process may take a long duration",
		Reason: "the root ca cert for issuing MEF south service cert will be expired in 15 days",
		Impact: "MEF Edge nodes will be unable to connect with MEF Center if the cert is expired",
	},
	{
		Type:              EventType,
		AlarmId:           MEFCenterCaCertUpdateAbnormal,
		AlarmName:         "MEF Center Ca Cert Update Abnormal",
		PerceivedSeverity: CriticalSeverity,
		DetailedInformation: "When MEF south ca cert auto update process is finished, " +
			"if any Edge node not successfully updates it's service cert, this event will be generated",
		Suggestion: "1. Export hub_svr root cert from MEF Center." +
			"2. Get the token for authentication with MEF Center." +
			"3. Do net-config operation on MEF Edge again.",
		Reason: "Unstable network or other reasons interfere the update process",
		Impact: "MEF Edge will not be able to connect MEF Center if current websocket connection is lost",
	},
	{
		Type:              EventType,
		AlarmId:           MEFCenterSvcCertUpdateAbnormal,
		AlarmName:         "MEF Center Service Cert Update Abnormal",
		PerceivedSeverity: CriticalSeverity,
		DetailedInformation: "When root ca cert for issuing MEF south service cert auto update process is finished, " +
			"if any Edge node not successfully updates it's root ca cert, this event will be generated",
		Suggestion: "1. Export hub_svr root cert from MEF Center." +
			"2. Get the token for authentication with MEF Center." +
			"3. Do net-config operation on MEF Edge again.",
		Reason: "Unstable network or other reasons interfere the update process",
		Impact: "MEF Edge will not be able to connect MEF Center if current websocket connection is lost",
	},
}

var alarmList map[string]requests.AlarmReq

func init() {
	alarmList = make(map[string]requests.AlarmReq, len(allAlarms))
	for _, alm := range allAlarms {
		alarmList[alm.AlarmId] = alm
	}
}

// CreateAlarm creates an alarm
func CreateAlarm(alarmId, resource, notifyType string) (*requests.AlarmReq, error) {
	template, ok := alarmList[alarmId]
	if !ok {
		hwlog.RunLog.Errorf("unknown alarm type, alarm id: [%s]", alarmId)
		return nil, errors.New("unknown alarm type")
	}

	return &requests.AlarmReq{
		Type:                template.Type,
		AlarmId:             alarmId,
		AlarmName:           template.AlarmName,
		Resource:            resource,
		PerceivedSeverity:   template.PerceivedSeverity,
		Timestamp:           time.Now().Format(time.RFC3339),
		NotificationType:    notifyType,
		DetailedInformation: template.DetailedInformation,
		Suggestion:          template.Suggestion,
		Reason:              template.Reason,
		Impact:              template.Impact,
	}, nil
}
