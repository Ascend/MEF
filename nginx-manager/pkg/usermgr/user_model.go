// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package usermgr this package is for manage user
package usermgr

import (
	"sync"

	"gorm.io/gorm"

	"nginx-manager/pkg/database"
)

var (
	userServiceSingleton sync.Once
	userServiceInstance  UserService
)

type baseReq struct {
	Ip string `json:"ip"`
}

type firstChangePwdReq struct {
	baseReq
	Username   string `json:"username"`
	Password   []byte `json:"password"`
	RePassword []byte `json:"rePassword"`
}

type changePwdReq struct {
	baseReq
	Username    string `json:"username"`
	Password    []byte `json:"password"`
	OldPassword []byte `json:"oldPassword"`
	RePassword  []byte `json:"rePassword"`
}

type loginReq struct {
	baseReq
	Username string `json:"username"`
	Password []byte `json:"password"`
}

type queryIpLockReq struct {
	baseReq
	targetIp string `json:"targetIp"`
}

type lockInfoResp struct {
	userLocked bool   `json:"userLocked"`
	ipLocked   bool   `json:"ipLocked"`
	userid     uint64 `json:"userid"`
	ip         string `json:"ip"`
}

// UserServiceImpl  an implement for UserService
type UserServiceImpl struct {
	db *gorm.DB
}

// UserService the userService interface
type UserService interface {
	createUser(*User) error
	getUserByName(string) (*User, error)
	getLockedUsers() (*[]User, error)
	updatePassword(*User) error
	updateUserLogin(*User) error
	updateUserLock(*User) error
	deleteForbiddenIp(string) error
	createForbiddenIp(*IpForbidden) error
	getForbiddenIp(string) (*IpForbidden, error)
	getForbiddenIps() (*[]IpForbidden, error)
	getHistoryPasswords(uint64) (*[]HistoryPassword, error)
	createHistoryPassword(*HistoryPassword) error
	deleteHistoryPassword(uint64) error
}

// UserServiceInstance get user db service
func UserServiceInstance() UserService {
	userServiceSingleton.Do(func() {
		userServiceInstance = &UserServiceImpl{db: database.GetDb()}
	})
	return userServiceInstance
}

func (u *UserServiceImpl) createUser(user *User) error {
	return u.db.Model(User{}).Create(user).Error
}

func (u *UserServiceImpl) getUserByName(username string) (*User, error) {
	var user User
	return &user, u.db.Model(User{}).Where("username = ?", username).First(&user).Error
}

func (u *UserServiceImpl) updatePassword(user *User) error {
	entity := map[string]interface{}{
		"Password":           user.Password,
		"Salt":               user.Salt,
		"FirstLogin":         user.FirstLogin,
		"PasswordWrongTimes": user.PasswordWrongTimes,
	}
	return u.updateUserByName(user.Username, entity)
}

func (u *UserServiceImpl) updateUserLogin(user *User) error {
	entity := map[string]interface{}{
		"PasswordWrongTimes": user.PasswordWrongTimes,
		"LoginTime":          user.LoginTime,
	}
	return u.updateUserByName(user.Username, entity)
}

func (u *UserServiceImpl) updateUserLock(user *User) error {
	entity := map[string]interface{}{
		"PasswordWrongTimes": user.PasswordWrongTimes,
		"LoginFailTime":      user.LoginFailTime,
		"LockTime":           user.LockTime,
		"LockState":          user.LockState,
	}
	return u.updateUserByName(user.Username, entity)
}

func (u *UserServiceImpl) updateUserByName(username string, columns map[string]interface{}) error {
	return u.db.Model(&User{}).Where("username = ?", username).UpdateColumns(columns).Error
}

func (u *UserServiceImpl) createForbiddenIp(forbidden *IpForbidden) error {
	return u.db.Model(IpForbidden{}).Create(forbidden).Error
}

func (u *UserServiceImpl) getForbiddenIp(ip string) (*IpForbidden, error) {
	var ipForbidden IpForbidden
	return &ipForbidden, u.db.Model(IpForbidden{}).Where("ip = ?", ip).First(&ipForbidden).Error
}

func (u *UserServiceImpl) deleteForbiddenIp(ip string) error {
	return u.db.Model(&IpForbidden{}).Where("ip = ?",
		ip).Delete(&IpForbidden{}).Error
}

func (u *UserServiceImpl) getLockedUsers() (*[]User, error) {
	var users []User
	return &users, u.db.Model(&User{}).Where("lock_state = ?", true).Find(&users).Error
}

func (u *UserServiceImpl) getForbiddenIps() (*[]IpForbidden, error) {
	var ips []IpForbidden
	return &ips, u.db.Model(&IpForbidden{}).Find(&ips).Error
}

func (u *UserServiceImpl) getHistoryPasswords(userId uint64) (*[]HistoryPassword, error) {
	var historyPassword []HistoryPassword
	return &historyPassword, u.db.Model(&HistoryPassword{}).Where("user_id = ?", userId).
		Order("created_at desc").Find(&historyPassword).Error
}

func (u *UserServiceImpl) createHistoryPassword(hisPass *HistoryPassword) error {
	return u.db.Model(HistoryPassword{}).Create(hisPass).Error
}

func (u *UserServiceImpl) deleteHistoryPassword(id uint64) error {
	return u.db.Model(HistoryPassword{}).Where("id = ?", id).Delete(&HistoryPassword{}).Error
}
