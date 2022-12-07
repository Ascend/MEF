// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base file utils used
package common

import (
	"huawei.com/mindx/common/utils"
	"os"
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
