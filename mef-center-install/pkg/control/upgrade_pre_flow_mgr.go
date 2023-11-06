// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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

	"huawei.com/mindx/mef/common/cmsverify"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// UpgradePreFlowMgr is a struct that uses to do upgrade, it is executed in the old version
type UpgradePreFlowMgr struct {
	tarPath string
	cmsPath string
	crlPath string
	util.SoftwareMgr
	unpackPath string
}

// GetUpgradePreMgr is a func to init an UpgradePreFlowMgr
func GetUpgradePreMgr(tarPath, cmsPath, crlPath string, components []string) (*UpgradePreFlowMgr, error) {
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
		cmsPath: cmsPath,
		crlPath: crlPath,
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
		upf.verifyPackage,
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
	pathMap := map[string]string{
		"tar file": upf.tarPath,
		"cms file": upf.cmsPath,
		"crl file": upf.crlPath,
	}
	for fileTag, filePath := range pathMap {
		if !fileutils.IsExist(filePath) {
			hwlog.RunLog.Errorf("%s does not exist", fileTag)
			fmt.Printf("%s does not exist\n", fileTag)
			return fmt.Errorf("%s does not exist", fileTag)
		}

		if _, err := fileutils.RealFileCheck(filePath, true, false, maxFileSize); err != nil {
			hwlog.RunLog.Errorf("check %s failed: %v", fileTag, err)
			fmt.Printf("check %s failed\n", fileTag)
			return fmt.Errorf("check %s failed", fileTag)
		}
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

func (upf *UpgradePreFlowMgr) verifyPackage() error {
	fmt.Println("start to verify package")
	updateCrlFlag, verifyCrl, err := prepareVerifyCrl(upf.crlPath)
	if err != nil {
		hwlog.RunLog.Errorf("prepare crl for verifying package failed, error: %v", err)
		return err
	}

	if err = cmsverify.VerifyPackage(verifyCrl, upf.cmsPath, upf.tarPath); err != nil {
		fmt.Println("verify package failed, the zip file might be tampered")
		hwlog.RunLog.Errorf("verify package failed,error:%v", err)
		return errors.New("verify package failed")
	}

	if updateCrlFlag {
		if err = updateLocalCrlFile(verifyCrl); err != nil {
			hwlog.RunLog.Errorf("update crl file failed, error: %v", err)
			return errors.New("update crl file failed")
		}
	}
	fmt.Println("update crl file success.")
	hwlog.RunLog.Info("update crl file success")

	hwlog.RunLog.Info("verify package succeeds")
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

func prepareVerifyCrl(crlFile string) (bool, string, error) {
	updateCrlFlag := true
	verifyCrl := crlFile
	var err error

	// when two input parameters are the same, the function can be used to check whether the CRL file is valid
	crlToUpdateValid, err := cmsverify.CompareCrls(verifyCrl, verifyCrl)
	if err != nil || int(crlToUpdateValid) != util.CompareSame {
		fmt.Println("crl file is invalid")
		hwlog.RunLog.Error("crl file is invalid")
		return true, "", errors.New("crl file is invalid")
	}

	crlOnDevicePath, err := getValidCrlOnDevice()
	if err != nil {
		hwlog.RunLog.Errorf("get valid crl on device failed, error: %v", err)
		return true, verifyCrl, errors.New("get valid crl on device failed")
	}

	if crlOnDevicePath == "" {
		return true, verifyCrl, nil
	}

	updateCrlFlag, verifyCrl, err = getUpdateCrlFlag(crlFile, crlOnDevicePath)
	if err != nil {
		hwlog.RunLog.Errorf("get update crl flag failed, error: %v", err)
		return true, "", err
	}

	return updateCrlFlag, verifyCrl, nil
}

func getValidCrlOnDevice() (string, error) {
	crlOnDevicePath := filepath.Join(util.CrlOnDeviceDir, util.CrlOnDeviceName)
	if fileutils.IsExist(crlOnDevicePath) {

		crlOnDeviceValid, err := cmsverify.CompareCrls(crlOnDevicePath, crlOnDevicePath)
		if err != nil || int(crlOnDeviceValid) != util.CompareSame {
			fmt.Println("Warning: crl file on device is invalid.")
			hwlog.RunLog.Warn("crl file on device is invalid")
			return "", nil
		}

		return crlOnDevicePath, nil
	}

	if err := fileutils.CreateDir(util.CrlOnDeviceDir, common.Mode755); err != nil {
		hwlog.RunLog.Errorf("create crl dir [%s] failed, error: %v", util.CrlOnDeviceDir, err)
		return crlOnDevicePath, fmt.Errorf("create crl dir [%s] failed", util.CrlOnDeviceDir)
	}

	if _, err := fileutils.RealDirCheck(util.CrlOnDeviceDir, true, false); err != nil {
		hwlog.RunLog.Errorf("check dir [%s] failed, error: %v", util.CrlOnDeviceDir, err)
		return crlOnDevicePath, fmt.Errorf("check dir [%s] failed", util.CrlOnDeviceDir)
	}

	return "", nil
}

func getUpdateCrlFlag(crlToUpdate, crlOnDevice string) (bool, string, error) {
	if crlToUpdate == "" || crlOnDevice == "" {
		hwlog.RunLog.Error("crl is invalid")
		return false, "", errors.New("crl is invalid")
	}

	var compareRes cmsverify.CrlCompareStatus
	var err error
	updateCrlFlag := true
	verifyCrl := crlToUpdate
	if compareRes, err = cmsverify.CompareCrls(crlToUpdate, crlOnDevice); err != nil {
		hwlog.RunLog.Errorf("compare crls failed, error: %v", err)
		return false, "", errors.New("compare crls failed failed")
	}

	switch int(compareRes) {
	case util.CompareSame:
		updateCrlFlag = false
		verifyCrl = crlOnDevice
		hwlog.RunLog.Info("the software package crl file is the same as the local crl file, " +
			"use the local crl file to verify and no update local crl file required")
	case util.CompareNew:
		hwlog.RunLog.Info("the software package crl file is newer than the local crl file, " +
			"use software package crl file to verify and update local crl file")
	case util.CompareOld:
		updateCrlFlag = false
		verifyCrl = crlOnDevice
		hwlog.RunLog.Info("the software package crl file is older than the local crl file, " +
			"use the local crl file to verify and no update local crl file required")
	default:
		hwlog.RunLog.Error("compare local crl file and the software package crl file failed, " +
			"use software package crl file to verify and update local crl file")
	}

	return updateCrlFlag, verifyCrl, nil
}

func updateLocalCrlFile(verifyCrl string) error {
	if err := fileutils.CreateDir(util.CrlOnDeviceDir, common.Mode755); err != nil {
		hwlog.RunLog.Errorf("create crl dir [%s] failed, error: %v", util.CrlOnDeviceDir, err)
		return fmt.Errorf("create crl dir [%s] failed", util.CrlOnDeviceDir)
	}
	crlOnDevicePath := filepath.Join(util.CrlOnDeviceDir, util.CrlOnDeviceName)
	if err := fileutils.CopyFile(verifyCrl, crlOnDevicePath); err != nil {
		hwlog.RunLog.Errorf("copy crl file [%s] failed, error: %v", verifyCrl, err)
		return errors.New("copy crl file failed")
	}

	const maxCrlSizeInMb = 10
	if _, err := fileutils.RealFileCheck(
		crlOnDevicePath, true, false, maxCrlSizeInMb); err != nil {
		hwlog.RunLog.Errorf("check file [%s] failed, error: %v", crlOnDevicePath, err)
		return fmt.Errorf("check file [%s] failed", crlOnDevicePath)
	}
	if err := fileutils.SetPathPermission(crlOnDevicePath, common.Mode640, false, false); err != nil {
		hwlog.RunLog.Errorf("set new crl permission failed, error: %v", err)
		return errors.New("set new crl permission failed")
	}

	return nil
}
