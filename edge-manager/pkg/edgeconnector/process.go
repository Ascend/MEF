// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket server basic process
package edgeconnector

import (
	"encoding/json"
	"errors"
	"strings"

	"edge-manager/pkg/database"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/util"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// IssueInfo struct for issuing service cert
type IssueInfo struct {
	NodeId      string
	ServiceCert []byte
}

// UpgradeInfo struct for upgrading software
type UpgradeInfo struct {
	NodeId          []string
	SoftwareName    string
	SoftwareVersion string
	baseInfo
}

// UpdateInfo struct for updating username and password
type UpdateInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateInfoToInstaller struct for updating username and password to installer
type UpdateInfoToInstaller struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

// IssueResp deals issue service cert response
type IssueResp struct {
	NodeId []string
	Result string
	Reason string
}

// RespFromInstaller response from edge-installer
type RespFromInstaller struct {
	NodeId string
	Result string
	Reason string
}

func getDestination(message *model.Message) string {
	var destination = ""
	switch message.GetOption() {
	case common.Issue:
		destination = common.CertManagerName
	case common.Upgrade:
		destination = common.RestfulServiceName
	case common.Download:
		destination = common.EdgeInstallerName
	case common.Get:
		destination = common.EdgeInstallerName
	default:
		hwlog.RunLog.Error("invalid option")
		return ""
	}
	return destination
}

func isDownloadResp(content string) *RespFromInstaller {
	var data RespFromInstaller
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		hwlog.RunLog.Error("parse to RespFromInstaller from edge-installer failed")
		return nil
	}

	return &data
}

func isDownloadReq(content string) *util.DownloadSfwReq {
	var data util.DownloadSfwReq
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		hwlog.RunLog.Error("parse to DownloadSfwReq from edge-installer failed")
		return nil
	}

	return &data
}

func checkUpgradeInfo(upgradeInfo util.DealSfwContent) error {
	sfwMgrBaseInfo := getSfwMgrInfo(upgradeInfo)
	defer common.ClearSliceByteMemory(sfwMgrBaseInfo.Password)
	if sfwMgrBaseInfo == nil {
		hwlog.RunLog.Error("get software base info failed")
		return errors.New("get software base info failed")
	}

	if err := sfwMgrBaseInfo.checkBaseInfo(); err != nil {
		return err
	}

	return nil
}

func getSfwMgrInfo(upgradeInfo util.DealSfwContent) *baseInfo {
	realUrl := strings.Split(upgradeInfo.Url, " ")[LocationUrl]
	dataBytes := strings.Split(realUrl, "/")
	if len(dataBytes) == 0 {
		hwlog.RunLog.Error("split upgradeInfo url failed")
		return nil
	}

	sfwMgrIP := strings.Split(dataBytes[LocationIpPort], ":")[LocationIP]
	sfwMgrPort := strings.Split(dataBytes[LocationIpPort], ":")[LocationPort]

	password := []byte(upgradeInfo.Password)
	defer common.ClearSliceByteMemory(password)

	sfwMgrBaseInfo := &baseInfo{
		Address:  sfwMgrIP,
		Port:     sfwMgrPort,
		Username: upgradeInfo.Username,
		Password: password,
	}
	return sfwMgrBaseInfo
}

func getUniqueNums() ([]string, error) {
	var uniqueNums []string
	if err := database.GetDb().Model(nodemanager.NodeInfo{}).Select("unique_name").Find(&uniqueNums).Error; err != nil {
		hwlog.RunLog.Errorf("get unique nums failed, error: %v", err)
		return []string{}, err
	}

	return uniqueNums, nil
}

func initGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	return gin.New()
}
