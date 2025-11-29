// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing check environment base task
package tasks

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"
)

var checkTask = CheckEnvironmentBaseTask{}

func TestCheckEnvironmentBaseTask(t *testing.T) {
	convey.Convey("check necessary tool success", t, checkNecessaryToolsSuccess)
	convey.Convey("check necessary tools failed", t, checkNecessaryToolsFailed)
	convey.Convey("check recommend tools failed", t, checkRecommendToolsFailed)
}

func checkNecessaryToolsSuccess() {
	p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil)
	defer p.Reset()
	err := checkTask.CheckNecessaryTools()
	convey.So(err, convey.ShouldBeNil)
}

func checkNecessaryToolsFailed() {
	p := gomonkey.ApplyFuncReturn(exec.LookPath, "", test.ErrTest)
	defer p.Reset()
	err := checkTask.CheckNecessaryTools()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("look path of [%s] failed, error: %v",
		necessaryTools[0], test.ErrTest))
}

func checkRecommendToolsFailed() {
	p := gomonkey.ApplyFuncSeq(exec.LookPath, []gomonkey.OutputCell{
		{Values: gomonkey.Params{"", nil}, Times: 6},
		{Values: gomonkey.Params{"", test.ErrTest}},
	})
	defer p.Reset()
	err := checkTask.CheckNecessaryTools()
	convey.So(err, convey.ShouldBeNil)
}
