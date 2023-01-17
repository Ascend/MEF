// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common for result
package common

import "fmt"

// Result [struct] to record result
type Result struct {
	ResultFlag bool
	Data       interface{}
	ErrorMsg   string
}

func (r Result) String() string {
	return fmt.Sprintf("result=%v; errorMsg=%s", r.ResultFlag, r.ErrorMsg)
}
