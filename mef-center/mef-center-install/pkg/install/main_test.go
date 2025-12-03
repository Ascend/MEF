// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
