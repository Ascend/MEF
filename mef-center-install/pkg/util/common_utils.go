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
	"syscall"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
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
		hwlog.RunLog.Errorf("get current absolute path failed, error: %v", err)
		fmt.Println("get current absolute path failed")
		return nil, err
	}

	paramJsonPath := path.Join(currentDir, InstallParamJson)
	paramJsonMgr, err := GetInstallParamJsonInfo(paramJsonPath)
	if err != nil {
		hwlog.RunLog.Errorf("get install info failed, error: %v", err)
		fmt.Println("get install info failed")
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

// GetLocalIps get local ips
func GetLocalIps() ([]string, error) {
	const maxIPLen = 20
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("get local ip address failed: %s", err.Error())
	}
	if len(addresses) > maxIPLen {
		return nil, fmt.Errorf("interfaceAddrs length cannot exceed %v", maxIPLen)
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
		fmt.Println("the existing dir is not the MEF working dir, cannot execute any command here")
		return fmt.Errorf("current sh path is not in the working path")
	}

	return nil
}

// GetBoolPointer get pointer based on bool value
// If the query or update value is 0 in db, the query or update fails. Use the pointer can solve the problem.
func GetBoolPointer(value bool) *bool {
	pointer := new(bool)
	*pointer = value
	return pointer
}

// ReducePriv reduce privilege
func ReducePriv() error {
	mefUid, mefGid, err := GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get mef uid and gid failed, error: %v", err)
		return errors.New("get mef uid and gid failed")
	}
	if err = syscall.Setegid(int(mefGid)); err != nil {
		hwlog.RunLog.Errorf("set egid failed, error: %v", err)
		return errors.New("set egid failed")
	}
	if err = syscall.Seteuid(int(mefUid)); err != nil {
		hwlog.RunLog.Errorf("set euid failed, error: %v", err)
		return errors.New("set euid failed")
	}

	return nil
}

// ResetPriv reset privilege
func ResetPriv() error {
	if err := syscall.Setegid(RootGid); err != nil {
		hwlog.RunLog.Errorf("set egid failed, error: %v", err)
		return errors.New("set egid failed")
	}
	if err := syscall.Seteuid(RootUid); err != nil {
		hwlog.RunLog.Errorf("set euid failed, error: %v", err)
		return errors.New("set euid failed")
	}

	return nil
}

// CheckParamOption check parameter in option slice
func CheckParamOption(optionalParam []string, inputParam string) error {
	for _, param := range optionalParam {
		if inputParam == param {
			return nil
		}
	}

	return errors.New("not support parameter")
}
