// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils for setting owner and group for path or file
package utils

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"syscall"
)

// SetPathOwnerGroup set owner and group for path or file
func SetPathOwnerGroup(path string, uid, gid uint32, recursive, ignoreFile bool) error {
	if uint64(uid) > math.MaxInt || uint64(gid) > math.MaxInt {
		return errors.New("user id or group id is out of range")
	}

	absPath, err := CheckOriginPath(path)
	if err != nil {
		return fmt.Errorf("check path failed, error: %s", err.Error())
	}
	if absPath == "" || !IsExist(absPath) {
		return errors.New("path does not exist")
	}

	if err = setOneOwnerGroup(absPath, uid, gid); err != nil {
		return err
	}

	if !recursive || IsFile(absPath) {
		return nil
	}

	return setWalkPathOwnerGroup(absPath, uid, gid, ignoreFile)
}

func setOneOwnerGroup(path string, uid, gid uint32) error {
	if _, err := CheckOriginPath(path); err != nil {
		return nil
	}

	pathInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("get info of path [%s] failed, error: %s", path, err.Error())
	}

	stat, ok := pathInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("get stat of path [%s] failed", path)
	}

	if stat.Uid == uid && stat.Gid == gid {
		return nil
	}

	if err = os.Lchown(path, int(uid), int(gid)); err != nil {
		return fmt.Errorf("set path [%s] owner and group failed, error: %s", path, err.Error())
	}
	return nil
}

func setWalkPathOwnerGroup(path string, uid, gid uint32, ignoreFile bool) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path [%s] failed, error: %s", path, err.Error())
		}
		if ignoreFile && !info.IsDir() {
			return nil
		}
		return setOneOwnerGroup(path, uid, gid)
	})
}
