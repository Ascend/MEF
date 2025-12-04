// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package edgeproxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-installer/pkg/common/constants"
)

func TestInitEdgeCoreStart(t *testing.T) {

	conn, server := CreateWebsocket(getAsyncMessage)
	defer func() {
		server.Close()
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()
	convey.Convey("Initialize the EdgeOmProxy object.\n", t, func() {
		fmt.Printf("Wait for %v and close after timeout.\n", WaitingDuration)
		ctx, cancel := context.WithTimeout(context.Background(), WaitingDuration)
		defer cancel()

		ecp := NewEdgeCoreProxy(true)
		convey.So(ecp.Name(), convey.ShouldResemble, constants.ModEdgeCore)
		convey.So(ecp.Enable(), convey.ShouldResemble, true)
		go ecp.Start()
		<-ctx.Done()
		err := ctx.Err()
		convey.So(err, convey.ShouldResemble, context.DeadlineExceeded)
	})
}
