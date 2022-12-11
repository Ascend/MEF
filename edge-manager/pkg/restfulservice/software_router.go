// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

func upgradeSoftware(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("upgrade software: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.EdgeInstallerName,
		Option:      common.Upgrade,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
