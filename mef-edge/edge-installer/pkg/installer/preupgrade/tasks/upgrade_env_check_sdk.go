// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package tasks

import (
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

// CheckOnlineEdgeInstallerEnv check edge installer environment
type CheckOnlineEdgeInstallerEnv struct {
	CheckOfflineEdgeInstallerEnv
	downloadPath string
}

// NewPrepareOnlineInstallEnv check environment before upgrade edge installer
func NewPrepareOnlineInstallEnv(downloadPath, extractPath, installPath string) *CheckOnlineEdgeInstallerEnv {
	tarFile, crlFile, cmsFile := getSoftwareFiles(downloadPath)
	installer := NewCheckOfflineEdgeInstallerEnv(tarFile, cmsFile, crlFile, extractPath, installPath)
	installer.extractMinDisk = constants.InstallerExtractOnlineMin
	installer.installMinDisk = constants.InstallerUpgradeSdkMin
	return &CheckOnlineEdgeInstallerEnv{
		CheckOfflineEdgeInstallerEnv: *installer,
		downloadPath:                 downloadPath,
	}
}

// Run check edge installer environment task
func (coe CheckOnlineEdgeInstallerEnv) Run() error {
	var checkFunc = []func() error{
		coe.checkDiskSpace,
		coe.changeFileOwner,
		coe.checkPackageValid,
		coe.unpackUgpTarPackage,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (coe CheckOnlineEdgeInstallerEnv) changeFileOwner() error {
	param := fileutils.SetOwnerParam{
		Path:       coe.downloadPath,
		Uid:        fileutils.RootUid,
		Gid:        fileutils.RootGid,
		Recursive:  true,
		IgnoreFile: false}

	return fileutils.SetPathOwnerGroup(param)
}

func getSoftwareFiles(extractPath string) (string, string, string) {
	var tarFile, crlFile, signFile string
	reader, entries, err := fileutils.ReadDir(extractPath)
	if err != nil {
		hwlog.RunLog.Errorf("read directory [%s] failed, error: %v", extractPath, err)
		return "", "", ""
	}
	defer fileutils.CloseFile(reader)

	for _, entry := range entries {
		fullPath := filepath.Join(extractPath, entry.Name())
		switch {
		case strings.HasSuffix(fullPath, constants.TarGzExt):
			tarFile = fullPath
		case strings.HasSuffix(fullPath, constants.CrlExt):
			crlFile = fullPath
		case strings.HasSuffix(fullPath, constants.SignExt) || strings.HasSuffix(fullPath, constants.CmsExt):
			signFile = fullPath
		default:
		}
	}

	return tarFile, crlFile, signFile
}
