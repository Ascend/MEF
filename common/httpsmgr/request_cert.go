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
	reqSvrUrl         = "certmanager/v1/certificates/service"
	getRootCaUrl      = "certmanager/v1/certificates/rootca"
	getCertContentUrl = "certmanager/v1/certificates/cert"
	updateCertUrl     = "edgemanager/v1/image/update"
)

type reqIssueCertBody struct {
	CertName string `json:"certName"`
	Csr      string `json:"csr"`
}

type respBody struct {
	Status string `json:"Status"`
	Msg    string `json:"Msg"`
	Data   string `json:"Data"`
}

// ReqCertParams [struct] for req cert params
type ReqCertParams struct {
	ClientTlsCert certutils.TlsCertInfo
}

// GetRootCa [method] for get root ca with certName
func (rcp *ReqCertParams) GetRootCa(certName string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s/?certName=%s", common.CertMgrDns, common.CertMgrPort,
		getRootCaUrl, certName)
	httpsReq := GetHttpsReq(url, rcp.ClientTlsCert)
	resp, err := httpsReq.Get()
	if err != nil {
		return "", err
	}
	return rcp.parseResp(resp)
}

// GetCertFile [method] for get cert content with certName
func (rcp *ReqCertParams) GetCertFile(certName string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s/?certName=%s", common.CertMgrDns, common.CertMgrPort,
		getCertContentUrl, certName)
	httpsReq := GetHttpsReq(url, rcp.ClientTlsCert)
	resp, err := httpsReq.Get()
	if err != nil {
		return "", err
	}
	return rcp.parseMsg(resp)
}

// UpdateCertFile [method] for update cert content
func (rcp *ReqCertParams) UpdateCertFile(certName string) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s", common.EdgeMgrDns, common.EdgeMgrPort, updateCertUrl)
	httpsReq := GetHttpsReq(url, rcp.ClientTlsCert)
	jsonBody, err := json.Marshal(certName)
	if err != nil {
		return "", err
	}
	respBytes, err := httpsReq.PostJson(jsonBody)
	if err != nil {
		return "", err
	}
	return rcp.parseMsg(respBytes)
}

// ReqIssueSvrCert [method] for issue server cert
func (rcp *ReqCertParams) ReqIssueSvrCert(certName string, csr []byte) (string, error) {
	url := fmt.Sprintf("https://%s:%d/%s", common.CertMgrDns, common.CertMgrPort, reqSvrUrl)
	httpsReq := GetHttpsReq(url, rcp.ClientTlsCert)
	issueCertBody := &reqIssueCertBody{CertName: certName, Csr: string(certutils.PemWrapCert(csr))}
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

func (rcp *ReqCertParams) parseMsg(respBytes []byte) (string, error) {
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
