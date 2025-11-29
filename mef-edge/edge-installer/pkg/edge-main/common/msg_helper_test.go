// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package common for test msg helper
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

func TestNewFDPodStatusMsg(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(statusmanager.LoadPodsDataForFd, []statusmanager.FdPod{}, nil)
	defer p.Reset()
	convey.Convey("new fd pod status msg success", t, func() {
		msg, err := NewFDPodStatusMsg(source, group, operation, resource)
		convey.So(msg.Router.Source, convey.ShouldEqual, source)
		convey.So(msg.Router.Destination, convey.ShouldEqual, constants.ModDeviceOm)
		convey.So(msg.Router.Option, convey.ShouldEqual, operation)
		convey.So(msg.Router.Resource, convey.ShouldEqual, resource)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("new fd pod status msg failed", t, func() {
		convey.Convey("create to FD message error", func() {
			p1 := gomonkey.ApplyFuncReturn(model.NewMessage, nil, test.ErrTest)
			defer p1.Reset()
			msg, err := NewFDPodStatusMsg(source, group, operation, resource)
			convey.So(msg, convey.ShouldBeNil)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("create to FD message error: %v", test.ErrTest))
		})

		convey.Convey("load pods data for fd failed", func() {
			p1 := gomonkey.ApplyFuncReturn(statusmanager.LoadPodsDataForFd, nil, test.ErrTest)
			defer p1.Reset()
			msg, err := NewFDPodStatusMsg(source, group, operation, resource)
			convey.So(msg, convey.ShouldBeNil)
			convey.So(err, convey.ShouldResemble, test.ErrTest)
		})

		convey.Convey("fill all pod status into content failed", func() {
			p1 := gomonkey.ApplyMethodReturn(&model.Message{}, "FillContent", test.ErrTest)
			defer p1.Reset()
			msg, err := NewFDPodStatusMsg(source, group, operation, resource)
			convey.So(msg, convey.ShouldBeNil)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("fill all pod status into content failed: %v", test.ErrTest))
		})
	})
}

func TestNewFDNodeStatusMsg(t *testing.T) {
	mockNodeStatus := &statusmanager.FdNode{
		Status: statusmanager.FdNodeStatus{Content: v1.Node{}},
	}
	p := gomonkey.ApplyFuncReturn(statusmanager.LoadNodeDataForFd, mockNodeStatus, nil)
	defer p.Reset()
	convey.Convey("new fd node status msg success", t, func() {
		msg, err := NewFDNodeStatusMsg(source, group, operation, resource)
		convey.So(msg.KubeEdgeRouter.Resource, convey.ShouldEqual,
			constants.ModifiedNodePrefix+mockNodeStatus.Status.Content.Name)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("new fd node status msg failed", t, func() {
		convey.Convey("create msg error", func() {
			p1 := gomonkey.ApplyFuncReturn(model.NewMessage, nil, test.ErrTest)
			defer p1.Reset()
			msg, err := NewFDNodeStatusMsg(source, group, operation, resource)
			convey.So(msg, convey.ShouldBeNil)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("create msg error: %v", test.ErrTest))
		})

		convey.Convey("get node status error", func() {
			p1 := gomonkey.ApplyFuncReturn(statusmanager.LoadNodeDataForFd, nil, test.ErrTest)
			defer p1.Reset()
			msg, err := NewFDNodeStatusMsg(source, group, operation, resource)
			convey.So(msg, convey.ShouldBeNil)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("get node status error: %v", test.ErrTest))
		})

		convey.Convey("fill node status into content failed", func() {
			p1 := gomonkey.ApplyMethodReturn(&model.Message{}, "FillContent", test.ErrTest)
			defer p1.Reset()
			msg, err := NewFDNodeStatusMsg(source, group, operation, resource)
			convey.So(msg, convey.ShouldBeNil)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("fill node status into content failed: %v", test.ErrTest))
		})
	})
}

func TestMsgOutProcess(t *testing.T) {
	convey.Convey("msg out process failed, msg is nil", t, func() {
		msgOut, err := MsgOutProcess(nil)
		convey.So(msgOut, convey.ShouldBeNil)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("invalid msg data"))
	})

	convey.Convey("msg out process success", t, func() {
		msg, err := model.NewMessage()
		if err != nil {
			return
		}
		msg.Header.ParentId = msg.Header.Id
		msg.SetIsSync(true)
		msg.SetRouter(source, group, operation, resource)
		msgOut, err := MsgOutProcess(msg)
		convey.So(err, convey.ShouldBeNil)
		fmt.Printf("msgOut: %+v\n", msgOut)
	})
}

