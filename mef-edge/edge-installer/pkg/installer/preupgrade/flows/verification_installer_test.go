// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package flows for testing verification installer
package flows

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

func TestVerificationInstaller(t *testing.T) {
	convey.Convey("test get verification flow", t, func() {
		verificationFlow := NewVerificationInstaller(testDir)
		convey.So(verificationFlow, convey.ShouldNotBeNil)
	})

	convey.Convey("test clear download path", t, func() {
		verification := verificationInstaller{downloadPath: constants.EdgeDownloadPath}
		convey.Convey("clear download path success", func() {
			p := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, nil)
			defer p.Reset()
			verification.clearDownloadPath()
		})

		convey.Convey("clear download path failed", func() {
			p := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
			defer p.Reset()
			verification.clearDownloadPath()
		})
	})
}
