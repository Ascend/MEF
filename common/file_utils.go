// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base file utils used
package common

import (
	"errors"
	"os"

	"huawei.com/mindx/common/utils"
)

const fileMode = 0600

// WriteData write data with path check
func WriteData(filePath string, fileData []byte) error {
	filePath, err := utils.CheckPath(filePath)
	if err != nil {
		return err
	}

	err = utils.MakeSureDir(filePath)
	if err != nil {
		return err
	}

	writer, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, fileMode)
	if err != nil {
		return err
	}
	defer func() {
		err := writer.Close()
		if err != nil {
			return
		}
	}()
	_, err = writer.Write(fileData)
	if err != nil {
		return err
	}
	return nil
}

// DeleteAllFile is used to delete all files into a path
func DeleteAllFile(filePath string) error {
	return os.RemoveAll(filePath)
}

// MakeSurePath is used to make sure a path exists by creating it if not
func MakeSurePath(tgtPath string) error {
	if utils.IsExist(tgtPath) {
		return nil
	}

	if err := os.MkdirAll(tgtPath, Mode700); err != nil {
		return errors.New("create directory failed")
	}

	return nil
}

// CopyDir is used to copy dir and all files into it
func CopyDir(srcPath string, dstPath string) error {
	if _, err := RunCommand(CommandCopy, true, "-r", srcPath, dstPath); err != nil {
		return err
	}
	return nil
}
