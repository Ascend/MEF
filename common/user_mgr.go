// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package common

import (
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"path"
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
	group, err := user.LookupGroupId(strconv.Itoa(u.gid))
	if err == nil && group.Name != u.group {
		hwlog.RunLog.Errorf("group id %s has already been occupied but the group is incorrect", u.group)
		return errors.New("group id has already been occupied but the group is incorrect")
	}

	if err != nil && strings.Contains(err.Error(), "group: unknown group") {
		hwlog.RunLog.Errorf("look up for group %d failed: %s ", u.gid, err.Error())
		return errors.New("look up for group id failed")
	}

	if utils.IsExist(path.Join("/home", u.user)) {
		hwlog.RunLog.Error("user home dir exists, please remove it")
		return errors.New("user home dir exists")
	}

	noLogin, err := exec.LookPath("nologin")
	if err != nil {
		hwlog.RunLog.Errorf("look path of nologin failed, error: %s", err.Error())
		return errors.New("look path of nologin failed")
	}

	if _, err = RunCommand("useradd", true, u.user, "-u", strconv.Itoa(u.uid), "-s", noLogin); err != nil {
		hwlog.RunLog.Errorf("exec useradd command failed, error: %s", err.Error())
		return errors.New("exec useradd command failed")
	}

	hwlog.RunLog.Infof("add user: %s, group: %s, uid: %d, gid: %d successfully", u.user, u.group, u.uid, u.gid)
	return nil
}

func (u *UserMgr) checkConflict(userInfo *user.User) error {
	if userInfo == nil {
		hwlog.RunLog.Error("pointer userInfo is nil")
		return errors.New("pointer userInfo is nil")
	}

	if userInfo.Uid != strconv.Itoa(u.uid) || userInfo.Gid != strconv.Itoa(u.gid) {
		hwlog.RunLog.Errorf("system already has user: %s, uid: %s, gid: %s, conflict detected",
			u.user, userInfo.Uid, userInfo.Gid)
		return errors.New("system already has user")
	}

	groupInfo, err := user.LookupGroupId(userInfo.Gid)
	if err != nil {
		hwlog.RunLog.Error("get group info failed")
		return errors.New("get group info failed")
	}

	if groupInfo.Name != u.group {
		hwlog.RunLog.Errorf("system already has another group for uid %s", userInfo.Gid)
		return errors.New("check group name failed")
	}

	if err = u.checkNoLogin(); err != nil {
		return err
	}

	hwlog.RunLog.Info("check user conflict success in device")
	return nil
}

func (u *UserMgr) checkNoLogin() error {
	cmdStr := fmt.Sprintf(UserGrepCommandPattern, strconv.Itoa(u.uid))
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
	userInfo, err := user.Lookup(u.user)
	if err != nil {
		hwlog.RunLog.Warnf("user [%s] not exists in device, begin creating", u.user)
		return u.createUser()
	}
	return u.checkConflict(userInfo)
}
