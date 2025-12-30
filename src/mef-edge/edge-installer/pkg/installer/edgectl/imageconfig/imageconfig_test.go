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

// Package imageconfig
package imageconfig

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

var (
	testAddress = "fd.fusion1.huawei.com"
	oldAddress  = "fd.fusion2.huawei.com"
	testErr     = errors.New("test error")
	expectErr   = errors.New("import image config failed")
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func TestImageCfgFlow(t *testing.T) {
	convey.Convey("test image config successful", t, imageCfgSuccess)
	convey.Convey("test image config failed", t, func() {
		convey.Convey("get image config failed", getImageConfigFailed)
		convey.Convey("create image config failed", createImageCfgFailed)
		convey.Convey("delete image config failed", DeleteImageCfgFailed)
	})
}

func imageCfgSuccess() {
	convey.Convey("image config does not exist, create image config success", func() {
		p := gomonkey.ApplyFuncReturn(config.GetImageCfg, nil, gorm.ErrRecordNotFound).
			ApplyFuncReturn(config.SetImageCfg, nil)
		defer p.Reset()

		imageCfg := NewImageCfgFlow(testAddress)
		err := imageCfg.RunTasks()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("image config exists, and old image config is same, no need to clear", func() {
		p := gomonkey.ApplyFuncReturn(config.GetImageCfg, &config.ImageConfig{ImageAddress: testAddress}, nil)
		defer p.Reset()

		imageCfg := NewImageCfgFlow(testAddress)
		err := imageCfg.RunTasks()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("image config exists, no need to delete old image file, update image config success", func() {
		p := gomonkey.ApplyFuncReturn(config.GetImageCfg, &config.ImageConfig{ImageAddress: oldAddress}, nil).
			ApplyFuncReturn(config.SetImageCfg, nil)
		defer p.Reset()
		imageCfg := NewImageCfgFlow(testAddress)
		err := imageCfg.RunTasks()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("image config exists, delete old image file success, update image config success", func() {
		imageLinkFile := filepath.Join(constants.DockerCertDir, oldAddress, constants.LinkCert)
		if err := fileutils.MakeSureDir(imageLinkFile); err != nil {
			hwlog.RunLog.Errorf("make sure image link dir failed, error: %v", err)
			return
		}
		if err := fileutils.CreateFile(imageLinkFile, constants.Mode600); err != nil {
			hwlog.RunLog.Errorf("create image link file failed, error: %v", err)
			return
		}
		p := gomonkey.ApplyFuncReturn(config.GetImageCfg, &config.ImageConfig{ImageAddress: oldAddress}, nil).
			ApplyFuncReturn(config.SetImageCfg, nil)
		defer p.Reset()
		imageCfg := NewImageCfgFlow(testAddress)
		err := imageCfg.RunTasks()
		convey.So(err, convey.ShouldBeNil)
	})
}

func getImageConfigFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetImageCfg, nil, testErr)
	defer p.Reset()

	imageCfg := NewImageCfgFlow(testAddress)
	err := imageCfg.RunTasks()
	convey.So(err, convey.ShouldResemble, expectErr)
}

func createImageCfgFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetImageCfg, nil, gorm.ErrRecordNotFound).
		ApplyFuncReturn(config.SetImageCfg, testErr)
	defer p.Reset()

	imageCfg := NewImageCfgFlow(testAddress)
	err := imageCfg.RunTasks()
	convey.So(err, convey.ShouldResemble, expectErr)
}

func DeleteImageCfgFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetImageCfg, &config.ImageConfig{ImageAddress: oldAddress}, nil).
		ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, testErr)
	defer p.Reset()

	imageCfg := NewImageCfgFlow(testAddress)
	err := imageCfg.RunTasks()
	convey.So(err, convey.ShouldResemble, expectErr)
}
