// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package fileutils for setting owner and group for path or file
package fileutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

var noEntityErrStr = "no such file or directory"

// SetOwnerParam is the struct of param used by SetPathOwnerGroup method
type SetOwnerParam struct {
	Path string
	Uid  uint32
	Gid  uint32
	// if to set the owner of files inside the path
	Recursive    bool
	IgnoreFile   bool
	CheckerParam []FileChecker
}

// SetPathOwnerGroup set owner and group for path or file
func SetPathOwnerGroup(param SetOwnerParam) error {
	if err := setOneOwnerGroup(param); err != nil {
		return err
	}

	if !param.Recursive {
		return nil
	}

	return setWalkPathOwnerGroup(param)
}

func setOneOwnerGroup(param SetOwnerParam) error {
	file, _, err := checkFile(param.Path, os.O_RDONLY, Mode400, param.CheckerParam...)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	// do not change the mode of a file that targeted by a softlink
	linkChecker := NewFileLinkChecker(false)
	if err = linkChecker.Check(file, param.Path); err != nil {
		return nil
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file info failed: %s", err.Error())
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("get stat of path [%s] failed", param.Path)
	}

	if stat.Uid == param.Uid && stat.Gid == param.Gid {
		return nil
	}

	if err = file.Chown(int(param.Uid), int(param.Gid)); err != nil {
		return fmt.Errorf("set path [%s] owner and group failed, error: %s", param.Path, err.Error())
	}
	return nil
}

func setWalkPathOwnerGroup(param SetOwnerParam) error {
	rootPath := param.Path
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if strings.HasSuffix(err.Error(), noEntityErrStr) {
				return nil
			}
			return fmt.Errorf("walk path [%s] failed, error: %s", path, err.Error())
		}
		if param.IgnoreFile && !info.IsDir() {
			return nil
		}
		param.Path = path
		return setOneOwnerGroup(param)
	})
}

// SetPathPermission set permission for path or file
// the variable recursive indicates whether to modify the permissions of all files with the path
func SetPathPermission(path string, mode os.FileMode, recursive, ignoreFile bool, checkerParam ...FileChecker) error {
	if !CheckMode(mode) {
		return errors.New("cannot set write permission for group or others")
	}
	if err := setOneMode(path, mode, checkerParam...); err != nil {
		return err
	}

	if !recursive {
		return nil
	}

	return setWalkPathMode(path, mode, ignoreFile, checkerParam...)
}

// SetParentPathPermission set permission for path and parent path recursively
func SetParentPathPermission(path string, mode os.FileMode, checkerParam ...FileChecker) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("get abs path failed: %s", err.Error())
	}

	if err = SetPathPermission(absPath, mode, false, true, checkerParam...); err != nil {
		return err
	}

	if absPath != "/" {
		return SetParentPathPermission(filepath.Dir(absPath), mode, checkerParam...)
	}
	return nil
}

func setOneMode(path string, mode os.FileMode, checkerParam ...FileChecker) error {
	file, _, err := checkFile(path, os.O_RDONLY, Mode400, checkerParam...)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	// do not change the mode of a file that targeted by a softlink
	linkChecker := NewFileLinkChecker(false)
	if err = linkChecker.Check(file, path); err != nil {
		return nil
	}

	if err = file.Chmod(mode); err != nil {
		return fmt.Errorf("set path [%s] mode failed, error: %s", path, err.Error())
	}
	return nil
}

func setWalkPathMode(path string, mode os.FileMode, ignoreFile bool, checkerParam ...FileChecker) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if strings.HasSuffix(err.Error(), noEntityErrStr) {
				return nil
			}
			return fmt.Errorf("walk path [%s] failed, error: %s", path, err.Error())
		}
		if ignoreFile && !info.IsDir() {
			return nil
		}
		return setOneMode(path, mode, checkerParam...)
	})
}
