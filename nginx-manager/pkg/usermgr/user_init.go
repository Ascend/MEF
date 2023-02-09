// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package usermgr this package is for manage user
package usermgr

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"nginx-manager/pkg/database"
	"nginx-manager/pkg/nginxcom"
)

const (
	unlockInterval = time.Second * 30
)

type handlerFunc func(req interface{}) common.RespMsg

type userManager struct {
	enable bool
	ctx    context.Context
}

// NewUserManager create app manager
func NewUserManager(enable bool, ctx context.Context) *userManager {
	am := &userManager{
		enable: enable,
		ctx:    ctx,
	}
	return am
}

func (u *userManager) Name() string {
	return nginxcom.UserManagerName
}

func (u *userManager) Enable() bool {
	if u.enable {
		if err := initTable(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init database table failed, cannot enable", u.Name())
			return !u.enable
		}
		if err := createDefaultUser(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init database table failed, cannot enable", u.Name())
			return !u.enable
		}
	}
	return u.enable
}

func (u *userManager) Start() {
	go u.intervalUnlockUserAndIp()
	for {
		select {
		case _, ok := <-u.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("user manager catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("user manager has listened stop signal")
			return
		default:
		}
		req, err := modulemanager.ReceiveMessage(u.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", u.Name())
			continue
		}
		msg := methodSelect(req)
		if msg == nil {
			hwlog.RunLog.Errorf("%s get method by option and resource failed", u.Name())
			continue
		}
		resp, err := req.NewResponse()
		if err != nil {
			hwlog.RunLog.Errorf("%s new response failed", u.Name())
			continue
		}
		resp.FillContent(msg)
		if err = modulemanager.SendMessage(resp); err != nil {
			hwlog.RunLog.Errorf("%s send response failed", u.Name())
			continue
		}
	}
}

func (u *userManager) intervalUnlockUserAndIp() {
	timer := time.NewTimer(unlockInterval)
	for {
		select {
		case _, ok := <-u.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		case <-timer.C:
			timer.Reset(unlockInterval)
			router := common.Router{
				Source:      common.RestfulServiceName,
				Destination: u.Name(),
				Option:      http.MethodPost,
				Resource:    filepath.Join(userMgrPath, "interval-unlock"),
			}
			common.SendSyncMessageByRestful("", &router)
		}
	}
}

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, exit := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !exit {
		hwlog.RunLog.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
		return nil
	}
	res = method(req.GetContent())
	return &res
}

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, filepath.Join(userMgrPath, "login")):           Login,
	common.Combine(http.MethodPatch, filepath.Join(userMgrPath, "first-change")):   FirstChange,
	common.Combine(http.MethodPatch, filepath.Join(userMgrPath, "change")):         Change,
	common.Combine(http.MethodPost, filepath.Join(userMgrPath, "interval-unlock")): intervalUnlock,
	common.Combine(http.MethodPost, filepath.Join(userMgrPath, "islocked")):        svcIpLocked,
}

func initTable() error {
	if err := database.CreateTableIfNotExists(User{}); err != nil {
		hwlog.RunLog.Error("create database table user failed")
		return err
	}
	if err := database.CreateTableIfNotExists(IpForbidden{}); err != nil {
		hwlog.RunLog.Error("create database table ip_forbidden failed")
		return err
	}
	if err := database.CreateTableIfNotExists(HistoryPassword{}); err != nil {
		hwlog.RunLog.Error("create database table history_password failed")
		return err
	}
	return nil
}
