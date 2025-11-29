//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package terminal provide a safe reader for password
package terminal

import "syscall"

var (
	errorEAGAINE error = syscall.EAGAIN
	errorEINVAL  error = syscall.EINVAL
	errorENOENT  error = syscall.ENOENT
)

func errnoConvert(e syscall.Errno) error {
	switch e {
	case syscall.EAGAIN:
		return errorEAGAINE
	case syscall.EINVAL:
		return errorEINVAL
	case syscall.ENOENT:
		return errorENOENT
	case 0:
		return nil
	default:
		return e
	}
}
