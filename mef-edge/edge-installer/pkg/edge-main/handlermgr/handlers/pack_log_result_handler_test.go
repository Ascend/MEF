// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
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
