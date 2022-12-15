// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"os"
	"path/filepath"
	"syscall"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

func setOneOwnerGroup(path string, uid, gid int) error {
	pathInfo, err := os.Stat(path)
	if err != nil {
		hwlog.RunLog.Errorf("get info of path [%s] failed, error: %s", path, err.Error())
		return err
	}
	stat, ok := pathInfo.Sys().(*syscall.Stat_t)
	if !ok {
		hwlog.RunLog.Errorf("get stat of path [%s] failed, error: %s", path, err.Error())
		return err
	}
	if int(stat.Uid) == uid && int(stat.Gid) == gid {
		return nil
	}
	if err = os.Chown(path, uid, gid); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed, error: %s", path, err.Error())
		return err
	}
	return nil
}

func setWalkPathOwnerGroup(path string, uid, gid int, ignoreFile bool) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			hwlog.RunLog.Errorf("walk path [%s] failed, error: %s", path, err.Error())
			return err
		}
		if ignoreFile && !info.IsDir() {
			return nil
		}
		return setOneOwnerGroup(path, uid, gid)
	})
}

// SetPathOwnerGroup set path owner and group
func SetPathOwnerGroup(path string, uid, gid int, recursive bool, ignoreFile bool) error {
	if _, err := utils.CheckPath(path); err != nil {
		hwlog.RunLog.Errorf("check path [%s] failed, error: %s", path, err.Error())
		return err
	}
	if err := setOneOwnerGroup(path, uid, gid); err != nil {
		return err
	}
	if !recursive || utils.IsFile(path) {
		return nil
	}
	return setWalkPathOwnerGroup(path, uid, gid, ignoreFile)
}
