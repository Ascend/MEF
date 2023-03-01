// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		dirAbsPath, err := filepath.EvalSymlinks(tempDir)
		if err != nil {
			hwlog.RunLog.Errorf("get %s's abspath failed: %s", tempDir, err.Error())
			return fmt.Errorf("get %s's abspath failed", tempDir)
		}
		if err = common.SetPathPermission(dirAbsPath, mode, isRecursive, ignoreFile); err != nil {
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

		if err = common.SetPathPermission(dirFilePath, mode, isRecursive, ignoreFile); err != nil {
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

		if err = common.SetPathPermission(dirFilePath, mode, false, false); err != nil {
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
		mode:        0700,
		fileType:    DirType,
		isRecursive: true,
		ignoreFile:  true,
		fileList: []string{
			cmm.pathMgr.GetMefPath(),
		},
	}
}

func (cmm *CenterModeMgr) getWorkMode500Dir() modeMgr {
	return modeMgr{
		mode:        0500,
		fileType:    DirType,
		isRecursive: false,
		ignoreFile:  true,
		fileList: []string{
			cmm.pathMgr.WorkPathMgr.GetRelativeBinDirPath(),
			cmm.pathMgr.WorkPathMgr.GetRelativeLibDirPath(),
			cmm.pathMgr.WorkPathMgr.GetRelativeKmcLibDirPath(),
			cmm.pathMgr.WorkPathMgr.GetRelativeLibDirPath(),
		},
	}
}

func (cmm *CenterModeMgr) getWorkMode600File() modeMgr {
	return modeMgr{
		mode:        0600,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetRelativeImagesDirPath(), "**/**/*.yaml"),
		},
	}
}

func (cmm *CenterModeMgr) getWorkMode500File() modeMgr {
	return modeMgr{
		mode:        0500,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			cmm.pathMgr.WorkPathMgr.GetRelativeRunShPath(),
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetRelativeLibDirPath(), "**/*.so"),
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetRelativeBinDirPath(), "**"),
		},
	}
}

func (cmm *CenterModeMgr) getWorkMode400File() modeMgr {
	return modeMgr{
		mode:        0400,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			cmm.pathMgr.WorkPathMgr.GetVersionXmlPath(),
			cmm.pathMgr.WorkPathMgr.GetInstallParamJsonPath(),
			filepath.Join(cmm.pathMgr.WorkPathMgr.GetRelativeImagesDirPath(), "**/**/*.tar.gz"),
		},
	}
}

func (cmm *CenterModeMgr) getConfigMode700Dir() modeMgr {
	return modeMgr{
		mode:        0700,
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
		mode:        0600,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/**/*.ks"),
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/**/*.key"),
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/**/*.crt"),
		},
	}
}

func (cmm *CenterModeMgr) getConfigMode400File() modeMgr {
	return modeMgr{
		mode:        0400,
		fileType:    FileType,
		isRecursive: false,
		ignoreFile:  false,
		fileList: []string{
			filepath.Join(cmm.pathMgr.GetConfigPath(), "**/**/*.json"),
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
