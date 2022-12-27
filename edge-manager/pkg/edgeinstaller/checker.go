// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller the checker used in edge-installer module
package edgeinstaller

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"edge-manager/pkg/util"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// CheckDataFromSfwMgr checks data from software manager
func CheckDataFromSfwMgr(downloadSfwContent *util.DealSfwContent, nodeId string) error {
	if downloadSfwContent == nil {
		hwlog.RunLog.Error("download software content is nil")
		return errors.New("download software content is nil")
	}
	dataBytes := strings.Split(downloadSfwContent.Url, "=")
	if len(dataBytes) == 0 {
		return errors.New("invalid download software content")
	}
	softwareName := strings.Split(dataBytes[1], "&")[LocationSfwName]

	if !CheckSfwInfo(softwareName) {
		return errors.New("check software info failed")
	}

	if !CheckNodeId(*downloadSfwContent, nodeId) {
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
func CheckNodeId(downloadSfwContent util.DealSfwContent, nodeId string) bool {
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
