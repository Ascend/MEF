// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers
package handlers

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

// TestSendRestfulResponse tests sendRestfulResponse
func TestSendRestfulResponse(t *testing.T) {
	convey.Convey("test send restful response", t, func() {
		patch := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
		defer patch.Reset()
		err := sendRestfulResponse(common.RespMsg{Status: common.Success}, &model.Message{})
		convey.So(err, convey.ShouldBeNil)
	})
}
