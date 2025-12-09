// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package handlers for testing pack log result handler
package handlers

import (
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

func TestPackLogResultHandler(t *testing.T) {
	convey.Convey("pack log result handler should be failed", t, func() {
		handler := packLogResultHandler{}
		packLogResponse, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
			Source:      constants.EdgeOm,
			Destination: constants.InnerClient,
			Operation:   constants.OptResponse,
			Resource:    constants.ResPackLogResponse,
			Content:     constants.OK,
		})
		convey.So(err, convey.ShouldBeNil)

		err = handler.Handle(packLogResponse)
		convey.So(err, convey.ShouldResemble, errors.New("failed to send pack result"))
	})
}
