// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils for setting permission for path or file
package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// SetPathPermission set permission for path or file
func SetPathPermission(path string, mode os.FileMode, recursive, ignoreFile bool) error {
	if !CheckMode(mode) {
		return errors.New("cannot set write permission for group or others")
	}

	absPath, err := CheckOriginPath(path)
	if err != nil {
		return fmt.Errorf("check path failed, error: %s", err.Error())
	}
	if absPath == "" || !IsExist(absPath) {
		return errors.New("path does not exist")
	}

	if err = setOneMode(absPath, mode); err != nil {
		return err
	}

	if !recursive || IsFile(absPath) {
		return nil
	}

	return setWalkPathMode(absPath, mode, ignoreFile)
}

// SetParentPathPermission set permission for path and parent path recursively
func SetParentPathPermission(path string, mode os.FileMode) error {
	absPath, err := CheckOriginPath(path)
	if err != nil {
		return fmt.Errorf("check path failed, error: %s", err.Error())
	}

	if err = SetPathPermission(absPath, mode, false, true); err != nil {
		return err
	}

	if absPath != "/" {
		return SetParentPathPermission(filepath.Dir(absPath), mode)
	}
	return nil
}

func setOneMode(path string, mode os.FileMode) error {
	if _, err := CheckOriginPath(path); err != nil {
		return nil
	}

	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("set path [%s] mode failed, error: %s", path, err.Error())
	}
	return nil
}

func setWalkPathMode(path string, mode os.FileMode, ignoreFile bool) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path [%s] failed, error: %s", path, err.Error())
		}
		if ignoreFile && !info.IsDir() {
			return nil
		}
		return setOneMode(path, mode)
	})
}
