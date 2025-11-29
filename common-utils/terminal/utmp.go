// Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package terminal provides get login user and ip
package terminal

/*
	#include <stdlib.h>
	#include <string.h>
	#include <utmp.h>
	#include "go_utmp.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// GetLoginUserAndIP get the terminal login info ,like who am i
func GetLoginUserAndIP() (string, string, error) {
	cUtName := C.CString(string(make([]byte, int32(C.UT_NAMESIZE))))
	defer C.free(unsafe.Pointer(cUtName))
	cUtHost := C.CString(string(make([]byte, int32(C.UT_HOSTSIZE))))
	defer C.free(unsafe.Pointer(cUtHost))
	if ret := C.GetSSHIP(cUtName, cUtHost); ret != 0 {
		return "", "", fmt.Errorf("get ssh ip failed,res code is:%d", int32(ret))
	}
	return C.GoString(cUtName), C.GoString(cUtHost), nil
}
