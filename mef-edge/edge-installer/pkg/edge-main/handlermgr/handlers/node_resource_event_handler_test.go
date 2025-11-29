// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlers for testing node resource event handler
package handlers

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

var (
	nodeResourceHandler = nodeResourceEventHandler{}
	nodeStr             = `{
          "metadata":  {
                    "name":  "2bf5b5ba-d71e-4d16-b00e-6531cc358bfc"
          },
          "status":  {
                    "capacity":  {
                              "cpu":  "4",
                              "huawei.com/Ascend310":  "100",
                              "memory":  "11856388Ki",
                              "pods":  "110"
                    }
          }
}`
)

func TestNodeResourceEventHandler(t *testing.T) {
	convey.Convey("test node resource event handler should be success", t, nodeResourceEventHandlerSuccess)
	convey.Convey("test node resource event handler should be failed", t, nodeResourceEventHandlerFailed)
}

func nodeResourceEventHandlerSuccess() {
	p := gomonkey.ApplyFuncReturn(util.SendInnerMsgResponse, nil).
		ApplyMethodReturn(statusmanager.GetNodeStatusMgr(), "GetAll",
			map[string]string{constants.ResourceTypeNode: nodeStr}, nil)
	defer p.Reset()
	err := nodeResourceHandler.Handle(&model.Message{})
	convey.So(err, convey.ShouldBeNil)
}

func nodeResourceEventHandlerFailed() {
	p := gomonkey.ApplyFuncReturn(util.SendInnerMsgResponse, test.ErrTest).
		ApplyMethodReturn(statusmanager.GetNodeStatusMgr(), "GetAll",
			map[string]string{"test1": nodeStr, "test2": nodeStr}, nil)
	defer p.Reset()
	err := nodeResourceHandler.Handle(&model.Message{})
	convey.So(err, convey.ShouldResemble, errors.New("exactly one node expected"))
}
