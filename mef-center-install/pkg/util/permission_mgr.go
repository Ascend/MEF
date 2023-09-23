// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

type modeMgr struct {
	mode        os.FileMode
	fileType    string
	isRecursive bool
	ignoreFile  bool
	fileList    []string
}

func (mi *modeMgr) setMode() error {
	if mi.fileType == DirType {
		if err := mi.setDirMode(mi.fileList, mi.mode, mi.isRecursive, mi.ignoreFile); err != nil {
			return err
		}

		return nil
	}

	if err := mi.setFileMode(mi.fileList, mi.mode, mi.isRecursive, mi.ignoreFile); err != nil {
		return err
	}

	return nil
}

func (mi *modeMgr) setDirMode(dirList []string, mode os.FileMode, isRecursive, ignoreFile bool) error {
	for _, tempDir := range dirList {
		if strings.Contains(tempDir, "**") {
			if err := mi.handleTypeFileMode(tempDir, mode); err != nil {
				return err
			}

			continue
		}

		dirAbsPath, err := filepath.EvalSymlinks(tempDir)
		if err != nil {
			hwlog.RunLog.Errorf("get %s's abspath failed: %s", tempDir, err.Error())
			return fmt.Errorf("get %s's abspath failed", tempDir)
		}
		if err = fileutils.SetPathPermission(dirAbsPath, mode, isRecursive, ignoreFile); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (mi *modeMgr) setFileMode(fileList []string, mode os.FileMode, isRecursive, ignoreFile bool) error {
	for _, tempFile := range fileList {
		if strings.Contains(tempFile, "**") {
			if err := mi.handleTypeFileMode(tempFile, mode); err != nil {
				return err
			}

			continue
		}

		dirFilePath, err := filepath.EvalSymlinks(tempFile)
		if err != nil {
			hwlog.RunLog.Errorf("get %s's abspath failed: %s", tempFile, err.Error())
			return fmt.Errorf("get %s's abspath failed", tempFile)
		}

		if err = fileutils.SetPathPermission(dirFilePath, mode, isRecursive, ignoreFile); err != nil {
			hwlog.RunLog.Error(err)
			return errors.New("set path permission failed")
		}
	}

	return nil
}

func (mi *modeMgr) handleTypeFileMode(typeFile string, mode os.FileMode) error {
	fileTypeList, err := filepath.Glob(typeFile)
	if err != nil {
		hwlog.RunLog.Errorf("execute glob func for path %s failed: %s", typeFile, err.Error())
		return fmt.Errorf("execute glob func for path %s failed", typeFile)
	}

	for _, file := range fileTypeList {
		dirFilePath, err := filepath.EvalSymlinks(file)
		if err != nil {
			hwlog.RunLog.Errorf("get %s's abspath failed: %s", file, err.Error())
			return fmt.Errorf("get %s's abspath failed", file)
		}

		if err = fileutils.SetPathPermission(dirFilePath, mode, false, false); err != nil {
			hwlog.RunLog.Error(err)
			return errors.New("set path permission failed")
		}
	}

	return nil
}

// CenterModeMgr is the struct to manager the MEF-Center Mode
type CenterModeMgr struct {
	pathMgr *InstallDirPathMgr
}

// GetCenterModeMgr is the func to init a CenterModeMgr struct
func GetCenterModeMgr(pathMgr *InstallDirPathMgr) *CenterModeMgr {
	return &CenterModeMgr{
		pathMgr: pathMgr,
	}
}

func (cmm *CenterModeMgr) getWorkMode700Dir() modeMgr {
	return modeMgr{
		mode:        common.Mode700,
		fileType:    DirType,
		isRecursive: true,
		ignoreFile:  true,
		fileList: []string{
			cmm.pathMgr.GetMefPath(),
		},
	}
}

func (cmm *CenterModeMgr) getWorkMode500Dir() modeMgr {
	imageDirPattern := fmt.Sprintf("**/**/%s", ImageDir)
	return modeMgr{
		mode:        common.Mode500,
		fileType:    DirType,
		isRecursive: false,
		ignoreFile:  true,
		fileList: []string{
			cmm.pathMgr.WorkPathMgr.GetBinDirPath(),
			cmm.pathMgr.WorkPathMgr.GetWorkLibDirPath(),
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetWorkLibDirPath(), "lib"),
			cmm.pathMgr.WorkPathMgr.GetKmcLibDirPath(),
			cmm.pathMgr.WorkPathMgr.GetWorkLibDirPath(),
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetWorkPath(), imageDirPattern),
		},
	}
}

func (cmm *CenterModeMgr) getWorkMode600File() modeMgr {
	return modeMgr{
		mode:        common.Mode600,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetImagesDirPath(), "**/**/*.yaml"),
		},
	}
}

func (cmm *CenterModeMgr) getWorkMode500File() modeMgr {
	return modeMgr{
		mode:        common.Mode500,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			cmm.pathMgr.WorkPathMgr.GetRunShPath(),
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetWorkLibDirPath(), "**/*.so"),
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetBinDirPath(), "**"),
		},
	}
}

func (cmm *CenterModeMgr) getWorkMode400File() modeMgr {
	return modeMgr{
		mode:        common.Mode400,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			cmm.pathMgr.WorkPathMgr.GetVersionXmlPath(),
			cmm.pathMgr.WorkPathMgr.GetInstallParamJsonPath(),
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetImagesDirPath(), "**/**/*.tar.gz"),
		},
	}
}

func (cmm *CenterModeMgr) getConfigMode700Dir() modeMgr {
	return modeMgr{
		mode:        common.Mode700,
		fileType:    DirType,
		isRecursive: true,
		ignoreFile:  true,
		fileList: []string{
			cmm.pathMgr.GetConfigPath(),
		},
	}
}

func (cmm *CenterModeMgr) getConfigMode600File() modeMgr {
	return modeMgr{
		mode:        common.Mode600,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/**/*.ks"),
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/*flag"),
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/*.json"),
		},
	}
}

func (cmm *CenterModeMgr) getConfigMode400File() modeMgr {
	return modeMgr{
		mode:        common.Mode400,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/**/*.crt"),
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/**/**/*.crt"),
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/**/*.key"),
		},
	}
}

func (cmm *CenterModeMgr) getWorkDirModeMgrs() []modeMgr {
	return []modeMgr{
		cmm.getWorkMode700Dir(),
		cmm.getWorkMode500Dir(),
		cmm.getWorkMode600File(),
		cmm.getWorkMode500File(),
		cmm.getWorkMode400File(),
	}
}

func (cmm *CenterModeMgr) getConfigDirModeMgrs() []modeMgr {
	return []modeMgr{
		cmm.getConfigMode700Dir(),
		cmm.getConfigMode600File(),
		cmm.getConfigMode400File(),
	}
}

// SetWorkDirMode is the func to set the mode of MEF-Center work Dir
func (cmm *CenterModeMgr) SetWorkDirMode() error {
	for _, sigModeMgr := range cmm.getWorkDirModeMgrs() {
		if err := sigModeMgr.setMode(); err != nil {
			return err
		}
	}
	return nil
}

// SetConfigDirMode is the func to set the mode of MEF-Center config Dir
func (cmm *CenterModeMgr) SetConfigDirMode() error {
	for _, sigModeMgr := range cmm.getConfigDirModeMgrs() {
		if err := sigModeMgr.setMode(); err != nil {
			return err
		}
	}
	return nil
}
