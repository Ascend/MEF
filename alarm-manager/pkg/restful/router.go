// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful this file is for setup router
package restful

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"alarm-manager/pkg/alarmmanager"

	"huawei.com/mindxedge/base/common/restfulmgr"
)

var innerAlarmRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/inner/v1/alarm": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/report",
			Method:       http.MethodPost,
			Destination:  alarmmanager.AlarmModuleName,
		},
	},
}

func setRouter(engine *gin.Engine) {
	restfulmgr.InitRouter(engine, innerAlarmRouterDispatchers)
}
