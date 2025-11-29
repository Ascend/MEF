// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common test for processing lock
package util

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"

	"edge-installer/pkg/common/constants"
)

var (
	lock = flagLock{
		flagPath:  filepath.Join(constants.FlagPath, constants.ProcessFlag),
		operation: constants.Upgrade,
	}
	testErr = errors.New("test error")
)

func TestLock(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", nil)
	defer p.Reset()

	convey.Convey("lock should be success", t, testLock)
	convey.Convey("lock should be failed, process flag is already locked", t, testLockErrLocked)
	convey.Convey("lock should be failed, check owner and permission & delete file error", t, testLockErrCheckDelete)
	convey.Convey("lock should be failed, read file error", t, testLockErrReadFile)
	convey.Convey("lock should be success, flag file lines error", t, testLockErrFlagFileLines)
	convey.Convey("lock should be success, Atoi error", t, testLockErrAtoi)
	convey.Convey("lock should be success, proc is not active", t, testLockErrProcActive)
	convey.Convey("lock should be failed, get proc name error", t, testLockErrGetProcName)
	convey.Convey("lock should be failed, write with lock error", t, testLockErrWriteWithLock)
}

func testLock() {
	err := lock.Lock()
	convey.So(err, convey.ShouldBeNil)
}

func testLockErrLocked() {
	var p1 = gomonkey.ApplyFunc(atomic.LoadUint32,
		func(addr *uint32) (val uint32) {
			return locking
		})
	defer p1.Reset()

	err := lock.Lock()
	convey.So(err, convey.ShouldResemble, errors.New("process flag is already locked"))
}

func testLockErrCheckDelete() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", testErr).
		ApplyFuncReturn(fileutils.DeleteFile, testErr)
	defer p1.Reset()

	atomic.StoreUint32(&lock.status, none)
	err := lock.Lock()
	convey.So(err, convey.ShouldResemble, testErr)
}

func testLockErrReadFile() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{[]byte(""), testErr}},

		{Values: gomonkey.Params{[]byte(""), nil}},
		{Values: gomonkey.Params{[]byte(""), nil}},
	}

	var p1 = gomonkey.ApplyFuncSeq(fileutils.LoadFile, outputs)
	defer p1.Reset()

	err := lock.Lock()
	convey.So(err, convey.ShouldResemble, testErr)

	err = lock.Lock()
	convey.So(err, convey.ShouldBeNil)
}

func testLockErrFlagFileLines() {
	var p1 = gomonkey.ApplyFunc(strings.Split,
		func(s, sep string) []string {
			return []string{"proc pid"}
		})
	defer p1.Reset()

	atomic.StoreUint32(&lock.status, none)
	err := lock.Lock()
	convey.So(err, convey.ShouldBeNil)
}

func testLockErrAtoi() {
	var p1 = gomonkey.ApplyFunc(strconv.Atoi,
		func(s string) (int, error) {
			return 0, testErr
		})
	defer p1.Reset()

	atomic.StoreUint32(&lock.status, none)
	err := lock.Lock()
	convey.So(err, convey.ShouldBeNil)
}

func testLockErrProcActive() {
	var p1 = gomonkey.ApplyFunc(IsProcessActive,
		func(pid int) bool {
			return false
		})
	defer p1.Reset()

	atomic.StoreUint32(&lock.status, none)
	err := lock.Lock()
	convey.So(err, convey.ShouldBeNil)
}

func testLockErrGetProcName() {
	var p1 = gomonkey.ApplyFunc(GetProcName,
		func(pid int) (string, error) {
			return "", testErr
		})
	defer p1.Reset()

	atomic.StoreUint32(&lock.status, none)
	err := lock.Lock()
	convey.So(err, convey.ShouldResemble, errors.New("get current process name failed"))
}

func testLockErrWriteWithLock() {
	var p1 = gomonkey.ApplyFunc(WriteWithLock,
		func(filePath string, data []byte) error {
			return testErr
		})
	defer p1.Reset()

	err := lock.Lock()
	convey.So(err, convey.ShouldNotBeNil)

	var p2 = gomonkey.ApplyFunc(GetProcName,
		func(pid int) (string, error) {
			return "", nil
		})
	defer p2.Reset()

	err = lock.Lock()
	convey.So(err, convey.ShouldResemble, errors.New("write pid, process name, and operation to process flag failed"))
}

func TestUnlock(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(fileutils.DeleteFile, nil)
	defer p.Reset()

	convey.Convey("unlock should be success", t, testUnlock)
	convey.Convey("unlock should be failed, delete error", t, testUnlockErrDelete)
}

func testUnlock() {
	atomic.StoreUint32(&lock.status, locking)
	err := lock.Unlock()
	convey.So(err, convey.ShouldBeNil)

	atomic.StoreUint32(&lock.status, none)
	err = lock.Unlock()
	convey.So(err, convey.ShouldBeNil)
}

func testUnlockErrDelete() {
	var p1 = gomonkey.ApplyFunc(fileutils.DeleteFile,
		func(path string, _ ...fileutils.FileChecker) error {
			return testErr
		})
	defer p1.Reset()

	atomic.StoreUint32(&lock.status, locking)
	err := lock.Unlock()
	convey.So(err, convey.ShouldResemble, errors.New("remove process flag file failed"))
}
