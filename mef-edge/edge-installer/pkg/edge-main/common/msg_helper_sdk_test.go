// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package common for test msg helper sdk
package common

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

func TestMEFOpLog(t *testing.T) {
	msg := getTestMsg()
	if msg == nil {
		return
	}
	convey.Convey("test MEFOpLog, ignore mef operation log", t, func() {
		msg.KubeEdgeRouter.Operation = constants.OptQuery
		MEFOpLog(msg)

		msg.KubeEdgeRouter.Operation = constants.OptResponse
		MEFOpLog(msg)
	})
	convey.Convey("test MEFOpLog, record mef operation log", t, func() {
		msg.KubeEdgeRouter.Operation = constants.OptGet
		MEFOpLog(msg)
	})
}

func TestMEFOpLogWithRes(t *testing.T) {
	convey.Convey("test MEFOpLogWithRes, ignore mef operation result log", t, ignoreMEFOpResLog)
	convey.Convey("test MEFOpLogWithRes, record mef operation result log success", t, recordMEFOpResLogSuccess)
}

func ignoreMEFOpResLog() {
	resp := &model.Message{}
	convey.Convey("resp.Header.ParentID is empty", func() {
		MEFOpLogWithRes(resp)
	})

	convey.Convey("resp.KubeEdgeRouter.Operation is not response", func() {
		resp.Header.ParentID = "12345"
		resp.KubeEdgeRouter.Operation = constants.OptReport
		MEFOpLogWithRes(resp)
	})

	convey.Convey("original operation is query", func() {
		msg := getTestMsg()
		if msg == nil {
			return
		}
		msg.KubeEdgeRouter.Operation = constants.OptQuery
		MEFOpLog(msg)
		resp.Header.ParentID = msg.Header.ID
		MEFOpLogWithRes(resp)
	})
}

func recordMEFOpResLogSuccess() {
	msg := getTestMsg()
	if msg == nil {
		return
	}
	resp := &model.Message{
		KubeEdgeRouter: model.MessageRoute{
			Operation: constants.OptResponse,
			Resource:  "/test/resource",
		},
	}
	resp.Header.ParentID = msg.Header.ID

	convey.Convey("record ok result", func() {
		resp.Content = model.RawMessage("OK")
		MEFOpLog(msg)
		MEFOpLogWithRes(resp)
	})

	convey.Convey("record failed result", func() {
		resp.Content = model.RawMessage("Failed")
		MEFOpLog(msg)
		MEFOpLogWithRes(resp)
	})
}
