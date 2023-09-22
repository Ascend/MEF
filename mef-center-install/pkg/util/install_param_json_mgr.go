// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

// InstallParamJsonTemplate is the struct to deal with install_param.json
type InstallParamJsonTemplate struct {
	InstallDir      string   `json:"install_dir"`
	LogDir          string   `json:"log_dir"`
	LogBackupDir    string   `json:"log_backup_dir"`
	OptionComponent []string `json:"option_component,omitempty"`
}

// GetInstallParamJsonInfo is used to get infos from install_param.json
func GetInstallParamJsonInfo(jsonPath string) (*InstallParamJsonTemplate, error) {
	installParam := InstallParamJsonTemplate{}
	if err := installParam.initFromFilePath(jsonPath); err == nil {
		if err := backuputils.NewBackupFileMgr(jsonPath).BackUp(); err != nil {
			fmt.Println("warning: create backup of install-param.json failed")
		}
		return &installParam, nil
	}
	fmt.Println("get install param json failed, try restore from backup")
	if err := backuputils.NewBackupFileMgr(jsonPath).Restore(); err != nil {
		return nil, fmt.Errorf("restore install-param.json from backup failed, %s", err.Error())
	}
	if err := installParam.initFromFilePath(jsonPath); err != nil {
		return nil, err
	}
	return &installParam, nil
}

func (ins *InstallParamJsonTemplate) initFromFilePath(jsonPath string) error {
	if !utils.IsExist(jsonPath) {
		return errors.New("install_param.json not exist")
	}
	file, err := utils.LoadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read component json failed: %s", err.Error())
	}
	err = json.Unmarshal(file, ins)
	if err != nil {
		return fmt.Errorf("parse json file failed: %s", err.Error())
	}
	return nil
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

	if err = ins.backupInstallParamJson(jsonPath); err != nil {
		hwlog.RunLog.Warnf("back up %s failed, %v", InstallParamJson, err.Error())
	}

	return nil
}

func (ins *InstallParamJsonTemplate) backupInstallParamJson(jsonPath string) error {
	realJsonPath, err := filepath.EvalSymlinks(jsonPath)
	if err != nil {
		return fmt.Errorf("get real path failed: %s", err.Error())
	}
	if err = backuputils.BackUpFiles(realJsonPath); err != nil {
		return fmt.Errorf("create backup file failed: %s", err.Error())
	}
	return nil
}

// AddComponentToInstallInfo add install option component in install info json
func AddComponentToInstallInfo(component, jsonPath string) error {
	installInfo, err := GetInstallInfo()
	if err != nil {
		return err
	}
	for _, c := range installInfo.OptionComponent {
		if c == component {
			return nil
		}
	}
	installInfo.OptionComponent = append(installInfo.OptionComponent, component)
	if err := installInfo.SetInstallParamJsonInfo(jsonPath); err != nil {
		return err
	}
	return nil
}

// DeleteComponentToInstallInfo delete uninstall option component in install info json
func DeleteComponentToInstallInfo(component, jsonPath string) error {
	installInfo, err := GetInstallInfo()
	if err != nil {
		return err
	}
	index := -1
	for i, c := range installInfo.OptionComponent {
		if c == component {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.New(ComponentNotInstalled)
	}
	installInfo.OptionComponent = append(installInfo.OptionComponent[:index], installInfo.OptionComponent[index+1:]...)
	if err := installInfo.SetInstallParamJsonInfo(jsonPath); err != nil {
		return err
	}
	return nil
}

// OptionComponentExist check if component is exit
func OptionComponentExist(component string) (bool, error) {
	installInfo, err := GetInstallInfo()
	if err != nil {
		return false, err
	}
	for _, c := range installInfo.OptionComponent {
		if c == component {
			return true, nil
		}
	}
	return false, nil
}
