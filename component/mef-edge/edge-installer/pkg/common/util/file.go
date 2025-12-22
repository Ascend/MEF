// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util this file for operate file
package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

var rwMutex sync.RWMutex

// WriteWithLock write file with flock
func WriteWithLock(filePath string, data []byte) error {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.FileMode(constants.Mode600))
	if err != nil {
		return fmt.Errorf("open file[%s] failed: %v", filePath, err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			hwlog.RunLog.Errorf("close file[%s] failed, error: %v", filePath, err)
		}
	}()
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("lock file[%s] failed: %v", filePath, err)
	}
	defer func() {
		if err = syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
			hwlog.RunLog.Errorf("unlock file[%s] failed: %v", filePath, err)
		}
	}()
	const (
		bufSize = 1024
		size10M = 10 * 1024 * 1024
	)
	buf := make([]byte, bufSize)
	size := 0
	for {
		n, err := f.Read(buf)
		size += n
		if err == io.EOF {
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return err
			}
			break
		}
		if err != nil {
			return err
		}
		if size > size10M {
			return errors.New("file is too big")
		}
	}
	if err = f.Truncate(int64(size)); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("write file[%s] failed: %v", filePath, err)
	}
	return nil
}

// InSamePartition is in same file partition
func InSamePartition(path1, path2 string) (bool, error) {
	fs1 := syscall.Stat_t{}
	fs2 := syscall.Stat_t{}
	if err := syscall.Stat(path1, &fs1); err != nil {
		return false, err
	}
	if err := syscall.Stat(path2, &fs2); err != nil {
		return false, err
	}
	return fs1.Dev == fs2.Dev, nil
}

func closeFile(file *os.File) {
	if file == nil {
		return
	}
	if err := file.Close(); err != nil {
		return
	}
	return
}

// SetImmutable [method] for set file path immutable
func SetImmutable(filepath string) error {
	_, err := envutils.RunCommand(constants.Chattr, envutils.DefCmdTimeoutSec, "+i", "-R", filepath)
	if err != nil {
		return err
	}
	return nil
}

// UnSetImmutable [method] for unset file path immutable
func UnSetImmutable(filepath string) error {
	_, err := envutils.RunCommand(constants.Chattr, envutils.DefCmdTimeoutSec, "-i", "-R", filepath)
	if err != nil {
		return err
	}
	return nil
}

// LoadJsonFile load json file to a map
func LoadJsonFile(jsonFilePath string) (map[string]interface{}, error) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	data, err := fileutils.LoadFile(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("read json file failed: %v", err)
	}
	jsonValue := make(map[string]interface{})
	if err = json.Unmarshal(data, &jsonValue); err != nil {
		return nil, fmt.Errorf("unmarshal json value failed: %v", err)
	}
	return jsonValue, nil
}

// SaveJsonValue save json value to file
func SaveJsonValue(jsonFilePath string, jsonValue map[string]interface{}) error {
	data, err := json.MarshalIndent(jsonValue, "", "    ")
	if err != nil {
		return errors.New("marshal json value failed")
	}
	rwMutex.Lock()
	defer rwMutex.Unlock()
	if err = fileutils.WriteData(jsonFilePath, data); err != nil {
		return errors.New("write json file failed")
	}
	return nil
}

// SetJsonValue set json value to map
func SetJsonValue(object map[string]interface{}, value interface{}, names ...string) error {
	if object == nil {
		return errors.New("map is nil")
	}
	length := len(names)
	if length == 0 {
		return errors.New("provide at least one name")
	}
	for i, name := range names {
		valueInterface, ok := object[name]
		if !ok && i != length-1 {
			return fmt.Errorf("name[%v] not found", name)
		}
		if i == length-1 {
			object[name] = value
			return nil
		}
		object, ok = valueInterface.(map[string]interface{})
		if !ok {
			return fmt.Errorf("convert property[%s] to map failed", name)
		}
	}
	return nil
}

// SetPathOwnerGroupToMEFEdge is the func to set path's owner and group to MEFEdge
func SetPathOwnerGroupToMEFEdge(path string, recursive, ignoreFile bool) error {
	uid, gid, err := GetMefId()
	if err != nil {
		return fmt.Errorf("get uid/gid of mef-edge failed: %v", err)
	}

	param := fileutils.SetOwnerParam{
		Path:       path,
		Uid:        uid,
		Gid:        gid,
		Recursive:  recursive,
		IgnoreFile: ignoreFile,
	}
	if err = fileutils.SetPathOwnerGroup(param); err != nil {
		return fmt.Errorf("set dir [%s] owner and group failed, error: %v", path, err)
	}
	return nil
}

// CreateBackupWithMefOwner a util func for create backup file for mef-edge
func CreateBackupWithMefOwner(originalPath string) error {
	mgr := NewEdgeUGidMgr()
	if err := mgr.SetEUGidToEdge(); err != nil {
		return fmt.Errorf("set euid/egid to mef-edge failed: %v", err)
	}
	defer func() {
		if err := mgr.ResetEUGid(); err != nil {
			hwlog.RunLog.Errorf("reset euid/egid failed, %v", err)
		}
	}()
	if err := backuputils.BackUpFiles(originalPath); err != nil {
		return fmt.Errorf("back up file with mef-edge owner failed, %v", err)
	}
	return nil
}
