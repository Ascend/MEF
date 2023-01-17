// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"encoding/json"
	"fmt"
	"os"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

// InstallParamJsonTemplate is the struct to deal with install_param.json
type InstallParamJsonTemplate struct {
	Components []string `json:"Components"`
	InstallDir string   `json:"install_dir"`
	LogDir     string   `json:"log_dir"`
}

// GetInstallParamJsonInfo is used to get infos from install_param.json
func GetInstallParamJsonInfo(jsonPath string) (*InstallParamJsonTemplate, error) {
	var componentsIns InstallParamJsonTemplate
	file, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("read component json failed: %s", err.Error())
	}
	err = json.Unmarshal(file, &componentsIns)
	if err != nil {
		return nil, fmt.Errorf("parse json file failed: %s", err.Error())
	}
	return &componentsIns, nil
}

// SetInstallParamJsonInfo is used to save infos into install_param.json
func (ins *InstallParamJsonTemplate) SetInstallParamJsonInfo(jsonPath string) error {
	file, err := os.OpenFile(jsonPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, common.Mode600)
	if err != nil {
		return fmt.Errorf("open %s failed: %s", InstallParamJson, err.Error())
	}
	defer func() {
		if err = file.Close(); err != nil {
			hwlog.RunLog.Errorf("close %s failed: %s", InstallParamJson, err.Error())
		}
	}()
	encoder := json.NewEncoder(file)

	if err = encoder.Encode(ins); err != nil {
		return fmt.Errorf("write content into %s failed: %s", InstallParamJson, err.Error())
	}

	return nil
}
