// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package httpsmgr for https manager
package httpsmgr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"huawei.com/mindxedge/base/common/certutils"
)

const (
	jsonContentType = "application/json"
	timeOut         = 60 * time.Second
)

// GetHttpsReq [method] for get https request
func GetHttpsReq(url string, tlsCert certutils.TlsCertInfo) *HttpsRequest {
	return &HttpsRequest{
		url:     url,
		tlsCert: tlsCert,
	}
}

// HttpsRequest [struct] for Https Request parameters
type HttpsRequest struct {
	url     string
	tlsCert certutils.TlsCertInfo
	client  *http.Client
}

func (hr *HttpsRequest) initClient() error {
	hr.tlsCert.SvrFlag = false
	tlsCfg, err := certutils.GetTlsCfgWithPath(hr.tlsCert)
	if err != nil {
		return err
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsCfg,
			TLSHandshakeTimeout: timeOut,
		},
		Timeout: timeOut,
	}
	hr.client = client
	return nil
}

// Get [method] for http get methods request
func (hr *HttpsRequest) Get() ([]byte, error) {
	if hr.client == nil {
		err := hr.initClient()
		if err != nil {
			return nil, fmt.Errorf("init https client failed: %v", err)
		}
	}
	resp, err := hr.client.Get(hr.url)
	if err != nil {
		return nil, err
	}
	defer hr.client.CloseIdleConnections()
	return hr.handleResp(resp)
}

// PostJson [method] for http Post request with json body
func (hr *HttpsRequest) PostJson(jsonBody []byte) ([]byte, error) {
	if hr.client == nil {
		err := hr.initClient()
		if err != nil {
			return nil, fmt.Errorf("init https client failed: %v", err)
		}
	}
	resp, err := hr.client.Post(hr.url, jsonContentType, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	defer hr.client.CloseIdleConnections()
	return hr.handleResp(resp)
}

func (hr *HttpsRequest) handleResp(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("http response is nil")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("https return error status code: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}
