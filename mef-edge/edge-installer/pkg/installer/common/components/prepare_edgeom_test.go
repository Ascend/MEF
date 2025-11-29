// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package components for testing prepare edge om
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

var prepareEdgeOmTask = NewPrepareEdgeOm(pathMgr, workAbsPathMgr)

func TestPrepareEdgeOmPrepareCfgDir(t *testing.T) {
	convey.Convey("prepare edge om config should be success", t, func() {
		err := prepareEdgeOmTask.PrepareCfgDir()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPrepareEdgeOmRun(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(fileutils.SetPathOwnerGroup, nil)
	defer p.Reset()
	convey.Convey("prepare edge om run should be success", t, func() {
		err := prepareEdgeOmTask.Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("prepare edge om run should be failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CopyDir, test.ErrTest)
		defer p1.Reset()
		err := prepareEdgeOmTask.Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("copy %s software dir failed, error: %v",
			constants.EdgeOm, test.ErrTest))
	})
}
