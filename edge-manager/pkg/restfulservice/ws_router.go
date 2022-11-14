// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"edge-manager/pkg/common"
	"edge-manager/pkg/util"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
)

func upgradeSfw(c *gin.Context) {
	var upgradeSfwReq util.UpgradeSfwReq
	if err := c.ShouldBindJSON(&upgradeSfwReq); err != nil {
		hwlog.OpLog.Error("update software: convert request body failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.EdgeInstallerName,
		option:      common.Upgrade,
		resource:    common.Software,
	}
	resp := sendSyncMessageByRestful(upgradeSfwReq, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
