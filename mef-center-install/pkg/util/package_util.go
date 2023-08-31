// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

type zipContent struct {
	TarPath, CmsPath, CrlPath string
}

// CheckZipFile check zip file
func CheckZipFile(unpackPath, zipPath string) error {
	const zipSizeMul int64 = 100
	if filepath.Dir(zipPath) == unpackPath {
		hwlog.RunLog.Errorf("zip path cannot be the unpack dir:%s", unpackPath)
		return errors.New("zip path cannot be inside the unpack path")
	}

	if zipPath == "" || !fileutils.IsExist(zipPath) {
		return errors.New("zip path does not exist")
	}

	if !path.IsAbs(zipPath) {
		return errors.New("zip path is not an absolute path")
	}

	if _, err := fileutils.RealFileCheck(zipPath, true, false, zipSizeMul); err != nil {
		return fmt.Errorf("check zip path failed: %s", err.Error())
	}
	return nil
}

// GetVerifyFileName get zip file content file name
func GetVerifyFileName(unpackZipPath, componentName string) (*zipContent, error) {
	unpackAbsPath, err := filepath.EvalSymlinks(unpackZipPath)
	if err != nil {
		hwlog.RunLog.Errorf("get unpack abs path failed: %s", unpackAbsPath)
		return nil, errors.New("get unpack abs path failed")
	}
	var tarName, cmsName, crlName string
	dir, err := common.ReadDir(unpackZipPath)
	if err != nil {
		hwlog.RunLog.Errorf("traversal unpack path failed: %s", err.Error())
		return nil, errors.New("traversal unpack path failed")
	}

	for _, file := range dir {
		if !strings.Contains(file.Name(), componentName) {
			continue
		}

		if strings.HasSuffix(file.Name(), common.TarGzSuffix) {
			if tarName != "" {
				hwlog.RunLog.Error("more than 1 tar.gz file in zip file")
				return nil, errors.New("more than 1 tar.gz file in zip file")
			}
			tarName = file.Name()
		}

		if strings.HasSuffix(file.Name(), common.CmsSuffix) {
			if cmsName != "" {
				hwlog.RunLog.Error("more than 1 cms file in zip file")
				return nil, errors.New("more than 1 cms file in zip file")
			}
			cmsName = file.Name()
		}

		if strings.HasSuffix(file.Name(), common.CrlSuffix) {
			if crlName != "" {
				hwlog.RunLog.Error("more than 1 crl file in zip file")
				return nil, errors.New("more than 1 crl file in zip file")
			}
			crlName = file.Name()
		}
	}

	if tarName == "" || cmsName == "" || crlName == "" {
		hwlog.RunLog.Error("the zip file does not contain all necessary file")
		return nil, errors.New("the zip file does not contain all necessary file")
	}
	return &zipContent{
		TarPath: filepath.Join(unpackAbsPath, tarName),
		CmsPath: filepath.Join(unpackAbsPath, cmsName),
		CrlPath: filepath.Join(unpackAbsPath, crlName)}, nil
}

// ClearPakEnv if install failed ,need to clear package file
func ClearPakEnv(path string) {
	fmt.Println("install failed, start to clear environment")
	hwlog.RunLog.Info("-----Start to clear environment-----")
	if err := common.DeleteAllFile(path); err != nil {
		fmt.Println("clear environment failed, please clear manually")
		hwlog.RunLog.Warnf("clear environment meets err:%s, need to do it manually", err.Error())
		hwlog.RunLog.Info("-----End to clear environment-----")
		return
	}
	fmt.Println("clear environment success")
	hwlog.RunLog.Info("clear environment success")
	hwlog.RunLog.Info("-----End to clear environment-----")
}
