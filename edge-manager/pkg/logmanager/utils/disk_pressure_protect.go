// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
		fileStat := syscall.Statfs_t{}
		if err := syscall.Statfs(w.filePath, &fileStat); err != nil {
			return 0, err
		}

		diskFree := fileStat.Bavail * uint64(fileStat.Bsize)
		freeBytes := diskFree - uint64(len(buffer))
		if w.reservedBytes != 0 && freeBytes <= w.reservedBytes {
			return 0, ErrDiskPressure
		}

		diskTotal := fileStat.Blocks * uint64(fileStat.Bsize)
		if diskTotal == 0 {
			return 0, ErrDiskPressure
		}
		freeRate := float64(freeBytes) / float64(diskTotal)
		if w.reservedRate != 0 && freeRate <= w.reservedRate {
			return 0, ErrDiskPressure
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
