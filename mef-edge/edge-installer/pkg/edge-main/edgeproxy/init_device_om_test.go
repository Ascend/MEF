// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package edgeproxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-installer/pkg/common/constants"
)

func TestInitDeviceOmStart(t *testing.T) {

	conn, server := CreateWebsocket(getAsyncMessage)
	defer func() {
		server.Close()
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()
	convey.Convey("Initialize the deviceOmProxy object\n", t, func() {
		fmt.Printf("Wait for %v and close after timeout.\n", WaitingDuration)
		ctx, cancel := context.WithTimeout(context.Background(), WaitingDuration)
		defer cancel()

		dop := NewDeviceOmProxy(true)
		convey.So(dop.Name(), convey.ShouldResemble, constants.ModDeviceOm)
		convey.So(dop.Enable(), convey.ShouldResemble, true)
		go dop.Start()
		<-ctx.Done()
		err := ctx.Err()
		convey.So(err, convey.ShouldResemble, context.DeadlineExceeded)
	})
}
