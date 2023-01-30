// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package common

import (
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"path"
	"strconv"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
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

func (u *UserMgr) createUser() error {
	if utils.IsExist(path.Join("/home", u.user)) {
		hwlog.RunLog.Error("user home dir exists, please remove it")
		return errors.New("user home dir exists")
	}
	noLogin, err := exec.LookPath("nologin")
	if err != nil {
		hwlog.RunLog.Errorf("look path of nologin failed, error: %s", err.Error())
		return errors.New("look path of nologin failed")
	}

	if _, err = user.LookupId(strconv.Itoa(u.uid)); err != nil {
		_, err = RunCommand("useradd", true, u.user, "-u", strconv.Itoa(u.uid), "-s", noLogin, "-M")
	} else {
		_, err = RunCommand("useradd", true, u.user, "-s", noLogin, "-M")
	}
	if err != nil {
		hwlog.RunLog.Errorf("exec useradd command failed, error: %s", err.Error())
		return errors.New("exec useradd command failed")
	}

	hwlog.RunLog.Infof("add user: %s, group: %s, uid: %d, gid: %d successfully", u.user, u.group, u.uid, u.gid)
	return nil
}

func (u *UserMgr) checkNoLogin() error {
	cmdStr := fmt.Sprintf(UserGrepCommandPattern, u.user)
	lines, err := RunCommand("sh", false, "-c", cmdStr)
	if err != nil {
		hwlog.RunLog.Errorf("exec check nologin command failed, error: %s", err.Error())
		return errors.New("exec check nologin command failed")
	}

	if lines != strconv.Itoa(NoLoginCount) {
		hwlog.RunLog.Errorf("the existing user [%s] is allowed to login", u.user)
		return fmt.Errorf("the existing user [%s] is allowed to login", u.user)
	}
	return nil
}

// AddUserAccount add a user account
func (u *UserMgr) AddUserAccount() error {
	var isUserExist, isGroupExist bool
	_, err := user.Lookup(u.user)
	if err == nil {
		hwlog.RunLog.Warnf("user [%s] exists in device", u.user)
		isUserExist = true
	}
	if _, err = user.LookupGroup(u.group); err == nil {
		hwlog.RunLog.Warnf("group [%s] exists in device", u.group)
		isGroupExist = true
	}
	if isUserExist && isGroupExist {
		if err = u.checkNoLogin(); err != nil {
			return err
		}
		hwlog.RunLog.Info("the existing user and group are valid,no need to create")
		return nil
	} else if isUserExist || isGroupExist {
		return errors.New("the user name or group name is in use")
	}
	return u.createUser()
}

// GetCurrentUser is used to get current username
func GetCurrentUser() (string, error) {
	userInfo, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("get current user info failed: %s", err)
	}

	return userInfo.Username, nil
}
