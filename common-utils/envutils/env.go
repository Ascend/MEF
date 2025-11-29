// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package envutils for get environment information
package envutils

import (
	"errors"
	"fmt"
	"math"
	"os/user"
	"strconv"
	"syscall"

	"huawei.com/mindx/common/terminal"
)

const (
	// tmpfsDevNum represents the dev number of tmpfs filesystem in linux stat struct
	tmpfsDevNum = 0x01021994
	oneMB       = 1024 * 1024
	parseBase10 = 10
	parseBit32  = 32
)

// linux command and path
const (
	cat      = "cat"
	uuidPath = "/proc/sys/kernel/random/uuid"
)

// GetFileSystem is used to get the file system of a path
// ret equals the type_t definition in linux statfs struct
func GetFileSystem(path string) (int64, error) {
	fileStat := syscall.Statfs_t{}
	if err := syscall.Statfs(path, &fileStat); err != nil {
		return 0, fmt.Errorf("get [%s]'s file system failed: %v", path, err)
	}

	return fileStat.Type, nil
}

// GetFileDevNum is used to get the dev info of a path
func GetFileDevNum(path string) (uint64, error) {
	fileStat := syscall.Stat_t{}
	if err := syscall.Stat(path, &fileStat); err != nil {
		return 0, fmt.Errorf("get [%s]'s file dev info failed: %v", path, err)
	}

	return fileStat.Dev, nil
}

// IsInTmpfs is used to check whether the path is in the temporary file system
func IsInTmpfs(path string) (bool, error) {
	dev, err := GetFileSystem(path)
	if err != nil {
		return false, err
	}
	return dev == tmpfsDevNum, nil
}

// GetDiskFree is used to get the free disk space of a path
func GetDiskFree(path string) (uint64, error) {
	fileStat := syscall.Statfs_t{}
	if err := syscall.Statfs(path, &fileStat); err != nil {
		return 0, err
	}

	diskFree := fileStat.Bavail * uint64(fileStat.Bsize)
	return diskFree, nil
}

// CheckDiskSpace is used to check whether the disk space on a path is enough
func CheckDiskSpace(path string, limit uint64) error {
	availSpace, err := GetDiskFree(path)
	if err != nil {
		return fmt.Errorf("get path [%s]'s disk available space failed: %v", path, err)
	}

	if availSpace < limit {
		fmt.Printf("warning: the path of [%s] disk space is not enough, at least %d MB is required\n",
			path, limit/oneMB)
		return errors.New("no enough space")
	}

	return nil
}

// GetUid is used to get user id by username
func GetUid(userName string) (uint32, error) {
	userInfo, err := user.Lookup(userName)
	if err != nil {
		return 0, fmt.Errorf("lookup user [%s] failed: %v", userName, err)
	}
	uid64, err := strconv.ParseUint(userInfo.Uid, parseBase10, parseBit32)
	if err != nil {
		return 0, fmt.Errorf("convert user [%s]'s uid to uint64 failed: %v", userName, err)
	}
	if uid64 > math.MaxUint32 {
		return 0, errors.New("can not convert uid to uint32")
	}
	return uint32(uid64), nil
}

// GetGid is used to get group id by group name
func GetGid(groupName string) (uint32, error) {
	groupInfo, err := user.LookupGroup(groupName)
	if err != nil {
		return 0, fmt.Errorf("lookup group [%s] failed: %v", groupName, err)
	}
	gid64, err := strconv.ParseUint(groupInfo.Gid, parseBase10, parseBit32)
	if err != nil {
		return 0, fmt.Errorf("convert group [%s]'s gid to uint64 failed: %v", groupName, err)
	}
	if gid64 > math.MaxUint32 {
		return 0, errors.New("can not convert uid to uint32")
	}
	return uint32(gid64), nil
}

// GetCurrentUser is used to get current username
func GetCurrentUser() (string, error) {
	userInfo, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("get current user info failed: %v", err)
	}

	return userInfo.Username, nil
}

// GetUserAndIP is used to get current user and ssh ip
func GetUserAndIP() (string, string, error) {
	curUser, err := GetCurrentUser()
	if err != nil {
		return "", "", err
	}
	_, sshIp, err := terminal.GetLoginUserAndIP()
	if err != nil {
		sshIp = "localhost"
	}
	return curUser, sshIp, nil
}

// CheckUserIsRoot is used to check whether the current user is root, returns error if not
func CheckUserIsRoot() error {
	curUser, err := GetCurrentUser()
	if err != nil {
		return err
	}

	if curUser != "root" {
		return fmt.Errorf("the current user must be root, can not be %s", curUser)
	}
	return nil
}

// GetUuid get uuid
func GetUuid() (string, error) {
	uuid, err := RunCommand(cat, DefCmdTimeoutSec, uuidPath)
	if err != nil {
		return "", fmt.Errorf("get uuid failed, error: %v", err)
	}
	return uuid, nil
}
