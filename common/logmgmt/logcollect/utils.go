// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logcollect  provides utils for log collection
package logcollect

import (
	"fmt"
	"time"
)

const (
	uploadTimeStamp = "2006-01-02T15-04-05.000"
)

// GetLogPackFileName get log pack file name
func GetLogPackFileName(module, node string) string {
	return fmt.Sprintf("%s_%s_%s.tar.gz", module, node, time.Now().Format(uploadTimeStamp))
}
