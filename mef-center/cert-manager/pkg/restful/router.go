// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package restful this file is for setup router
package restful

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/restfulmgr"

	"cert-manager/pkg/certmanager"
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
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/info",
			Method:       http.MethodGet,
			Destination:  common.CertManagerName}, "certName"},
	},
	"/certmanager/v1/crl": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/import",
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
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/crl",
			Method:       http.MethodGet,
			Destination:  common.CertManagerName}, "crlName"},
		restfulmgr.GenericDispatcher{
			RelativePath: "/update-result",
			Method:       http.MethodPost,
			Destination:  common.CertManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/imported-certs",
			Method:       http.MethodGet,
			Destination:  common.CertManagerName},
	},
}

func setRouter(engine *gin.Engine) {
	engine.GET("/certmanager/v1/export", certmanager.ExportRootCa)
	restfulmgr.InitRouter(engine, certRouterDispatchers)
	restfulmgr.InitRouter(engine, innerCertRouterDispatchers)
}

type queryDispatcher struct {
	restfulmgr.GenericDispatcher
	name string
}

func (query queryDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	return c.Query(query.name), nil
}
