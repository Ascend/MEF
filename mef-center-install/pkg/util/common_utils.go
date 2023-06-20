// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util constants public file for install/upgrade/run function
package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
)

// GetArch is used to get the arch info
func GetArch() (string, error) {
	arch, err := envutils.RunCommand(ArchCommand, envutils.DefCmdTimeoutSec, "-i")
	if err != nil {
		return "", err
	}
	return arch, nil
}

// GetInstallInfo is used to get the information from install-param.json
func GetInstallInfo() (*InstallParamJsonTemplate, error) {
	currentDir, err := filepath.Abs(filepath.Dir(filepath.Dir(os.Args[0])))
	if err != nil {
		fmt.Printf("get current absolute path error: %s\n", err.Error())
		return nil, err
	}

	paramJsonPath := path.Join(currentDir, InstallParamJson)
	paramJsonMgr, err := GetInstallParamJsonInfo(paramJsonPath)
	if err != nil {
		fmt.Printf("get current absolute path error: %s\n", err.Error())
		return nil, err
	}

	return paramJsonMgr, nil
}

// GetMefId is used to get uid/gid for user/group MEFCenter
func GetMefId() (uint32, uint32, error) {
	uid, err := envutils.GetUid(MefCenterName)
	if err != nil {
		return 0, 0, fmt.Errorf("get uid failed, error: %v", err)
	}

	gid, err := envutils.GetGid(MefCenterGroup)
	if err != nil {
		return 0, 0, fmt.Errorf("get gid failed, error: %v", err)
	}

	return uid, gid, nil
}

// GetPublicIps get local ips
func GetPublicIps() ([]string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("get local ip address failed: %s", err.Error())
	}
	var validIps []string
	for _, address := range addresses {
		ipNet, ok := address.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ipNet.IP.To4() == nil {
			continue
		}
		if ipNet.IP.IsPrivate() {
			continue
		}
		validIps = append(validIps, ipNet.IP.String())
	}
	if len(validIps) == 0 {
		return nil, errors.New("no public ip found")
	}
	return validIps, nil
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
	}
}

// CheckCurrentPath is the public check func for run.sh
// it checks if the current run.sh is in the mef-center softlink
func CheckCurrentPath(linkPath string) error {
	curPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("get current path failed: %s", err.Error())
	}

	curAbsPath, err := filepath.EvalSymlinks(curPath)
	if err != nil {
		return fmt.Errorf("get current abs path failed: %s", err.Error())
	}

	workingAbsPath, err := filepath.EvalSymlinks(linkPath)
	if err != nil {
		return fmt.Errorf("get working abs path failed: %s", err.Error())
	}

	if filepath.Dir(curAbsPath) != workingAbsPath {
		return fmt.Errorf("current sh path is not in the working path")
	}

	return nil
}
