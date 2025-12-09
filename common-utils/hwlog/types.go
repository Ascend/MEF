// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import "errors"

// ContextKey especially for context value
// to solve problem of "should not use basic type untyped string as key in context.WithValue"
type ContextKey string

// String  the implement of String method
func (c ContextKey) String() string {
	return string(c)
}

const (
	// UserID used for context value key of "ID"
	UserID ContextKey = "UserID"
	// ReqID used for context value key of "requestID"
	ReqID ContextKey = "RequestID"
)

// SelfLogWriter used this to replace some opensource log
type SelfLogWriter struct {
}

// Write  implement the interface of io.writer
func (l *SelfLogWriter) Write(p []byte) (int, error) {
	if RunLog == nil {
		return -1, errors.New("hwlog is not initialized")
	}
	RunLog.Info(string(p))
	return len(p), nil
}
