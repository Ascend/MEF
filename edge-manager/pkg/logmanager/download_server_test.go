// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logmanager
package logmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
)

// TestHandleDownload test handle download
func TestHandleDownload(t *testing.T) {
	convey.Convey("test handle download", t, func() {
		ctx := &gin.Context{}
		patch := gomonkey.ApplyMethodReturn(ctx, "File").
			ApplyMethodReturn(ctx, "AbortWithStatusJSON")
		defer patch.Reset()

		HandleDownload(ctx)
	})
}
