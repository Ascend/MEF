// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package common test for set node ip
package common

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

func getNetMgrTemplate() *config.NetManager {
	return &config.NetManager{
		NetType:  constants.MEF,
		IP:       "127.0.0.1",
		Port:     0,
		AuthPort: 0,
		WithOm:   false,
		Token:    nil,
	}
}

func TestSetNodeIPToEdgeCore(t *testing.T) {
	netManager := getNetMgrTemplate()
	var p = gomonkey.ApplyFuncReturn(config.GetNetManager, netManager, nil).
		ApplyFuncReturn(net.Dial, fakeNetConn{}, nil).
		ApplyMethodReturn(&net.TCPAddr{}, "String", "127.0.0.1").
		ApplyFuncReturn(config.SetNodeIP, nil)
	defer p.Reset()

	convey.Convey("set node ip should be success, netType is not MEF", t, func() {
		netManagerWithFD := getNetMgrTemplate()
		netManagerWithFD.NetType = constants.FD
		var p1 = gomonkey.ApplyFuncReturn(config.GetNetManager, netManagerWithFD, nil)
		defer p1.Reset()
		err := SetNodeIPToEdgeCore()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("set node ip should be success, set node ip success", t, func() {
		err := SetNodeIPToEdgeCore()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("get node ip should be failed, get install root dir error", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, test.ErrTest)
		defer p1.Reset()

		err := SetNodeIPToEdgeCore()
		expErr := fmt.Errorf("get config path manager failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("get node ip should be failed, get net manager error", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(config.GetNetManager, nil, test.ErrTest)
		defer p1.Reset()

		err := SetNodeIPToEdgeCore()
		innerErr := fmt.Errorf("get net config failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("get node ip failed: %v", innerErr))
	})

	convey.Convey("get node ip should be failed, set node ip error", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(config.SetNodeIP, test.ErrTest)
		defer p1.Reset()

		err := SetNodeIPToEdgeCore()
		expErr := fmt.Errorf("set node ip failed: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

type fakeNetConn struct{}

func (c fakeNetConn) Read([]byte) (int, error)         { return 0, nil }
func (c fakeNetConn) Write([]byte) (int, error)        { return 0, nil }
func (c fakeNetConn) Close() error                     { return nil }
func (c fakeNetConn) LocalAddr() net.Addr              { return &net.TCPAddr{} }
func (c fakeNetConn) RemoteAddr() net.Addr             { return &net.TCPAddr{} }
func (c fakeNetConn) SetDeadline(time.Time) error      { return nil }
func (c fakeNetConn) SetReadDeadline(time.Time) error  { return nil }
func (c fakeNetConn) SetWriteDeadline(time.Time) error { return nil }
