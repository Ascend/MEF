// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package httpsmgr for http
package httpsmgr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"
)

const (
	contentType        = "Content-Type"
	jsonContentType    = "application/json"
	binaryContentType  = "binary/octet-stream"
	timeOut            = 2 * time.Minute
	maxBodySize        = 2 * 1024 * 1024
	maxAllowedSize     = 10 * 1024
	defaultReadTimeOut = 2 * time.Minute
)

// HttpsRequest [struct] for Https Request parameters
type HttpsRequest struct {
	url         string
	tlsCert     certutils.TlsCertInfo
	client      *http.Client
	reqHeader   map[string]interface{}
	readTimeout time.Duration
}

// GetHttpsReq [method] for get https request
func GetHttpsReq(url string, tlsCert certutils.TlsCertInfo, headers ...map[string]interface{}) *HttpsRequest {
	req := &HttpsRequest{
		url:         url,
		tlsCert:     tlsCert,
		readTimeout: defaultReadTimeOut,
	}
	if len(headers) > 0 {
		req.reqHeader = headers[0]
	}
	return req
}

// Get [method] for http get methods request
func (hr *HttpsRequest) Get(body io.Reader) ([]byte, error) {
	if hr.client == nil {
		if err := hr.initClient(); err != nil {
			return nil, fmt.Errorf("init https client failed: %v", err)
		}
	}
	req, err := http.NewRequest(http.MethodGet, hr.url, body)
	if err != nil {
		return nil, err
	}
	if len(hr.reqHeader) > 0 {
		for k, v := range hr.reqHeader {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	resp, err := hr.client.Do(req)
	if err != nil {
		return nil, utils.TrimInfoFromError(err)
	}
	defer hr.client.CloseIdleConnections()
	return hr.handleResp(resp)
}

// GetWithTimeout method send get request with timeout available to set
func (hr *HttpsRequest) GetWithTimeout(body io.Reader, timeout time.Duration) ([]byte, error) {
	if hr.client == nil {
		if err := hr.initClient(); err != nil {
			return nil, fmt.Errorf("init https client failed: %v", err)
		}
	}
	hr.client.Timeout = timeout
	transport, ok := hr.client.Transport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("client transport type is incorrect")
	}
	transport.TLSHandshakeTimeout = timeout

	defer func() {
		hr.client.Timeout = timeOut
		transport.TLSHandshakeTimeout = timeOut
	}()
	return hr.Get(body)
}

// SetReadTimeout method set the read time out for https request
func (hr *HttpsRequest) SetReadTimeout(timeout time.Duration) *HttpsRequest {
	hr.readTimeout = timeout
	return hr
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
		Timeout: hr.readTimeout,
	}
	hr.client = client
	return nil
}

func (hr *HttpsRequest) handleResp(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("http response is nil")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("https return error status code: %d", resp.StatusCode)
	}
	readBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, utils.TrimInfoFromError(err)
	}
	return readBytes, nil
}

// PostFile [method] for http Post file
func (hr *HttpsRequest) PostFile(filePath string) ([]byte, error) {
	if hr.client == nil {
		if err := hr.initClient(); err != nil {
			return nil, fmt.Errorf("init https client failed: %v", err)
		}
	}
	if _, err := fileutils.RealFileCheck(filePath, false, false, maxAllowedSize); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filePath, os.O_RDONLY, fileutils.Mode400)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			hwlog.RunLog.Errorf("failed to close the uploaded file, %v", err)
		}
	}()
	req, err := http.NewRequest(http.MethodPost, hr.url, file)
	if err != nil {
		return nil, err
	}
	req.Header.Set(contentType, binaryContentType)
	if len(hr.reqHeader) > 0 {
		for k, v := range hr.reqHeader {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	resp, err := hr.client.Do(req)
	if err != nil {
		return nil, utils.TrimInfoFromError(err)
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
	if err != nil {
		return nil, err
	}
	req.Header.Set(contentType, jsonContentType)
	if len(hr.reqHeader) > 0 {
		for k, v := range hr.reqHeader {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	resp, err := hr.client.Do(req)
	if err != nil {
		return nil, utils.TrimInfoFromError(err)
	}

	defer hr.client.CloseIdleConnections()
	return hr.handleResp(resp)
}

// GetRespToFileWithLimit [method] for http get resp to file
func (hr *HttpsRequest) GetRespToFileWithLimit(writer io.Writer, limit int64) error {
	if hr.client == nil {
		if err := hr.initClient(); err != nil {
			return fmt.Errorf("init https client failed: %v", err)
		}
	}
	req, err := http.NewRequest(http.MethodGet, hr.url, nil)
	if err != nil {
		return err
	}
	if len(hr.reqHeader) > 0 {
		for k, v := range hr.reqHeader {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	resp, err := hr.client.Do(req)
	if err != nil {
		return utils.TrimInfoFromError(err)
	}
	defer hr.client.CloseIdleConnections()
	return hr.handleRespToFile(resp, writer, limit)
}

func (hr *HttpsRequest) handleRespToFile(resp *http.Response, writer io.Writer, limit int64) error {
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
		return utils.TrimInfoFromError(err)
	}

	return nil
}
