// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package terminal provide a safe reader for password
package terminal

import (
	"errors"
	"io"
	"syscall"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

const maxReadLen = 3

func TestReadPassword(t *testing.T) {
	mock := gomonkey.ApplyFunc(ioctlGetTermios, func(fd int, req uint) (*termios, error) {
		return &termios{}, nil
	})
	defer mock.Reset()
	mock3 := gomonkey.ApplyFunc(readPasswordLine, func(reader io.Reader, len int) ([]byte, error) {
		return []byte("111"), nil
	})
	defer mock3.Reset()
	convey.Convey("test password reader", t, func() {
		mock2 := gomonkey.ApplyFunc(ioctlSetTermios, func(fd int, req uint, value *termios) error {
			return nil
		})
		defer mock2.Reset()
		pd, err := ReadPassword(0, maxReadLen)
		convey.So(string(pd), convey.ShouldEqual, "111")
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("test password reader err", t, func() {
		mock2 := gomonkey.ApplyFunc(ioctlSetTermios, func(fd int, req uint, value *termios) error {
			return errors.New("mock error")
		})
		defer mock2.Reset()
		pd, err := ReadPassword(0, maxReadLen)
		convey.So(pd, convey.ShouldEqual, nil)
		convey.So(err.Error(), convey.ShouldEqual, "mock error")
	})
}

func TestReadPasswordLine(t *testing.T) {
	convey.Convey("test readPasswordLine", t, func() {
		r := localReader(0)
		mock := gomonkey.ApplyMethodFunc(r, "Read", func(buf []byte) (int, error) {
			buf[0] = byte('\n')
			return 1, nil
		})
		defer mock.Reset()
		_, err := readPasswordLine(r, 1)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("test readPasswordLine and return EOF", t, func() {
		r := localReader(0)
		mock := gomonkey.ApplyMethodFunc(r, "Read", func(buf []byte) (int, error) {
			return 0, io.EOF
		})
		defer mock.Reset()
		_, err := readPasswordLine(r, 1)
		convey.So(err, convey.ShouldEqual, io.EOF)
	})

	convey.Convey("test readPasswordLine err", t, func() {
		r := localReader(0)
		mock := gomonkey.ApplyMethodFunc(r, "Read", func(buf []byte) (int, error) {
			return 0, errors.New("mock error")
		})
		defer mock.Reset()
		_, err := readPasswordLine(r, 1)
		convey.So(err.Error(), convey.ShouldEqual, "mock error")
	})
}

func TestIoctlGetTermios(t *testing.T) {
	convey.Convey("test ioctlGetTermios, ", t, func() {
		mock := gomonkey.ApplyFunc(syscallIOctl, func(fd int, req uint, arg uintptr) (err error) {
			return errors.New("mock error")
		})
		defer mock.Reset()
		tos, err := ioctlGetTermios(0, syscall.TCGETS)
		convey.So(tos.iflag, convey.ShouldEqual, 0)
		convey.So(err.Error(), convey.ShouldEqual, "mock error")
	})
}

func TestIoctlSetTermios(t *testing.T) {
	convey.Convey("test ioctlSetTermios, ", t, func() {
		mock := gomonkey.ApplyFunc(syscallIOctl, func(fd int, req uint, arg uintptr) (err error) {
			return errors.New("mock error")
		})
		defer mock.Reset()
		err := ioctlSetTermios(0, syscall.TCSETS, &termios{})
		convey.So(err.Error(), convey.ShouldEqual, "mock error")
	})
}

// TestReadPasswordWithTimeout tests ReadPasswordWithTimeout
func TestReadPasswordWithTimeout(t *testing.T) {
	convey.Convey("test ReadPasswordWithTimeout, ", t, func() {
		convey.Convey("ReadPasswordWithTimeout should return password when succeed", testReadPasswordSucceed)
		convey.Convey("ReadPasswordWithTimeout should report error when timeout", testReadPasswordTimeout)
	})
}

func testReadPasswordSucceed() {
	const testPassword = "123"
	mock := gomonkey.ApplyFuncReturn(ioctlSetTermios, nil).
		ApplyFuncReturn(ioctlGetTermios, &termios{}, nil).
		ApplyFunc(getPwd, func(pwdChannel chan<- []byte, fd int, maxReadLength int) {
			if pwdChannel != nil {
				pwdChannel <- []byte(testPassword)
			}
		})
	defer mock.Reset()

	password, err := ReadPasswordWithTimeout(0, maxReadLen, time.Second)
	convey.So(err, convey.ShouldBeNil)
	convey.So(string(password), convey.ShouldEqual, testPassword)
}

func testReadPasswordTimeout() {
	mock := gomonkey.ApplyFuncReturn(ioctlSetTermios, nil).
		ApplyFuncReturn(ioctlGetTermios, &termios{}, nil).
		ApplyFunc(getPwd, func(pwdChannel chan<- []byte, fd int, maxReadLength int) {})
	defer mock.Reset()

	_, err := ReadPasswordWithTimeout(0, maxReadLen, time.Millisecond)
	convey.So(err, convey.ShouldNotBeNil)
	convey.So(err.Error(), convey.ShouldContainSubstring, "time out")
}
