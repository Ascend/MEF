// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package envutils
package envutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
)

type dirEntry string

func (e dirEntry) Name() string {
	return string(e)
}

func (e dirEntry) Info() (os.FileInfo, error) {
	return nil, nil
}

func (e dirEntry) IsDir() bool {
	return false
}

func (e dirEntry) Type() os.FileMode {
	return 0
}

func TestGetPortByPid(t *testing.T) {
	convey.Convey("test get port by pid", t, func() {
		ppm := &ProcessPortMgr{}
		patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, []byte(
			`sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
0: 0100007F:0001 00000000:0004 01 00000000:00000000 00:00000000 00000000     0        0 1111 1 0 100 0 0 10 0
1: 0100007F:0002 00000000:0005 0A 00000000:00000000 00:00000000 00000000     0        0 2222 1 0 100 0 0 10 0
2: 0100007F:0003 00000000:0006 0A 00000000:00000000 00:00000000 00000000   123        0 3333 1 0 100 0 0 10 0
3: 0100007F:0007 00000000:0008 01 00000000:00000000 00:00000000 00000000   123        0 4444 1 0 100 0 0 10 0
3: 0100007F:0009 00000000:000A 0A 00000000:00000000 00:00000000 00000000   123        0 5555 1 0 100 0 0 10 0`), nil).
			ApplyFunc(fileutils.ReadLink, func(filePath string) (string, error) {
				return map[string]string{
					"1": "socket:[1111]",
					"2": "socket:[2222]",
					"3": "socket:[3333]",
				}[filepath.Base(filePath)], nil
			}).
			ApplyFuncReturn(fileutils.ReadDir, nil, []os.DirEntry{
				dirEntry("1"),
				dirEntry("2"),
				dirEntry("3"),
			}, nil)
		defer patches.Reset()

		ports, err := ppm.GetPortByPid(TcpProtocol, EstablishedState)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ports, convey.ShouldResemble, []int{1})

		ports, err = ppm.GetPortByPid(TcpProtocol, ListenState)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ports, convey.ShouldResemble, []int{2, 3})

		ports, err = ppm.GetPortByPid(TcpProtocol, AllState)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ports, convey.ShouldResemble, []int{1, 2, 3})
	})
}
