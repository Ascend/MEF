// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package websocket test for init.go
package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/websocketmgr"

	"huawei.com/mindxedge/base/common"
)

var (
	wsClient *alarmWsClient
	ctx      context.Context
	cancel   context.CancelFunc
)

func TestAlarmManager(t *testing.T) {
	ctx, cancel = context.WithCancel(context.Background())
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(NewAlarmWsClient(true, ctx)); err != nil {
		panic(err)
	}
	wsClient = &alarmWsClient{
		enable: true,
		ctx:    ctx,
	}
	convey.Convey("test alarmWsClient method 'NewAlarmWsClient', 'Name', 'Enable'", t, testAlarmManager)
	convey.Convey("test alarmWsClient method 'Start'", t, testAlarmMgrStart)
}

func testAlarmManager() {
	if wsClient == nil {
		panic("alarm websocket client is nil")
	}
	convey.So(wsClient.Name(), convey.ShouldEqual, common.AlarmManagerWsMoudle)
	convey.So(wsClient.Enable(), convey.ShouldBeTrue)
}

func testAlarmMgrStart() {
	msg, err := model.NewMessage()
	convey.So(err, convey.ShouldBeNil)
	var p1 = gomonkey.ApplyFuncReturn(initClient, nil).
		ApplyFuncReturn(modulemgr.ReceiveMessage, msg, nil).
		ApplyMethodReturn(&websocketmgr.WsClientProxy{}, "Send", nil)
	defer p1.Reset()
	if wsClient == nil {
		panic("alarm websocket client is nil")
	}
	go wsClient.Start()
	const sleepTime = 200 * time.Millisecond
	time.Sleep(sleepTime)
	cancel()
}
