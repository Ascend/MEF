// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package almutils
package almutils

// all known alarms
const (
	DockerAbnormal        = "0x00131001"
	LivenessProbeAbnormal = "0x00131002"
	ApplyNPUAbnormal      = "0x00131003"
	ApplicationRestart    = "0x00131004"
	MsgParameterInvalid   = "0x00135001"
	OperationInvalid      = "0x00135002"
	EdgeLogAbnormal       = "0x00131011"
	NPUAbnormal           = "0x00131012"
	CertAbnormal          = "0x00131013"
	EdgeDBAbnormal        = "0x00131014"

	// MefAlarmIdPrefix is Alarm ID related to "mef" start with "0x00131"
	MefAlarmIdPrefix = "0x00131"
)

var allAlarms = []Alarm{

	{
		Type:              TypeAlarm,
		AlarmId:           DockerAbnormal,
		AlarmName:         "Docker Engine Abnormal",
		PerceivedSeverity: CRITICAL,
		DetailedInformation: "This alarm is generated when the Docker engine is not running properly. " +
			"This alarm is cleared when the Docker engine is running properly.",
		Suggestion: "1. Log in to the device and check whether the Docker engine is running properly." +
			"2. If the alarm cannot be cleared, please contact Huawei technical support.",
		Reason: "The Docker engine is not running properly.",
		Impact: "Containers cannot be deployed or run.",
	},

	{
		Type:              TypeAlarm,
		AlarmId:           LivenessProbeAbnormal,
		AlarmName:         "Application Liveness Probe Abnormal",
		PerceivedSeverity: MINOR,
		DetailedInformation: "This alarm is generated when the container liveness probe detects an exception, " +
			"for example, a container deadlock or exit. " +
			"This alarm is cleared when the liveness probe succeeds or the container is deleted.",
		Suggestion: "Check whether the probe configuration is correct and " +
			"locate the cause of the container application exception.",
		Reason: "The application in the container is abnormal.",
		Impact: "Containerized applications may fail to provide services.",
	},

	{
		Type:              TypeAlarm,
		AlarmId:           ApplyNPUAbnormal,
		AlarmName:         "Apply NPU Resource Failed",
		PerceivedSeverity: MAJOR,
		DetailedInformation: "This alarm is generated when NPU resources fail to be applied from a container. " +
			"This alarm is cleared when NPU resources are successfully applied or the container is deleted.",
		Suggestion: "Check whether the NPU resource is normal and try to deploy the container again.",
		Reason:     "The NPU is abnormal or the number of applied NPU resources exceeds the upper limit.",
		Impact:     "Containerized applications cannot use NPU resources.",
	},

	{
		Type:                TypeEvent,
		AlarmId:             ApplicationRestart,
		AlarmName:           "Application Restart",
		PerceivedSeverity:   OK,
		DetailedInformation: "This event is generated when a container restarts.",
		Suggestion:          "Locate the cause of the container application restart.",
		Reason: "An application in the container exits abnormally, " +
			"or the container is restarted for configuration change.",
		Impact: "Services provided by containerized applications may be interrupted.",
	},

	{
		Type:                TypeAlarm,
		AlarmId:             EdgeDBAbnormal,
		AlarmName:           "MEFEdge Database Abnormal",
		PerceivedSeverity:   MAJOR,
		DetailedInformation: "This alarm is generated when the MEFEdge detects its database is malformed.",
		Suggestion: "1. Log in to the device and check whether the database is malformed." +
			" 2. Contact Huawei technical support",
		Reason: "The MEFEdge database is malformed.",
		Impact: "The MEFEdge may work improperly.",
	},

	{
		Type:                TypeEvent,
		AlarmId:             MsgParameterInvalid,
		AlarmName:           "Invalid parameter",
		PerceivedSeverity:   OK,
		DetailedInformation: "Check the input parameters is invalid.",
		Suggestion:          "Please check the input parameters.",
		Reason:              "",
		Impact:              "",
	},

	{
		Type:                TypeEvent,
		AlarmId:             OperationInvalid,
		AlarmName:           "Operation Invalid",
		PerceivedSeverity:   OK,
		DetailedInformation: "This event is generated when the EdgeCore receives some invalid operation requests.",
		Suggestion: "Some invalid operations happened, eg:" +
			"operate too often. Check whether the operation is allowed.",
		Reason: "The received operation request is invalid.",
		Impact: "The operation request may not be handled.",
	},

	{
		Type:                TypeAlarm,
		AlarmId:             EdgeLogAbnormal,
		AlarmName:           "MEFEdge Log Abnormal",
		PerceivedSeverity:   MINOR,
		DetailedInformation: "This alarm is generated when the MEFEdge's log space is almost full.",
		Suggestion: "1. Check the directories of MEFEdge's log have sufficient storage space. " +
			"2. When data is backed up, files with insufficient directory space are processed to free up space. " +
			"3. Contact Huawei technical support",
		Reason: "1. The directories of MEFEdge's log have insufficient storage space.",
		Impact: "1. MEFEdge cannot write log. 2. EdgeCore cannot work properly.",
	},

	{
		Type:              TypeAlarm,
		AlarmId:           NPUAbnormal,
		AlarmName:         "NPU Abnormal",
		PerceivedSeverity: CRITICAL,
		DetailedInformation: "This alarm is generated when the npu chip health is not ok." +
			"This alarm is cleared when the npu chip health is ok properly.",
		Suggestion: "1. Log in to the device and check whether the npu chip health is ok properly." +
			"2. Please contact Huawei technical support and provide corresponding information.",
		Reason: "The npu chip health is not ok properly.",
		Impact: "Inference containers cannot be deployed or run.",
	},

	{
		Type:                TypeAlarm,
		AlarmId:             CertAbnormal,
		AlarmName:           "Cert Abnormal",
		PerceivedSeverity:   MAJOR,
		DetailedInformation: "This alarm is generated when the MEF Center certificate is about to expire or has expired.",
		Suggestion: "1. Check whether the certificate is about to expire or has expired." +
			"2. If the certificate has expired, set the net configuration again." +
			"3. Contact Vendor technical support.",
		Reason: "The MEF Center certificate is expired or about to expire.",
		Impact: "After the certificate has expired, the interconnection between MEF Center and Edge will be affected.",
	},
}

var idToAlarms map[string]Alarm

func init() {
	idToAlarms = make(map[string]Alarm, len(allAlarms))
	for _, alm := range allAlarms {
		idToAlarms[alm.AlarmId] = alm
	}
}
