// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
		hwlog.RunLog.Errorf("failed to handle download, file check failed, %v", err)
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	ctx.File(targetFile)
}
