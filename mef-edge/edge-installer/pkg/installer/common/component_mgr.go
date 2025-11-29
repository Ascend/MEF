// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common this file for get component information
package common

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

// ComponentMgr is the struct to manager component operation
type ComponentMgr struct {
	workPathMgr   *pathmgr.WorkPathMgr
	configPathMgr *pathmgr.ConfigPathMgr
}

// NewComponentMgr new component manager
func NewComponentMgr(installRootDir string) *ComponentMgr {
	return &ComponentMgr{
		workPathMgr:   pathmgr.NewWorkPathMgr(installRootDir),
		configPathMgr: pathmgr.NewConfigPathMgr(installRootDir),
	}
}

// CheckAllServiceActive is the func to check all service status and to warn if not
func (c ComponentMgr) CheckAllServiceActive() {
	components := c.GetComponents()
	for _, component := range components {
		if !component.IsExist() {
			fmt.Printf("The component [%s] does not exist, please check it.\n", component.Name)
			hwlog.RunLog.Warnf("the component [%s] does not exist", component.Name)
			continue
		}
		if !component.isServiceActive() {
			fmt.Printf("warning: service [%s] is not active.\n", component.Service.Name)
			hwlog.RunLog.Warnf("service [%s] is not active", component.Service.Name)
		}
	}
}

// Start component
func (c ComponentMgr) Start(name string) error {
	component, err := c.Get(name)
	if err != nil {
		return fmt.Errorf("get component failed, error: %v", err)
	}
	return component.Start()
}

// Stop component
func (c ComponentMgr) Stop(name string) error {
	component, err := c.Get(name)
	if err != nil {
		return fmt.Errorf("get component failed, error: %v", err)
	}
	return component.Stop()
}

// Restart component
func (c ComponentMgr) Restart(name string) error {
	component, err := c.Get(name)
	if err != nil {
		return fmt.Errorf("get component failed, error: %v", err)
	}
	return component.Restart()
}

// Get component
func (c ComponentMgr) Get(name string) (Component, error) {
	for _, item := range c.GetComponents() {
		if item.Name == name {
			return item, nil
		}
	}
	return Component{}, fmt.Errorf("component [%s] does not exist", name)
}

// StartAll start all component
func (c ComponentMgr) StartAll() error {
	if err := util.CheckNecessaryCommands(); err != nil {
		fmt.Println(err)
		return errors.New("check necessary commands failed")
	}

	if err := checkK8sProc(); err != nil {
		fmt.Println(err)
		hwlog.RunLog.Error(err.Error())
		return err
	}
	if err := SetNodeIPToEdgeCore(); err != nil {
		hwlog.RunLog.Warnf("set nodeIP to edge core config file failed: %v, will use default nodeIP", err)
	}
	if err := c.makeSureCerts(); err != nil {
		return errors.New("make sure certs exist failed")
	}
	if err := c.changeDocker(); err != nil {
		return errors.New("change docker isolation failed")
	}
	if err := c.RegisterAllServices(); err != nil {
		hwlog.RunLog.Errorf("register all service failed, error: %v", err)
		return errors.New("register all service failed")
	}
	if err := util.StartService(constants.MefEdgeTargetFile); err != nil {
		hwlog.RunLog.Errorf("start target [%s] failed, error: %v", constants.MefEdgeTargetFile, err)
		return fmt.Errorf("start target [%s] failed", constants.MefEdgeTargetFile)
	}
	c.CheckAllServiceActive()
	hwlog.RunLog.Info("start all services success")
	return nil
}

// StopAll stop all component
func (c ComponentMgr) StopAll() error {
	stopFailed := false
	infos := c.GetComponents()
	for _, c := range infos {
		if !c.IsExist() {
			continue
		}
		if err := c.Stop(); err != nil {
			fmt.Printf("stop component [%s] failed, error: %v.\n", c.Name, err)
			hwlog.RunLog.Errorf("stop component [%s] failed, error: %v", c.Name, err)
			stopFailed = true
		}
	}
	if err := RemoveLimitPortRule(); err != nil {
		return fmt.Errorf("delete port limit rule failed when stop components, %v", err)
	}
	if stopFailed {
		return fmt.Errorf("stop components failed")
	}
	hwlog.RunLog.Info("stop all services success")
	return nil
}

