// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

//go:build linux || freebsd || dragonfly || solaris
// +build linux freebsd dragonfly solaris

// Package rand implement the security rand
package rand

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	maxReadSize = 1<<25 - 1
)

// A randomReader satisfies reads by reading the file named name.
type randomReader struct {
	f  io.Reader
	mu sync.Mutex
}

func init() {
	Reader = &randomReader{}
}

func warnBlocked() {
	fmt.Println("mindx-security/rand: blocked for 60 seconds waiting to read random data from the kernel")
}

var supportOs = "linux"

// Read implements the interface of io.Reader
func (r *randomReader) Read(b []byte) (int, error) {
	t := time.AfterFunc(time.Minute, warnBlocked)
	defer t.Stop()
	if len(b) > maxReadSize {
		return 0, errors.New("byte size is too large")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if runtime.GOOS != supportOs {
		return 0, errors.New("not supported")
	}
	f, err := os.Open("/dev/random")
	if err != nil {
		return 0, err
	}
	defer func() {
		if f == nil {
			return
		}
		err = f.Close()
		if err != nil {
			fmt.Println("close random file failed")
		}
	}()
	if f == nil {
		return 0, errors.New("invalid random reader")
	}
	return f.Read(b)
}
