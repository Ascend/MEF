// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util this file for process lock
package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

// FlagLock lock flag
type FlagLock interface {
	Lock() error
	Unlock() error
}

type flagLock struct {
	status     uint32
	statusLock sync.Mutex
	flagPath   string
	operation  string
}

const (
	locking   = 1
	none      = 0
	flagLines = 3

	pidIndex      = 0
	procNameIndex = 1
	optIndex      = 2
)

var (
	instance FlagLock
	onceInit sync.Once
)

// FlagLockInstance lock flag instance
func FlagLockInstance(flagPath string, lockFlag string, operation string) FlagLock {
	onceInit.Do(func() {
		instance = &flagLock{flagPath: filepath.Join(flagPath, lockFlag), operation: operation}
	})
	return instance
}

// Lock process flag
func (f *flagLock) Lock() error {
	if atomic.LoadUint32(&f.status) == locking {
		hwlog.RunLog.Error("process flag is already locked")
		return errors.New("process flag is already locked")
	}
	f.statusLock.Lock()
	defer f.statusLock.Unlock()
	if f.status == locking {
		hwlog.RunLog.Error("process flag is already locked")
		return errors.New("process flag is already locked")
	}
	atomic.StoreUint32(&f.status, locking)

	if err := f.checkFlag(); err != nil {
		atomic.StoreUint32(&f.status, none)
		return err
	}
	if err := f.writeFlag(); err != nil {
		atomic.StoreUint32(&f.status, none)
		return err
	}

	hwlog.RunLog.Info("lock process success")
	return nil
}

// Unlock unlock process flag
func (f *flagLock) Unlock() error {
	if atomic.LoadUint32(&f.status) == none {
		return nil
	}
	defer func() {
		f.statusLock.Lock()
		atomic.StoreUint32(&f.status, none)
		f.statusLock.Unlock()
	}()
	if err := fileutils.DeleteFile(f.flagPath); err != nil {
		hwlog.RunLog.Warnf("remove process flag file failed: %v", err)
		return errors.New("remove process flag file failed")
	}
	hwlog.RunLog.Info("unlock process success")
	return nil
}

func (f *flagLock) writeFlag() error {
	pid := os.Getpid()
	pidString := strconv.Itoa(pid)
	procName, err := GetProcName(pid)
	if err != nil {
		hwlog.RunLog.Errorf("get current process name failed, error: %v", err)
		return errors.New("get current process name failed")
	}
	flagContent := fmt.Sprintf("%s\n%s\n%s", pidString, procName, f.operation)
	if err = WriteWithLock(f.flagPath, []byte(flagContent)); err != nil {
		hwlog.RunLog.Errorf("write pid to process flag failed, error: %v", err)
		return errors.New("write pid, process name, and operation to process flag failed")
	}
	hwlog.RunLog.Info("write process flag success")
	return nil
}

func (f *flagLock) checkFlag() error {
	if !fileutils.IsExist(f.flagPath) {
		return nil
	}
	if _, err := fileutils.CheckOwnerAndPermission(f.flagPath, constants.ProcessFlagUmask,
		constants.ProcessFlagUid); err != nil {
		hwlog.RunLog.Warnf("existed process flag file is invalid, %v", err)
		if err = fileutils.DeleteFile(f.flagPath); err != nil {
			hwlog.RunLog.Errorf("remove existed process flag file failed, error: %v", err)
			return err
		}
		return nil
	}
	hwlog.RunLog.Warn("process flag file is already existed")
	return f.checkProcessLocked()
}

func (f *flagLock) checkProcessLocked() error {
	data, err := fileutils.LoadFile(f.flagPath)
	if err != nil {
		hwlog.RunLog.Errorf("load flag file failed, error: %v", err)
		return err
	}
	content := string(data)
	if content == "" {
		hwlog.RunLog.Warn("existed process flag file is empty")
		return nil
	}
	parts := strings.Split(content, "\n")
	if len(parts) != flagLines {
		hwlog.RunLog.Warn("existed process flag file is not expected")
		return nil
	}
	pid, err := strconv.Atoi(parts[pidIndex])
	if err != nil {
		hwlog.RunLog.Warnf("convert pid bytes to int failed, error: %v", err)
		return nil
	}
	if !IsProcessActive(pid) {
		hwlog.RunLog.Warn("the previous process has exited")
		return nil
	}
	activeProcName, err := GetProcName(pid)
	if err != nil {
		hwlog.RunLog.Warnf("get previous process name failed, error: %v", err)
		return nil
	}
	if activeProcName == parts[procNameIndex] {
		operation := parts[optIndex]
		errMsg := fmt.Errorf("another [%s] process is running, pid: %d, "+
			"please operate after the process is complete", operation, pid)
		fmt.Println(errMsg.Error())
		hwlog.RunLog.Errorf(errMsg.Error())
		return errMsg
	}
	return nil
}
