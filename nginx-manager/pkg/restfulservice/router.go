// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

const (
	urlNgxEdgeMgrCert = "/inner/v1/ngxmanager/cert/edge-manager"
	maxBodySize       = common.MB * 5
)

func setRouter(engine *gin.Engine) {
	engine.POST(urlNgxEdgeMgrCert, updateEdgeMgrSouthCert)
}

func updateEdgeMgrSouthCert(ctx *gin.Context) {
	errMsg := "handle cert update operation error"
	bodyData, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxBodySize))
	if err != nil {
		hwlog.RunLog.Errorf("read http request body error: %v", err)
		ctx.JSON(http.StatusBadRequest, errMsg)
		return
	}
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("generate module message error: %v", err)
		ctx.JSON(http.StatusBadRequest, errMsg)
		return
	}
	msg.SetRouter(
		common.RestfulServiceName,
		common.CertUpdaterName,
		common.OptPost,
		common.ResEdgeMgrCertUpdate)
	if err = msg.FillContent(bodyData); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		ctx.JSON(http.StatusBadRequest, errMsg)
		return
	}
	// this sync message timeout must be greater than nginxReloadConfTimeout
	resp, err := modulemgr.SendSyncMessage(msg, time.Minute)
	if err != nil {
		hwlog.RunLog.Errorf("forward cert update message to module cert-updater error: %v", err)
		ctx.JSON(http.StatusBadRequest, errMsg)
		return
	}
	var content string
	if err = resp.ParseContent(&content); err != nil {
		hwlog.RunLog.Error("response message from module cert-manager content type error, expect string")
		ctx.JSON(http.StatusBadRequest, errMsg)
		return
	}
	if content != common.OK {
		hwlog.RunLog.Errorf("update nginx south cert error: %v", content)
		ctx.JSON(http.StatusBadRequest, errMsg)
		return
	}
	ctx.JSON(http.StatusOK, common.NewOkRespMsg())
}
