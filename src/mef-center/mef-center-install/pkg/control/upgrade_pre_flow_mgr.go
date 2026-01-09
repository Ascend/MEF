// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package control

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// UpgradePreFlowMgr is a struct that uses to do upgrade, it is executed in the old version
type UpgradePreFlowMgr struct {
	tarPath string
	util.SoftwareMgr
	unpackPath string
}

// GetUpgradePreMgr is a func to init an UpgradePreFlowMgr
func GetUpgradePreMgr(tarPath string, components []string) (*UpgradePreFlowMgr, error) {
	pathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		return nil, fmt.Errorf("init upgrade pre mgr failed: %v", err)
	}
	mgr := &UpgradePreFlowMgr{
		SoftwareMgr: util.SoftwareMgr{
			Components:     components,
			InstallPathMgr: pathMgr,
		},
		tarPath: tarPath,
	}
	mgr.unpackPath = mgr.InstallPathMgr.WorkPathMgr.GetTempTarPath()

	return mgr, nil
}

// DoUpgrade is the main func that to upgrade mef-center
func (upf *UpgradePreFlowMgr) DoUpgrade() error {
	if err := upf.preCheck(); err != nil {
		return err
	}

	var upgradeTasks = []func() error{
		upf.checkUpgradePaths,
		upf.prepareUnpackDir,
		upf.unzipTarFile,
		upf.copyInstallJson,
	}

	for _, function := range upgradeTasks {
		err := function()
		if err == nil {
			continue
		}

		util.ClearPakEnv(upf.InstallPathMgr.WorkPathMgr.GetVarDirPath())
		return err
	}

	if err := upf.execNewSh(); err != nil {
		return err
	}

	return nil
}

func (upf *UpgradePreFlowMgr) preCheck() error {
	hwlog.RunLog.Info("start to exec environment check")
	var checkTasks = []func() error{
		upf.checkUser,
		util.CheckNecessaryCommands,
		upf.checkCurrentPath,
		upf.checkDiskSpace,
	}

	for _, function := range checkTasks {
		err := function()
		if err == nil {
			continue
		}
		return err
	}
	hwlog.RunLog.Info("environment check succeeds")
	return nil
}

func (upf *UpgradePreFlowMgr) checkUser() error {
	if err := envutils.CheckUserIsRoot(); err != nil {
		fmt.Println("the current user is not root, cannot upgrade")
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("check user successful")
	return nil
}

func (upf *UpgradePreFlowMgr) checkCurrentPath() error {
	if err := util.CheckCurrentPath(upf.InstallPathMgr.GetWorkPath()); err != nil {
		fmt.Println("the existing dir is not the MEF working dir")
		hwlog.RunLog.Error(err)
		return errors.New("check current path failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) checkDiskSpace() error {
	if err := envutils.CheckDiskSpace(upf.InstallPathMgr.GetRootPath(), util.UpgradeDiskSpace); err != nil {
		hwlog.RunLog.Errorf("check upgrade disk space failed: %s", err.Error())
		return errors.New("check upgrade disk space failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) checkUpgradePaths() error {
	const maxFileSize = 512
	if !fileutils.IsExist(upf.tarPath) {
		hwlog.RunLog.Errorf("tar file does not exist")
		fmt.Printf("tar file does not exist\n")
		return fmt.Errorf("tar file does not exist")
	}

	if _, err := fileutils.RealFileCheck(upf.tarPath, true, false, maxFileSize); err != nil {
		hwlog.RunLog.Errorf("check tar file failed: %v", err)
		fmt.Printf("check tar file failed\n")
		return fmt.Errorf("check tar file failed")
	}

	return nil
}

func (upf *UpgradePreFlowMgr) prepareUnpackDir() error {
	if err := fileutils.CreateDir(upf.unpackPath, fileutils.Mode700); err != nil {
		hwlog.RunLog.Errorf("create unpack tar dir failed: %s", err.Error())
		return errors.New("create unpack tar dir failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) unzipTarFile() error {
	hwlog.RunLog.Info("start to unzip tar file")
	fmt.Println("start to unzip tar file")
	if upf.tarPath == "" {
		hwlog.RunLog.Error("tarPath is nil")
		return errors.New("tarPath is nil")
	}

	if err := fileutils.ExtraTarGzFile(upf.tarPath, upf.unpackPath, true); err != nil {
		hwlog.RunLog.Errorf("unzip tar file failed: %s", err.Error())
		return errors.New("unzip tar file failed")
	}
	hwlog.RunLog.Info("unzip tar file succeeds")
	fmt.Println("unzip tar file succeeds")
	return nil
}

func (upf *UpgradePreFlowMgr) copyInstallJson() error {
	tgtDir := filepath.Join(upf.unpackPath, util.InstallDirName)
	tgtAbsDir, err := filepath.EvalSymlinks(tgtDir)
	if err != nil {
		hwlog.RunLog.Errorf("get [%s]'s abs path failed: %s", tgtDir, err.Error())
		return errors.New("get install_param.json's abs path failed")
	}
	tgtAbsPath := path.Join(tgtAbsDir, util.InstallParamJson)

	srcPath := upf.InstallPathMgr.WorkPathMgr.GetInstallParamJsonPath()
	srcAbsPath, err := filepath.EvalSymlinks(srcPath)
	if err != nil {
		hwlog.RunLog.Errorf("get [%s]'s abs path failed: %s", srcPath, err.Error())
		return errors.New("get install_param.json's abs path failed")
	}

	if err = fileutils.CopyFile(srcAbsPath, tgtAbsPath); err != nil {
		hwlog.RunLog.Errorf("copy install_param.json failed: %s", err.Error())
		return errors.New("copy install_param.json failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) execNewSh() error {
	upgradeShPath := filepath.Join(upf.unpackPath, util.InstallDirName, util.ScriptsDirName, util.UpgradeShName)
	if err := envutils.RunCommandWithOsStdout(upgradeShPath, util.UpgradeTimeoutSec); err != nil {
		upf.newShErrDeal(err)
		hwlog.RunLog.Errorf("upgrade failed, exec new version upgrade sh meet error: %v", err)
		return errors.New("exec new version upgrade sh meet error")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) newShErrDeal(returnErr error) {
	if strings.Contains(returnErr.Error(), "invalid arch") {
		fmt.Println("the upgrading zip is for another CPU architecture")
		hwlog.RunLog.Error("upgrade failed: the upgrading zip is for another CPU architecture")
		util.ClearPakEnv(upf.InstallPathMgr.WorkPathMgr.GetVarDirPath())
		return
	}

	tempVarDir := upf.InstallPathMgr.WorkPathMgr.GetVarDirPath()
	if fileutils.IsExist(tempVarDir) {
		if err := fileutils.DeleteAllFileWithConfusion(upf.InstallPathMgr.WorkPathMgr.GetVarDirPath()); err != nil {
			hwlog.RunLog.Warnf("delete temp dir %s failed, need to clear it manually", err.Error())
			return
		}
		hwlog.RunLog.Info("clear environment success")
	}
}
