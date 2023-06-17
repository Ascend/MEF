// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller the checker used in edge-installer module
package edgeinstaller

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
)

func getValidSfwName() []string {
	return []string{common.EdgeCore, common.EdgeInstaller, common.DevicePlugin}
}

type sfwParaPattern struct {
	patterns map[string]string
}

var sfwPattern = sfwParaPattern{patterns: map[string]string{
	"sfwVersion":          `^[\w]+`,
	"downloadUrlFromUser": "^[a-z-%?]{0,2048}$",
	"username":            "^[a-zA-Z0-9_-]{1,16}$",
	"password":            `^[\S]{16,32}$`,
},
}

func (s *sfwParaPattern) getPatternFromMap(mapKey string) (string, bool) {
	pattern, ok := s.patterns[mapKey]
	return pattern, ok
}

func (ur *UpgradeSfwReq) checkUpgradeSfwReq() error {
	var checkUpgradeSfwReqInfo = []func() error{
		ur.checkNodeNums,
		ur.checkSfwName,
		ur.checkSfwVersion,
		ur.checkUrlFromUser,
		ur.checkUsername,
		ur.checkPassword,
	}

	for _, function := range checkUpgradeSfwReqInfo {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (ur *UpgradeSfwReq) checkNodeNums() error {
	_, err := getNodeNum(ur.NodeNums)
	if err != nil {
		hwlog.RunLog.Errorf("check nodeNums failed, get node unique name when upgrading failed, error: %v", err)
		return fmt.Errorf("check nodeNums failed, get node unique name when upgrading failed, error: %v", err)
	}

	hwlog.RunLog.Info("check nodeNums success")
	return nil
}

func (ur *UpgradeSfwReq) checkSfwName() error {
	validSfwNames := getValidSfwName()
	for _, sfwName := range validSfwNames {
		if ur.SoftwareName == sfwName {
			hwlog.RunLog.Info("check software name success")
			return nil
		}
	}

	hwlog.RunLog.Error("check software name failed, software name is invalid")
	return errors.New("check software name failed, software name is invalid")
}

func (ur *UpgradeSfwReq) checkSfwVersion() error {
	sfwVersionPattern, ok := sfwPattern.getPatternFromMap("sfwVersion")
	if !ok {
		hwlog.RunLog.Error("check software version failed, regex is not exist")
		return errors.New("check software version failed, regex is not exist")
	}

	if !util.RegexStringChecker(ur.Password, sfwVersionPattern) {
		hwlog.RunLog.Error("check software version failed, doesn't match regex")
		return errors.New("check software version failed, doesn't match regex")
	}

	hwlog.RunLog.Info("check software version success")
	return nil
}

func (ur *UpgradeSfwReq) checkUrlFromUser() error {
	downloadUrlFromUserPattern, ok := sfwPattern.getPatternFromMap("downloadUrlFromUser")
	if !ok {
		hwlog.RunLog.Error("check download url from user failed, regex is not exist")
		return errors.New("check download url from user failed, regex is not exist")
	}

	if !util.RegexStringChecker(ur.Password, downloadUrlFromUserPattern) {
		hwlog.RunLog.Error("check download url from user failed, doesn't match regex")
		return errors.New("check download url from user failed, doesn't match regex")
	}

	hwlog.RunLog.Info("check download url from user success")
	return nil
}

func (ur *UpgradeSfwReq) checkUsername() error {
	usernamePattern, ok := sfwPattern.getPatternFromMap("username")
	if !ok {
		hwlog.RunLog.Error("check username failed, username regex is not exist")
		return errors.New("check username failed, username regex is not exist")
	}

	if !util.RegexStringChecker(ur.Password, usernamePattern) {
		hwlog.RunLog.Error("check username failed, username doesn't match regex")
		return errors.New("check username failed, username doesn't match regex")
	}

	hwlog.RunLog.Info("check username success")
	return nil
}

func (ur *UpgradeSfwReq) checkPassword() error {
	passwordPattern, ok := sfwPattern.getPatternFromMap("password")
	if !ok {
		hwlog.RunLog.Error("check password failed, password regex is not exist")
		return errors.New("check password failed, password regex is not exist")
	}

	if !util.RegexStringChecker(ur.Password, passwordPattern) {
		hwlog.RunLog.Error("check password failed, password doesn't match regex")
		return errors.New("check password failed, password doesn't match regex")
	}

	hwlog.RunLog.Info("check password success")
	return nil
}
