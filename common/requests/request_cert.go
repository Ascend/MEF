// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package httpsmgr for https manager
package requests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
)

const (
	reqSvrUrl     = "inner/v1/certificates/service"
	getRootCaUrl  = "inner/v1/certificates/rootca"
	getCrlUrl     = "inner/v1/certificates/crl"
	updateCertUrl = "inner/v1/image/update"
)

type reqIssueCertBody struct {
	CertName string `json:"certName"`
	Csr      string `json:"csr"`
}

// ReqCertParams [struct] for req cert params
type ReqCertParams struct {
	ClientTlsCert certutils.TlsCertInfo
}

// GetRootCa [method] for get root ca with certName
func (rcp *ReqCertParams) GetRootCa(certName string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s/?certName=%s", common.CertMgrDns, common.CertMgrPort,
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
	url := fmt.Sprintf("https://%s:%d/%s/?crlName=%s", common.CertMgrDns, common.CertMgrPort,
		getCrlUrl, crlName)
	httpsReq := httpsmgr.GetHttpsReq(url, rcp.ClientTlsCert)
	resp, err := httpsReq.Get(nil)
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
