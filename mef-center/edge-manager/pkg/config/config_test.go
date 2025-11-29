// Copyright (c) 2024. Huawei Technologies Co., Ltd.  All rights reserved.

package config

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func TestPodConfig(t *testing.T) {
	convey.Convey("test CheckAndModifyHostPath", t, testCheckAndModifyHostPath)
	convey.Convey("test testCheckAndModifyMaxLimitNumber", t, testCheckAndModifyMaxLimitNumber)

}

func testCheckAndModifyHostPath() {
	testPaths := []string{
		"/tmp",
		"./errorPath1",
		"/../errorPath2",
	}
	path := CheckAndModifyHostPath(testPaths)
	convey.So(path, convey.ShouldResemble, []string{"/tmp"})
}

func testCheckAndModifyMaxLimitNumber() {
	const defaultMaxLimitNumber = 20
	errorNumber := int64(0)
	number := CheckAndModifyMaxLimitNumber(errorNumber)
	convey.So(number, convey.ShouldEqual, defaultMaxLimitNumber)
}

func TestAuthConfig(t *testing.T) {
	convey.Convey("test auth config", t, testCheckAuthConfig)
}

func testCheckAuthConfig() {
	testConfig := AuthInfo{TokenExpireTime: minTokenExpireTime}
	SetConfig(testConfig)
	config := GetAuthConfig()
	convey.So(config, convey.ShouldResemble, testConfig)

	errorConfig := AuthInfo{TokenExpireTime: 0}
	err := CheckAuthConfig(errorConfig)
	convey.So(err, convey.ShouldNotBeNil)
}
