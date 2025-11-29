// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr for testing init
package handlermgr

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

var handlerManager = handlerManger{}

func TestHandlerManger(t *testing.T) {
	convey.Convey("test handler manager [Name] method", t, func() {
		convey.So(NewHandlerMgrModule(true).Name(), convey.ShouldEqual, constants.ModEdgeOm)
	})

	convey.Convey("test handler manager [Enable] method", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
			ApplyFuncReturn(fileutils.CreateDir, nil).
			ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
			ApplyFuncReturn(fileutils.SetPathPermission, nil).
			ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, nil)
		defer p.Reset()

		convey.Convey("enable success", func() {
			p1 := gomonkey.ApplyFuncReturn(config.LoadPodConfig, &config.PodConfig{}, nil)
			defer p1.Reset()
			convey.So(NewHandlerMgrModule(true).Enable(), convey.ShouldBeTrue)
		})

		convey.Convey("enable failed", func() {
			p1 := gomonkey.ApplyFuncReturn(config.LoadPodConfig, &config.PodConfig{}, test.ErrTest)
			defer p1.Reset()
			convey.So(NewHandlerMgrModule(true).Enable(), convey.ShouldBeFalse)
		})
	})

	convey.Convey("test handler manager [Stop] method", t, func() {
		handlerManager.ctx, handlerManager.cancel = context.WithCancel(context.Background())
		convey.So(handlerManager.Stop(), convey.ShouldBeTrue)
	})

	convey.Convey("test handler manager [dispatchMsg] method", t, func() {
		handlerManager.dispatchMsg(&model.Message{})
	})
}
