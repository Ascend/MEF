// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package nodemanager for node init test
package nodemanager

import (
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr/model"
)

func TestDispatchMsg(t *testing.T) {
	convey.Convey("selectMethod functional test", t, func() {
		convey.Convey("innerGetNodeInfoByUniqueName success", func() {
			input := model.Message{}
			_, err := selectMethod(&input)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("innerGetNodeInfoByUniqueName param error", func() {
			input := model.Message{}
			input.SetRouter("", "", http.MethodGet, nodeUrlRootPath)
			_, err := selectMethod(&input)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
