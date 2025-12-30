// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
