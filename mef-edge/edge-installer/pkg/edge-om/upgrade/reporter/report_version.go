// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package reporter to send msg to cloud
package reporter

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/edge-om/common/cloudconnect"
)

const (
	connectCheckInterval = 3 * time.Second
)

// Handler report software version handler
type Handler struct{}

// Handle entry
func (ch *Handler) Handle(msg *model.Message) error {
	ReportSoftwareVersion(1)
	return nil
}

// ReportSoftwareVersion [method] report software info to cloud
func ReportSoftwareVersion(count int) {
	hwlog.RunLog.Infof("start to report software version to cloud")
	const defaultReCheckTimes = 20
	reCheckCount := 0
	for {
		if reCheckCount == defaultReCheckTimes {
			hwlog.RunLog.Infof("report software version to cloud timeout")
			return
		}
		if cloudconnect.GetCloudConnectStatus() {
			break
		}
		time.Sleep(connectCheckInterval)
		reCheckCount++
	}

	for i := 0; i < count; i++ {
		reportSoftwareVersionInfo()
		if i == count-1 {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Minute)
	}
}

type softwareInfo struct {
	Name            string
	Version         string
	InactiveVersion string
}

func getSoftwareInfo() ([]softwareInfo, error) {
	softwareName := constants.MEFEdgeName

	var softwareInfos []softwareInfo

	installerActiveVersion, err := getSoftwareActiveVersion(softwareName)
	if err != nil {
		return nil, fmt.Errorf("get package version failed, error: %v", err)
	}

	installerInActiveVersion, err := getSoftwareInActiveVersion(softwareName)
	if err != nil {
		hwlog.RunLog.Warnf("get package version failed, error: %v", err)
	}

	softwareInfos = append(softwareInfos, softwareInfo{Name: softwareName,
		Version:         installerActiveVersion,
		InactiveVersion: installerInActiveVersion})

	return softwareInfos, nil
}

func getSoftwareActiveVersion(softwareName string) (string, error) {
	if softwareName != constants.MEFEdgeName {
		return "", errors.New("invalid software name")
	}
	compWorkDir, err := path.GetCompWorkDir()
	if err != nil {
		return "", err
	}

	versionPath := filepath.Join(filepath.Dir(compWorkDir), constants.VersionXml)
	if !fileutils.IsExist(versionPath) {
		return "", nil
	}

	versionXmlManager := config.NewVersionXmlMgr(versionPath)
	packageVersion, err := versionXmlManager.GetVersion()
	if err != nil {
		return "", fmt.Errorf("get package active version failed, error: %v", err)
	}
	return packageVersion, nil
}

func getSoftwareInActiveVersion(softwareName string) (string, error) {
	if softwareName != constants.MEFEdgeName {
		return "", errors.New("invalid software name")
	}

	versionDir := constants.EdgeInstaller
	versionPath := filepath.Join(constants.UnpackPath, versionDir, constants.VersionXml)
	if !fileutils.IsExist(versionPath) {
		return "", nil
	}

	versionXmlManager := config.NewVersionXmlMgr(versionPath)
	packageVersion, err := versionXmlManager.GetVersion()
	if err != nil {
		return "", fmt.Errorf("get package inactive version failed, error: %v", err)
	}
	return packageVersion, nil
}

// EdgeReportSoftwareInfo [struct] to report edge software info
type EdgeReportSoftwareInfo struct {
	SerialNumber string         `json:"serialNumber"`
	SoftwareInfo []softwareInfo `json:"softwareInfo"`
}

// reportSoftwareVersionInfo [method] report software version to cloud
func reportSoftwareVersionInfo() {
	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create new message failed, error: %v", err)
		return
	}
	edgeOmCfg, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Errorf("get config dir failed: %v", err)
		return
	}
	dbMgr := config.NewDbMgr(edgeOmCfg, constants.DbEdgeOmPath)
	installCfg, err := config.GetInstall(dbMgr)
	if err != nil {
		hwlog.RunLog.Errorf("get install config failed,error:%v", err)
		return
	}

	softwareInfos, err := getSoftwareInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get package version failed, error: %v", err)
		return
	}

	var info = EdgeReportSoftwareInfo{
		SerialNumber: installCfg.SerialNumber,
		SoftwareInfo: softwareInfos,
	}

	sendMsg.SetRouter(constants.UpgradeManagerName, constants.InnerClient, constants.OptReport,
		constants.ResSoftwareVersion)
	if err = sendMsg.FillContent(info, true); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return
	}
	if err = modulemgr.SendMessage(sendMsg); err != nil {
		hwlog.RunLog.Errorf("%s sends message to %s failed", constants.UpgradeManagerName, constants.InnerClient)
	}
}
