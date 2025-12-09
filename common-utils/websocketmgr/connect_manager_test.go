// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package websocketmgr for
package websocketmgr

import (
	"bytes"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gorilla/websocket"
	"github.com/smartystreets/goconvey/convey"
)

// TestReadMsg test read message
func TestReadMsg(t *testing.T) {
	convey.Convey("read msg should avoid zip zoom", t, func() {
		const (
			msgSize      = 10 * 1024 * 1024
			msgReadLimit = 1 * 1024 * 1024
		)
		convey.Convey("zip bomb msg", func() {
			cm := wsConnectMgr{conn: &websocket.Conn{}}

			patches := gomonkey.ApplyMethodReturn(&websocket.Conn{}, "NextReader",
				websocket.TextMessage, bytes.NewReader(make([]byte, msgSize)), nil)
			defer patches.Reset()

			msg, err := cm.readMsg()
			convey.So(err != nil || len(msg) <= int(defaultReadSizeLimit), convey.ShouldBeTrue)
		})

		convey.Convey("normal msg", func() {
			cm := wsConnectMgr{conn: &websocket.Conn{}}
			patches := gomonkey.ApplyMethodReturn(&websocket.Conn{}, "NextReader",
				websocket.TextMessage, bytes.NewReader(make([]byte, msgReadLimit)), nil)
			defer patches.Reset()

			msg, err := cm.readMsg()
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(msg), convey.ShouldEqual, msgReadLimit)
		})
	})
}
