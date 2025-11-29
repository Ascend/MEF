// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package install

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

var ErrTest = errors.New("test error")

func TestMain(m *testing.M) {
	patches := gomonkey.ApplyFuncReturn(util.CheckNecessaryCommands, nil)
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, patches)
}

func ResetAndClearDir(p *gomonkey.Patches, clearPath string) {
	if p != nil {
		p.Reset()
	}
	err := fileutils.DeleteAllFileWithConfusion(clearPath)
	convey.So(err, convey.ShouldBeNil)
}
