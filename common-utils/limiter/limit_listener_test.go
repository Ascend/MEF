// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package limiter implement a token bucket limiter
package limiter

import (
	"errors"
	"net"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

const (
	len2 = 2
)

func TestLimitListenerAccept(t *testing.T) {
	convey.Convey("test Accept function", t, func() {

		limitLor, err := LimitListener(&mockLicener{}, len2, len2, DefaultCacheSize)
		if err != nil {
			return
		}
		l, ok := limitLor.(*localLimitListener)
		if !ok {
			return
		}
		mock2 := gomonkey.ApplyFunc(getIpAndKey, func(net.Conn) (string, string) {
			return "127.0.0.1", "key-127.0.0.1"
		})
		defer mock2.Reset()
		convey.Convey("acquire token success", func() {
			_, err = l.Accept()
			convey.So(err, convey.ShouldEqual, nil)
		})

		convey.Convey("accept failed", func() {
			mock := gomonkey.ApplyMethodFunc(l.Listener, "Accept", func() (net.Conn, error) {
				return nil, errors.New("mock error")
			})
			defer mock.Reset()
			con, err := l.Accept()
			convey.So(err, convey.ShouldNotEqual, nil)
			convey.So(con, convey.ShouldEqual, nil)
		})

		convey.Convey("acquire token failed", func() {
			mock := gomonkey.ApplyPrivateMethod(l, "acquire", func(*localLimitListener) bool {
				return false
			})
			defer mock.Reset()
			con, err := l.Accept()
			convey.So(err, convey.ShouldEqual, nil)
			conm, ok := con.(*limitListenerConn)
			if !ok {
				return
			}
			convey.So(conm.release, convey.ShouldNotEqual, nil)
		})

	})
}

type mockLicener struct {
}

func (l *mockLicener) Accept() (net.Conn, error) {
	return &net.TCPConn{}, nil
}

func (l *mockLicener) Addr() net.Addr {
	return &net.IPAddr{
		IP:   []byte("127.0.0.1"),
		Zone: "",
	}
}

func (l *mockLicener) Close() error {
	return nil
}

func TestGetIpAndKey(t *testing.T) {
	convey.Convey("test getIp function", t, func() {
		c := net.TCPConn{}
		mock := gomonkey.ApplyMethodFunc(&c, "RemoteAddr", func() net.Addr {
			return &net.IPAddr{
				IP:   []byte("127.0.0.1"),
				Zone: "",
			}
		})
		defer mock.Reset()
		ip, _ := getIpAndKey(&c)
		convey.So(ip, convey.ShouldNotEqual, "")
	})
}

func TestLimitListener(t *testing.T) {
	convey.Convey("test new listener function success", t, func() {
		l, err := LimitListener(&mockLicener{}, maxConnection, maxIPConnection, DefaultDataLimit)
		convey.So(l, convey.ShouldNotEqual, nil)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("test new listener function", t, func() {
		_, err := LimitListener(&mockLicener{}, maxConnection+1, maxIPConnection, DefaultDataLimit)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("test new listener function", t, func() {
		_, err := LimitListener(&mockLicener{}, maxConnection, maxIPConnection+1, DefaultDataLimit)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}
