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

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

// printHelper helper function for log printing
func printHelper(lg *log.Logger, msg string, maxLogLength int, escape bool, ctx ...context.Context) {
	str := getCallerInfo(ctx...)
	trimMsg := ""
	if escape {
		trimMsg = strings.TrimRight(msg, "\r\n")
		trimMsg = escapeHtml(trimMsg)
	} else {
		trimMsg = strings.Replace(msg, "\r", " ", -1)
		trimMsg = strings.Replace(trimMsg, "\n", " ", -1)
	}
	runeArr := []rune(trimMsg)
	if length := len(runeArr); length > maxLogLength {
		trimMsg = string(runeArr[:maxLogLength])
	}

	zone, _ := time.Now().Zone()
	lg.Println(zone + ` ` + str + trimMsg)
}

func escapeHtml(msg string) string {
	const (
		minLen      = 3
		headRmCount = 1
		tailRmCount = 2
	)

	msg = strings.Replace(msg, "\u007f", "", -1)
	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	// if encode error happens, do nothing
	if err := encoder.Encode(msg); err != nil {
		return msg
	}
	encStr := bf.String()
	if len(encStr) < minLen {
		return msg
	}
	// rm the head " and tail "\n which added by json encoder
	trimMsg := encStr[headRmCount : len(encStr)-tailRmCount]
	return trimMsg
}

// getCallerInfo gets the caller's information
func getCallerInfo(ctx ...context.Context) string {
	var deep = stackDeep
	var userID interface{}
	var traceID interface{}
	for _, c := range ctx {
		if c == nil {
			deep++
			continue
		}
		userID = c.Value(UserID)
		traceID = c.Value(ReqID)
	}
	var funcName string
	pc, codePath, codeLine, ok := runtime.Caller(deep)
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}
	p := strings.Split(codePath, "/")
	l := len(p)
	if l == pathLen {
		funcName = p[l-1]
	} else if l > pathLen {
		funcName = fmt.Sprintf("%s/%s", p[l-pathLen], p[l-1])
	}
	callerPath := fmt.Sprintf("%s:%d", funcName, codeLine)
	goroutineID := getGoroutineID()
	str := fmt.Sprintf("%-8s%s    ", goroutineID, callerPath)
	if userID != nil || traceID != nil {
		str = fmt.Sprintf("%s{%#v}-{%#v} ", str, userID, traceID)
	}
	return str
}

// getCallerGoroutineID gets the goroutineID
func getGoroutineID() string {
	b := make([]byte, bitsize, bitsize)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	return string(b)
}
