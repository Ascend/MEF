// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers
package handlers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/logmanager/tasks"
	"edge-manager/pkg/logmanager/testutils"
	"edge-manager/pkg/types"
)

// TestMain sets up the hwlog
func TestMain(m *testing.M) {
	if err := testutils.PrepareHwlog(); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
	}
	m.Run()
}

// TestGetNodeSerialNumbersByID tests getNodeSerialNumbersByID func
func TestGetNodeSerialNumbersByID(t *testing.T) {
	convey.Convey("test getNodeSerialNumbersByID", t, func() {
		resp := common.RespMsg{
			Status: common.Success,
			Data:   types.InnerGetNodeInfosResp{NodeInfos: []types.NodeInfo{{SerialNumber: "123"}}},
		}
		patch := gomonkey.ApplyFuncReturn(common.SendSyncMessageByRestful, resp)
		defer patch.Reset()

		serialNumbers, err := getNodeSerialNumbersByID(nil)
		convey.So(err, convey.ShouldBeNil)
		convey.So(serialNumbers, convey.ShouldResemble, []string{"123"})
	})
}

// TestCreateTaskHandle tests the createTaskHandler
func TestCreateTaskHandle(t *testing.T) {
	convey.Convey("test createTaskHandler.Handle's argument check", t, func() {
		patch := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
		defer patch.Reset()
		var handler createTaskHandler
		err := handler.Handle(&model.Message{Content: `{}`})
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("test createTaskHandler.Handle's get serial numbers", t, func() {
		patch := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil).
			ApplyFuncReturn(getNodeSerialNumbersByID, nil, errors.New("get serial number failed"))
		defer patch.Reset()
		var handler createTaskHandler
		err := handler.Handle(&model.Message{Content: `{"module":"edgeNode", "edgeNodes": [1,2]}`})
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("test createTaskHandler.Handle's submit task", t, func() {
		patch := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil).
			ApplyFuncReturn(getNodeSerialNumbersByID, []string{"1", "2"}, nil).
			ApplyFuncReturn(tasks.SubmitLogDumpTask, "", errors.New("submit task failed"))
		defer patch.Reset()
		var handler createTaskHandler
		err := handler.Handle(&model.Message{Content: `{"module":"edgeNode", "edgeNodes": [1,2]}`})
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("test createTaskHandler.Handle", t, func() {
		patch := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil).
			ApplyFuncReturn(getNodeSerialNumbersByID, []string{"1", "2"}, nil).
			ApplyFuncReturn(tasks.SubmitLogDumpTask, "", nil)
		defer patch.Reset()
		var handler createTaskHandler
		err := handler.Handle(&model.Message{Content: `{"module":"edgeNode", "edgeNodes": [1,2]}`})
		convey.So(err, convey.ShouldBeNil)
	})
}