func checkK8sProc() error {
	processes, err := util.GetProcesses()
	if err != nil {
		return fmt.Errorf("get running processes on device failed: %v", err)
	}
	for _, process := range processes {
		procName, err := util.GetProcName(process)
		if err != nil {
			continue
		}
		if procName == "kubelet" && util.CheckProcUser(process, constants.RootUserName) {
			return errors.New("kubelet should not running on edge node, please check and stop it")
		}
		if procName == "kube-proxy" && util.CheckProcUser(process, constants.RootUserName) {
			return errors.New("kube-proxy should not running on edge node, please check and stop it")
		}
	}
	return nil
}

// RemoveLimitPortRule remove port limit rule
func RemoveLimitPortRule() error {
	res, err := filepath.EvalSymlinks(constants.IptablesPath)
	if err != nil {
		hwlog.RunLog.Error("cannot get iptables command")
		return err
	}
	if _, err = envutils.RunCommand(res, envutils.DefCmdTimeoutSec, constants.Iptables,
		"-D", "INPUT", "-p", "tcp", "-j", constants.PortLimitIptablesRuleName); err != nil {
		hwlog.RunLog.Warnf("clean input iptables rule err: %v", err)
	}
	if _, err = envutils.RunCommand(res, envutils.DefCmdTimeoutSec, constants.Iptables,
		"-F", constants.PortLimitIptablesRuleName); err != nil {
		hwlog.RunLog.Warnf("clean port limit rule err: %v", err)
	}
	if _, err = envutils.RunCommand(res, envutils.DefCmdTimeoutSec, constants.Iptables,
		"-X", constants.PortLimitIptablesRuleName); err != nil {
		hwlog.RunLog.Warnf("delete port limit rule err: %v", err)
		return nil
	}
	hwlog.RunLog.Info("remove port limit rule success")
	return nil
}

// RestartAll restart all component
func (c ComponentMgr) RestartAll() error {
	if err := util.CheckNecessaryCommands(); err != nil {
		fmt.Println(err)
		return errors.New("check necessary commands failed")
	}

	if err := checkK8sProc(); err != nil {
		fmt.Println(err)
		hwlog.RunLog.Error(err.Error())
		return err
	}
	if err := SetNodeIPToEdgeCore(); err != nil {
		hwlog.RunLog.Warnf("set nodeIP to edge core config file failed: %v, will use default nodeIP", err)
	}
	if err := c.makeSureCerts(); err != nil {
		return errors.New("make sure certs exist failed")
	}
	if err := c.changeDocker(); err != nil {
		return errors.New("change docker isolation failed")
	}
	if err := c.RegisterAllServices(); err != nil {
		hwlog.RunLog.Errorf("register all service failed, error: %v", err)
		return errors.New("register all service failed")
	}
	if err := util.RestartService(constants.MefEdgeTargetFile); err != nil {
		hwlog.RunLog.Errorf("restart target [%s] failed, error: %v", constants.MefEdgeTargetFile, err)
		return fmt.Errorf("restart target [%s] failed", constants.MefEdgeTargetFile)
	}
	c.CheckAllServiceActive()
	hwlog.RunLog.Info("restart all services success")
	return nil
}

func (c ComponentMgr) makeSureCerts() error {
	installRootDir, err := path.GetInstallRootDir()
	if err != nil {
		hwlog.RunLog.Errorf("get install root dir failed, error: %v", err)
		return errors.New("get install root dir failed")
	}
	generateCertsTask, err := NewGenerateCertsTask(installRootDir)
	if err != nil {
		hwlog.RunLog.Errorf("get generate certs task failed, error: %v", err)
		return err
	}
	if err = generateCertsTask.MakeSureEdgeCerts(); err != nil {
		hwlog.RunLog.Errorf("regenerate certs failed, error: %v", err)
		return err
	}
	return nil
}

// RegisterAllServices register all component
func (c ComponentMgr) RegisterAllServices() error {
	targetPath := getServicePath(c.workPathMgr.GetServicePath(constants.MefEdgeTargetFile))
	if err := registerTarget(targetPath); err != nil {
		hwlog.RunLog.Errorf("register target failed, error: %v", err)
		return errors.New("register target failed")
	}
	infos := c.GetComponents()
	for _, c := range infos {
		if err := c.RegisterService(); err != nil {
			hwlog.RunLog.Errorf("register service [%s] failed, error: %v", c.Service.Name, err)
			return fmt.Errorf("register service [%s] failed", c.Service.Name)
		}
	}
	hwlog.RunLog.Info("register all services success")
	return nil
}