func TestMsgInProcess(t *testing.T) {
	convey.Convey("msg in process failed, msg is nil", t, func() {
		msgIn, err := MsgInProcess(nil)
		convey.So(msgIn, convey.ShouldBeNil)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("invalid msg data"))
	})

	convey.Convey("msg in process success", t, func() {
		msg := getTestMsg()
		if msg == nil {
			return
		}
		msg.Header.ParentID = msg.Header.Id
		msg.Header.Sync = true
		msg.Header.ResourceVersion = "1.0"
		msgIn, err := MsgInProcess(msg)
		convey.So(err, convey.ShouldBeNil)
		fmt.Printf("msgIn: %+v\n", msgIn)
	})
}

func TestMsgOptLog(t *testing.T) {
	msg := getTestMsg()
	if msg == nil {
		return
	}
	convey.Convey("test MsgOptLog, skip minXOM inner message", t, func() {
		msg.KubeEdgeRouter.Operation = constants.OptUpdate
		msg.KubeEdgeRouter.Resource = constants.ResImageCertInfo
		MsgOptLog(msg)
	})

	convey.Convey("test MsgOptLog, record restart pod operation log", t, func() {
		msg.KubeEdgeRouter.Operation = constants.OptRestart
		msg.KubeEdgeRouter.Resource = constants.ActionPod
		MsgOptLog(msg)
	})

	convey.Convey("test MsgOptLog, record other operation log", t, func() {
		msg.KubeEdgeRouter.Operation = constants.OptUpdate
		msg.KubeEdgeRouter.Resource = constants.ActionContainerInfo
		MsgOptLog(msg)
	})

	convey.Convey("test MsgOptLog, over max msg opt log cache", t, func() {
		msg.KubeEdgeRouter.Operation = constants.OptUpdate
		msg.KubeEdgeRouter.Resource = constants.ActionContainerInfo
		const maxMsgOptLogCache = 600
		for i := 0; i <= maxMsgOptLogCache; i++ {
			MsgOptLog(msg)
		}
	})
}

func TestMsgResultOptLog(t *testing.T) {
	convey.Convey("test MsgResultOptLog, ignore operation result log", t, ignoreOptResLog)
	convey.Convey("test MsgResultOptLog, record operation result log by content", t, recordOpLogByContent)
	convey.Convey("test MsgResultOptLog, record operation result log by original msg", t, recordOpLogByOriginMsg)
}

func ignoreOptResLog() {
	resp := &model.Message{}
	convey.Convey("resp is nil", func() {
		MsgResultOptLog(nil)
	})

	convey.Convey("resp.KubeEdgeRouter.Operation is report", func() {
		resp.KubeEdgeRouter.Operation = constants.OptReport
		MsgResultOptLog(resp)
	})

	convey.Convey("resp.KubeEdgeRouter.Operation is queryAllAlarm", func() {
		resp.KubeEdgeRouter.Resource = constants.QueryAllAlarm
		MsgResultOptLog(resp)
	})

	convey.Convey("msg no need to write opt log", func() {
		resp.KubeEdgeRouter.Operation = constants.OptUpdate
		resp.KubeEdgeRouter.Resource = constants.ResPodStatus
		MsgResultOptLog(resp)
	})
}

func recordOpLogByContent() {
	resp, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new test response message failed: %v", err)
		return
	}
	content := config.ProgressTip{
		Percentage: "100%",
		Result:     "success",
		Reason:     "",
	}
	resp.Header.ID = resp.Header.Id
	resp.SetRouter(constants.ModHandlerMgr, constants.ModDeviceOm, constants.OptUpdate, constants.ResConfigResult)
	resp.SetKubeEdgeRouter(constants.HardwareModule, constants.GroupHub, constants.OptUpdate,
		constants.ResConfigResult)

	convey.Convey("record operation res log for pods_data", func() {
		content.Topic = constants.ResourceTypePodsData
		if err = resp.FillContent(content, true); err != nil {
			hwlog.RunLog.Errorf("fill data into test content failed: %v", err)
			return
		}
		MsgResultOptLog(resp)
	})

	convey.Convey("record operation res log for pod", func() {
		content.Topic = constants.ActionPod
		if err = resp.FillContent(content, true); err != nil {
			hwlog.RunLog.Errorf("fill data into test content failed: %v", err)
			return
		}
		MsgResultOptLog(resp)
	})
}

