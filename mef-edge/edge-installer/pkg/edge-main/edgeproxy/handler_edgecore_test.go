// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package edgeproxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

func TestEdgeCoreStart(t *testing.T) {

	conn, server := CreateWebsocket(getEdgeCoreMessage)
	defer func() {
		server.Close()
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()
	convey.Convey("Create a EdgeCoreProxy object\n", t, func() {
		ecp := &EdgecoreProxy{}
		fmt.Printf("Wait for %v and close after timeout.\n", WaitingDuration)
		ctx, cancel := context.WithTimeout(context.Background(), WaitingDuration)
		defer cancel()
		go func() {
			err := ecp.Start(conn)
			if err != nil {
				panic(err)
			}
		}()
		<-ctx.Done()
		err := ctx.Err()
		convey.So(err, convey.ShouldResemble, context.DeadlineExceeded)
	})
}

func getEdgeCoreMessage() *model.Message {
	cntBytes := []byte(`test`)

	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, %v", err)
		return nil
	}

	respMsg.KubeEdgeRouter = model.MessageRoute{
		Source:    constants.SourceHardware,
		Group:     constants.GroupHub,
		Operation: constants.OptQuery,
		Resource:  constants.ActionDefaultNodeStatus,
	}
	respMsg.Header.ID = respMsg.Header.Id
	respMsg.Header.Sync = false
	respMsg.SetRouter(constants.CfgRestore, constants.ModDeviceOm, constants.OptQuery, constants.ActionDefaultNodeStatus)
	if err = respMsg.FillContent(cntBytes); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return nil
	}
	return respMsg
}
