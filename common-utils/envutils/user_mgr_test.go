// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package envutils
package envutils

import (
	"errors"
	"fmt"
	"os/user"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
)

func TestAddUserAccount(t *testing.T) {
	convey.Convey("test add user account", t, func() {
		convey.Convey("test add user account when user and group both exist",
			testAddUserAccountWhenUserAndGroupBothExist)
		convey.Convey("test create user", testCreateUser)
		convey.Convey("test check nologin", testCheckNoLogin)
		convey.Convey("test check group contain other user", testCheckGroupContainOtherUser)
		convey.Convey("test is info in file", testIsInfoInFile)
	})
}

const (
	testUid = 1022
	testGid = 1024
)

func testAddUserAccountWhenUserAndGroupBothExist() {
	u := &user.User{Uid: "1022", Gid: "1024", Username: "test-user"}
	g := &user.Group{Gid: "1024", Name: "test-group"}
	patches := gomonkey.ApplyFuncReturn(user.Lookup, u, nil).ApplyFuncReturn(user.LookupGroup, g, nil)
	defer patches.Reset()

	mgr := NewUserMgr(u.Username, g.Name, testUid, testGid)
	convey.Convey("test home dir exists", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, true)
		defer patches.Reset()
		convey.So(mgr.AddUserAccount(), convey.ShouldNotBeNil)
	})
	convey.Convey("test GroupIds failed", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", nil, errors.New("not found"))
		defer patches.Reset()
		convey.So(mgr.AddUserAccount(), convey.ShouldNotBeNil)
	})
	convey.Convey("test GroupIds not match", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"1023"}, nil)
		defer patches.Reset()
		convey.So(mgr.AddUserAccount(), convey.ShouldNotBeNil)
	})
	convey.Convey("test checkNoLogin failed", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"1024"}, nil).
			ApplyPrivateMethod(&UserMgr{}, "checkNoLogin", mockCheckLoginReturns(errors.New("can log in")))
		defer patches.Reset()
		convey.So(mgr.AddUserAccount(), convey.ShouldNotBeNil)
	})
	convey.Convey("test checkGroupContainOtherUser failed", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"1024"}, nil).
			ApplyPrivateMethod(&UserMgr{}, "checkNoLogin", mockCheckLoginReturns(nil)).
			ApplyPrivateMethod(&UserMgr{}, "checkGroupContainOtherUser",
				mockCheckGroupContainOtherUserReturns(errors.New("contains other user")))
		defer patches.Reset()
		convey.So(mgr.AddUserAccount(), convey.ShouldNotBeNil)
	})
	convey.Convey("test check exist user/group successfully", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"1024"}, nil).
			ApplyPrivateMethod(&UserMgr{}, "checkNoLogin", mockCheckLoginReturns(nil)).
			ApplyPrivateMethod(&UserMgr{}, "checkGroupContainOtherUser",
				mockCheckGroupContainOtherUserReturns(nil))
		defer patches.Reset()
		convey.So(mgr.AddUserAccount(), convey.ShouldBeNil)
	})
}

func testCreateUser() {
	patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).ApplyFuncReturn(RunCommand, "", nil)
	defer patches.Reset()

	mgr := UserMgr{}
	err := mgr.createUser("")
	convey.So(err, convey.ShouldBeNil)
}

func testCheckNoLogin() {
	patches := gomonkey.ApplyFuncSeq(isInfoInFile, []gomonkey.OutputCell{
		{Times: 1, Values: []interface{}{true, nil}},
		{Times: 1, Values: []interface{}{false, nil}},
	})
	defer patches.Reset()

	mgr := UserMgr{}
	err := mgr.checkNoLogin("")
	convey.So(err, convey.ShouldBeNil)
	err = mgr.checkNoLogin("")
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckGroupContainOtherUser() {
	patches := gomonkey.ApplyFuncSeq(isInfoInFile, []gomonkey.OutputCell{
		{Times: 1, Values: []interface{}{true, nil}},
		{Times: 1, Values: []interface{}{false, nil}},
	})
	defer patches.Reset()

	mgr := UserMgr{}
	err := mgr.checkGroupContainOtherUser()
	convey.So(err, convey.ShouldBeNil)
	err = mgr.checkGroupContainOtherUser()
	convey.So(err, convey.ShouldNotBeNil)
}

func testIsInfoInFile() {
	testcases := []struct {
		content string
		reg     string
		result  bool
		err     error
	}{
		{content: `root:x:0:`, reg: fmt.Sprintf(noOtherInGroupPattern, "root"), result: true},
		{content: `root:x:0:other`, reg: fmt.Sprintf(noOtherInGroupPattern, "root")},
		{content: `root:x:0:0:root:/root:/sbin/nologin`,
			reg: fmt.Sprintf(noLoginPattern, "root", "/sbin/nologin"), result: true},
		{content: `root:x:0:0:root:/root:`, reg: fmt.Sprintf(noLoginPattern, "root", "/sbin/nologin")},
	}
	var patches *gomonkey.Patches
	defer func() {
		if patches != nil {
			patches.Reset()
		}
	}()

	for _, tc := range testcases {
		fmt.Printf("reg=%s, content=%s\n", tc.reg, tc.content)
		if patches != nil {
			patches.Reset()
		}
		patches = gomonkey.ApplyFuncReturn(fileutils.LoadFile, []byte(tc.content), nil)
		result, err := isInfoInFile("", tc.reg)
		if tc.err != nil {
			convey.So(err, convey.ShouldNotBeNil)
			continue
		}
		convey.So(err, convey.ShouldBeNil)
		if tc.result {
			convey.So(result, convey.ShouldBeTrue)
			continue
		} else {
			convey.So(result, convey.ShouldBeFalse)
		}
	}
}

func mockCheckLoginReturns(err error) func(string) error {
	return func(string) error {
		return err
	}
}

func mockCheckGroupContainOtherUserReturns(err error) func() error {
	return func() error {
		return err
	}
}
