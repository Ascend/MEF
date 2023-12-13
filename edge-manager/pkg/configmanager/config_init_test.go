// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configmanager for config init test
package configmanager

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"
)

func TestMethodSelect(t *testing.T) {
	convey.Convey("method select functional test", t, func() {
		convey.Convey("config manager method select failed without url", func() {
			input, _ := model.NewMessage()
			msg := methodSelect(input)
			convey.So(msg, convey.ShouldBeNil)
		})
		convey.Convey("config manager method select failed with root url", func() {
			input, _ := model.NewMessage()
			input.SetRouter("", "", http.MethodPost, configUrlRootPath)
			msg := methodSelect(input)
			convey.So(msg, convey.ShouldBeNil)
		})
		convey.Convey("config manager method select success with image config url", func() {
			input, _ := model.NewMessage()
			input.SetRouter("", "", http.MethodPost, filepath.Join(configUrlRootPath, "config"))
			msg := methodSelect(input)
			convey.So(msg, convey.ShouldNotBeNil)
		})
		convey.Convey("config manager method select success with update url", func() {
			input, _ := model.NewMessage()
			input.SetRouter("", "", http.MethodPost, filepath.Join(innerConfigUrlRootPath, "update"))
			msg := methodSelect(input)
			convey.So(msg, convey.ShouldNotBeNil)
		})
	})
}
