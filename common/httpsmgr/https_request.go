// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package httpsmgr for https manager
package httpsmgr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common/certutils"
)

const (
	jsonContentType = "application/json"
	timeOut         = 60 * time.Second
	maxBodySize     = 2 * 1024 * 1024
)

// GetHttpsReq [method] for get https request
func GetHttpsReq(url string, tlsCert certutils.TlsCertInfo, headers ...map[string]interface{}) *HttpsRequest {
	req := &HttpsRequest{
		url:     url,
		tlsCert: tlsCert,
	}
	if len(headers) > 0 {
		req.reqHeader = headers[0]
	}
	return req
}

// HttpsRequest [struct] for Https Request parameters
type HttpsRequest struct {
	url       string
	tlsCert   certutils.TlsCertInfo
	client    *http.Client
	reqHeader map[string]interface{}
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
		if err := hr.initClient(); err != nil {
			return nil, fmt.Errorf("init https client failed: %v", err)
		}
	}
	req, err := http.NewRequest(http.MethodGet, hr.url, nil)
	if len(hr.reqHeader) > 0 {
		for k, v := range hr.reqHeader {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	resp, err := hr.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer hr.client.CloseIdleConnections()
	return hr.handleResp(resp)
}

// PostJson [method] for http Post request with json body
func (hr *HttpsRequest) PostJson(jsonBody []byte) ([]byte, error) {
	if hr.client == nil {
		if err := hr.initClient(); err != nil {
			return nil, fmt.Errorf("init https client failed: %v", err)
		}
	}
	req, err := http.NewRequest(http.MethodPost, hr.url, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", jsonContentType)
	if len(hr.reqHeader) > 0 {
		for k, v := range hr.reqHeader {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	resp, err := hr.client.Do(req)
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
	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return readBytes, nil
}

// GetRespToFileByLimit [method] for http get resp to file
func (hr *HttpsRequest) GetRespToFileByLimit(writer io.Writer, limit int64) error {
	if hr.client == nil {
		if err := hr.initClient(); err != nil {
			return fmt.Errorf("init https client failed: %v", err)
		}
	}
	req, err := http.NewRequest(http.MethodGet, hr.url, nil)
	if len(hr.reqHeader) > 0 {
		for k, v := range hr.reqHeader {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	resp, err := hr.client.Do(req)
	if err != nil {
		return err
	}
	defer hr.client.CloseIdleConnections()
	return hr.handleRespToWriter(resp, writer, limit)
}

func (hr *HttpsRequest) handleRespToWriter(resp *http.Response, writer io.Writer, limit int64) error {
	if resp == nil {
		return fmt.Errorf("http response is nil")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}()

	if resp.ContentLength > limit {
		return fmt.Errorf("response content length up to limit")
	}

	if _, err := io.Copy(writer, io.LimitReader(resp.Body, limit)); err != nil {
		hwlog.RunLog.Errorf("write https resp to writer failed, error: %v", err)
		return err
	}

	return nil
}
