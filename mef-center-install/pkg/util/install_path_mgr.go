// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"fmt"
	"path"
	"path/filepath"
)

// InstallDirPathMgr is a struct that controls all dir/file path in installed pkg dir
// paths are distributed by the workPath and config path
// all dir/file path in installed pkg dir should be got by specified func in this struct
type InstallDirPathMgr struct {
	rootPath      string
	WorkPathMgr   *WorkPathMgr
	WorkPathAMgr  *WorkPathAMgr
	ConfigPathMgr *ConfigPathMgr
}

// GetRootPath returns the installation root path
func (idm *InstallDirPathMgr) GetRootPath() string {
	return idm.rootPath
}

// GetMefPath returns the MEF-Center dir path
func (idm *InstallDirPathMgr) GetMefPath() string {
	return path.Join(idm.rootPath, OutMefDirName)
}

// GetWorkAPath returns mef-center-A dir path
func (idm *InstallDirPathMgr) GetWorkAPath() string {
	return path.Join(idm.GetMefPath(), MefWorkA)
}

// GetWorkPath returns mef-center softlink path
func (idm *InstallDirPathMgr) GetWorkPath() string {
	return path.Join(idm.GetMefPath(), MefSoftLink)
}

// GetInstallPkgDir returns install_package dir path
func (idm *InstallDirPathMgr) GetInstallPkgDir() string {
	return path.Join(idm.GetMefPath(), InstallPackageDir)
}

// GetConfigPath returns mef-config dir path
func (idm *InstallDirPathMgr) GetConfigPath() string {
	return path.Join(idm.GetMefPath(), MefConfigDir)
}

// GetTargetWorkPath returns the target upgrade path, if the existing work dir is A, B would be returned and vice versa
func (idm *InstallDirPathMgr) GetTargetWorkPath() (string, error) {
	normPath := idm.GetWorkPath()
	realPath, err := filepath.EvalSymlinks(normPath)
	if err != nil {
		return "", fmt.Errorf("get realpath of work dir failed: %s", err.Error())
	}

	workAPath := idm.GetWorkAPath()
	if realPath == workAPath {
		return MefWorkB, nil
	}

	return MefWorkA, nil
}

// WorkPathMgr is a struct that controls all dir/file path in the mef-center softlink dir
// all dir/file path in the mef-center softlink dir should be got by specified func in this struct
type WorkPathMgr struct {
	workPath string
}

// GetRelativeLibDirPath returns the lib dir path in mef-center softlink
func (wpm *WorkPathMgr) GetRelativeLibDirPath() string {
	return path.Join(wpm.workPath, MefLibDir)
}

// GetRelativeKmcLibDirPath returns the kmc-lib dir path in mef-center softlink
func (wpm *WorkPathMgr) GetRelativeKmcLibDirPath() string {
	return path.Join(wpm.GetRelativeLibDirPath(), MefKmcLibDir)
}

// GetInstallParamJsonPath returns the imstall_param json path in mef-center softlink
func (wpm *WorkPathMgr) GetInstallParamJsonPath() string {
	return path.Join(wpm.workPath, InstallParamJson)
}

// GetRelativeImagesDirPath returns the image dir path in mef-center softlink
func (wpm *WorkPathMgr) GetRelativeImagesDirPath() string {
	return path.Join(wpm.workPath, ImagesDirName)
}

// GetRelativeImageConfigPath returns the image config dir path in mef-center softlink
func (wpm *WorkPathMgr) GetRelativeImageConfigPath(component string) string {
	return path.Join(wpm.GetRelativeImagesDirPath(), component, ImageConfigDir)
}

// GetRelativeYamlPath returns the relative component's yaml path in mef-center softlink
func (wpm *WorkPathMgr) GetRelativeYamlPath(component string) string {
	return path.Join(wpm.GetRelativeImageConfigPath(component), fmt.Sprintf("%s.yaml", component))
}

