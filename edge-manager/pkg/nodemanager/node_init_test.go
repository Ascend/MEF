// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package nodemanager for node init test
package nodemanager

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func TestDispatchMsg(t *testing.T) {
	Convey("dispatchMsg functional test", t, func() {
		Convey("innerGetNodeInfoByUniqueName success", func() {
			input := model.Message{}
			_, err := dispatchMsg(&input)
			So(err, ShouldNotBeNil)
		})
		Convey("innerGetNodeInfoByUniqueName param error", func() {
			input := model.Message{}
			input.SetRouter("", "", http.MethodGet, nodeUrlRootPath)
			_, err := dispatchMsg(&input)
			So(err, ShouldBeNil)
		})
	})
}
