// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
// Package alarmmanager for
package alarmmanager

import (
	"encoding/json"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

const firstAlarmIdx = 0
const testIp = "10.10.10.10"

var alarmRequestExample = requests.AlarmReq{
	Type:              "alarm",
	AlarmId:           "0x01000003",
	AlarmName:         "Image Repository Cert Abnormal",
	Resource:          "cert",
	PerceivedSeverity: "MAJOR",
	Timestamp:         "2023-12-25T11:49:32+08:00",
	NotificationType:  "alarm",
	DetailedInformation: "This alarm is generated when the Image Repository " +
		"certificate is about to expire or has expired.",
	Suggestion: "1. Check whether the certificate is about to expire or has expired.2. If the certificate" +
		" has expired, import the Image Repository certificate again.3. Contact Vendor technical support.",
	Reason: "The Image Repository certificate is expired or about to expire.",
	Impact: "After the certificate has expired, the interconnection between MEF Center and " +
		"Image Repository will be affected.",
}

func TestDealAlarmReq(t *testing.T) {
	alarmsReq := requests.AddAlarmReq{
		Alarms: []requests.AlarmReq{alarmRequestExample},
		Sn:     "sn",
		Ip:     testIp,
	}
	convey.Convey("Test DealAlarmReq", t, func() {
		convey.Convey("test add alarm", func() {
			bytes, err := json.Marshal(alarmsReq)
			convey.So(err, convey.ShouldBeNil)
			inter, err := dealAlarmReq(&model.Message{Content: bytes})
			convey.So(inter, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Test DealAlarmReq remove alarm", func() {
			alarmsReq.Alarms[firstAlarmIdx].NotificationType = "clear"
			bytes, err := json.Marshal(alarmsReq)
			convey.So(err, convey.ShouldBeNil)
			inter, err := dealAlarmReq(&model.Message{Content: bytes})
			convey.So(inter, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("test adding event", func() {
			alarmsReq.Alarms[firstAlarmIdx].Type = "event"
			bytes, err := json.Marshal(alarmsReq)
			convey.So(err, convey.ShouldBeNil)
			inter, err := dealAlarmReq(&model.Message{Content: bytes})
			convey.So(inter, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDealNodeClearReq(t *testing.T) {
	convey.Convey("Test DealNode Clear", t, func() {
		req := requests.ClearNodeAlarmReq{Sn: NoneExistSn}
		bytes, err := json.Marshal(req)
		convey.So(err, convey.ShouldBeNil)
		ret, err := dealNodeClearReq(&model.Message{Content: bytes})
		convey.So(ret, convey.ShouldEqual, common.OK)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDealEvent(t *testing.T) {
	alarmRequestExample.Type = alarms.EventType
	dealer := GetAlarmReqDealer(&alarmRequestExample, testSn2, testIp)
	err := dealer.dealEvent()

	patch := gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "getNodeEventCount",
		func(sn string) (int, error) {
			return maxOneNodeEventCount, nil
		}).
		ApplyPrivateMethod(&AlarmDbHandler{}, "deleteAlarmInfos",
			func(data []AlarmInfo) error {
				return nil
			})
	defer patch.Reset()
	convey.Convey("test DealEvent", t, func() {
		convey.So(err, convey.ShouldBeNil)
		err = dealer.dealEvent()
		convey.So(err, convey.ShouldBeNil)
	})
}
