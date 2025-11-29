//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

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
