// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package envutils
package envutils

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/terminal"
)

func TestGetFileSystem(t *testing.T) {
	convey.Convey("test get file system", t, func() {
		_, err := GetFileSystem("/")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetFileDevNum(t *testing.T) {
	convey.Convey("test get file dev num", t, func() {
		_, err := GetFileDevNum("/")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestIsInTmpfs(t *testing.T) {
	convey.Convey("test whether path is in tempfs", t, func() {
		patches := gomonkey.ApplyFuncReturn(GetFileSystem, int64(tmpfsDevNum), nil)
		defer patches.Reset()

		tempfs, err := IsInTmpfs("/")
		convey.So(err, convey.ShouldBeNil)
		convey.So(tempfs, convey.ShouldBeTrue)
	})
}

func TestGetDiskFree(t *testing.T) {
	convey.Convey("test get disk free space", t, func() {
		_, err := GetDiskFree("/")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCheckDiskSpace(t *testing.T) {
	convey.Convey("test check disk free space", t, func() {
		const available = 20
		patches := gomonkey.ApplyFuncReturn(GetDiskFree, uint64(available), nil)
		defer patches.Reset()

		err := CheckDiskSpace("/", available)
		convey.So(err, convey.ShouldBeNil)
		err = CheckDiskSpace("/", available+1)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestGetUid(t *testing.T) {
	convey.Convey("test get uid", t, func() {
		uid, err := GetUid("root")
		convey.So(err, convey.ShouldBeNil)
		convey.So(uid, convey.ShouldEqual, 0)
	})
}

func TestGetGid(t *testing.T) {
	convey.Convey("test get gid", t, func() {
		uid, err := GetGid("root")
		convey.So(err, convey.ShouldBeNil)
		convey.So(uid, convey.ShouldEqual, 0)
	})
}

func TestGetCurrentUser(t *testing.T) {
	convey.Convey("test get current user", t, func() {
		_, err := GetCurrentUser()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetUserAndIP(t *testing.T) {
	convey.Convey("test get current user and ip", t, func() {
		patches := gomonkey.ApplyFuncSeq(terminal.GetLoginUserAndIP, []gomonkey.OutputCell{
			{Times: 1, Values: []interface{}{"", "ssh-ip", nil}},
			{Times: 1, Values: []interface{}{"", "ssh-ip", errors.New("no ssh")}},
		})
		defer patches.Reset()

		_, ip, err := GetUserAndIP()
		convey.So(err, convey.ShouldBeNil)
		convey.So(ip, convey.ShouldEqual, "ssh-ip")
		_, ip, err = GetUserAndIP()
		convey.So(err, convey.ShouldBeNil)
		convey.So(ip, convey.ShouldNotEqual, "ssh-ip")
	})
}

func TestCheckUserIsRoot(t *testing.T) {
	convey.Convey("test get current user and ip", t, func() {
		patches := gomonkey.ApplyFuncSeq(GetCurrentUser, []gomonkey.OutputCell{
			{Times: 1, Values: []interface{}{"root", nil}},
			{Times: 1, Values: []interface{}{"normal-user", nil}},
			{Times: 1, Values: []interface{}{"", errors.New("no user")}},
		})
		defer patches.Reset()

		err := CheckUserIsRoot()
		convey.So(err, convey.ShouldBeNil)
		err = CheckUserIsRoot()
		convey.So(err, convey.ShouldNotBeNil)
		err = CheckUserIsRoot()
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestGetUuid(t *testing.T) {
	convey.Convey("test get uuid", t, func() {
		_, err := GetUuid()
		convey.So(err, convey.ShouldBeNil)
	})
}
