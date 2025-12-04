// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package websocket test for init_client.go
package websocket

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"alarm-manager/pkg/alarmmanager"
)

func TestInitClient(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(certutils.GetTlsCfgWithPath, nil, nil).
		ApplyMethodReturn(&websocketmgr.WsClientProxy{}, "Start", nil)
	defer patches.Reset()
	convey.Convey("test func initClient success", t, testInitClient)
	convey.Convey("test func initClient failed, init proxy failed", t, testInitClientErrInitProxy)
	convey.Convey("test func initClient failed, set rps limiter config failed", t, testInitClientErrSetLimiterCfg)
	convey.Convey("test func initClient failed, proxy start failed", t, testInitClientErrStart)
}

func testInitClient() {
	err := initClient()
	convey.So(err, convey.ShouldBeNil)
}

func testInitClientErrInitProxy() {
	var p1 = gomonkey.ApplyFuncReturn(certutils.GetTlsCfgWithPath, nil, test.ErrTest)
	defer p1.Reset()
	err := initClient()
	expErr := errors.New("init proxy config failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testInitClientErrSetLimiterCfg() {
	var p1 = gomonkey.ApplyMethodReturn(&websocketmgr.ProxyConfig{}, "SetRpsLimiterCfg", test.ErrTest)
	defer p1.Reset()
	err := initClient()
	expErr := fmt.Errorf("init websocket rps limiter config failed: %v", test.ErrTest)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testInitClientErrStart() {
	var p1 = gomonkey.ApplyMethodReturn(&websocketmgr.WsClientProxy{}, "Start", test.ErrTest)
	defer p1.Reset()
	err := initClient()
	expErr := errors.New("init alarm-manager client failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func TestClearAllAlarms(t *testing.T) {
	convey.Convey("test func clearAllAlarms success", t, testClearAllAlarms)
	convey.Convey("test func clearAllAlarms failed, delete from db failed", t, testClearAllAlarmsErrDelete)
}

func testClearAllAlarms() {
	var p1 = gomonkey.ApplyMethodReturn(&alarmmanager.AlarmDbHandler{}, "DeleteEdgeAlarm", nil)
	defer p1.Reset()
	clearAllAlarms(websocketmgr.WebsocketPeerInfo{})
}

func testClearAllAlarmsErrDelete() {
	var p1 = gomonkey.ApplyMethodReturn(&alarmmanager.AlarmDbHandler{}, "DeleteEdgeAlarm", test.ErrTest)
	defer p1.Reset()
	clearAllAlarms(websocketmgr.WebsocketPeerInfo{})
}
