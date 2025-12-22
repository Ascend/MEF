// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package almutils test for alarm
package almutils

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func TestCreateAndSendAlarm(t *testing.T) {
	convey.Convey("create and send alarm should success", t, testCreateAndSendAlarm)
	convey.Convey("send alarm should failed", t, testCreateAlarmErr)
	convey.Convey("send alarm should failed", t, testSendAlarmErr)
}

func testCreateAndSendAlarm() {
	var p1 = gomonkey.ApplyFunc(modulemgr.SendAsyncMessage,
		func(m *model.Message) error {
			return nil
		})
	defer p1.Reset()
	err := CreateAndSendAlarm(DockerAbnormal, "", "", "", "")
	convey.So(err, convey.ShouldResemble, nil)
}

func testCreateAlarmErr() {
	var p1 = gomonkey.ApplyFunc(CreateAlarm,
		func(alarmId, resource, notifyType string) (*Alarm, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()
	err := CreateAndSendAlarm(DockerAbnormal, "", "", "", "")
	convey.So(err, convey.ShouldEqual, test.ErrTest)
}

func testSendAlarmErr() {
	var p1 = gomonkey.ApplyFunc(SendAlarm,
		func(source, destination string, alarm ...*Alarm) error {
			return test.ErrTest
		})
	defer p1.Reset()
	err := CreateAndSendAlarm(DockerAbnormal, "", "", "", "")
	convey.So(err, convey.ShouldEqual, test.ErrTest)
}
