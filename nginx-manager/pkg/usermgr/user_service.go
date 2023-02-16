// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package usermgr this package is for manage user
package usermgr

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/passutils"

	"nginx-manager/pkg/nginxcom"
)

var defaultUsername = os.Getenv(nginxcom.DefaultUsernameKey)

func createDefaultUser() error {
	user, err := UserServiceInstance().getUserByName(defaultUsername)
	// user already exist
	if err == nil {
		hwlog.RunLog.Warn("user exist, no need to create")
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		hwlog.RunLog.Errorf("createDefaultUser query db error, user: %s", defaultUsername)
		return err
	}
	now := time.Now().Format(common.TimeFormat)
	user = &User{
		Username:  defaultUsername,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err = UserServiceInstance().createUser(user); err != nil {
		hwlog.RunLog.Errorf("insert user to db failed, user: %s", user.Username)
		return fmt.Errorf("insert user to db failed")
	}
	hwlog.RunLog.Info("insert user to db success")
	return nil
}

// FirstChange change default user password after deployment the mef system
func FirstChange(input interface{}) common.RespMsg {
	var req firstChangePwdReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("first change pwd convert param error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
	}
	if checkResult := newFirstChangeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("first change pwd check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: ""}
	}
	defer func() {
		common.ClearStringMemory(*req.Password)
		common.ClearStringMemory(*req.RePassword)
	}()
	if *req.Password != *req.RePassword {
		hwlog.RunLog.Error("password and rePassword not equal")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "", Data: nil}
	}
	cachedUser, resp := checkGetUserFirstLogin(*req.Username)
	if resp.Status != common.Success {
		return resp
	}
	if err := passutils.CheckPassWord(cachedUser.Username, req.Password); err != nil {
		hwlog.RunLog.Errorf("change password err: %s", err.Error())
		return common.RespMsg{Status: common.ErrorChangePassword, Msg: "", Data: nil}
	}
	updateSuccess, encryptPassWord, saltString := updatePwd(*req.Username, req.Password)
	if !updateSuccess {
		return common.RespMsg{Status: common.ErrorChangePassword, Msg: "", Data: nil}
	}
	saveHistoryPassword(encryptPassWord, saltString, cachedUser.ID)
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// Login handle user login action
func Login(input interface{}) common.RespMsg {
	var req loginReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("user login convert param error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
	}
	if checkResult := newLoginChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("login check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorLogin}
	}
	defer func() {
		common.ClearStringMemory(*req.Password)
	}()
	user, resp := checkGetUser(*req.Username)
	if resp.Status != common.Success {
		return resp
	}
	_, err := UserServiceInstance().getForbiddenIp(*req.Ip)
	// ip is frozen
	if err == nil {
		return common.RespMsg{Status: common.ErrorLockState, Msg: "", Data: nil}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		hwlog.RunLog.Infof("user %s login query ip failed, error: %s", user.Username, err)
		return common.RespMsg{Status: common.ErrorLogin, Msg: "", Data: nil}
	}
	// password is wrong
	if resp = dealLockAndComparePwd(*req.Ip, req.Password, user); resp.Status != common.Success {
		return resp
	}
	if err = updateUserLogin(user); err != nil {
		return common.RespMsg{Status: common.ErrorLogin, Msg: "", Data: nil}
	}
	dbUser, resp := checkGetUser(user.Username)
	if resp.Status != common.Success {
		return resp
	}
	// login success
	hwlog.RunLog.Infof("user %s login success", user.Username)
	return common.RespMsg{Status: common.Success, Msg: "", Data: dbUser}
}

// Change change user's password
func Change(input interface{}) common.RespMsg {
	var req changePwdReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("change pwd convert param error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
	}
	if checkResult := newChangeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("change pwd check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: ""}
	}
	defer func() {
		common.ClearStringMemory(*req.OldPassword)
		common.ClearStringMemory(*req.Password)
		common.ClearStringMemory(*req.RePassword)
	}()
	if *req.Password != *req.RePassword {
		hwlog.RunLog.Error("password and rePassword not equal")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "", Data: nil}
	}
	cachedUser, resp := checkGetUser(*req.Username)
	if resp.Status != common.Success {
		return resp
	}
	if resp = dealLockAndComparePwd(*req.Ip, req.OldPassword, cachedUser); resp.Status != common.Success {
		return resp
	}
	if err := passutils.CheckPassWord(cachedUser.Username, req.Password); err != nil {
		hwlog.RunLog.Errorf("change password err: %s", err.Error())
		return common.RespMsg{Status: common.ErrorChangePassword, Msg: "", Data: nil}
	}
	if resp = checkHistoryPassword(req.Password, cachedUser); resp.Status != common.Success {
		return resp
	}
	updateSuccess, encryptPassWord, saltString := updatePwd(*req.Username, req.Password)
	if !updateSuccess {
		return common.RespMsg{Status: common.ErrorChangePassword, Msg: "", Data: nil}
	}
	saveHistoryPassword(encryptPassWord, saltString, cachedUser.ID)
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// Logout handle user logout action
func Logout(input interface{}) common.RespMsg {
	var req logoutReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("logout convert param error: %s", err.Error())
		hwlog.OpLog.Errorf("[%v]user admin logout fail", *req.Ip)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
	}
	hwlog.RunLog.Info("user admin logout success")
	hwlog.OpLog.Infof("[%v]user admin logout success", *req.Ip)
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func updatePwd(username string, password *string) (bool, string, string) {
	encryptPassWord, saltString, err := passutils.GetEncryptPassword(password)
	if err != nil {
		hwlog.RunLog.Errorf("change password err: %s", err.Error())
		return false, "", ""
	}
	user := User{
		Username:           username,
		Password:           encryptPassWord,
		Salt:               saltString,
		FirstLogin:         false,
		PasswordWrongTimes: 0,
	}
	if err := UserServiceInstance().updatePassword(&user); err != nil {
		hwlog.RunLog.Errorf("change password error, user: %s", username)
		return false, "", ""
	}
	hwlog.RunLog.Info("change password success")
	return true, encryptPassWord, saltString

}

func checkHistoryPassword(newPassword *string, cacheUser *User) common.RespMsg {
	historyPasswords, err := UserServiceInstance().getHistoryPasswords(cacheUser.ID)
	if err != nil {
		hwlog.RunLog.Errorf("query history password err when check, user: %s", cacheUser.Username)
		return common.RespMsg{Status: common.ErrorQueryHisPassword, Msg: "", Data: nil}
	}
	for _, v := range *historyPasswords {
		if passutils.ComparePassword(newPassword, v.HistoryPassword, v.Salt) {
			hwlog.RunLog.Errorf("check history password err, user: %s", cacheUser.Username)
			return common.RespMsg{Status: common.ErrorPasswordRepeat, Msg: "", Data: nil}
		}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func saveHistoryPassword(hashVal string, newSalt string, userId uint64) {
	historyPasswords, err := UserServiceInstance().getHistoryPasswords(userId)
	if err != nil {
		hwlog.RunLog.Error("save password err")
		return
	}
	var deleteHistory HistoryPassword
	hisPassLen := len(*historyPasswords)
	if hisPassLen >= nginxcom.HistoryPasswordSaveCount {
		tmpPasses := *historyPasswords
		deleteHistory = tmpPasses[hisPassLen-1]
		if err := UserServiceInstance().deleteHistoryPassword(deleteHistory.ID); err != nil {
			hwlog.RunLog.Errorf("delete history password fail, userId: %d", deleteHistory.UserId)
		}
	}
	toUpdateHisPass := &HistoryPassword{
		UserId:          userId,
		HistoryPassword: hashVal,
		Salt:            newSalt,
		CreatedAt:       time.Now().Format(common.TimeFormat),
	}
	if err := UserServiceInstance().createHistoryPassword(toUpdateHisPass); err != nil {
		hwlog.RunLog.Errorf("save history password fail, userId: d%", userId)
		return
	}
	hwlog.RunLog.Info("save history password success")
}

func updateUserLogin(user *User) error {
	now := time.Now().Format(common.TimeFormat)
	updateUser := &User{
		Username:           user.Username,
		LoginTime:          now,
		PasswordWrongTimes: 0,
	}
	if err := UserServiceInstance().updateUserLogin(updateUser); err != nil {
		hwlog.RunLog.Errorf("update login error, user: %s", user.Username)
		return fmt.Errorf("update login error, user: %s", user.Username)
	}
	return nil
}

func checkGetUser(username string) (*User, common.RespMsg) {
	cachedUser, err := UserServiceInstance().getUserByName(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hwlog.RunLog.Errorf("user %s not found", username)
		return nil, common.RespMsg{Status: common.ErrorPassOrUser, Msg: "", Data: nil}
	}
	if err != nil {
		hwlog.RunLog.Errorf("query user %s error", username)
		return nil, common.RespMsg{Status: common.ErrorLogin, Msg: "", Data: nil}
	}
	if cachedUser.LockState {
		hwlog.RunLog.Errorf("user %s in lock state", username)
		return nil, common.RespMsg{Status: common.ErrorLockState, Msg: "", Data: nil}
	}
	if cachedUser.FirstLogin {
		hwlog.RunLog.Errorf("user %s is in first login, cannot operate", username)
		return nil, common.RespMsg{Status: common.ErrorNeedFirstLogin, Msg: "", Data: nil}
	}
	return cachedUser, common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func checkGetUserFirstLogin(username string) (*User, common.RespMsg) {
	cachedUser, err := UserServiceInstance().getUserByName(username)
	if err != nil {
		hwlog.RunLog.Errorf("first change password query user: %s", username)
		return nil, common.RespMsg{Status: common.ErrorChangePassword, Msg: "", Data: nil}
	}
	if !cachedUser.FirstLogin {
		hwlog.RunLog.Errorf("user %s not first login, cannot change password from interface firstChange", username)
		return nil, common.RespMsg{Status: common.ErrorUserAlreadyFirstLogin, Msg: "", Data: nil}
	}
	return cachedUser, common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
