// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package logmanager
package logmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/logmanager/tasks"
	"edge-manager/pkg/logmanager/utils"
)

// TestStart test module start
func TestStart(t *testing.T) {
	convey.Convey("test logMgr start", t, func() {

		ctx, cancel := context.WithCancel(context.Background())
		var count int
		mockReceiveMessage := func(string) (*model.Message, error) {
			count++
			if count == 1 {
				return model.NewMessage()
			}
			cancel()
			return nil, errors.New("test error")
		}
		patch := gomonkey.ApplyFunc(modulemgr.ReceiveMessage, mockReceiveMessage)
		defer patch.Reset()

		mgr := &logMgr{ctx: ctx}
		mgr.Start()
	})
}

// TestEnable test enable
func TestEnable(t *testing.T) {
	convey.Convey("test logMgr enable", t, func() {
		patch := gomonkey.ApplyFuncReturn(tasks.InitTasks, nil).
			ApplyFuncReturn(utils.CleanTempFiles, true, nil)
		defer patch.Reset()

		mgr := &logMgr{}
		mgr.Enable()
	})
}
