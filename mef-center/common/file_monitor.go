// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common for
package common

import (
	"bytes"
	"context"
	"reflect"
	"time"

	"huawei.com/mindx/common/fileutils"
)

// FileMonitor implements file modification detection by periodical hash comparison.
type FileMonitor struct {
	filePaths     []string
	checkInterval time.Duration
	callback      func([]string)
	fileStates    map[string]fileState
	recheckChan   chan struct{}
}

type fileState struct {
	err      error
	checksum []byte
}

// NewFileMonitor creates a new FileMonitor
func NewFileMonitor(checkInterval time.Duration, onChangedFn func([]string), filePaths ...string) *FileMonitor {
	return &FileMonitor{
		filePaths:     filePaths,
		checkInterval: checkInterval,
		callback:      onChangedFn,
		recheckChan:   make(chan struct{}, 1),
	}
}

// Run runs the monitor. This method won't return until ctx was cancelled.
func (fm *FileMonitor) Run(ctx context.Context) {
	if fm == nil {
		return
	}

	fileStates := make(map[string]fileState)
	for _, filePath := range fm.filePaths {
		fileStates[filePath] = fileState{}
	}
	fm.fileStates = fileStates

	ticker := time.NewTicker(fm.checkInterval)
	defer ticker.Stop()

	fm.check()
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			fm.check()
		case _, ok := <-fm.recheckChan:
			if !ok {
				return
			}
			ticker.Reset(fm.checkInterval)
			fm.check()
		case <-ctx.Done():
			return
		}
	}
}

// Recheck checks modification immediately. This function also resets the ticker.
func (fm *FileMonitor) Recheck() {
	if fm == nil {
		return
	}

	select {
	case fm.recheckChan <- struct{}{}:
	default:
	}
}

func (fm *FileMonitor) check() {
	var changedFiles []string
	for _, filePath := range fm.filePaths {
		if fm.compareAndUpdateFileState(filePath) {
			changedFiles = append(changedFiles, filePath)
		}
	}
	if len(changedFiles) > 0 {
		fm.callback(changedFiles)
	}
}

// compareAndUpdateFileState returns true if file is changed
func (fm *FileMonitor) compareAndUpdateFileState(filePath string) bool {
	currentState, ok := fm.fileStates[filePath]
	if !ok {
		return false
	}
	checksum, err := fileutils.GetFileSha256(filePath)
	var checksumBytes []byte
	if err == nil {
		checksumBytes = []byte(checksum)
	}

	fm.fileStates[filePath] = fileState{checksum: checksumBytes, err: err}
	return !(bytes.Equal(currentState.checksum, checksumBytes) && reflect.DeepEqual(err, currentState.err))
}
