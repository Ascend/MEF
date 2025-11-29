// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package config this file for edge core config file manager
package config

import (
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

// SetDatabase set database to edge core config file
func SetDatabase(edgeCoreConfigPath, dataSource string) error {
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore config failed, error: %v", err)
		return errors.New("get edgecore config failed")
	}
	err = util.SetJsonValue(edgeCoreConfig, dataSource, constants.ConfigDatabase, constants.ConfigDataSource)
	if err != nil {
		hwlog.RunLog.Errorf("set database value failed, error: %v", err)
		return errors.New("set database value failed")
	}

	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edgecore config failed, error: %v", err)
		return errors.New("save edgecore config failed")
	}
	return nil
}

// SetCertPath set cert path to edge core config file
func SetCertPath(edgeCoreConfigPath string, configPathMgr *pathmgr.ConfigPathMgr) error {
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore config failed, error: %v", err)
		return errors.New("get edgecore config failed")
	}

	tlsCaFile := configPathMgr.GetCompInnerRootCertPath(constants.EdgeCore)
	if err = util.SetJsonValue(edgeCoreConfig, tlsCaFile, constants.ConfigModules, constants.ConfigEdgeHub,
		constants.ConfigTlsCaFile); err != nil {
		hwlog.RunLog.Errorf("set value for modules.edgeHub.tlsCaFile failed, error: %v", err)
		return errors.New("set value for modules.edgeHub.tlsCaFile failed")
	}

	tlsCertFile := configPathMgr.GetCompInnerSvrCertPath(constants.EdgeCore)
	if err = util.SetJsonValue(edgeCoreConfig, tlsCertFile, constants.ConfigModules, constants.ConfigEdgeHub,
		constants.ConfigTlsCertFile); err != nil {
		hwlog.RunLog.Errorf("set json value for modules.edgeHub.tlsCertFile failed, error: %v", err)
		return errors.New("set value for modules.edgeHub.tlsCertFile failed")
	}

	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edgecore config failed, error: %v", err)
		return errors.New("save edgecore config failed")
	}
	return nil
}

// SetHostname set hostname config to edge core config file
func SetHostname(edgeCoreConfigPath, hostnameOverride string) error {
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore config failed, error: %v", err)
		return errors.New("get edgecore config failed")
	}
	if err = util.SetJsonValue(edgeCoreConfig, hostnameOverride, constants.ConfigModules, constants.ConfigEdged,
		constants.ConfigHostnameOverride); err != nil {
		hwlog.RunLog.Errorf("set value for modules.edged.hostnameOverride failed, error: %v", err)
		return errors.New("set value for modules.edged.hostnameOverride failed")
	}

	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edgecore config failed, error: %v", err)
		return errors.New("save edgecore config failed")
	}
	return nil
}

// SetSerialNumber set Serial Number to edge core config file
func SetSerialNumber(edgeCoreConfigPath, serialNumber string) error {
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore config failed, error: %v", err)
		return errors.New("get edgecore config failed")
	}
	if err = util.SetJsonValue(edgeCoreConfig, serialNumber, constants.ConfigModules, constants.ConfigEdged,
		constants.ConfigNodeLabels, constants.ConfigSerialNumber); err != nil {
		hwlog.RunLog.Errorf("set value for modules.edged.nodeLabels.serialNumber failed, error: %v", err)
		return errors.New("set value for modules.edged.nodeLabels.serialNumber failed")
	}

	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edgecore config failed, error: %v", err)
		return errors.New("save edgecore config failed")
	}
	return nil
}

// SetCgroupDriver set cgroup driver to edge core config file
func SetCgroupDriver(edgeCoreConfigPath, SetCgroupDriver string) error {
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore config failed, error: %v", err)
		return errors.New("get edgecore config failed")
	}
	if err = util.SetJsonValue(edgeCoreConfig, SetCgroupDriver, constants.ConfigModules, constants.ConfigEdged,
		constants.ConfigTailoredKubelet, constants.ConfigCgroupDriver); err != nil {
		hwlog.RunLog.Errorf("set value for modules.edged.tailoredKubeletConfig.cgroupDriver failed, error: %v", err)
		return errors.New("set value for modules.edged.tailoredKubeletConfig.cgroupDriver failed")
	}

	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edgecore config failed, error: %v", err)
		return errors.New("save edgecore config failed")
	}
	return nil
}

// SetNodeIP set nodeIP to edge core config file
func SetNodeIP(edgeCoreConfigPath, nodeIP string) error {
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		return fmt.Errorf("get edge core config failed, error: %v", err)
	}

	if err = util.SetJsonValue(edgeCoreConfig, nodeIP, constants.ConfigModules, constants.ConfigEdged,
		constants.ConfigNodeIP); err != nil {
		return fmt.Errorf("set value for modules.edged.nodeIP failed, error: %v", err)
	}

	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		return fmt.Errorf("save edge core config failed, error: %v", err)
	}
	return nil
}

// SmoothEdgeCoreConfigPipePath smooth edgecore pipe config to new edgecore config file
func SmoothEdgeCoreConfigPipePath(installRootDir, pipePath string) error {
	edgeCoreConfigPath := pathmgr.NewConfigPathMgr(installRootDir).GetEdgeCoreConfigPath()
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore config failed, error: %v", err)
		return errors.New("get edgecore config failed")
	}
	if err = util.SetJsonValue(edgeCoreConfig, pipePath, constants.ConfigModules, constants.ConfigEdgeHub,
		constants.ConfigTlsPrivateKeyFile); err != nil {
		hwlog.RunLog.Errorf("set value for modules.edgeHub.tlsPrivateKeyFile failed, error: %v", err)
		return errors.New("set value for modules.edgeHub.tlsPrivateKeyFile failed")
	}
	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edgecore config failed, error: %v", err)
		return errors.New("save edgecore config failed")
	}
	return nil
}

