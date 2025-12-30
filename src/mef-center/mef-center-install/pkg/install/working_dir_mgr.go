// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package install

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// WorkingDirCtl is a struct that used to manager the working dir prepare actions in installation and upgrading
type WorkingDirCtl struct {
	pathMgr     util.WorkPathItf
	mefLinkPath string
	components  []string
}

// GetWorkingDirMgr is the func to init a WorkingDirCtl struct
func GetWorkingDirMgr(pathMgr util.WorkPathItf, mefLinkPath string, components []string) *WorkingDirCtl {
	return &WorkingDirCtl{
		pathMgr:     pathMgr,
		mefLinkPath: mefLinkPath,
		components:  components,
	}
}

// DoUpgradePrepare is the main flow-control func to prepare a working dir in upgrading flow
func (wdc *WorkingDirCtl) DoUpgradePrepare() error {
	var prepareWorkingDirTasks = []func() error{
		wdc.prepareRootWorkDir,
		wdc.prepareLibDir,
		wdc.prepareRunSh,
		wdc.prepareBinDir,
		wdc.prepareVersionXml,
		wdc.prepareComponentWorkDir,
		wdc.prepareInstallParamJson,
	}

	fmt.Println("start to prepare working dir")
	hwlog.RunLog.Info("-----Start to prepare working dir-----")
	for _, function := range prepareWorkingDirTasks {
		if err := function(); err != nil {
			return err
		}
	}
	fmt.Println("prepare working dir success")
	hwlog.RunLog.Info("-----Prepare working dir successful-----")
	return nil

}

// DoInstallPrepare is the main flow-control func to prepare a working dir in installation flow
func (wdc *WorkingDirCtl) DoInstallPrepare() error {
	var prepareWorkingDirTasks = []func() error{
		wdc.prepareRootWorkDir,
		wdc.prepareRunSh,
		wdc.prepareBinDir,
		wdc.prepareLibDir,
		wdc.prepareVersionXml,
		wdc.prepareComponentWorkDir,
		wdc.prepareSymlinks,
	}

	fmt.Println("start to prepare working dir")
	hwlog.RunLog.Info("-----Start to prepare working dir-----")
	for _, function := range prepareWorkingDirTasks {
		if err := function(); err != nil {
			return err
		}
	}
	fmt.Println("prepare working dir success")
	hwlog.RunLog.Info("-----Prepare working dir successful-----")
	return nil
}

func (wdc *WorkingDirCtl) prepareRootWorkDir() error {
	hwlog.RunLog.Info("start to prepare root work directories")

	mefWorkPath := wdc.pathMgr.GetWorkPath()
	if err := fileutils.CreateDir(mefWorkPath, fileutils.Mode700); err != nil {
		hwlog.RunLog.Errorf("create mef root work path failed: %v", err.Error())
		return errors.New("create mef root work path failed")
	}

	hwlog.RunLog.Info("prepare root work directories successful")
	return nil
}

func (wdc *WorkingDirCtl) prepareLibDir() error {
	hwlog.RunLog.Info("start to prepare lib dir")
	currentPath, err := wdc.getCurrentPath()
	if err != nil {
		return err
	}

	libDst := wdc.pathMgr.GetWorkLibDirPath()
	if err = fileutils.CreateDir(libDst, fileutils.Mode700); err != nil {
		hwlog.RunLog.Errorf("create lib path failed: %v", err.Error())
		return errors.New("create lib path failed")
	}

	libSrc := path.Join(currentPath, util.MefLibDir)
	if err = fileutils.CopyDirWithSoftlink(libSrc, libDst); err != nil {
		hwlog.RunLog.Errorf("copy lib dir failed, error: %v", err.Error())
		return errors.New("copy lib dir failed")
	}

	for _, component := range wdc.components {
		componentMgr := util.GetComponentMgr(component)
		if err = componentMgr.PrepareLibDir(libSrc, wdc.pathMgr); err != nil {
			return err
		}
	}

	hwlog.RunLog.Info("prepare lib dir successful")
	return nil
}

func (wdc *WorkingDirCtl) prepareRunSh() error {
	hwlog.RunLog.Info("start to copy run.sh")
	currentPath, err := wdc.getCurrentPath()
	if err != nil {
		return err
	}

	scriptSrc := path.Join(currentPath, util.MefScriptsDir, util.MefRunScript)
	if err = fileutils.CopyFile(scriptSrc, wdc.pathMgr.GetRunShPath()); err != nil {
		hwlog.RunLog.Errorf("copy run scripts dir failed, error: %v", err.Error())
		return errors.New("copy run scripts dir failed")
	}

	runScripPath := wdc.pathMgr.GetRunShPath()
	if err = fileutils.SetPathPermission(runScripPath, common.Mode500, false, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] mode failed, error: %s", runScripPath, err.Error())
		return errors.New("set run script path mode failed")
	}

	hwlog.RunLog.Info("copy run.sh successful")
	return nil
}

