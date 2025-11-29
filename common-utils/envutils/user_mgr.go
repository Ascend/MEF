// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package envutils for add user account
package envutils

import (
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

const (
	noLoginPattern        = "^%s:.*:%s$"
	noOtherInGroupPattern = "^%s:.*:$"
	userAddCommand        = "useradd"
	homeDir               = "/home"
	noLogin               = "nologin"
	etcPasswdFile         = "/etc/passwd"
	etcGroupFile          = "/etc/group"
	errUnknownUid         = "user: unknown userid"
)

// UserMgr is used to mgr user
type UserMgr struct {
	user  string
	group string
	uid   int
	gid   int
}

// NewUserMgr create user manager instance
func NewUserMgr(user, group string, uid, gid int) *UserMgr {
	return &UserMgr{
		user:  user,
		group: group,
		uid:   uid,
		gid:   gid,
	}
}

// AddUserAccount add a user account
func (u *UserMgr) AddUserAccount() error {
	noLoginPath, err := exec.LookPath(noLogin)
	if err != nil {
		hwlog.RunLog.Errorf("look path of nologin failed, error: %s", err.Error())
		return errors.New("look path of nologin failed")
	}

	var isUserExist, isGroupExist bool
	userInfo, err := user.Lookup(u.user)
	if err == nil {
		hwlog.RunLog.Warnf("user [%s] exists in device", u.user)
		isUserExist = true
	}
	groupInfo, err := user.LookupGroup(u.group)
	if err == nil {
		hwlog.RunLog.Warnf("group [%s] exists in device", u.group)
		isGroupExist = true
	}
	if isUserExist && isGroupExist {
		if fileutils.IsExist(userInfo.HomeDir) {
			hwlog.RunLog.Error("user home dir exists, please remove it")
			return errors.New("user home dir exists")
		}
		gIds, err := userInfo.GroupIds()
		if err != nil {
			hwlog.RunLog.Errorf("get user groups failed, error: %s", err.Error())
			return fmt.Errorf("get user groups failed, error: %s", err.Error())
		}
		if !isInGroup(groupInfo.Gid, gIds) {
			hwlog.RunLog.Errorf("the existing user [%s] is not in group [%s]", u.user, u.group)
			return fmt.Errorf("the existing user [%s] is not in group [%s]", u.user, u.group)
		}
		if err = u.checkNoLogin(noLoginPath); err != nil {
			return err
		}
		if err = u.checkGroupContainOtherUser(); err != nil {
			return err
		}
		hwlog.RunLog.Info("the existing user and group are valid, do not need to create")
		return nil
	}
	if isUserExist || isGroupExist {
		hwlog.RunLog.Error("the user name or group name is in use")
		return errors.New("the user name or group name is in use")
	}
	return u.createUser(noLoginPath)
}

func (u *UserMgr) createUser(noLoginPath string) error {
	if fileutils.IsExist(filepath.Join(homeDir, u.user)) {
		hwlog.RunLog.Error("user home dir exists, please remove it")
		return errors.New("user home dir exists")
	}

	var runCmdErr error
	if _, lookupErr := user.LookupId(strconv.Itoa(u.uid)); lookupErr == nil {
		_, runCmdErr = RunCommand(userAddCommand, DefCmdTimeoutSec, u.user, "-s", noLoginPath, "-M")
	} else if strings.Contains(lookupErr.Error(), errUnknownUid) {
		_, runCmdErr = RunCommand(userAddCommand, DefCmdTimeoutSec,
			u.user, "-u", strconv.Itoa(u.uid), "-s", noLoginPath, "-M")
	} else {
		hwlog.RunLog.Errorf("lookup uid [%d] failed, error: %s", u.uid, lookupErr.Error())
		return errors.New("lookup uid failed")
	}
	if runCmdErr != nil {
		hwlog.RunLog.Errorf("exec useradd command failed, error: %s", runCmdErr.Error())
		return errors.New("exec useradd command failed")
	}

	hwlog.RunLog.Infof("add user: %s, group: %s successfully", u.user, u.group)
	return nil
}

func (u *UserMgr) checkNoLogin(noLoginPath string) error {
	noLoginReg := fmt.Sprintf(noLoginPattern, u.user, noLoginPath)
	found, err := isInfoInFile(etcPasswdFile, noLoginReg)
	if err != nil {
		hwlog.RunLog.Errorf("check user is nologin failed: %s", err.Error())
		return errors.New("check user is nologin failed")
	}
	if found {
		return nil
	}

	hwlog.RunLog.Error("user does not have nologin attribute")
	return errors.New("user does not have nologin attribute")
}

func (u *UserMgr) checkGroupContainOtherUser() error {
	noOtherInGroupReg := fmt.Sprintf(noOtherInGroupPattern, u.group)
	found, err := isInfoInFile(etcGroupFile, noOtherInGroupReg)
	if err != nil {
		hwlog.RunLog.Errorf("check no others in group failed: %s", err.Error())
		return errors.New("check no others in group failed")
	}
	if found {
		return nil
	}

	hwlog.RunLog.Error("the existing group contains other user")
	return errors.New("the existing group contains other user")
}

func isInfoInFile(file string, pattern string) (bool, error) {
	data, err := fileutils.LoadFile(file)
	if err != nil {
		hwlog.RunLog.Errorf("load file failed, error: %s", err.Error())
		return false, errors.New("load file failed")
	}
	if data == nil {
		hwlog.RunLog.Error("content of file is nil")
		return false, errors.New("content of file is nil")
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		found, err := regexp.MatchString(pattern, line)
		if err != nil {
			hwlog.RunLog.Errorf("regexp match failed: %s", err.Error())
			return false, errors.New("regexp match failed")
		}
		if found {
			return true, nil
		}
	}
	return false, nil
}

func isInGroup(groupId string, groupIds []string) bool {
	for _, gid := range groupIds {
		if gid == groupId {
			return true
		}
	}
	return false
}
