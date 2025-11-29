// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr this file for saving image repository ca when connecting with FD
package handlermgr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

// CertInfo image repository cert info
type CertInfo struct {
	Ip        string `json:"ip"`
	Port      string `json:"port"`
	Domain    string `json:"domain"`
	CaContent string `json:"ca_content"`
}

// saveCertHandler edge_om static info handlers
type saveCertHandler struct {
	fdIp string
}

// Handle Handler handle entry
func (ch *saveCertHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start save image repository cert info")
	var opResult bool
	// defer must be placed in an anonymous function here. Otherwise, opResult is always false.
	defer func() {
		ch.PrintOpLog(opResult)
	}()

	var certInfo CertInfo
	if err := msg.ParseContent(&certInfo); err != nil {
		hwlog.RunLog.Errorf("parse image repository cert info para failed: %v", err)
		return errors.New("parse image repository cert info para failed")
	}

	if checkResult := newCertParaChecker().Check(certInfo); !checkResult.Result {
		hwlog.RunLog.Errorf("check image repository cert info failed: %v", checkResult.Reason)
		return errors.New("check image repository cert info failed")
	}
	ch.fdIp = certInfo.Ip

	hwlog.OpLog.Infof("[%s@%s] save [FD:%s] image repository cert start",
		constants.ModDeviceOm, constants.LocalIp, ch.fdIp)

	if err := saveCert(&certInfo); err != nil {
		hwlog.RunLog.Errorf("save image repository cert info failed: %v", err)
		return errors.New("save image repository cert info failed")
	}

	opResult = true
	hwlog.RunLog.Info("save image repository cert info success")
	return nil
}

func (ch *saveCertHandler) PrintOpLog(opResult bool) {
	if ch.fdIp == "" {
		hwlog.RunLog.Warn("get fd ip error")
	}

	if opResult {
		hwlog.OpLog.Infof("[%s@%s] save [FD:%s] image repository cert success",
			constants.ModDeviceOm, constants.LocalIp, ch.fdIp)
		return
	}
	hwlog.OpLog.Errorf("[%s@%s] save [FD:%s] image repository cert failed",
		constants.ModDeviceOm, constants.LocalIp, ch.fdIp)
}

func saveCert(certInfo *CertInfo) error {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		return fmt.Errorf("get config path manager, error: %v", err)
	}
	certPath := configPathMgr.GetImageCertPath()
	if err = fileutils.WriteData(certPath, []byte(certInfo.CaContent)); err != nil {
		return fmt.Errorf("write cert to file failed, error: %v", err)
	}
	if err = fileutils.SetPathPermission(certPath, constants.Mode400, false, false); err != nil {
		return fmt.Errorf("set cert mode failed, error: %v", err)
	}

	if err = backuputils.BackUpFiles(certPath); err != nil {
		hwlog.RunLog.Warnf("create backup for image cert failed, %v", err)
	}

	var dockerCertDirs = []string{filepath.Join(constants.DockerCertDir, certInfo.Domain),
		filepath.Join(constants.DockerCertDir, certInfo.Domain+":"+certInfo.Port)}
	for _, certDir := range dockerCertDirs {
		if err = fileutils.CreateDir(certDir, constants.Mode700); err != nil {
			return fmt.Errorf("create docker cert %s failed, error: %v", certDir, err)
		}
		if err = copyCertToDocker(certPath, filepath.Join(certDir, constants.ImageCertFileName)); err != nil {
			return err
		}
	}
	return nil
}

func copyCertToDocker(certPath, dockerCertPath string) error {
	if err := fileutils.DeleteFile(dockerCertPath); err != nil {
		hwlog.RunLog.Errorf("delete docker cert path [%s] failed: %v", dockerCertPath, err)
		return err
	}
	if err := fileutils.CopyFile(certPath, dockerCertPath); err != nil {
		return fmt.Errorf("copy cert [%s] to docker cert dir [%s] to failed, error: %v",
			certPath, dockerCertPath, err)
	}

	// docker cert path may contain character ':'
	linkChecker := fileutils.NewFileLinkChecker(false)
	linkChecker.SetNext(fileutils.NewFileOwnerChecker(true, false, constants.RootUserUid, constants.RootUserGid))
	linkChecker.SetNext(fileutils.NewFileModeChecker(true, constants.ModeUmask022, true, true))

	file, err := os.Open(dockerCertPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			hwlog.RunLog.Errorf("failed to docker cert file, %v", err)
		}
	}()
	if err := linkChecker.Check(file, dockerCertPath); err != nil {
		return fmt.Errorf("check docker cert file [%s] to failed, error: %v", dockerCertPath, err)
	}
	if err := fileutils.SetPathPermission(dockerCertPath, constants.Mode400, false, true); err != nil {
		return fmt.Errorf("set permission for  docker cert file [%s] to failed, error: %v", dockerCertPath, err)
	}
	return nil
}
