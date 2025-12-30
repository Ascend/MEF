// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
