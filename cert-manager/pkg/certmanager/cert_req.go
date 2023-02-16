// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager cert manager module
package certmanager

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
