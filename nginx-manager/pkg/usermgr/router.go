// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package usermgr this package is for manage user
package usermgr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/restfulmgr"

	"nginx-manager/pkg/nginxcom"
)

var userMgrPath = "/usermanager/v1"

var routers = map[string][]restfulmgr.DispatcherItf{
	userMgrPath: {
		userDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/login",
			Method:       http.MethodPost,
			Destination:  nginxcom.UserManagerName}},
		userDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/first-change",
			Method:       http.MethodPatch,
			Destination:  nginxcom.UserManagerName}},
		userDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/change",
			Method:       http.MethodPatch,
			Destination:  nginxcom.UserManagerName}},
		userDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/islocked",
			Method:       http.MethodPatch,
			Destination:  nginxcom.UserManagerName}},
	},
}

type userDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (dispatcher userDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	data, err := c.GetRawData()
	if err != nil {
		hwlog.RunLog.Error("gin get raw data failed")
		return "", fmt.Errorf("gin get raw data failed")
	}
	req := make(map[string]interface{})
	if err := common.ParamConvert(string(data), &req); err != nil {
		return "", err
	}
	req["ip"] = c.ClientIP()
	reqBytes, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Error("get input parameter failed")
		return "", fmt.Errorf("get input parameter failed")
	}
	reqStr := string(reqBytes)
	return reqStr, err
}
