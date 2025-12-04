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
	"github.com/gorilla/websocket"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

func TestEdgeOmStart(t *testing.T) {
	convey.Convey("Send the update container_info message\n", t, func() {
		conn, server := CreateWebsocket(getEdgeOMMessage)
		defer func() {
			server.Close()
			err := conn.Close()
			if err != nil {
				panic(err)
			}

		}()
		p1 := gomonkey.ApplyFuncReturn(SendMsgToWs, nil)
		defer p1.Reset()
		startEdgeOmProxy(conn)
	})

}

func TestEdgeOmStart2(t *testing.T) {
	convey.Convey("Send the sys_info message\n", t, func() {
		conn, server := CreateWebsocket(getReportSysInfoMessage)
		defer func() {
			server.Close()
			err := conn.Close()
			if err != nil {
				panic(err)
			}
		}()
		startEdgeOmProxy(conn)
	})
}

func startEdgeOmProxy(conn *websocket.Conn) {
	eop := &EdgeOmProxy{}
	fmt.Printf("Wait for %v and close after timeout.\n", WaitingDuration)
	ctx, cancel := context.WithTimeout(context.Background(), WaitingDuration)
	defer cancel()
	go func() {
		err := eop.Start(conn)
		if err != nil {
			panic(err)
		}
	}()
	<-ctx.Done()
	err := ctx.Err()
	convey.So(err, convey.ShouldResemble, context.DeadlineExceeded)
}
func getEdgeOMMessage() *model.Message {
	cntBytes := []byte("Hi")

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
	respMsg.Header.Sync = false
	respMsg.SetRouter(constants.CfgRestore, constants.ModDeviceOm, constants.OptUpdate, constants.ActionContainerInfo)
	if err = respMsg.FillContent(cntBytes); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return nil
	}
	return respMsg
}

func getReportSysInfoMessage() *model.Message {
	jsonStr := `{"product_capability_edge": ["edge1", "edge2", "edge3"]}`
	cntBytes := []byte(jsonStr)

	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, %v", err)
		return nil
	}

	respMsg.KubeEdgeRouter = model.MessageRoute{
		Source:    constants.SourceHardware,
		Group:     constants.GroupHub,
		Operation: constants.OptReport,
		Resource:  constants.ResStatic,
	}
	respMsg.Header.ID = respMsg.Header.Id
	respMsg.Header.Sync = false
	respMsg.SetRouter(constants.CfgRestore, constants.ModDeviceOm, constants.OptUpdate, constants.ResStatic)
	if err = respMsg.FillContent(cntBytes); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return nil
	}
	return respMsg
}
