// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

// LoadPodConfig [method] loading pod config files for edge-om
func LoadPodConfig() (*PodConfig, error) {
	config, err := loadPodConfigWithBackup()
	if err != nil {
		return nil, fmt.Errorf("load pod config failed, err: %s", err.Error())
	}
	checkAndModifyPodConfigPara(&config)
	return &config, nil
}

func loadPodConfigWithBackup() (PodConfig, error) {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		return PodConfig{}, err
	}

	var podCfg PodConfig
	if err := backuputils.InitConfig(configPathMgr.GetContainerConfigPath(), podCfg.initContainerCfgFromPath); err != nil {
		hwlog.RunLog.Errorf("init container config failed, err: %v", err)
		return PodConfig{}, err
	}
	if err := backuputils.InitConfig(configPathMgr.GetPodConfigPath(), podCfg.initPodSecurityCfgFromPath); err != nil {
		hwlog.RunLog.Errorf("init pod security config failed, err: %v", err)
		return PodConfig{}, err
	}
	return podCfg, nil
}

func (p *PodConfig) initContainerCfgFromPath(filePath string) error {
	var containerCfg ContainerConfig
	containerCfgFile, err := fileutils.LoadFile(filePath)
	if err != nil {
		return fmt.Errorf("load container config file failed, %v", err)
	}
	if err = json.Unmarshal(containerCfgFile, &containerCfg); err != nil {
		return fmt.Errorf("unmarshal container config failed, %v", err)
	}
	p.ContainerConfig = containerCfg
	return nil
}

func (p *PodConfig) initPodSecurityCfgFromPath(filePath string) error {
	var podSecCfg PodSecurityConfig
	podCfgFile, err := fileutils.LoadFile(filePath)
	if err != nil {
		return fmt.Errorf("load pod security config file failed, %v", err)
	}
	if err = json.Unmarshal(podCfgFile, &podSecCfg); err != nil {
		return fmt.Errorf("unmarshal pod security config failed, %v", err)
	}
	p.PodSecurityConfig = podSecCfg
	return nil
}

func checkAndModifyPodConfigPara(podConfig *PodConfig) {
	podConfig.HostPath = checkAndModifyHostPath(podConfig.HostPath)
	podConfig.MaxContainerNumber = checkAndModifyMaxContainerNumber(podConfig.MaxContainerNumber)
	podConfig.SystemReservedCPUQuota = checkAndModifySystemReservedCPUQuota(podConfig.SystemReservedCPUQuota)
	podConfig.SystemReservedMemoryQuota = checkAndModifySystemReservedMemoryQuota(podConfig.SystemReservedMemoryQuota)
}

func checkAndModifyHostPath(hostPath []string) []string {
	const (
		defaultMaxHostPathNumber = 256
	)
	whiteListNumber := len(hostPath)
	if whiteListNumber > defaultMaxHostPathNumber {
		whiteListNumber = defaultMaxHostPathNumber
	}

	hostPathTmp := make([]string, 0, whiteListNumber)
	for i := 0; i < whiteListNumber; i++ {
		path := filepath.Clean(hostPath[i])
		if err := checkSingleHostPath(filepath.Clean(path)); err != nil {
			hwlog.RunLog.Errorf("checking pod config host path [%s] failed, and it won't be effective", hostPath[i])
			continue
		}
		hostPathTmp = append(hostPathTmp, filepath.Clean(path))
	}
	return hostPathTmp
}

func checkSingleHostPath(verifyPath string) error {
	_, err := fileutils.CheckOriginPath(verifyPath)
	if err != nil {
		return err
	}
	checkers := []func(string) error{
		checkModelFilePath,
		checkHostPathMode,
		checkHostPathParentOwner,
	}

	for _, checker := range checkers {
		if err := checker(verifyPath); err != nil {
			return err
		}
	}
	return nil
}

func checkModelFilePath(verifyPath string) error {
	// mode of modelfile path need to be 700
	if verifyPath == constants.ModeFileActiveDir {
		if _, err := fileutils.CheckOwnerAndPermission(verifyPath, constants.ModeUmask077,
			constants.RootUserUid); err != nil {
			return err
		}
	}
	return nil
}

func checkHostPathMode(verifyPath string) error {
	modeChecker := fileutils.NewFileModeChecker(true, fileutils.DefaultWriteFileMode, true, false)
	file, err := os.OpenFile(verifyPath, os.O_RDONLY, fileutils.Mode400)
	if err != nil {
		return fmt.Errorf("open file %s failed: %s", verifyPath, err.Error())
	}
	defer fileutils.CloseFile(file)
	if err = modeChecker.Check(file, verifyPath); err != nil {
		return fmt.Errorf("check file %s failed: %s", verifyPath, err.Error())
	}
	return nil
}

func checkHostPathParentOwner(verifyPath string) error {
	parentPath := filepath.Dir(verifyPath)
	ownerChecker := fileutils.NewFileOwnerChecker(true, false, constants.RootUserUid, constants.RootUserGid)
	file, err := os.OpenFile(parentPath, os.O_RDONLY, fileutils.Mode400)
	if err != nil {
		return fmt.Errorf("open file %s failed: %s", parentPath, err.Error())
	}
	defer fileutils.CloseFile(file)
	if err = ownerChecker.Check(file, parentPath); err != nil {
		return fmt.Errorf("check file %s failed: %s", verifyPath, err.Error())
	}
	return nil
}

func checkAndModifyMaxContainerNumber(containerNum int) int {
	const (
		defaultMaxContainerNumber = 20
		minContainerNumber        = 1
		maxContainerNumber        = 128
	)

	if containerNum < minContainerNumber || containerNum > maxContainerNumber {
		return defaultMaxContainerNumber
	}
	return containerNum
}

func checkAndModifySystemReservedCPUQuota(quota float64) float64 {
	const (
		defaultCpuQuota = 1.0
		maxCpuQuota     = 4.0
		minCpuQuota     = 0.5
	)
	if quota < minCpuQuota || quota > maxCpuQuota {
		quota = defaultCpuQuota
	}
	return quota
}

func checkAndModifySystemReservedMemoryQuota(quota int64) int64 {
	const (
		defaultMemoryQuota = 1024
		maxMemoryQuota     = 4096
		minMemoryQuota     = 512
	)
	if quota < minMemoryQuota || quota > maxMemoryQuota {
		quota = defaultMemoryQuota
	}
	return quota
}
