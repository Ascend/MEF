// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package edgeproxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-installer/pkg/common/constants"
)

func TestInitEdgeOmStart(t *testing.T) {

	conn, server := CreateWebsocket(getAsyncMessage)
	defer func() {
		server.Close()
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()
	convey.Convey("Initialize the EdgeOmProxy object\n", t, func() {
		fmt.Printf("Wait for %v and close after timeout.\n", WaitingDuration)
		ctx, cancel := context.WithTimeout(context.Background(), WaitingDuration)
		defer cancel()

		eop := NewEdgeOmProxy(true)
		convey.So(eop.Name(), convey.ShouldResemble, constants.ModEdgeOm)
		convey.So(eop.Enable(), convey.ShouldResemble, true)
		go eop.Start()
		<-ctx.Done()
		err := ctx.Err()
		convey.So(err, convey.ShouldResemble, context.DeadlineExceeded)
	})
}
