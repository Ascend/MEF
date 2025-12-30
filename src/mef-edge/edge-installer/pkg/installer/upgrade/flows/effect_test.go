// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package flows for testing effect flow
package flows

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/upgrade/tasks"
)

var (
	testEffectDir = "/tmp/test_effect_flow_dir"
	flowEffect    = NewEffectFlow(pathmgr.NewPathMgr(testEffectDir, testEffectDir, testEffectDir, testEffectDir))
)

func TestEffectFlow(t *testing.T) {
	convey.Convey("effect flow should be success", t, effectFlowSuccess)
	convey.Convey("effect flow should be failed", t, effectFlowFailed)
}

func effectFlowSuccess() {
	p := gomonkey.ApplyMethodReturn(&tasks.PostEffectProcessTask{}, "Run", nil)
	defer p.Reset()
	err := flowEffect.RunTasks()
	convey.So(err, convey.ShouldBeNil)
}

func effectFlowFailed() {
	p := gomonkey.ApplyMethodReturn(&tasks.PostEffectProcessTask{}, "Run", test.ErrTest)
	defer p.Reset()
	err := flowEffect.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("upgrade post process task failed"))
}
