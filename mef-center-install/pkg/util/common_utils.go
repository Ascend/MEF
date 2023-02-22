// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util constants public file for install/upgrade/run function
package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"

	"path"
	"path/filepath"
	"strconv"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// GetArch is used to get the arch info
func GetArch() (string, error) {
	arch, err := common.RunCommand(ArchCommand, true, common.DefaultCmdWaitTime, "-i")
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

// CheckDiskSpace is used to check if the disk space on a path is enough to a limit
func CheckDiskSpace(path string, limit uint64) error {
	availUpgradeSpace, err := common.GetDiskFree(path)
	if err != nil {
		return fmt.Errorf("get path [%s]'s disk available space failed: %s", path, err.Error())
	}

	if availUpgradeSpace < limit {
		return errors.New("no enough space")
	}

	return nil
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

// GetLocalIp get local ip
func GetLocalIp() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("get local ip address failed: %s", err.Error())
	}
	privateIp := ""
	for _, address := range addresses {
		ipNet, ok := address.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if !ipNet.IP.IsPrivate() {
			return ipNet.IP.String(), nil
		}
		privateIp = ipNet.IP.String()
	}
	if privateIp != "" {
		return privateIp, nil
	}
	return "", errors.New("get local ip address failed")
}

// GetCenterUid is used to get the MEFCenter UID
func GetCenterUid() (string, error) {
	userInfo, err := user.Lookup(MefCenterName)
	if err != nil {
		hwlog.RunLog.Errorf("get %s uid failed: %s", MefCenterName, err.Error())
		return "", err
	}

	return userInfo.Uid, nil
}

// GetNecessaryTools is used to get the necessary tools of MEF-Center
func GetNecessaryTools() []string {
	return []string{
		"sh",
		"kubectl",
		"docker",
		"uname",
		"cp",
		"grep",
		"useradd",
		"wc",
	}
}