// UnregisterAllServices register all component
func (c ComponentMgr) UnregisterAllServices() error {
	if err := util.ResetFailedService(); err != nil {
		return fmt.Errorf("system service reset-failed failed, error: %v", err)
	}
	infos := c.GetComponents()
	for _, c := range infos {
		if err := c.UnregisterService(); err != nil {
			return fmt.Errorf("unregister service [%s] failed, error: %v", c.Service.Name, err)
		}
	}
	if err := unregisterTarget(); err != nil {
		return fmt.Errorf("unregister target failed, error: %v", err)
	}
	hwlog.RunLog.Info("unregister all services success")
	return nil
}

// UpdateServiceFiles update all service file values
func (c ComponentMgr) UpdateServiceFiles(logDir, logBackupDir string) error {
	infos := c.GetComponents()
	workDir := c.workPathMgr.GetWorkDir()
	workAbsDir, err := fileutils.EvalSymlinks(workDir)
	if err != nil {
		return fmt.Errorf("get work [%s] abs dir failed, error: %v", workDir, err)
	}
	dic := map[string]string{
		constants.InstallEdgeDir:     c.workPathMgr.GetMefEdgeDir(),
		constants.LogEdgeDir:         logDir,
		constants.LogBackupDirName:   logBackupDir,
		constants.InstallSoftWareDir: workAbsDir,
	}

	for _, c := range infos {
		if err = util.ReplaceValueInService(c.Service.Path, c.Service.ModeUmask, c.Service.UserName, dic); err != nil {
			hwlog.RunLog.Errorf("replace marks in service file [%s] failed, error: %v", c.Service.Path, err)
			return fmt.Errorf("replace marks in service file [%s] failed", c.Service.Path)
		}
	}
	hwlog.RunLog.Info("update service files success")
	return nil
}

// GetComponents get all components
func (c ComponentMgr) GetComponents() []Component {
	return []Component{
		c.GetEdgeInit(),
		c.GetEdgeOm(),
		c.GetEdgeMain(),
		c.GetEdgeCore(),
		c.GetDevicePlugin(),
	}
}

// GetEdgeInit get component edge-init
func (c ComponentMgr) GetEdgeInit() Component {
	componentDir := getComponentDir(c.workPathMgr.GetCompWorkDir(constants.EdgeInstaller))
	binPath := c.workPathMgr.GetCompBinaryPath(constants.EdgeInstaller, constants.MefInitScriptName)
	servicePath := getServicePath(c.workPathMgr.GetServicePath(constants.EdgeInitServiceFile))
	return Component{
		Name:    constants.MefInitServiceName,
		Dir:     componentDir,
		Service: newFileInfo(constants.EdgeInitServiceFile, servicePath, constants.ModeUmask077, constants.RootUserName),
		Bin:     newFileInfo(constants.MefInitScriptName, binPath, constants.ModeUmask077, constants.RootUserName),
	}
}

// GetEdgeMain get component edge-main
func (c ComponentMgr) GetEdgeMain() Component {
	componentDir := getComponentDir(c.workPathMgr.GetCompWorkDir(constants.EdgeMain))
	binPath := c.workPathMgr.GetCompBinaryPath(constants.EdgeMain, constants.EdgeMainFileName)
	servicePath := getServicePath(c.workPathMgr.GetServicePath(constants.EdgeMainServiceFile))
	return Component{
		Name:    constants.EdgeMainFileName,
		Dir:     componentDir,
		Service: newFileInfo(constants.EdgeMainServiceFile, servicePath, constants.ModeUmask077, constants.RootUserName),
		Bin:     newFileInfo(constants.EdgeMainFileName, binPath, constants.ModeUmask077, constants.EdgeUserName),
	}
}

