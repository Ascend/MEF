// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful this file is for setup router
package restful

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"cert-manager/pkg/certmanager"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/restfulmgr"
)

var certRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/certmanager/v1/certificates": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/import",
			Method:       http.MethodPost,
			Destination:  common.CertManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/delete-cert",
			Method:       http.MethodPost,
			Destination:  common.CertManagerName},
	},
}

var innerCertRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/inner/v1/certificates": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/service",
			Method:       http.MethodPost,
			Destination:  common.CertManagerName},
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/rootca",
			Method:       http.MethodGet,
			Destination:  common.CertManagerName}, "certName"},
	},
}

func setRouter(engine *gin.Engine) {
	engine.GET("/certmanager/v1/version", versionQuery)
	engine.GET("/certmanager/v1/export", certmanager.ExportRootCa)
	restfulmgr.InitRouter(engine, certRouterDispatchers)
	restfulmgr.InitRouter(engine, innerCertRouterDispatchers)
}

func versionQuery(c *gin.Context) {
	msg := fmt.Sprintf("%s version: %s", BuildNameStr, BuildVersionStr)
	common.ConstructResp(c, common.Success, "", msg)
}

type queryDispatcher struct {
	restfulmgr.GenericDispatcher
	name string
}

func (query queryDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	return c.Query(query.name), nil
}
