// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
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
