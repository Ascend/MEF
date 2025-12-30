// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package requests for https requests about cert
package requests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
)

const (
	reqSvrUrl               = "inner/v1/certificates/service"
	getRootCaUrl            = "inner/v1/certificates/rootca"
	getCrlUrl               = "inner/v1/certificates/crl"
	getImportedCertsInfoUrl = "inner/v1/certificates/imported-certs"
	updateCertUrl           = "inner/v1/image/update"
)

type reqIssueCertBody struct {
	CertName string `json:"certName"`
	Csr      string `json:"csr"`
}

// ImportedCertsInfo [struct] for getting imported certs info req params
type ImportedCertsInfo struct {
	NorthCert    []byte `json:"northCert"`
	SoftwareCert []byte `json:"softwareCert"`
	ImageCert    []byte `json:"imageCert"`
}

// ReqCertParams [struct] for req cert params
type ReqCertParams struct {
	ClientTlsCert certutils.TlsCertInfo
}

// GetRootCa [method] for get root ca with certName
func (rcp *ReqCertParams) GetRootCa(certName string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s?certName=%s", common.CertMgrDns, common.CertMgrPort,
		getRootCaUrl, certName)
	httpsReq := httpsmgr.GetHttpsReq(url, rcp.ClientTlsCert)
	resp, err := httpsReq.Get(nil)
	if err != nil {
		return "", err
	}
	return rcp.parseResp(resp)
}

// GetCrl [method] for get crl with crlName
func (rcp *ReqCertParams) GetCrl(crlName string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s?crlName=%s", common.CertMgrDns, common.CertMgrPort,
		getCrlUrl, crlName)
	httpsReq := httpsmgr.GetHttpsReq(url, rcp.ClientTlsCert)
	const timeout = 3 * time.Second
	resp, err := httpsReq.GetWithTimeout(nil, timeout)
	if err != nil {
		return "", err
	}
	return rcp.parseResp(resp)
}

// UpdateCertFile [method] for update cert content
func (rcp *ReqCertParams) UpdateCertFile(cert certutils.UpdateClientCert) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s", common.EdgeMgrDns, common.EdgeMgrPort, updateCertUrl)
	httpsReq := httpsmgr.GetHttpsReq(url, rcp.ClientTlsCert)
	jsonBody, err := json.Marshal(&cert)
	if err != nil {
		return "", err
	}
	respBytes, err := httpsReq.PostJson(jsonBody)
	if err != nil {
		return "", err
	}
	return rcp.parseResp(respBytes)
}

// ReqIssueSvrCert [method] for issue server cert
func (rcp *ReqCertParams) ReqIssueSvrCert(certName string, csr []byte) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s", common.CertMgrDns, common.CertMgrPort, reqSvrUrl)
	httpsReq := httpsmgr.GetHttpsReq(url, rcp.ClientTlsCert)
	issueCertBody := &reqIssueCertBody{
		CertName: certName,
		Csr:      base64.StdEncoding.EncodeToString(certutils.PemWrapCert(csr)),
	}
	jsonBody, err := json.Marshal(issueCertBody)
	if err != nil {
		return "", err
	}
	respBytes, err := httpsReq.PostJson(jsonBody)
	if err != nil {
		return "", err
	}
	return rcp.parseResp(respBytes)
}

// GetImportedCertsInfo [method] for getting imported certs info
func (rcp *ReqCertParams) GetImportedCertsInfo() (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s", common.CertMgrDns, common.CertMgrPort, getImportedCertsInfoUrl)
	httpsReq := httpsmgr.GetHttpsReq(url, rcp.ClientTlsCert)
	resp, err := httpsReq.Get(nil)
	if err != nil {
		return "", err
	}
	return rcp.parseResp(resp)
}

func (rcp *ReqCertParams) parseResp(respBytes []byte) (string, error) {
	var resp common.RespMsg
	err := json.Unmarshal(respBytes, &resp)
	if err != nil {
		return "", err
	}

	status := resp.Status
	if status != common.Success {
		return "", fmt.Errorf("parse cert response failed: status=%s, msg=%s", status, resp.Msg)
	}
	data, ok := resp.Data.(string)
	if !ok {
		return "", fmt.Errorf("param data is not string")
	}
	return data, nil
}
