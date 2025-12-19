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
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/constants"
)

// HandleDownload handle download request
func HandleDownload(ctx *gin.Context) {
	targetFile := filepath.Join(constants.LogDumpPublicDir, constants.EdgeNodesTarGzFileName)
	if _, err := fileutils.CheckOriginPath(targetFile); err != nil {
		hwlog.RunLog.Errorf("failed to handle download, file is abnormal, %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, "file is abnormal")
		return
	}
	ctx.File(targetFile)
}
