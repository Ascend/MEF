// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package httpsmgr for
package httpsmgr

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"huawei.com/mindx/common/hwlog"
)

const kilo = 1000.0

// LoggerAdapter  for gin framework
func LoggerAdapter() gin.HandlerFunc {
	return func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		startTime := time.Now()
		c.Next()
		stopTime := time.Since(startTime)
		duration := int(math.Ceil(float64(stopTime.Nanoseconds()) / kilo / kilo))
		urlStatus := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		dataLength := c.Writer.Size()

		if dataLength < 0 {
			dataLength = 0
		}
		if len(c.Errors) > 0 {
			hwlog.RunLog.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s: %s <%3d> (%dms) | %15s | %s ",
				c.Request.Method, urlPath, urlStatus, duration, clientIP, clientUserAgent)
			if urlStatus >= http.StatusInternalServerError {
				hwlog.OpLog.Error(msg)
			} else if urlStatus >= http.StatusBadRequest {
				hwlog.OpLog.Warn(msg)
			} else {
				hwlog.OpLog.Info(msg)
			}
		}
	}
}