// SmoothEdgeCoreSafeConfig smooth edgecore config to new edgecore config file
func SmoothEdgeCoreSafeConfig(installRootDir string) error {
	edgeCoreConfigPath := pathmgr.NewConfigPathMgr(installRootDir).GetEdgeCoreConfigPath()
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edge core config failed, error: %v", err)
		return errors.New("get edge core config failed")
	}

	var needUpdateFields = []struct {
		value  interface{}
		fields []string
	}{
		{constants.KubeletRootDir, []string{constants.ConfigModules, constants.ConfigEdged,
			constants.ConfigRootDirectory}},
		{0, []string{constants.ConfigModules, constants.ConfigEdged,
			constants.ConfigTailoredKubelet, constants.ConfigReadOnlyPort}},
		{true, []string{constants.ConfigModules, constants.ConfigEdged,
			constants.ConfigTailoredKubelet, constants.ConfigServerTLSBootstrap}},
		{"10%", []string{constants.ConfigModules, constants.ConfigEdged,
			constants.ConfigTailoredKubelet, constants.ConfigEvictionHard, constants.SignalNodeFsAvailable}},
		{"5%", []string{constants.ConfigModules, constants.ConfigEdged,
			constants.ConfigTailoredKubelet, constants.ConfigEvictionHard, constants.SignalNodeFsInodesFree}},
		{map[string]interface{}{"enable": false}, []string{constants.ConfigModules,
			constants.ConfigDeviceTwin}},
		{map[string]interface{}{"enable": false}, []string{constants.ConfigModules,
			constants.ConfigEventBus}},
	}

	for _, v := range needUpdateFields {
		err = util.SetJsonValue(edgeCoreConfig, v.value, v.fields...)
		if err != nil {
			hwlog.RunLog.Errorf("set edge core filed [%s] failed: %v", v.fields[len(v.fields)-1], err)
			return errors.New("set edge core filed failed")
		}
	}

	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edge core config failed, error: %v", err)
		return errors.New("save edge core config failed")
	}
	return nil
}

// SmoothEdgeCoreConfigSystemReserve smooth edgecore system-reserved config to new edgecore config file
func SmoothEdgeCoreConfigSystemReserve(installRootDir string, isRollback bool) error {
	modifiedJson := map[string]string{}
	if !isRollback {
		config, err := LoadPodConfig()
		if err != nil {
			return fmt.Errorf("set edgecore config failed, error: %v", err)
		}
		modifiedJson["cpu"] = fmt.Sprintf("%.2f", config.SystemReservedCPUQuota)
		modifiedJson["memory"] = fmt.Sprintf("%dMi", config.SystemReservedMemoryQuota)
	}

	edgeCoreConfigPath := pathmgr.NewConfigPathMgr(installRootDir).GetEdgeCoreConfigPath()
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore config failed, error: %v", err)
		return errors.New("get edgecore config failed")
	}

	if err = util.SetJsonValue(edgeCoreConfig, modifiedJson, constants.ConfigModules, constants.ConfigEdged,
		constants.ConfigTailoredKubelet, "systemReserved"); err != nil {
		hwlog.RunLog.Errorf("set value for modules.edged.tailoredKubeletConfig failed, error: %v", err)
		return errors.New("set value for modules.edged.tailoredKubeletConfig failed")
	}
	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edgecore config failed, error: %v", err)
		return errors.New("save edgecore config failed")
	}
	return nil
}

// setOldKubeletRootDir smooth kubelet root dir
func setOldKubeletRootDir(installRootDir string) error {
	edgeCoreConfigPath := pathmgr.NewConfigPathMgr(installRootDir).GetEdgeCoreConfigPath()
	edgeCoreConfig, err := util.LoadJsonFile(edgeCoreConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore config failed, error: %v", err)
		return errors.New("get edgecore config failed")
	}

	var needUpdateFields = []struct {
		value  interface{}
		fields []string
	}{
		{constants.OldKubeletRootDir, []string{constants.ConfigModules, constants.ConfigEdged,
			constants.ConfigRootDirectory}},
	}

	for _, v := range needUpdateFields {
		err = util.SetJsonValue(edgeCoreConfig, v.value, v.fields...)
		if err != nil {
			hwlog.RunLog.Errorf("set edge core filed [%s] failed: %v", v.fields[len(v.fields)-1], err)
			return errors.New("set edge core filed failed")
		}
	}

	if err = util.SaveJsonValue(edgeCoreConfigPath, edgeCoreConfig); err != nil {
		hwlog.RunLog.Errorf("save edge core config failed, error: %v", err)
		return errors.New("save edge core config failed")
	}

	return nil
}

// EffectToOldestVersionSmooth do smooth job for config file when upgrade old version
func EffectToOldestVersionSmooth(upgradeVersion string, installRootDir string) error {
	if upgradeVersion == constants.Version5Rc1 {
		if err := SmoothEdgeCoreConfigPipePath(installRootDir, constants.OldTlsPrivateKeyFile); err != nil {
			return fmt.Errorf("smooth old config to edge core config file failed, error: %v", err)
		}
		if err := SmoothEdgeCoreConfigSystemReserve(installRootDir, true); err != nil {
			return fmt.Errorf("smooth old system reserved config failed, error: %v", err)
		}
	}

	if strings.Compare(upgradeVersion, constants.Version5Rc3) < 0 &&
		!strings.HasPrefix(upgradeVersion, constants.Version5) {
		if err := setOldKubeletRootDir(installRootDir); err != nil {
			return fmt.Errorf("smooth old config to edge core config file failed, error: %v", err)
		}
	}

	return nil
}
