// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
