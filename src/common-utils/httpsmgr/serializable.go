// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
