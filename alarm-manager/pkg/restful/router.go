// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful this file is for setup router
package restful

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"alarm-manager/pkg/types"
	"alarm-manager/pkg/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/restfulmgr"
)

const (
	pageNumberKey = "pageNum"
	pageSizeKey   = "pageSize"
	snKey         = "sn"
	groupIdKey    = "groupId"
	ifCenterKey   = "ifCenter"
)

var northAlarmDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/alarmmanager/v1": {
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/alarms",
			Method:       http.MethodGet,
			Destination:  common.AlarmManagerName}},
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/alarm",
			Method:       http.MethodGet,
			Destination:  common.AlarmManagerName}, "id", false},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/events",
			Method:       http.MethodGet,
			Destination:  common.AlarmManagerName}},
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/event",
			Method:       http.MethodGet,
			Destination:  common.AlarmManagerName}, "id", false},
	},
}

var edgeAlarmRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edge/alarm": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/report",
			Method:       http.MethodPost,
			Destination:  utils.AlarmModuleName,
		},
	},
}

func setRouter(engine *gin.Engine) {
	restfulmgr.InitRouter(engine, edgeAlarmRouterDispatchers)
	restfulmgr.InitRouter(engine, northAlarmDispatchers)
}

type queryDispatcher struct {
	restfulmgr.GenericDispatcher
	name     string
	isString bool
}

type listDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (query queryDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	if query.isString {
		return getStringReqPara(c, query.name)
	}
	return getUintReqPara(c, query.name)
}

func getStringReqPara(c *gin.Context, paraName string) (string, error) {
	value := c.Query(paraName)
	if value == "" {
		return "", fmt.Errorf("req string para [%s] is invalid", paraName)
	}
	return value, nil
}

func getUintReqPara(c *gin.Context, paraName string) (uint64, error) {
	value, err := strconv.ParseUint(c.Query(paraName), common.BaseHex, common.BitSize64)
	if err != nil {
		return 0, fmt.Errorf("req int para [%s] is invalid", paraName)
	}
	return value, nil
}

func (list listDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	pageNum, pageNumErr := strconv.ParseUint(c.Query(pageNumberKey), common.BaseHex, common.BitSize64)
	pageSize, pageSizeErr := strconv.ParseUint(c.Query(pageSizeKey), common.BaseHex, common.BitSize64)
	if pageSizeErr != nil || pageNumErr != nil {
		return nil, fmt.Errorf("pageNum[%s] or pageSize[%s] is invalid",
			c.Query(pageNumberKey), c.Query(pageSizeKey))
	}
	values := c.Request.URL.Query()
	ifCenter := values.Get(ifCenterKey)
	groupIdStr := values.Get(groupIdKey)
	if groupIdStr == "0" {
		return nil, fmt.Errorf("groupId cannot be 0")
	}
	if groupIdStr == "" {
		groupIdStr = "0"
	}

	snStr := values.Get(snKey)
	if snStr == "" && len(values[snKey]) != 0 {
		return nil, fmt.Errorf("sn cannot be empty string")
	}

	if len(values[snKey]) == 0 {
		snStr = alarms.CenterSn
	}

	groupId, err2 := strconv.ParseUint(groupIdStr, common.BaseHex, common.BitSize64)
	if err2 != nil {
		return nil, fmt.Errorf("groupId[%s] is invalid", c.Query(groupIdKey))
	}
	return types.ListAlarmOrEventReq{PageNum: pageNum, PageSize: pageSize, Sn: snStr, GroupId: groupId,
		IfCenter: ifCenter}, nil
}
