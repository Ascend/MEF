// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package terminal provide a safe reader for password
package terminal

import (
	"syscall"
	"unsafe"
)

// localReader is an io.Reader that reads from a specific file descriptor.
type localReader int

// Read read specical file from syscall
func (r localReader) Read(buf []byte) (int, error) {
	return read(int(r), buf)
}

func read(fd int, p []byte) (int, error) {
	buf := unsafe.Pointer(&_zero)
	if len(p) != 0 {
		buf = unsafe.Pointer(&p[0])
	}
	res, _, errNo := syscall.Syscall(syscall.SYS_READ, uintptr(fd), uintptr(buf), uintptr(len(p)))
	return int(res), errnoConvert(errNo)
}
