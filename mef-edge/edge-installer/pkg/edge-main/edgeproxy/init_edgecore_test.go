// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
