// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package envutils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindx/common/hwlog"
)

const (
	processUid   = 0
	processUmask = 077
	lockBaseDir  = "/run"
)

type flock struct {
	lockPath string
	handle   *os.File
}

const (
	flagLines = 2

	pidIndex    = 0
	reasonIndex = 1
)

var (
	instance *flock
	onceInit sync.Once
)

// GetFlock get the instance
func GetFlock(lockName string) *flock {
	onceInit.Do(func() {
		instance = &flock{lockPath: filepath.Join(lockBaseDir, lockName)}
	})
	return instance
}

// Lock lock process
func (f *flock) Lock(reason string) error {
	if err := f.checkLockFileExist(); err != nil {
		return err
	}
	file, err := os.OpenFile(f.lockPath, os.O_RDWR|os.O_CREATE, os.FileMode(fileutils.Mode600))
	if err != nil {
		return fmt.Errorf("open file[%s] failed:%v", f.lockPath, err)
	}
	f.handle = file
	linkChecker := fileutils.NewFileLinkChecker(false)
	if err = linkChecker.Check(file, f.lockPath); err != nil {
		f.closeHandle()
		return fmt.Errorf("lock file [%s] failed: %v", f.lockPath, err)
	}
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.printLockInfo()
		f.closeHandle()
		return fmt.Errorf("lock file[%s] failed:%v", f.lockPath, err)
	}
	if err = f.writeLockInfo(reason); err != nil {
		f.Unlock()
		return err
	}
	hwlog.RunLog.Info("lock process success")
	return nil
}

func (f *flock) closeHandle() {
	if err := f.handle.Close(); err != nil {
		hwlog.RunLog.Errorf("close file[%s] failed,error:%v", f.lockPath, err)
	}
}

// Unlock unlock process
func (f *flock) Unlock() {
	if f.handle == nil {
		return
	}
	if err := syscall.Flock(int(f.handle.Fd()), syscall.LOCK_UN); err != nil {
		hwlog.RunLog.Errorf("unlock file[%s] failed:%v", f.lockPath, err)
	}
	f.closeHandle()
}

func (f *flock) writeLockInfo(reason string) error {
	pid := os.Getpid()
	pidString := strconv.Itoa(pid)
	flagContent := fmt.Sprintf("%s\n%s", pidString, reason)
	if err := f.write([]byte(flagContent)); err != nil {
		hwlog.RunLog.Errorf("write pid to process flag failed, error: %v", err)
		return errors.New("write pid, process name to process flag failed")
	}
	hwlog.RunLog.Info("write process flag success")
	return nil
}

func (f *flock) checkLockFileExist() error {
	if !fileutils.IsExist(f.lockPath) {
		return nil
	}
	_, err := fileutils.CheckOwnerAndPermission(f.lockPath, processUmask, processUid)
	if err == nil {
		return nil
	}
	hwlog.RunLog.Warnf("existed process flag file is invalid, %v", err)
	if err = fileutils.DeleteFile(f.lockPath); err != nil {
		hwlog.RunLog.Errorf("remove existed process flag file failed, error: %v", err)
		return err
	}
	return nil
}

func (f *flock) printLockInfo() {
	data, err := fileutils.LoadFile(f.lockPath)
	if err != nil {
		hwlog.RunLog.Warnf("load lock file error: %v", err)
		return
	}
	content := string(data)
	if content == "" {
		hwlog.RunLog.Warn("existed process flag file is empty")
		return
	}
	parts := strings.Split(content, "\n")
	if len(parts) != flagLines {
		hwlog.RunLog.Warn("existed process flag file is not expected")
		return
	}
	pid, err := strconv.Atoi(parts[pidIndex])
	if err != nil {
		hwlog.RunLog.Warnf("convert pid bytes to int failed, error: %v", err)
		return
	}
	reason := parts[reasonIndex]
	hwlog.RunLog.Errorf("another process is running, pid: %d, reason: [%s]."+
		"please operate after the process is complete", pid, reason)
}

func (f *flock) write(data []byte) error {
	if err := f.handle.Truncate(0); err != nil {
		return err
	}
	if _, err := f.handle.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := f.handle.Write(data); err != nil {
		return fmt.Errorf("write file[%s] failed:%v", f.lockPath, err)
	}
	return nil
}