func (wdc *WorkingDirCtl) prepareBinDir() error {
	hwlog.RunLog.Info("start to prepare bin dir")
	currentPath, err := wdc.getCurrentPath()
	if err != nil {
		return err
	}

	sbinDst := wdc.pathMgr.GetBinDirPath()
	if err = fileutils.CreateDir(sbinDst, fileutils.Mode700); err != nil {
		hwlog.RunLog.Errorf("create sbin work path failed: %v", err.Error())
		return errors.New("create sbin work path failed")
	}

	sbinSrc := filepath.Join(currentPath, util.MefBinDir, util.ControllerBin)
	controllerPath := wdc.pathMgr.GetControllerBinPath()
	if err = fileutils.CopyFile(sbinSrc, controllerPath); err != nil {
		hwlog.RunLog.Errorf("copy mef controller failed, error: %v", err.Error())
		return errors.New("copy mef controller failed")
	}

	hwlog.RunLog.Info("prepare bin dir successful")
	return nil
}

func (wdc *WorkingDirCtl) prepareVersionXml() error {
	hwlog.RunLog.Info("start to copy version.xml")
	currentPath, err := wdc.getCurrentPath()
	if err != nil {
		return err
	}

	srcFile := path.Join(currentPath, util.VersionXml)
	if err = fileutils.CopyFile(srcFile, wdc.pathMgr.GetVersionXmlPath()); err != nil {
		hwlog.RunLog.Errorf("copy version.xml failed, error: %v", err.Error())
		return errors.New("copy version.xml failed")
	}

	versionXmlPath := wdc.pathMgr.GetVersionXmlPath()
	if err = fileutils.SetPathPermission(versionXmlPath, common.Mode400, false, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] mode failed, error: %s", versionXmlPath, err.Error())
		return errors.New("set version.xml path mode failed")
	}

	hwlog.RunLog.Info("copy version.xml successful")
	return nil
}

func (wdc *WorkingDirCtl) getCurrentPath() (string, error) {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		hwlog.RunLog.Error("get current path failed")
		return "", errors.New("get current path failed")
	}
	currentPath := path.Dir(currentDir)

	return currentPath, nil
}

func (wdc *WorkingDirCtl) prepareComponentWorkDir() error {
	hwlog.RunLog.Info("start to prepare component work directories")
	workPath := wdc.pathMgr.GetImagesDirPath()
	if err := fileutils.CreateDir(workPath, fileutils.Mode700); err != nil {
		hwlog.RunLog.Errorf("create component root work path [%s] failed: %v", workPath, err.Error())
		return errors.New("create component root work path failed")
	}

	// prepare components' working directory
	for _, component := range wdc.components {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.PrepareSingleComponentDir(wdc.pathMgr); err != nil {
			return err
		}
	}

	hwlog.RunLog.Info("prepare component work directories successful")
	return nil
}

func (wdc *WorkingDirCtl) prepareSymlinks() error {
	hwlog.RunLog.Info("start to prepare softlinks")

	configSrc := wdc.pathMgr.GetWorkPath()
	configDst := wdc.mefLinkPath
	if err := os.Symlink(configSrc, configDst); err != nil {
		hwlog.RunLog.Errorf("create work dir symlink failed, error: %v", err.Error())
		return errors.New("create work dir symlink failed")
	}

	hwlog.RunLog.Info("prepare softlinks successful")
	return nil
}

func (wdc *WorkingDirCtl) prepareInstallParamJson() error {
	curDirPath, err := filepath.Abs(filepath.Dir(filepath.Dir(os.Args[0])))
	if err != nil {
		hwlog.RunLog.Errorf("get current dir abs path failed: %s", err.Error())
		return errors.New("get current dir abs path failed")
	}

	srcPath := path.Join(curDirPath, util.InstallParamJson)
	dstPath := wdc.pathMgr.GetInstallParamJsonPath()
	if err = fileutils.CopyFile(srcPath, dstPath); err != nil {
		hwlog.RunLog.Errorf("prepare install-param.json failed: %s", err.Error())
		return errors.New("prepare install-param.json failed")
	}

	if err := backuputils.BackUpFiles(dstPath); err != nil {
		hwlog.RunLog.Warnf("back up install-param.json failed: %s", err.Error())
	}

	return nil
}
