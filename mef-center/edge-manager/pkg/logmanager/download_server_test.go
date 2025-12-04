// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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
