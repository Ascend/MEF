// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package kmc

/*
#cgo CFLAGS: -I./include -Wall -Wno-unused-function  -fstack-protector-all -fPIE -fPIC
#cgo LDFLAGS: -ldl -Wl,-z,relro -Wl,-z,noexecstack -fPIE
#include "kmc.h"
*/
import "C"
import (
	"unsafe"

	"huawei.com/mindx/common/hwlog"
)

var logger hwlog.CryptoLogger

//export goLoggerCallback
func goLoggerCallback(_ unsafe.Pointer, level C.LogLevel, msg *C.char) {
	if logger == nil {
		return
	}

	// msg memory managed by C, no need free manually
	s := C.GoString(msg)
	switch level {
	case C.LOG_ERROR:
		logger.Error(s)
	case C.LOG_WARN:
		logger.Warn(s)
	case C.LOG_INFO:
		logger.Info(s)
	case C.LOG_DEBUG:
		logger.Debug(s)
	case C.LOG_TRACE:
		logger.Trace(s)
	default:
	}
}
