// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
