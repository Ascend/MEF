// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// GetArch is used to get the arch info
func GetArch() (string, error) {
	arch, err := common.RunCommand(ArchCommand, true, "-i")
	if err != nil {
		return "", err
	}
	return arch, nil
}

func GetInstallInfo() (*InstallParamJsonTemplate, error) {
	currentDir, err := filepath.Abs(filepath.Dir(filepath.Dir(os.Args[0])))
	if err != nil {
		fmt.Printf("get current absolute path error: %s", err)
		hwlog.RunLog.Errorf("get current absolute path error: %s", err)
		return nil, err
	}

	paramJsonPath := path.Join(currentDir, InstallParamJson)
	paramJsonMgr, err := GetInstallParamJsonInfo(paramJsonPath)
	if err != nil {
		fmt.Printf("get current absolute path error: %s", err)
		hwlog.RunLog.Errorf("get current absolute path error: %s", err)
		return nil, err
	}

	return paramJsonMgr, nil
}