func recordOpLogByOriginMsg() {
	msg := getTestMsg()
	if msg == nil {
		return
	}
	resp, err := model.NewMessage()
	if err != nil {
		return
	}
	resp.Header.ParentID = msg.Header.ID
	resp.SetRouter(constants.ModHandlerMgr, constants.ModDeviceOm, constants.OptUpdate, constants.ResConfigResult)
	resp.SetKubeEdgeRouter(constants.HardwareModule, constants.GroupHub, constants.OptUpdate,
		constants.ActionPod)

	convey.Convey("record operation res log when operation is OptError", func() {
		MsgOptLog(msg)
		resp.KubeEdgeRouter.Operation = constants.OptError
		MsgResultOptLog(resp)
	})

	convey.Convey("record operation res log when operation is OptResponse", func() {
		resp.KubeEdgeRouter.Operation = constants.OptResponse

		MsgOptLog(msg)
		resp.Content = model.RawMessage(constants.OK)
		MsgResultOptLog(resp)

		MsgOptLog(msg)
		resp.Content = model.RawMessage(constants.Failed)
		MsgResultOptLog(resp)
	})

	convey.Convey("operation type error", func() {
		MsgOptLog(msg)
		resp.KubeEdgeRouter.Operation = constants.OptUpdate
		MsgResultOptLog(resp)
	})
}

func TestUpdateFdAddrInfo(t *testing.T) {
	convey.Convey("test update fd address info should be success", t, updateFdAddrInfoSuccess)
	convey.Convey("test update fd address info should be failed", t, updateFdAddrInfoFailed)
}

func updateFdAddrInfoSuccess() {
	msg := getTestMsg()
	if msg == nil {
		return
	}

	convey.Convey("skip update fd address info", func() {
		msg.KubeEdgeRouter.Operation = constants.OptResponse
		err := UpdateFdAddrInfo(msg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("update fd address info success", func() {
		msg.KubeEdgeRouter.Resource = constants.ResImageCertInfo
		testIp, testPort := "127.0.0.1", "0"
		content := map[string]string{"ip": testIp, "port": testPort}
		if err := msg.FillContent(content, true); err != nil {
			hwlog.RunLog.Errorf("fill data into test content failed: %v", err)
			return
		}

		err := UpdateFdAddrInfo(msg)
		convey.So(err, convey.ShouldBeNil)

		fdIp, err = GetFdIp()
		convey.So(err, convey.ShouldBeNil)
		convey.So(fdIp, convey.ShouldEqual, testIp)
	})
}

func updateFdAddrInfoFailed() {
	msg := getTestMsg()
	if msg == nil {
		return
	}
	msg.KubeEdgeRouter.Resource = constants.ResImageCertInfo

	convey.Convey("msg is nil", func() {
		err := UpdateFdAddrInfo(nil)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("invalid message"))
	})

	convey.Convey("parse content failed", func() {
		p := gomonkey.ApplyMethodReturn(&model.Message{}, "ParseContent", test.ErrTest)
		defer p.Reset()
		err := UpdateFdAddrInfo(msg)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("get json data failed: %v", test.ErrTest))
	})

	convey.Convey("unmarshal content failed", func() {
		p := gomonkey.ApplyFuncReturn(json.Unmarshal, test.ErrTest)
		defer p.Reset()
		err := UpdateFdAddrInfo(msg)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("content ip key does not exist", func() {
		if err := msg.FillContent(map[string]string{}, true); err != nil {
			hwlog.RunLog.Errorf("fill data into test content failed: %v", err)
			return
		}
		err := UpdateFdAddrInfo(msg)
		convey.So(err, convey.ShouldResemble, errors.New("content ip key does not exist"))
	})

	convey.Convey("content port key does not exist", func() {
		if err := msg.FillContent(map[string]string{"ip": "127.0.0.1"}, true); err != nil {
			hwlog.RunLog.Errorf("fill data into test content failed: %v", err)
			return
		}
		err := UpdateFdAddrInfo(msg)
		convey.So(err, convey.ShouldResemble, errors.New("content port key does not exist"))
	})

	convey.Convey("invalid fd ip or port", func() {
		if err := msg.FillContent(map[string]string{"ip": "", "port": ""}, true); err != nil {
			hwlog.RunLog.Errorf("fill data into test content failed: %v", err)
			return
		}
		err := UpdateFdAddrInfo(msg)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("invalid fd ip or port"))
	})
}
