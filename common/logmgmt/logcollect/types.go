// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logcollect provides utils for log collection
package logcollect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindxedge/base/common"
)

const (
	// ProgressAbort indicates an error
	ProgressAbort = -1
	// ProgressMax is the maximum value of progress
	ProgressMax = 100

	// EdgeMaxPackSize max size of pack size
	EdgeMaxPackSize = 200 * common.MB
	// EdgeMaxFileSize max size of file size
	EdgeMaxFileSize = 50 * common.MB
	// CenterMaxPackSize max size of pack size
	CenterMaxPackSize = 10 * 1024 * common.MB
	// CenterMaxFileSize max size of file size
	CenterMaxFileSize = 1 * 1024 * common.MB
	// CenterLogExportDir center log exports dir
	CenterLogExportDir = "/var/log_exports/center"
	// EdgeLogExportDir edge log exports dir
	EdgeLogExportDir = "/var/log_exports/edge"

	urlSplitCount = 2
)

const (
	// ModuleCenter center
	ModuleCenter = "center"
	// ModuleEdge edge
	ModuleEdge = "edge"
)

// TaskProgress defines log collection progress
type TaskProgress struct {
	Progress int    `json:"progress"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// UploadConfig defines configuration for uploading logs
type UploadConfig struct {
	MethodAndUrl MethodAndUrl `json:"url"`
}

// MethodAndUrl defines method and url
type MethodAndUrl struct {
	Method string
	Url    string
}

// MarshalJSON implements Marshaller for MethodAndUrl
func (s MethodAndUrl) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	if err := encoder.Encode(fmt.Sprintf("%s %s", s.Method, s.Url)); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// UnmarshalJSON implements Unmarshaler for MethodAndUrl
func (s *MethodAndUrl) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var methodAndUrl string
	if err := decoder.Decode(&methodAndUrl); err != nil {
		return err
	}
	parts := strings.Split(methodAndUrl, " ")
	if len(parts) != urlSplitCount {
		return errors.New("bad format for method and url")
	}
	s.Method, s.Url = parts[0], parts[1]
	return nil
}

// CreateTaskReq defines request for creating tasks
type CreateTaskReq struct {
	Module      string       `json:"module"`
	EdgeNodes   []string     `json:"nodes"`
	HttpsServer UploadConfig `json:"httpsServer"`
}

// BatchQueryTaskReq defines request for querying task info
type BatchQueryTaskReq struct {
	Module    string   `json:"module"`
	EdgeNodes []string `json:"nodes"`
}

// QueryTaskResp defines response for querying task info
type QueryTaskResp struct {
	Module   string      `json:"module"`
	EdgeNode string      `json:"node"`
	Data     interface{} `json:"data"`
}
