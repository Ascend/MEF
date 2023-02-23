// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

func downloadCert(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.RunLog.Error("download cert: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeMsgManagerName,
		Option:      common.OptGet,
		Resource:    common.ResDownLoadCert,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
