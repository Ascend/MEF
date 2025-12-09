// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package components for testing prepare device plugin
package components

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

func TestPrepareDevicePluginRun(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(fileutils.SetPathOwnerGroup, nil)
	defer p.Reset()
	convey.Convey("prepare device plugin run should be success", t, func() {
		err := NewPrepareDevicePlugin(pathMgr, workAbsPathMgr).Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("prepare device plugin run should be failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CopyDir, test.ErrTest)
		defer p1.Reset()
		err := NewPrepareDevicePlugin(pathMgr, workAbsPathMgr).Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("copy %s software dir failed, error: %v",
			constants.DevicePlugin, test.ErrTest))
	})
}