// GetEdgeCore get component edge core
func (c ComponentMgr) GetEdgeCore() Component {
	componentDir := getComponentDir(c.workPathMgr.GetCompWorkDir(constants.EdgeCore))
	binPath := c.workPathMgr.GetCompBinaryPath(constants.EdgeCore, constants.EdgeCoreFileName)
	servicePath := getServicePath(c.workPathMgr.GetServicePath(constants.EdgeCoreServiceFile))
	return Component{
		Name:    constants.EdgeCoreFileName,
		Dir:     componentDir,
		Service: newFileInfo(constants.EdgeCoreServiceFile, servicePath, constants.ModeUmask077, constants.RootUserName),
		Bin:     newFileInfo(constants.EdgeCoreFileName, binPath, constants.ModeUmask077, constants.RootUserName),
	}
}

// GetDevicePlugin get component device-plugin
func (c ComponentMgr) GetDevicePlugin() Component {
	componentDir := getComponentDir(c.workPathMgr.GetCompWorkDir(constants.DevicePlugin))
	binPath := c.workPathMgr.GetCompBinaryPath(constants.DevicePlugin, constants.DevicePluginFileName)
	servicePath := getServicePath(c.workPathMgr.GetServicePath(constants.DevicePluginServiceFile))
	return Component{
		Name:    constants.DevicePluginFileName,
		Dir:     componentDir,
		Service: newFileInfo(constants.DevicePluginServiceFile, servicePath, constants.ModeUmask077, constants.RootUserName),
		Bin:     newFileInfo(constants.DevicePluginFileName, binPath, constants.ModeUmask077, constants.RootUserName),
	}
}

// GetEdgeOm get component edge-om
func (c ComponentMgr) GetEdgeOm() Component {
	componentDir := getComponentDir(c.workPathMgr.GetCompWorkDir(constants.EdgeOm))
	binPath := c.workPathMgr.GetCompBinaryPath(constants.EdgeOm, constants.EdgeOmFileName)
	servicePath := getServicePath(c.workPathMgr.GetServicePath(constants.EdgeOmServiceFile))
	return Component{
		Name:    constants.EdgeOmFileName,
		Dir:     componentDir,
		Service: newFileInfo(constants.EdgeOmServiceFile, servicePath, constants.ModeUmask077, constants.RootUserName),
		Bin:     newFileInfo(constants.EdgeOmFileName, binPath, constants.ModeUmask077, constants.RootUserName),
	}
}

func newFileInfo(name, filePath string, umask uint32, userName string) FileInfo {
	return FileInfo{
		Name:      name,
		Path:      filePath,
		ModeUmask: umask,
		UserName:  userName,
	}
}

func getServicePath(svcPath string) string {
	absSvcPath, err := filepath.EvalSymlinks(svcPath)
	if err != nil {
		hwlog.RunLog.Errorf("get srvPath's abs path failed: %s", err.Error())
		return svcPath
	}
	return absSvcPath
}

func getComponentDir(compWorkDir string) string {
	absCompWorkDir, err := filepath.EvalSymlinks(compWorkDir)
	if err != nil {
		hwlog.RunLog.Errorf("get component path's abs path failed: %s", err.Error())
		return compWorkDir
	}
	return absCompWorkDir
}

func registerTarget(targetPath string) error {
	if err := util.CopyServiceFileToSystemd(targetPath, constants.ModeUmask077, constants.RootUserName); err != nil {
		hwlog.RunLog.Errorf("copy target file [%s] failed, error: %v", targetPath, err)
		return fmt.Errorf("copy target file [%s] failed", targetPath)
	}
	hwlog.RunLog.Infof("register target [%s] success", constants.MefEdgeTargetFile)
	return nil
}

func unregisterTarget() error {
	if !util.IsServiceInSystemd(constants.MefEdgeTargetFile) {
		return nil
	}
	if active := util.IsServiceActive(constants.MefEdgeTargetFile); active {
		if err := util.StopService(constants.MefEdgeTargetFile); err != nil {
			hwlog.RunLog.Errorf("stop target [%s] failed, error: %v", constants.MefEdgeTargetFile, err)
			return fmt.Errorf("stop target [%s] failed", constants.MefEdgeTargetFile)
		}
	}
	if err := util.RemoveServiceFileInSystemd(constants.MefEdgeTargetFile); err != nil {
		hwlog.RunLog.Errorf("remove target [%s] failed, error: %v", constants.MefEdgeTargetFile, err)
		return fmt.Errorf("remove target [%s] failed", constants.MefEdgeTargetFile)
	}
	hwlog.RunLog.Infof("unregister target [%s] success", constants.MefEdgeTargetFile)
	return nil
}
