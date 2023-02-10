// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package usermgr this package is for manage user
package usermgr

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/passutils"

	"nginx-manager/pkg/nginxcom"
)

// dealLockAndComparePwd 比较密码并锁定
func dealLockAndComparePwd(clientIp string, targetPass []byte, user *User) common.RespMsg {
	unlockUser(user)
	forbiddenIp, err := UserServiceInstance().getForbiddenIp(clientIp)
	if err == nil {
		unlockIp(forbiddenIp)
	}
	return comparePwdAndLock(clientIp, targetPass, user)
}

func comparePwdAndLock(clientIp string, targetPass []byte, user *User) common.RespMsg {
	if passutils.ComparePassword(targetPass, user.Password, user.Salt) {
		return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
	}
	exceedMaxCount := false
	if user.PasswordWrongTimes+1 >= nginxcom.MaxPwdWrongTimes {
		exceedMaxCount = true
	}
	lockUser(user, exceedMaxCount)
	lockIp(clientIp, exceedMaxCount)
	if exceedMaxCount {
		lockInfo := lockInfoResp{UserLocked: false, IpLocked: true, Userid: user.ID, Ip: clientIp}
		hwlog.RunLog.Errorf("compare password fail, lock user: %d, lock ip: %s", lockInfo.Userid, clientIp)
		return common.RespMsg{Status: common.ErrorPassOrUser, Msg: "", Data: lockInfo}
	}
	hwlog.RunLog.Errorf("compare password fail, user: %s", user.Username)
	return common.RespMsg{Status: common.ErrorPassOrUser, Msg: "", Data: nil}
}

func lockUser(user *User, exceedMaxCount bool) {
	now := time.Now().Format(common.TimeFormat)
	updateUser := &User{}
	updateUser.Username = user.Username
	updateUser.LoginFailTime = now
	updateUser.PasswordWrongTimes = user.PasswordWrongTimes + 1
	if exceedMaxCount {
		updateUser.LockTime = now
		updateUser.LockState = true
	} else {
		updateUser.LockTime = user.LockTime
		updateUser.LockState = false
	}
	if err := UserServiceInstance().updateUserLock(updateUser); err != nil {
		hwlog.RunLog.Errorf("update User Lock error, user: %s", user.Username)
	}
}

func lockIp(ip string, exceedMaxCount bool) {
	if !exceedMaxCount {
		return
	}
	hwlog.RunLog.Infof("lock ip: %s", ip)
	now := time.Now().Format(common.TimeFormat)
	updateIp := &IpForbidden{}
	updateIp.Ip = ip
	updateIp.LockTime = now
	if err := UserServiceInstance().createForbiddenIp(updateIp); err != nil {
		hwlog.RunLog.Error("update forbiddenIp error")
	}
}

func unlockUser(user *User) {
	if !user.LockState {
		return
	}
	// 解锁
	lockTime, err := time.Parse(common.TimeFormat, user.LockTime)
	if err != nil {
		hwlog.RunLog.Errorf("unlock user parse lock time error, %s", err.Error())
		return
	}
	if time.Now().Sub(lockTime) >= nginxcom.UserLockTime {
		hwlog.RunLog.Infof("unlock user: %s", user.Username)
		toUpdateUser := &User{
			Username:           user.Username,
			LoginFailTime:      user.LoginFailTime,
			PasswordWrongTimes: user.PasswordWrongTimes,
			LockTime:           user.LockTime,
			LockState:          false,
		}
		if err := UserServiceInstance().updateUserLock(toUpdateUser); err != nil {
			hwlog.RunLog.Errorf("update unlock user error, user: %s", user.Username)
		}
	}
}

func unlockIp(forbidden *IpForbidden) {
	lockTime, err := time.Parse(common.TimeFormat, forbidden.LockTime)
	if err != nil {
		hwlog.RunLog.Errorf("unlock user parse lock time error, %s", err.Error())
		return
	}
	if time.Now().Sub(lockTime) >= nginxcom.IpLockTime {
		hwlog.RunLog.Infof("unlock ip: %s", forbidden.Ip)
		if err := UserServiceInstance().deleteForbiddenIp(forbidden.Ip); err != nil {
			hwlog.RunLog.Error("delete forbidden ip error")
		}
	}
}

func intervalUnlock(input interface{}) common.RespMsg {
	intervalUnlockUser()
	intervalUnlockIp()
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func svcLocked(input interface{}) common.RespMsg {
	var req queryIpLockReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("query lock convert param error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
	}
	ipLocked := true
	if len(req.TargetIp) > 0 {
		_, err := UserServiceInstance().getForbiddenIp(req.TargetIp)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ipLocked = false
		} else if err != nil {
			hwlog.RunLog.Errorf("query lock ip error: %s", err.Error())
			return common.RespMsg{Status: common.ErrorQueryLock, Msg: "", Data: nil}
		}
	}
	user, err := UserServiceInstance().getUserById(1)
	if err != nil {
		hwlog.RunLog.Errorf("query lock user error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorQueryLock, Msg: "", Data: nil}
	}
	lockInfo := lockInfoResp{
		UserLocked: user.LockState,
		IpLocked:   ipLocked,
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: lockInfo}
}

func intervalUnlockUser() {
	lockedUsers, err := UserServiceInstance().getLockedUsers()
	if err != nil {
		hwlog.RunLog.Error("query locked users error")
		return
	}
	if lockedUsers != nil {
		for _, user := range *lockedUsers {
			unlockUser(&user)
		}
	}
}

func intervalUnlockIp() {
	lockedIps, err := UserServiceInstance().getForbiddenIps()
	if err != nil {
		hwlog.RunLog.Error("query locked ips error")
		return
	}
	if lockedIps != nil {
		for _, ip := range *lockedIps {
			unlockIp(&ip)
		}
	}
}
