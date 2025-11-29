// Copyright (c) 2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package downloadmgr

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

func moduleInit() {
	modulemgr.ModuleInit()
	modules := []model.Module{
		NewDownloadMgr(true),
	}
	for _, mod := range modules {
		if err := modulemgr.Registry(mod); err != nil {
			panic(err)
		}
	}
}

func TestStartDownloadMrg(t *testing.T) {
	convey.Convey("test start download mrg case", t, func() {
		moduleInit()
		mgr := NewDownloadMgr(true)
		go mgr.Start()
		convey.So(mgr.Name(), convey.ShouldResemble, constants.DownloadManagerName)
		convey.So(mgr.Enable(), convey.ShouldResemble, true)
	})
	convey.Convey("test process case", t, func() {
		mgr := &downloadMgr{}
		p1 := gomonkey.ApplyPrivateMethod(mgr, "processDownloadSoftware", func(msg model.Message) error {
			return nil
		})
		defer p1.Reset()
		mgr.process(model.Message{})
	})
}