// WorkPathAMgr is a struct that controls all dir/file path in the mef-center-A dir
// all dir/file path in the mef-center-A dir should be got by specified func in this struct
type WorkPathAMgr struct {
	workAPath string
}

// GetWorkAPath returns the mef-center-A dir path
func (wam *WorkPathAMgr) GetWorkAPath() string {
	return wam.workAPath
}

// GetImagesDirPath returns the images dir path in work-A dir
func (wam *WorkPathAMgr) GetImagesDirPath() string {
	return path.Join(wam.workAPath, ImagesDirName)
}

// GetBinDirPath returns the bin dir path in work-A dir
func (wam *WorkPathAMgr) GetBinDirPath() string {
	return path.Join(wam.workAPath, MefBinDir)
}

// GetInstallerBinPath returns the installation binary path in work-A dir
func (wam *WorkPathAMgr) GetInstallerBinPath() string {
	return path.Join(wam.GetBinDirPath(), InstallBin)
}

// GetRunShPath returns the run.sh path in work-A dir
func (wam *WorkPathAMgr) GetRunShPath() string {
	return path.Join(wam.workAPath, MefRunScript)
}

// GetVersionXmlPath returns the version.xml path in work-A dir
func (wam *WorkPathAMgr) GetVersionXmlPath() string {
	return path.Join(wam.workAPath, VersionXml)
}

// GetWorkAKmcLibDirPath returns kmc-lib path in work-A dir
func (wam *WorkPathAMgr) GetWorkAKmcLibDirPath() string {
	return path.Join(wam.workAPath, MefLibDir, MefKmcLibDir)
}

// GetImageConfigPath returns the image-config path in work-A dir
func (wam *WorkPathAMgr) GetImageConfigPath(component string) string {
	return path.Join(wam.GetImagesDirPath(), component, ImageConfigDir)
}

// GetDockerFilePath returns single component's Dockerfile path by component's name in work-A dir
func (wam *WorkPathAMgr) GetDockerFilePath(component string) string {
	return path.Join(wam.GetImageConfigPath(component), DockerFileName)
}

// GetNginxDirPath returns the nginx dir path in nginx module in work-A dir
func (wam *WorkPathAMgr) GetNginxDirPath() string {
	return path.Join(wam.GetImageConfigPath(NginxManagerName), NginxDirName)
}

// GetComponentBinaryPath returns single component's binary path by component's name in work-A dir
func (wam *WorkPathAMgr) GetComponentBinaryPath(component string) string {
	return path.Join(wam.GetImageConfigPath(component), component)
}

// GetImagePath returns single component's image dir path by component's name in work-A dir
func (wam *WorkPathAMgr) GetImagePath(component string) string {
	return path.Join(wam.GetImagesDirPath(), component, ImageDir)
}

// GetComponentLibPath returns component's lib dir that would be deleted after docker build
func (wam *WorkPathAMgr) GetComponentLibPath(component string) string {
	return path.Join(wam.GetImageConfigPath(component), ComponentLibDir)
}

// ConfigPathMgr is a struct that controls all dir/file path in the mef-config dir
// all dir/file path in the mef-config dir should be got by specified func in this struct
type ConfigPathMgr struct {
	configPath string
}

// GetConfigPath returns the mef-config dir path
func (cpm *ConfigPathMgr) GetConfigPath() string {
	return cpm.configPath
}

// GetComponentConfigPath returns single component's config dir path by component's name
func (cpm *ConfigPathMgr) GetComponentConfigPath(component string) string {
	return path.Join(cpm.configPath, component)
}

// GetMefCertsDirPath returns single component's certs dir path by component's name
func (cpm *ConfigPathMgr) GetMefCertsDirPath(component string) string {
	return path.Join(cpm.GetComponentConfigPath(component), CertsDir)
}

// GetComponentCertPath returns single component's certs file path by component's name
func (cpm *ConfigPathMgr) GetComponentCertPath(component string) string {
	return path.Join(cpm.GetMefCertsDirPath(component), component+CertSuffix)
}

