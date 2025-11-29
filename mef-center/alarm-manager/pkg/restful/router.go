// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful this file is for setup router
package restful

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"alarm-manager/pkg/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/restfulmgr"
)

const (
	pageNumberKey = "pageNum"
	pageSizeKey   = "pageSize"
	snKey         = "sn"
	groupIdKey    = "groupId"
	ifCenterKey   = "ifCenter"
)

var alarmRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
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

func setRouter(engine *gin.Engine) {
	restfulmgr.InitRouter(engine, alarmRouterDispatchers)
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
	// don't allow empty values
	if isKeyAssignedToEmpty(values, ifCenterKey) || isKeyAssignedToEmpty(values, groupIdKey) ||
		isKeyAssignedToEmpty(values, snKey) {
		return nil, fmt.Errorf("params in [%s,%s,%s] cannot be assigned to empty string",
			ifCenterKey, groupIdKey, snKey)
	}
	ifCenter := values.Get(ifCenterKey)
	if ifCenter == utils.TrueStr {
		return utils.ListAlarmOrEventReq{PageNum: pageNum, PageSize: pageSize, IfCenter: ifCenter}, nil
	}

	groupIdStr := values.Get(groupIdKey)
	if groupIdStr == "0" {
		return nil, fmt.Errorf("groupId cannot be assigned to 0")
	}
	// ensure construction of ListAlarmOrEventReq{},"" cannot parse to int
	if groupIdStr == "" {
		groupIdStr = "0"
	}
	groupId, err := strconv.ParseUint(groupIdStr, common.BaseHex, common.BitSize64)
	if err != nil {
		return nil, fmt.Errorf("groupId[%s] is invalid", c.Query(groupIdKey))
	}

	snStr := values.Get(snKey)
	return utils.ListAlarmOrEventReq{PageNum: pageNum, PageSize: pageSize, Sn: snStr, GroupId: groupId,
		IfCenter: ifCenter}, nil
}

func isKeyAssignedToEmpty(values url.Values, keyName string) bool {
	// values.Get returns the first result in []string or "" if key not found
	// values[key] returns []string
	return len(values[keyName]) != 0 && values.Get(keyName) == ""
}
