// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package path get the specified path based on the current execution path
package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

// GetCompWorkDir get component work dir. e.g. /usr/local/mindx/MEFEdge/software_A/edge_main
// Cannot be recorded log because one of the call points is before the init log
func GetCompWorkDir() (string, error) {
	// currentPath: e.g. /usr/local/mindx/MEFEdge/software_A/edge_main/bin/edge-main
	currentPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get current path failed, %v", err)
	}

	// compWorkDir: e.g. /usr/local/mindx/MEFEdge/software_A/edge_main
	compWorkDir, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(currentPath)))
	if err != nil {
		return "", fmt.Errorf("eval comp work dir symlinks failed, %v", err)
	}
	return compWorkDir, nil
}

// GetCompConfigDir get component config dir. e.g. /usr/local/mindx/MEFEdge/config/edge_main
func GetCompConfigDir() (string, error) {
	// compWorkDir: e.g. /usr/local/mindx/MEFEdge/software_A/edge_main
	compWorkDir, err := GetCompWorkDir()
	if err != nil {
		hwlog.RunLog.Errorf("get comp work dir failed, error: %v", err)
		return "", errors.New("get comp work dir failed")
	}

	// compCfgDir: e.g. /usr/local/mindx/MEFEdge/software_A/edge_main/config -> /usr/local/mindx/MEFEdge/config/edge_main
	compCfgDir, err := filepath.EvalSymlinks(filepath.Join(compWorkDir, constants.Config))
	if err != nil {
		return "", fmt.Errorf("eval comp config dir symlink failed, %v", err)
	}
	return compCfgDir, nil
}

// GetCompSpecificDir get comp specific dir (e.g. inner_certs, peer_certs...)
func GetCompSpecificDir(dirName string) (string, error) {
	compConfigDir, err := GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Errorf("get comp config dir failed, error: %v", err)
		return "", errors.New("get comp config dir failed")
	}
	return filepath.Join(compConfigDir, dirName), nil
}
