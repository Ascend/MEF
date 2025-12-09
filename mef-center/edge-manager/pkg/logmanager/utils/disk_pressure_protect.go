// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils
package utils

import (
	"errors"
	"io"
	"syscall"

	"huawei.com/mindxedge/base/common"
)

const (
	defaultReservedRate  = 0.2
	defaultReservedBytes = 200 * common.MB
)

var (
	// ErrDiskPressure indicates disk pressure
	ErrDiskPressure = errors.New("disk pressure")
)

// WithDiskPressureProtect creates a new diskPressureProtectWriter
func WithDiskPressureProtect(writer io.Writer, filePath string) io.Writer {
	return &diskPressureProtectWriter{
		reservedBytes: defaultReservedBytes,
		reservedRate:  defaultReservedRate,
		writer:        writer,
		filePath:      filePath,
	}
}

type diskPressureProtectWriter struct {
	reservedBytes uint64
	reservedRate  float64
	writer        io.Writer
	filePath      string
	lastCheckPos  uint64
	currentPos    uint64
}

func (w *diskPressureProtectWriter) Write(buffer []byte) (int, error) {
	var checked bool
	if w.currentPos == 0 || (uint64(len(buffer))+w.currentPos)-w.lastCheckPos > common.MB {
		if err := checkDiskSpace(w.filePath, uint64(len(buffer)), w.reservedBytes, w.reservedRate); err != nil {
			return 0, err
		}
		checked = true
	}

	nRead, err := w.writer.Write(buffer)
	if err != nil && err != io.EOF {
		return 0, err
	}
	w.currentPos += uint64(nRead)
	if checked {
		w.lastCheckPos = w.currentPos
	}
	return nRead, err
}

// CheckDiskSpace checks whether disk space is enough
func CheckDiskSpace(filePath string, requiredSpace uint64) error {
	return checkDiskSpace(filePath, requiredSpace, defaultReservedBytes, defaultReservedRate)
}

func checkDiskSpace(filePath string, requiredSpace, reservedBytes uint64, reservedRate float64) error {
	fileStat := syscall.Statfs_t{}
	if err := syscall.Statfs(filePath, &fileStat); err != nil {
		return err
	}

	diskFree := fileStat.Bavail * uint64(fileStat.Bsize)
	if diskFree < requiredSpace {
		return ErrDiskPressure
	}
	freeBytes := diskFree - requiredSpace
	if reservedBytes != 0 && freeBytes <= reservedBytes {
		return ErrDiskPressure
	}

	diskTotal := fileStat.Blocks * uint64(fileStat.Bsize)
	if diskTotal == 0 {
		return ErrDiskPressure
	}
	freeRate := float64(freeBytes) / float64(diskTotal)
	if reservedRate != 0 && freeRate <= reservedRate {
		return ErrDiskPressure
	}
	return nil
}
