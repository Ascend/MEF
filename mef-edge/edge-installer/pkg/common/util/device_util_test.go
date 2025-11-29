// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
package util

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/test"
)

const testRunCommandRes = "run command result"

func TestGetA500Sn(t *testing.T) {
	convey.Convey("test func getA500Sn success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, testRunCommandRes, nil)
		defer p1.Reset()
		sn, err := getA500Sn()
		arr := strings.Split(testRunCommandRes, " ")
		if len(arr) < snIndex+1 {
			panic("run command res len is invalid")
		}
		expSn := strings.TrimRight(arr[snIndex], "\r\n")
		convey.So(sn, convey.ShouldResemble, expSn)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func getA500Sn failed, sn error", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", nil)
		defer p1.Reset()
		sn, err := getA500Sn()
		convey.So(sn, convey.ShouldResemble, "")
		expErr := errors.New("get a500 serial number failed,error:output is not expected format")
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func getA500Sn failed, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		sn, err := getA500Sn()
		convey.So(sn, convey.ShouldResemble, "")
		expErr := fmt.Errorf("get a500 serial number failed,error:%v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetCgroupDriver(t *testing.T) {
	convey.Convey("test func GetCgroupDriver success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, testRunCommandRes, nil)
		defer p1.Reset()
		cgroupDriver, err := GetCgroupDriver()
		convey.So(cgroupDriver, convey.ShouldResemble, testRunCommandRes)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetCgroupDriver failed, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		cgroupDriver, err := GetCgroupDriver()
		convey.So(cgroupDriver, convey.ShouldResemble, "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