// GetNginxServerCrtPath returns nginx module's server cert file path
func (cpm *ConfigPathMgr) GetNginxServerCrtPath() string {
	return path.Join(cpm.GetMefCertsDirPath(NginxManagerName), NginxManagerName+NginxServerSuffix+CertSuffix)
}

// GetComponentKeyPath returns single component's certs key path by component's name
func (cpm *ConfigPathMgr) GetComponentKeyPath(component string) string {
	return path.Join(cpm.GetMefCertsDirPath(component), component+KeySuffix)
}

// GetNginxServerKeyPath returns nginx module's server key file path
func (cpm *ConfigPathMgr) GetNginxServerKeyPath() string {
	return path.Join(cpm.GetMefCertsDirPath(NginxManagerName), NginxManagerName+NginxServerSuffix+KeySuffix)
}

// GetRootCaDirPath returns the root ca dir path
func (cpm *ConfigPathMgr) GetRootCaDirPath() string {
	return path.Join(cpm.configPath, RootCaDir)
}

// GetRootCaCertDirPath returns the root ca cert dir path
func (cpm *ConfigPathMgr) GetRootCaCertDirPath() string {
	return path.Join(cpm.GetRootCaDirPath(), RootCaFileDir)
}

// GetRootCaKeyDirPath returns the root ca key dir path
func (cpm *ConfigPathMgr) GetRootCaKeyDirPath() string {
	return path.Join(cpm.GetRootCaDirPath(), RootCaKeyDir)
}

// GetRootCaCertPath returns the root ca cert file path
func (cpm *ConfigPathMgr) GetRootCaCertPath() string {
	return path.Join(cpm.GetRootCaCertDirPath(), RootCaFile)
}

// GetRootCaKeyPath returns the root ca key file path
func (cpm *ConfigPathMgr) GetRootCaKeyPath() string {
	return path.Join(cpm.GetRootCaKeyDirPath(), RootKeyFile)
}

// GetRootKmcDirPath returns the kmc dir path for root ca
func (cpm *ConfigPathMgr) GetRootKmcDirPath() string {
	return path.Join(cpm.GetRootCaDirPath(), KmcDir)
}

// GetRootMasterKmcPath returns the kmc master key file path for root ca
func (cpm *ConfigPathMgr) GetRootMasterKmcPath() string {
	return path.Join(cpm.GetRootKmcDirPath(), MasterKeyFile)
}

// GetRootBackKmcPath returns the kmc backup key file path for root ca
func (cpm *ConfigPathMgr) GetRootBackKmcPath() string {
	return path.Join(cpm.GetRootKmcDirPath(), BackUpKeyFile)
}

// GetComponentMasterKmcPath returns the kmc master key file path for single component by component's name
func (cpm *ConfigPathMgr) GetComponentMasterKmcPath(component string) string {
	return path.Join(cpm.GetComponentConfigPath(component), KmcDir, MasterKeyFile)
}

// GetComponentBackKmcPath returns the kmc backup key file path for single component by component's name
func (cpm *ConfigPathMgr) GetComponentBackKmcPath(component string) string {
	return path.Join(cpm.GetComponentConfigPath(component), KmcDir, BackUpKeyFile)
}

// InitInstallDirPathMgr returns the InstallDirPathMgr construct by the root path
// Each call to this func returns a distinct mgr value even if the root path is identical
func InitInstallDirPathMgr(rootPath string) *InstallDirPathMgr {
	mgrIns := InstallDirPathMgr{rootPath: rootPath}
	mgrIns.WorkPathMgr = &WorkPathMgr{workPath: mgrIns.GetWorkPath()}
	mgrIns.WorkPathAMgr = &WorkPathAMgr{workAPath: mgrIns.GetWorkAPath()}
	mgrIns.ConfigPathMgr = &ConfigPathMgr{configPath: mgrIns.GetConfigPath()}
	return &mgrIns
}
