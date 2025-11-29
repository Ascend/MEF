// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package httpsmgr
package httpsmgr

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

func serializable() func(ctx *gin.Context) {
	var lock sync.Mutex
	return func(ctx *gin.Context) {
		locked := lock.TryLock()
		if !locked {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		defer lock.Unlock()

		ctx.Next()
	}
}
