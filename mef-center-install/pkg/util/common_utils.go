// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util constants public file for install/upgrade/run function
package util

import (
	"fmt"
	"os"
	"os/user"

	"path"
	"path/filepath"
	"strconv"
	"syscall"

	"huawei.com/mindx/common/hwlog"
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
	usr, err := common.GetCurrentUser()
	if err != nil {
		return err
	}
	if usr != RootUserName {
		return fmt.Errorf("install failed: the install user must be root, can not be %s", usr)
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

// GetInstallInfo is used to get the information from install-param.json
func GetInstallInfo() (*InstallParamJsonTemplate, error) {
	currentDir, err := filepath.Abs(filepath.Dir(filepath.Dir(os.Args[0])))
	if err != nil {
		fmt.Printf("get current absolute path error: %s", err)
		hwlog.RunLog.Errorf("get current absolute path error: %s", err)
		return nil, err
	}

	paramJsonPath := path.Join(currentDir, InstallParamJson)
	paramJsonMgr, err := GetInstallParamJsonInfo(paramJsonPath)
	if err != nil {
		fmt.Printf("get current absolute path error: %s", err)
		hwlog.RunLog.Errorf("get current absolute path error: %s", err)
		return nil, err
	}

	return paramJsonMgr, nil
}

// GetMefId is used to get uid/gid for user/group MEFCenter
func GetMefId() (int, int, error) {
	mefUser, err := user.Lookup(MefCenterName)
	if err != nil {
		return 0, 0, fmt.Errorf("get MEFCenter uid failed： %s", err.Error())
	}
	uid, err := strconv.Atoi(mefUser.Uid)
	if err != nil {
		return 0, 0, fmt.Errorf("transfer %s uid into int failed: %s", MefCenterName, err.Error())
	}

	mefGroup, err := user.LookupGroup(MefCenterGroup)
	if err != nil {
		return 0, 0, fmt.Errorf("get MEFCenter gid failed: %s", err.Error())
	}
	gid, err := strconv.Atoi(mefGroup.Gid)
	if err != nil {
		return 0, 0, fmt.Errorf("transfer %s gid into int failed: %s", MefCenterGroup, err.Error())
	}

	return uid, gid, nil
}
