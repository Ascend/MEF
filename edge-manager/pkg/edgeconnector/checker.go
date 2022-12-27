// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the checker used to check baseInfo
package edgeconnector

import (
	"errors"
	"net"
	"strconv"

	"edge-manager/pkg/util"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"
)

func (base *baseInfo) checkBaseInfo() error {
	var checkBaseInfo = []func() error{
		base.checkAddr,
		base.checkPort,
		base.checkName,
		base.checkPwd,
	}

	for _, function := range checkBaseInfo {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (base *baseInfo) checkAddr() error {
	if base.Address == "" || base.Address == common.ZeroAddr || base.Address == common.BroadCastAddr {
		hwlog.RunLog.Error("check address failed, address is not allowed")
		return errors.New("check address failed")
	}
	addr := net.ParseIP(base.Address)
	if addr == nil || addr.To4() == nil {
		hwlog.RunLog.Error("check address failed, address invalid")
		return errors.New("check address failed")
	}
	hwlog.RunLog.Info("check address success")
	return nil
}

func (base *baseInfo) checkPort() error {
	portInt, err := strconv.Atoi(base.Port)
	if err != nil {
		return errors.New("convert port to int failed")
	}
	if !util.CheckInt(portInt, common.MinPort, common.MaxPort) {
		hwlog.RunLog.Errorf("check port failed, port is out of range [%d,%d]",
			common.MinPort, common.MaxPort)
		return errors.New("check port failed")
	}
	hwlog.RunLog.Info("check port success")
	return nil
}

func (base *baseInfo) checkName() error {
	if !util.CheckInt(len(base.Username), MinNameLength, MaxNameLength) {
		hwlog.RunLog.Errorf("check username length failed, username length is out of range [%d,%d]",
			MinNameLength, MaxNameLength)
		return errors.New("check username length failed")
	}
	if !util.CheckNameFormat(base.Username) {
		hwlog.RunLog.Error("check username format failed, it contains invalid characters")
		return errors.New("check username format failed")
	}
	hwlog.RunLog.Info("check username success")
	return nil
}

func (base *baseInfo) checkPwd() error {
	if !util.CheckInt(len(base.Password), MinPwdLength, MaxPwdLength) {
		hwlog.RunLog.Errorf("check password failed, password length is out of range [%d,%d]",
			MinPwdLength, MaxPwdLength)
		return errors.New("check password failed")
	}
	if err := utils.CheckPassWordComplexity(base.Password); err != nil {
		hwlog.RunLog.Error("check password failed, password complex not meet the requirement")
		return errors.New("check password failed")
	}
	hwlog.RunLog.Info("check password success")
	return nil
}
