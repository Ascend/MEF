// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certmanager cert manager module
package certmanager

const (
	updateSuccessCode int64 = 1
	updateFailedCode  int64 = 2
)

// importCertReq import ca req
type importCertReq struct {
	CertName string `json:"certName" binding:"required,oneof=software image res_file apig alarm edge_core device_plugin"`
	Cert     string `json:"cert" binding:"required,gte=1,lte=2000000"`
}

// deleteCaReq delete ca req
type deleteCaReq struct {
	Type string `json:"type"`
}

type csrJson struct {
	CertName string `json:"certName"`
	Csr      string `json:"csr"`
}

type importCrlReq struct {
	CrlName string `json:"crlName"`
	Crl     string `json:"crl"`
}

type certUpdateResult struct {
	CertType   string `json:"certType"`
	ResultCode int64  `json:"resultCode"`
	Desc       string `json:"desc"`
}
