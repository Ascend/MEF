// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package config test for capability_cache.go
package config

import (
	"context"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

func TestCapabilityCache(t *testing.T) {
	convey.Convey("test CapabilityCache method SetEdgeOmCaps", t, func() {
		info := StaticInfo{
			ProductCapabilityEdge: []string{constants.CapabilityNpuSharingConfig, constants.CapabilityNpuSharing},
		}
		GetCapabilityCache().SetEdgeOmCaps(info)
		// clear the capabilities to avoid affecting subsequent test cases
		GetCapabilityCache().capabilities = sync.Map{}
	})

	convey.Convey("test CapabilityCache method Notify / StartReportJob", t, func() {
		GetCapabilityCache().Notify()
		go GetCapabilityCache().StartReportJob(context.Background())
	})

	convey.Convey("test func reportCapabilities", t, func() {
		// no capability
		GetCapabilityCache().capabilities = sync.Map{}
		reportCapabilities()

		GetCapabilityCache().Set(constants.CapabilityNpuSharing, true)

		// report success
		var p1 = gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
		defer p1.Reset()
		reportCapabilities()

		// send message failed
		var p2 = gomonkey.ApplyFuncReturn(modulemgr.SendMessage, test.ErrTest)
		defer p2.Reset()
		reportCapabilities()

		// fill content failed
		var p3 = gomonkey.ApplyMethodReturn(&model.Message{}, "FillContent", test.ErrTest)
		defer p3.Reset()
		reportCapabilities()

		// new message failed
		var p4 = gomonkey.ApplyFuncReturn(model.NewMessage, nil, test.ErrTest)
		defer p4.Reset()
		reportCapabilities()

	})
}
