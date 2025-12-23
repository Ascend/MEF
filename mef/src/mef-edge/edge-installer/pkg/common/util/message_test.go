// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

const content = "test content"

func TestNewInnerMsgWithFullParas(t *testing.T) {
	convey.Convey("normal case", t, func() {
		_, err := NewInnerMsgWithFullParas(InnerMsgParams{Content: content})
		convey.So(err, convey.ShouldResemble, nil)

		// content is nil
		_, err = NewInnerMsgWithFullParas(InnerMsgParams{Content: nil})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("new message failed case", t, func() {
		p1 := gomonkey.ApplyFuncReturn(model.NewMessage, nil, errors.New("get id failed"))
		defer p1.Reset()
		_, err := NewInnerMsgWithFullParas(InnerMsgParams{Content: content})
		convey.So(fmt.Sprintf("%v", err), convey.ShouldContainSubstring, "new message error")
	})
	convey.Convey("new message failed case", t, func() {
		p1 := gomonkey.ApplyFuncReturn(json.Marshal, nil, errors.New("marshal failed"))
		defer p1.Reset()
		_, err := NewInnerMsgWithFullParas(InnerMsgParams{Content: true})
		convey.So(fmt.Sprintf("%v", err), convey.ShouldContainSubstring, "marshal msg content failed")
	})
}

func TestSendSyncMsg(t *testing.T) {

	resp := &model.Message{Content: model.RawMessage(constants.Success)}
	patches := gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, resp, nil)
	defer patches.Reset()

	convey.Convey("test func SendSyncMsg success", t, func() {
		data, err := SendSyncMsg(InnerMsgParams{Content: content})
		convey.So(data, convey.ShouldEqual, constants.Success)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func SendSyncMsg failed", t, func() {
		convey.Convey("new model message failed", func() {
			p := gomonkey.ApplyFuncReturn(NewInnerMsgWithFullParas, nil, test.ErrTest)
			defer p.Reset()
			data, err := SendSyncMsg(InnerMsgParams{Content: content})
			convey.So(data, convey.ShouldBeBlank)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("new model message failed, error: %v", test.ErrTest))
		})

		convey.Convey("send sync message failed", func() {
			p := gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, nil, test.ErrTest)
			defer p.Reset()
			data, err := SendSyncMsg(InnerMsgParams{Content: content})
			convey.So(data, convey.ShouldBeBlank)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("send sync message failed, error: %v", test.ErrTest))
		})

		convey.Convey("parse content failed", func() {
			p := gomonkey.ApplyMethodReturn(&model.Message{}, "ParseContent", test.ErrTest)
			defer p.Reset()
			data, err := SendSyncMsg(InnerMsgParams{Content: content})
			expErr := fmt.Errorf("fill data into content failed: %v", test.ErrTest)
			convey.So(data, convey.ShouldBeBlank)
			convey.So(err, convey.ShouldResemble, expErr)
		})
	})
}

func TestSendInnerMsgResponse(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
	defer patches.Reset()

	msg, err := model.NewMessage()
	if err != nil {
		panic(err)
	}

	convey.Convey("test func SendInnerMsgResponse success", t, func() {
		err = SendInnerMsgResponse(msg, content)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func SendInnerMsgResponse failed, new response failed", t, func() {
		var p1 = gomonkey.ApplyMethodReturn(&model.Message{}, "NewResponse", nil, test.ErrTest)
		defer p1.Reset()
		err = SendInnerMsgResponse(msg, content)
		expErr := fmt.Errorf("create response message failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func SendInnerMsgResponse failed, fill content failed", t, func() {
		var p1 = gomonkey.ApplyMethodReturn(&model.Message{}, "FillContent", test.ErrTest)
		defer p1.Reset()
		err = SendInnerMsgResponse(msg, content)
		expErr := fmt.Errorf("fill response message failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func SendInnerMsgResponse failed, send message failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(modulemgr.SendMessage, test.ErrTest)
		defer p1.Reset()
		err = SendInnerMsgResponse(msg, content)
		expErr := fmt.Errorf("send response message failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
