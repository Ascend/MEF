// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package control

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/mef/common/cmsverify"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// UpgradePreFlowMgr is a struct that uses to do upgrade, it is executed in the old version
type UpgradePreFlowMgr struct {
	zipPath string
	util.SoftwareMgr
	unpackZipPath string
	unpackTarPath string
}

// GetUpgradePreMgr is a func to init an UpgradePreFlowMgr
func GetUpgradePreMgr(zipPath string, components []string, installPath string) *UpgradePreFlowMgr {
	mgr := &UpgradePreFlowMgr{
		SoftwareMgr: util.SoftwareMgr{
			Components:     components,
			InstallPathMgr: util.InitInstallDirPathMgr(installPath),
		},
		zipPath: zipPath,
	}
	mgr.unpackZipPath = mgr.InstallPathMgr.WorkPathMgr.GetTempZipPath()
	mgr.unpackTarPath = mgr.InstallPathMgr.WorkPathMgr.GetTempTarPath()
	return mgr
}

// DoUpgrade is the main func that to upgrade mef-center
func (upf *UpgradePreFlowMgr) DoUpgrade() error {
	if err := upf.preCheck(); err != nil {
		return err
	}

	var upgradeTasks = []func() error{
		upf.prepareUnpackDir,
		upf.unzipZipFile,
		upf.verifyPackage,
		upf.unzipTarFile,
		upf.copyInstallJson,
	}

	for _, function := range upgradeTasks {
		err := function()
		if err == nil {
			continue
		}

		upf.clearEnv()
		return err
	}

	if err := upf.execNewSh(); err != nil {
		return err
	}

	return nil
}

func (upf *UpgradePreFlowMgr) preCheck() error {
	hwlog.RunLog.Info("start to exec environment check")
	fmt.Println("start to exec environment check")
	var checkTasks = []func() error{
		upf.checkUser,
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
	fmt.Println("environment check succeeds")
	return nil
}

func (upf *UpgradePreFlowMgr) checkUser() error {
	if err := util.CheckUser(); err != nil {
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("check user successful")
	return nil
}

func (upf *UpgradePreFlowMgr) checkCurrentPath() error {
	if err := util.CheckCurrentPath(upf.InstallPathMgr.GetWorkPath()); err != nil {
		hwlog.RunLog.Error(err)
		return errors.New("check current path failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) checkDiskSpace() error {
	if err := util.CheckDiskSpace(upf.InstallPathMgr.GetRootPath(), util.UpgradeDiskSpace); err != nil {
		hwlog.RunLog.Errorf("check upgrade disk space failed: %s", err.Error())
		return errors.New("check upgrade disk space failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) prepareUnpackDir() error {
	if err := common.MakeSurePath(upf.unpackZipPath); err != nil {
		hwlog.RunLog.Errorf("create unpack zip dir failed: %s", err.Error())
		return errors.New("create unpack zip dir failed")
	}

	if err := common.MakeSurePath(upf.unpackTarPath); err != nil {
		hwlog.RunLog.Errorf("create unpack tar dir failed: %s", err.Error())
		return errors.New("create unpack tar dir failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) unzipZipFile() error {
	hwlog.RunLog.Info("start to unzip zip file")
	fmt.Println("start to unzip zip file")
	if err := common.ExtraUpgradeZipFile(upf.zipPath, upf.unpackZipPath); err != nil {
		hwlog.RunLog.Errorf("unzip zip file failed: %s", err.Error())
		return errors.New("unzip zip file failed")
	}
	hwlog.RunLog.Info("unzip zip file succeeds")
	fmt.Println("unzip zip file succeeds")
	return nil
}

func (upf *UpgradePreFlowMgr) verifyPackage() error {
	hwlog.RunLog.Info("start to verify package")
	fmt.Println("start to verify package")
	unpackAbsPath, err := filepath.EvalSymlinks(upf.unpackZipPath)
	if err != nil {
		hwlog.RunLog.Errorf("get unpack abs path failed: %s", unpackAbsPath)
		return errors.New("get unpack abs path failed")
	}
	crlPath := path.Join(unpackAbsPath, upf.getCrlFileName())
	cmsPath := path.Join(unpackAbsPath, upf.getCmsFileName())
	tarPath := path.Join(unpackAbsPath, upf.getTarFileName())
	if err = cmsverify.VerifyPackage(crlPath, cmsPath, tarPath); err != nil {
		hwlog.RunLog.Errorf("verify package failed,error:%v", err)
		return errors.New("verify package failed")
	}
	fmt.Println("verify package succeeds")
	hwlog.RunLog.Info("verify package succeeds")
	return nil
}

func (upf *UpgradePreFlowMgr) unzipTarFile() error {
	hwlog.RunLog.Info("start to unzip tar file")
	fmt.Println("start to unzip tar file")
	tarName := path.Join(upf.unpackZipPath, upf.getTarFileName())

	if err := common.ExtraTarGzFile(tarName, upf.unpackTarPath, true); err != nil {
		hwlog.RunLog.Errorf("unzip tar file failed: %s", err.Error())
		return errors.New("unzip tar file failed")
	}
	hwlog.RunLog.Info("unzip tar file succeeds")
	fmt.Println("unzip tar file succeeds")
	return nil
}

func (upf *UpgradePreFlowMgr) copyInstallJson() error {
	tgtDir := filepath.Join(upf.unpackTarPath, util.InstallDirName)
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

	if err = utils.CopyFile(srcAbsPath, tgtAbsPath); err != nil {
		hwlog.RunLog.Errorf("copy install_param.json failed: %s", err.Error())
		return errors.New("copy install_param.json failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) getPureFileName() string {
	_, zipName := filepath.Split(upf.zipPath)
	return strings.TrimRight(zipName, common.ZipSuffix)
}

func (upf *UpgradePreFlowMgr) getTarFileName() string {
	return upf.getPureFileName() + common.TarGzSuffix
}

func (upf *UpgradePreFlowMgr) getCmsFileName() string {
	return upf.getPureFileName() + common.CmsSuffix
}

func (upf *UpgradePreFlowMgr) getCrlFileName() string {
	return upf.getPureFileName() + common.CrlSuffix
}

func (upf *UpgradePreFlowMgr) execNewSh() error {
	shPath := filepath.Join(upf.unpackTarPath, util.InstallDirName, util.ScriptsDirName, util.UpgradeShName)
	_, err := common.RunCommand(shPath, true, util.UpgradeTimeoutSec)
	if err != nil {
		hwlog.RunLog.Error("upgrade failed: exec new version upgrade sh meet error")
		return errors.New("upgrade failed")
	}
	return nil
}

func (upf *UpgradePreFlowMgr) clearEnv() {
	fmt.Println("install failed, start to clear environment")
	hwlog.RunLog.Info("-----Start to clear environment-----")
	if err := common.DeleteAllFile(upf.InstallPathMgr.WorkPathMgr.GetRelativeVarDirPath()); err != nil {
		fmt.Println("clear environment failed, please clear manually")
		hwlog.RunLog.Warnf("clear environment meets err:%s, need to do it manually", err.Error())
		hwlog.RunLog.Info("-----End to clear environment-----")
		return
	}
	fmt.Println("clear environment success")
	hwlog.RunLog.Info("-----End to clear environment-----")
}
