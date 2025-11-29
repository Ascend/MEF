// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configmanager for config init test
package configmanager

import (
	"errors"
	"net/http"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

func TestMethodSelect(t *testing.T) {
	convey.Convey("method select functional test", t, func() {
		convey.Convey("config manager method select failed without url", func() {
			input, _ := model.NewMessage()
			msg := methodSelect(input)
			convey.So(msg, convey.ShouldBeNil)
		})
		convey.Convey("config manager method select failed with root url", func() {
			input, _ := model.NewMessage()
			input.SetRouter("", "", http.MethodPost, configUrlRootPath)
			msg := methodSelect(input)
			convey.So(msg, convey.ShouldBeNil)
		})
		convey.Convey("config manager method select success with image config url", func() {
			input, _ := model.NewMessage()
			input.SetRouter("", "", http.MethodPost, filepath.Join(configUrlRootPath, "config"))
			msg := methodSelect(input)
			convey.So(msg, convey.ShouldNotBeNil)
		})
		convey.Convey("config manager method select success with update url", func() {
			input, _ := model.NewMessage()
			input.SetRouter("", "", http.MethodPost, filepath.Join(innerConfigUrlRootPath, "update"))
			msg := methodSelect(input)
			convey.So(msg, convey.ShouldNotBeNil)
		})
	})
}

func TestPeriodCheckToken(t *testing.T) {
	convey.Convey("periodically check token test", t, func() {
		var (
			checkTimes int
			wg         sync.WaitGroup
			tickerCh   = make(chan time.Time, 1)
		)
		patches := gomonkey.ApplyFuncReturn(time.NewTicker, &time.Ticker{C: tickerCh}).
			ApplyFunc(checkAndUpdateToken, func() { checkTimes++ }).
			ApplyMethod(&time.Ticker{}, "Stop", func(ticker *time.Ticker) {})
		defer patches.Reset()

		wg.Add(1)
		go func() {
			periodCheckToken()
			wg.Done()
		}()

		tickerCh <- time.Time{}
		close(tickerCh)
		wg.Wait()

		const times2 = 2
		convey.So(checkTimes, convey.ShouldEqual, times2)
	})
}

func TestDispatch(t *testing.T) {
	var cm configManager
	req, err := model.NewMessage()
	if err != nil {
		t.Fatalf("create message failed")
	}

	convey.Convey("dispatch succeeded test", t, func() {
		patches := gomonkey.ApplyFuncReturn(methodSelect, &common.RespMsg{}).
			ApplyFuncReturn(modulemgr.SendMessage, nil)
		defer patches.Reset()

		cm.dispatch(req)
	})

	convey.Convey("dispatch handle failed test", t, func() {
		cm.dispatch(req)
	})

	convey.Convey("dispatch create response test", t, func() {
		patches := gomonkey.ApplyFuncReturn(methodSelect, &common.RespMsg{}).
			ApplyMethodReturn(&model.Message{}, "NewResponse", nil, errors.New("test error"))
		defer patches.Reset()

		cm.dispatch(req)
	})

	convey.Convey("dispatch fill content failed test", t, func() {
		patches := gomonkey.ApplyFuncReturn(methodSelect, &common.RespMsg{}).
			ApplyMethodReturn(&model.Message{}, "FillContent", errors.New("test error"))
		defer patches.Reset()

		cm.dispatch(req)
	})

	convey.Convey("dispatch fill content failed test", t, func() {
		patches := gomonkey.ApplyFuncReturn(methodSelect, &common.RespMsg{}).
			ApplyFuncReturn(modulemgr.SendMessage, errors.New("test error"))
		defer patches.Reset()

		cm.dispatch(req)
	})
}
