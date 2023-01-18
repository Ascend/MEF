// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util constants public file for install/upgrade/run function
package util

import (
	"fmt"
	"os/user"
	"syscall"

	"huawei.com/mindxedge/base/common"
)

// GetArch is used to get the arch info
func GetArch() (string, error) {
	arch, err := common.RunCommand(ArchCommand, true, "-i")
	if err != nil {
		return "", err
	}
	return arch, nil
}

// CheckUser is used to check if the current user is root and returns error if not
func CheckUser() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("get current user info failed : %s", err)
	}
	if usr.Username != RootUserName {
		return fmt.Errorf("install failed: the install user must be root, can not be %s", usr.Username)
	}

	return nil
}

// GetDiskFree is used to get the free disk space of a path
func GetDiskFree(path string) (uint64, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return 0, err
	}
	diskFree := fs.Bfree * uint64(fs.Bsize)
	return diskFree, nil
}
