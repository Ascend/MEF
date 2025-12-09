// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package appmanager

import (
	"context"
	"net/http"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

var (
	mockMsgChan = make(chan model.Message)
	ctx         context.Context
	cancelFunc  context.CancelFunc
)

func TestAppManager(t *testing.T) {
	convey.Convey("test app manager", t, testAppManager)
}

func testAppManager() {
	ctx, cancelFunc = context.WithCancel(context.Background())
	patches := gomonkey.ApplyPrivateMethod(&appStatusServiceImpl{}, "initAppStatusService", func() error { return nil }).
		ApplyFuncReturn(database.CreateTableIfNotExist, nil).
		ApplyFunc(modulemgr.ReceiveMessage, mockReceiveMsg).
		ApplyFunc(modulemgr.SendMessage, mockSendMsg)
	defer patches.Reset()

	appMgr := appManager{
		enable: true,
		ctx:    ctx,
	}
	modulemgr.ModuleInit()
	err := modulemgr.Registry(&appMgr)
	if err != nil {
		panic(err)
	}
	modulemgr.Start()
	convey.So(appMgr.Name(), convey.ShouldEqual, common.AppManagerName)
	convey.So(appMgr.Enable(), convey.ShouldBeTrue)

	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
	}
	msg := newMsgWithContentForUT(reqData)
	msg.SetRouter("test-module", common.AppManagerName, http.MethodGet, "/edgemanager/v1/app/list")
	mockMsgChan <- *msg
	<-ctx.Done()
}

func mockReceiveMsg(_ string) (*model.Message, error) {
	msg, ok := <-mockMsgChan
	if !ok {
		panic("mock channel is close")
	}
	return &msg, nil
}

func mockSendMsg(m *model.Message) error {
	if m.Router.Resource != "/edgemanager/v1/app/list" {
		panic("app manager sending wrong message when testing")
	}
	cancelFunc()
	return nil
}
