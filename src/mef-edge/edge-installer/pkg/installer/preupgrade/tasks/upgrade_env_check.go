// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks this file for check environment before upgrade
package tasks

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

// CheckEnvironmentBase check environment base
type CheckEnvironmentBase struct {
	extractPath    string
	installPath    string
	extractMinDisk uint64
	installMinDisk uint64
}

// NewCheckOfflineEdgeInstallerEnv check environment before upgrade edge installer
func NewCheckOfflineEdgeInstallerEnv(tarFile, extractPath,
	installPath string) *CheckOfflineEdgeInstallerEnv {
	return &CheckOfflineEdgeInstallerEnv{
		CheckEnvironmentBase: CheckEnvironmentBase{
			extractPath:    extractPath,
			installPath:    installPath,
			extractMinDisk: constants.InstallerExtractMin,
			installMinDisk: constants.InstallerUpgradeMin,
		},
		tarPath: tarFile,
	}
}

func (ceb CheckEnvironmentBase) cleanEnv() error {
	if err := fileutils.DeleteAllFileWithConfusion(ceb.extractPath); err != nil {
		hwlog.RunLog.Errorf("clean extract path[%s] failed,error:%v", ceb.extractPath, err)
		return err
	}
	upgradeDir := filepath.Join(ceb.installPath, constants.SoftwareDirTemp)
	err := util.UnSetImmutable(upgradeDir)
	if err != nil {
		hwlog.RunLog.Warnf("unset temp dir[%s] immutable find error, maybe include link file", upgradeDir)
	}

	if err := fileutils.DeleteAllFileWithConfusion(upgradeDir); err != nil {
		hwlog.RunLog.Errorf("clean upgrade temp dir[%s] failed,error:%v", upgradeDir, err)
		return err
	}

	hwlog.RunLog.Info("clean environment success")
	return nil
}

func (ceb CheckEnvironmentBase) checkDiskSpace() error {
	if err := fileutils.CreateDir(ceb.extractPath, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("make sure [%s] exist failed,error:%v", ceb.extractPath, err)
		return err
	}
	isSamePart, err := util.InSamePartition(ceb.extractPath, ceb.installPath)
	if err != nil {
		hwlog.RunLog.Errorf("check is same partition failed,error:%v", err)
		return err
	}
	if isSamePart {
		if err = envutils.CheckDiskSpace(ceb.extractPath, ceb.extractMinDisk+ceb.installMinDisk); err != nil {
			fmt.Println("disk space is not enough")
			hwlog.RunLog.Error(err)
			return err
		}
	}
	if err = envutils.CheckDiskSpace(ceb.extractPath, ceb.extractMinDisk); err != nil {
		fmt.Println("disk space is not enough")
		hwlog.RunLog.Error(err)
		return err
	}
	if err = envutils.CheckDiskSpace(ceb.installPath, ceb.installMinDisk); err != nil {
		fmt.Println("disk space is not enough")
		hwlog.RunLog.Error(err)
		return err
	}
	hwlog.RunLog.Info("check disk space success")
	return nil
}

func (ceb CheckEnvironmentBase) checkNecessaryCommands() error {
	if err := util.CheckNecessaryCommands(); err != nil {
		fmt.Println(err)
		return errors.New("check necessary commands failed")
	}
	hwlog.RunLog.Info("check necessary commands success")
	return nil
}

// CheckOfflineEdgeInstallerEnv check edge installer environment
type CheckOfflineEdgeInstallerEnv struct {
	CheckEnvironmentBase
	tarPath string
}

// Run check edge installer environment task
func (coe CheckOfflineEdgeInstallerEnv) Run() error {
	var checkFunc = []func() error{
		coe.cleanEnv,
		coe.checkDiskSpace,
		coe.checkNecessaryCommands,
		coe.unpackUgpTarPackage,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (coe CheckOfflineEdgeInstallerEnv) unpackUgpTarPackage() error {
	if _, err := fileutils.RealDirCheck(coe.extractPath, true, false); err != nil {
		hwlog.RunLog.Errorf("check extractPath failed: %v", err)
		return errors.New("check extractPath failed")
	}

	if err := fileutils.ExtraTarGzFile(coe.tarPath, coe.extractPath, true); err != nil {
		hwlog.RunLog.Errorf("extract tar package file failed: %v", err)
		return errors.New("extract tar package file failed")
	}

	hwlog.RunLog.Info("extract tar package file success")
	return nil
}
