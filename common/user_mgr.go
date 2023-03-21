// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package common

import (
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"

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
		_, err = RunCommand("useradd", true, DefCmdTimeoutSec,
			u.user, "-u", strconv.Itoa(u.uid), "-s", noLogin, "-M")
	} else {
		_, err = RunCommand("useradd", true, DefCmdTimeoutSec, u.user, "-s", noLogin, "-M")
	}
	if err != nil {
		hwlog.RunLog.Errorf("exec useradd command failed, error: %s", err.Error())
		return errors.New("exec useradd command failed")
	}

	hwlog.RunLog.Infof("add user: %s, group: %s, uid: %d, gid: %d successfully", u.user, u.group, u.uid, u.gid)
	return nil
}

func (u *UserMgr) checkNoLogin() error {
	userReg := fmt.Sprintf(UserGrepCommandPattern, u.user)
	ret, err := RunCommand(GrepCommand, true, DefCmdTimeoutSec, userReg, EtcPasswdFile)
	if err != nil {
		hwlog.RunLog.Errorf("exec check nologin command failed, error: %s", err.Error())
		return errors.New("exec check nologin command failed")
	}

	lines := strings.Split(ret, "\n")
	for _, line := range lines {
		found, err := regexp.MatchString(NoLoginFlag, line)
		if err != nil {
			hwlog.RunLog.Errorf("check if user is no login on reg match failed: %s", err.Error())
			return errors.New("check if user is no login failed")
		}
		if found {
			return nil
		}
	}

	return errors.New("user does not have nologin attribute")
}

// AddUserAccount add a user account
func (u *UserMgr) AddUserAccount() error {
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
		gIds, err := userInfo.GroupIds()
		if err != nil {
			hwlog.RunLog.Errorf("get user groups failed,error:%v", err)
			return err
		}
		if !isInGroup(groupInfo.Gid, gIds) {
			hwlog.RunLog.Errorf("the existing user[%s] is not in group[%s]", u.user, u.group)
			return fmt.Errorf("the existing user[%s] is not in group[%s]", u.user, u.group)
		}
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
		return "", fmt.Errorf("get current user info failed: %s", err.Error())
	}

	return userInfo.Username, nil
}

func isInGroup(groupId string, groupIds []string) bool {
	for _, gid := range groupIds {
		if gid == groupId {
			return true
		}
	}
	return false
}
