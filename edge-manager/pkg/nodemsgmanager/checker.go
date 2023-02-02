// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package nodemsgmanager the checker used in edge-installer module
package nodemsgmanager

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"edge-manager/pkg/util"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

// CheckRespDataFromSfwMgr checks data from software manager
func CheckRespDataFromSfwMgr(respDataFromSfwMgr *RespDataFromSfwMgr, nodeId string) error {
	if respDataFromSfwMgr == nil {
		hwlog.RunLog.Error("download software content is nil")
		return errors.New("download software content is nil")
	}

	dataBytes := strings.Split(respDataFromSfwMgr.DownloadUrl, "=")
	if len(dataBytes) == 0 {
		return errors.New("invalid download software content")
	}
	softwareName := strings.Split(dataBytes[1], "&")[LocationSfwName]

	if !CheckSfwInfo(softwareName) {
		return errors.New("check software info failed")
	}

	if !CheckNodeId(*respDataFromSfwMgr, nodeId) {
		return errors.New("check nodeId failed")
	}

	return nil
}

// CheckSfwInfo checks whether software info from software manager is correct
func CheckSfwInfo(softwareName string) bool {
	if softwareName != common.EdgeCore && softwareName != common.EdgeInstaller && softwareName != common.DevicePlugin {
		hwlog.RunLog.Error("check software name failed")
		return false
	}

	return true
}

// CheckNodeId checks whether nodeId is correct
func CheckNodeId(downloadSfwContent RespDataFromSfwMgr, nodeId string) bool {
	return downloadSfwContent.NodeId == nodeId
}

// CheckUpdateTableSfwInfo checks whether ip and port is correct in SoftwareMgrInfo
func CheckUpdateTableSfwInfo(updateTableSfwInfo *SoftwareMgrInfo) common.RespMsg {
	if err := CheckAddr(updateTableSfwInfo.Address); err != nil {
		return common.RespMsg{Status: "", Msg: "check address of updating table software info error", Data: nil}
	}

	if err := CheckPort(updateTableSfwInfo.Port); err != nil {
		return common.RespMsg{Status: "", Msg: "check port of updating table software info error", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// CheckAddr checks whether address is correct
func CheckAddr(address string) error {
	if address == "" || address == common.ZeroAddr || address == common.BroadCastAddr {
		hwlog.RunLog.Error("check address failed, address is not allowed")
		return errors.New("check address failed")
	}
	addr := net.ParseIP(address)
	if addr == nil || addr.To4() == nil {
		hwlog.RunLog.Error("check address failed, address invalid")
		return errors.New("check address failed")
	}

	hwlog.RunLog.Info("check address success")
	return nil
}

// CheckPort checks whether port is correct
func CheckPort(port string) error {
	portInt, err := strconv.Atoi(port)
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
