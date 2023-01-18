// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package httpsmgr for https manager
package httpsmgr

import (
	"encoding/json"
	"fmt"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
)

const (
	reqSvrUrl    = "certmanager/v1/certificates/service"
	getRootCaUrl = "certmanager/v1/certificates/rootca"
)

type reqIssueCertBody struct {
	certName string
	csr      string
}

type respBody struct {
	Status string `json:"Status"`
	Msg    string `json:"Msg"`
	Data   string `json:"Data"`
}

// ReqCertParams [struct] for req cert params
type ReqCertParams struct {
	Address       string
	Port          int
	ClientTlsCert certutils.TlsCertInfo
}

// GetRootCa [method] for get root ca with certName
func (rcp *ReqCertParams) GetRootCa(certName string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s/?certName=%s", rcp.Address, rcp.Port, getRootCaUrl, certName)
	httpsReq := GetHttpsReq(url, rcp.ClientTlsCert)
	resp, err := httpsReq.Get()
	if err != nil {
		return "", err
	}
	return rcp.parseResp(resp)
}

// ReqIssueSvrCert [method] for issue server cert
func (rcp *ReqCertParams) ReqIssueSvrCert(certName string, csr []byte) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s", rcp.Address, rcp.Port, reqSvrUrl)
	httpsReq := GetHttpsReq(url, rcp.ClientTlsCert)
	issueCertBody := &reqIssueCertBody{certName: certName, csr: string(certutils.PemWrapCert(csr))}
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
	var resp respBody
	err := json.Unmarshal(respBytes, &resp)
	if err != nil {
		return "", err
	}

	status := resp.Status
	if status != common.Success {
		return "", fmt.Errorf("parse cert response failed: status=%s, msg=%s", status, resp.Msg)
	}
	return resp.Data, nil
}
