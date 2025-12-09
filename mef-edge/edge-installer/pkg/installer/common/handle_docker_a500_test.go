// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package common for testing handle docker
package common

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
)

func TestChangeDocker(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(config.CheckIsA500, true).
		ApplyFuncReturn(fileutils.EvalSymlinks, "", nil)
	defer p.Reset()
	convey.Convey("change docker should be success", t, changeDockerSuccess)
	convey.Convey("change docker should be failed", t, changeDockerFailed)
}

func changeDockerSuccess() {
	convey.Convey("no need change docker", func() {
		p1 := gomonkey.ApplyFuncReturn(config.CheckIsA500, false)
		defer p1.Reset()
		err := componentMgr.changeDocker()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("change docker success", func() {
		p1 := gomonkey.ApplyFuncReturn(envutils.RunCommand, "", nil)
		defer p1.Reset()
		err := componentMgr.changeDocker()
		convey.So(err, convey.ShouldBeNil)
	})
}

func changeDockerFailed() {
	convey.Convey("eval symlink failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()
		err := componentMgr.changeDocker()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("change docker failed", func() {
		p1 := gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		err := componentMgr.changeDocker()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
