// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.

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
