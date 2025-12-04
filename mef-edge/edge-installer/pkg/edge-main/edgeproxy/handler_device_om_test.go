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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

func TestDeviceOmStart(t *testing.T) {

	convey.Convey("Send a message for updating container_info\n", t, func() {
		conn, server := CreateWebsocket(getDeviceOMMessage)
		defer func() {
			server.Close()
			err := conn.Close()
			if err != nil {
				panic(err)
			}
		}()
		dop := &DeviceOmProxy{}
		fmt.Printf("Wait for %v and close after timeout.\n", WaitingDuration)
		ctx, cancel := context.WithTimeout(context.Background(), WaitingDuration)
		defer cancel()
		go func() {
			err := dop.Start(conn)
			if err != nil {
				panic(err)
			}
		}()
		<-ctx.Done()
		err := ctx.Err()
		convey.So(err, convey.ShouldResemble, context.DeadlineExceeded)
	})

}

func getDeviceOMMessage() *model.Message {

	cntBytes := []byte(`test`)

	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, %v", err)
		return nil
	}

	respMsg.KubeEdgeRouter = model.MessageRoute{
		Source:    constants.SourceHardware,
		Group:     constants.GroupHub,
		Operation: constants.OptUpdate,
		Resource:  constants.ActionContainerInfo,
	}
	respMsg.Header.ID = respMsg.Header.Id
	respMsg.Header.Sync = true
	respMsg.SetRouter(constants.CfgRestore, constants.ModDeviceOm, constants.OptUpdate, constants.ActionContainerInfo)
	if err = respMsg.FillContent(cntBytes); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return nil
	}
	return respMsg
}

func TestFeedbackError(t *testing.T) {
	convey.Convey("When the SendMsgToWs interface is successfully invoked,"+
		" an error response message is sent.\n", t, func() {
		p1 := gomonkey.ApplyFuncReturn(SendMsgToWs, nil)
		defer p1.Reset()
		FeedbackError(nil, "", getDeviceOMMessage())
	})
}
