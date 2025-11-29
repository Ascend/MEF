// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

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
