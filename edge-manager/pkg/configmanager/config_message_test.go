// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configmanager for inner message test
package configmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindxedge/base/common"
)

func TestGetAllNodeInfo(t *testing.T) {
	convey.Convey("get all node info functional test", t, func() {
		convey.Convey("get all node info success", func() {
			mockPath1 := gomonkey.ApplyFuncReturn(common.SendSyncMessageByRestful, common.RespMsg{
				Status: common.Success, Msg: "", Data: nil})
			defer mockPath1.Reset()
			_, err := getAllNodeInfo()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
