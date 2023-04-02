// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"fmt"
	"path/filepath"
)

// WorkPathItf is an interface contains the path on mef-center-X dir
type WorkPathItf interface {
	GetWorkPath() string
	GetWorkLibDirPath() string
	GetRunShPath() string
	GetBinDirPath() string
	GetControllerBinPath() string
	GetVersionXmlPath() string
	GetImagesDirPath() string
	GetImageConfigPath(string) string
	GetImagePath(string) string
	GetDockerFilePath(string) string
	GetNginxDirPath() string
	GetComponentBinaryPath(string) string
	GetComponentLibPath(string) string
	GetInstallParamJsonPath() string
	GetComponentYamlPath(string) string
}

// InstallDirPathMgr is a struct that controls all dir/file path in installed pkg dir
// paths are distributed by the workPath and config path
// all dir/file path in installed pkg dir should be got by specified func in this struct
type InstallDirPathMgr struct {
	rootPath      string
	WorkPathMgr   *WorkPathMgr
	WorkPathAMgr  *WorkPathAMgr
	TmpPathMgr    *TmpUpgradeMgr
	ConfigPathMgr *ConfigPathMgr
}

// GetRootPath returns the installation root path
func (idm *InstallDirPathMgr) GetRootPath() string {
	return idm.rootPath
}

// GetMefPath returns the MEF-Center dir path
func (idm *InstallDirPathMgr) GetMefPath() string {
	return filepath.Join(idm.rootPath, OutMefDirName)
}

// GetWorkAPath returns mef-center-A dir path
func (idm *InstallDirPathMgr) GetWorkAPath() string {
	return filepath.Join(idm.GetMefPath(), MefWorkA)
}

// GetWorkBPath returns mef-center-B dir path
func (idm *InstallDirPathMgr) GetWorkBPath() string {
	return filepath.Join(idm.GetMefPath(), MefWorkB)
}

// GetTmpUpgradePath returns temp-upgrade dir path, which is a tmp dir on upgrade flow
func (idm *InstallDirPathMgr) GetTmpUpgradePath() string {
	return filepath.Join(idm.GetMefPath(), TempUpgradeDir)
}

// GetTmpCertsPath returns temp cert dir path, which is a tmp dir on exchange ca flow
func (idm *InstallDirPathMgr) GetTmpCertsPath() string {
	return filepath.Join(idm.GetMefPath(), TempCertDir)
}

// GetWorkPath returns mef-center softlink path
func (idm *InstallDirPathMgr) GetWorkPath() string {
	return filepath.Join(idm.GetMefPath(), MefSoftLink)
}

// GetInstallPkgDir returns install_package dir path
func (idm *InstallDirPathMgr) GetInstallPkgDir() string {
	return filepath.Join(idm.GetMefPath(), InstallPackageDir)
}

// GetConfigPath returns mef-config dir path
func (idm *InstallDirPathMgr) GetConfigPath() string {
	return filepath.Join(idm.GetMefPath(), MefConfigDir)
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
		return idm.GetWorkBPath(), nil
	}

	return idm.GetWorkAPath(), nil
}

// WorkPathMgr is a struct that controls all dir/file path in the mef-center softlink dir
// all dir/file path in the mef-center softlink dir should be got by specified func in this struct
type WorkPathMgr struct {
	workPath string
}

// GetWorkPath returns the work dir path in mef-center softlink
func (wpm *WorkPathMgr) GetWorkPath() string {
	return wpm.workPath
}

// GetWorkLibDirPath returns the lib dir path in mef-center softlink
func (wpm *WorkPathMgr) GetWorkLibDirPath() string {
	return filepath.Join(wpm.workPath, MefLibDir)
}

// GetVarDirPath returns the var dir path in mef-center softlink
// the var dir is a temporary path that used to storage temp files
func (wpm *WorkPathMgr) GetVarDirPath() string {
	return filepath.Join(wpm.workPath, MefVarDir)
}

// GetTempZipPath returns the path to unzip zip file in mef-center softlink
func (wpm *WorkPathMgr) GetTempZipPath() string {
	return filepath.Join(wpm.GetVarDirPath(), MefZipDir)
}

// GetTempTarPath returns the path to unzip tar file in mef-center softlink
func (wpm *WorkPathMgr) GetTempTarPath() string {
	return filepath.Join(wpm.GetVarDirPath(), MefTarDir)
}

// GetBinDirPath returns the bin dir path in mef-center softlink
func (wpm *WorkPathMgr) GetBinDirPath() string {
	return filepath.Join(wpm.workPath, MefBinDir)
}

// GetRunShPath returns the run.sh path in mef-center softlink
func (wpm *WorkPathMgr) GetRunShPath() string {
	return filepath.Join(wpm.workPath, MefRunScript)
}

// GetKmcLibDirPath returns the kmc-lib dir path in mef-center softlink
func (wpm *WorkPathMgr) GetKmcLibDirPath() string {
	return filepath.Join(wpm.GetWorkLibDirPath(), MefKmcLibDir)
}

// GetOtherLibDirPath returns the other lib dir path in mef-center softlink
func (wpm *WorkPathMgr) GetOtherLibDirPath() string {
	return filepath.Join(wpm.GetWorkLibDirPath(), OtherLibDir)
}

// GetInstallParamJsonPath returns the imstall_param json path in mef-center softlink
func (wpm *WorkPathMgr) GetInstallParamJsonPath() string {
	return filepath.Join(wpm.workPath, InstallParamJson)
}

// GetControllerBinPath returns the controller binary path in mef-center softlink
func (wpm *WorkPathMgr) GetControllerBinPath() string {
	return filepath.Join(wpm.GetBinDirPath(), ControllerBin)
}

// GetVersionXmlPath returns the version.xml path in mef-center softlink
func (wpm *WorkPathMgr) GetVersionXmlPath() string {
	return filepath.Join(wpm.workPath, VersionXml)
}

// GetImagesDirPath returns the image dir path in mef-center softlink
func (wpm *WorkPathMgr) GetImagesDirPath() string {
	return filepath.Join(wpm.workPath, ImagesDirName)
}

// GetImageConfigPath returns the image config dir path in mef-center softlink
func (wpm *WorkPathMgr) GetImageConfigPath(component string) string {
	return filepath.Join(wpm.GetImagesDirPath(), component, ImageConfigDir)
}

// GetImagePath returns single component's image dir path by component's name in mef-center softlink
func (wpm *WorkPathMgr) GetImagePath(component string) string {
	return filepath.Join(wpm.GetImagesDirPath(), component, ImageDir)
}

// GetDockerFilePath returns single component's Dockerfile path by component's name in mef-center softlink
func (wpm *WorkPathMgr) GetDockerFilePath(component string) string {
	return filepath.Join(wpm.GetImageConfigPath(component), DockerFileName)
}

// GetNginxDirPath returns the nginx dir path in nginx module in mef-center softlink
func (wpm *WorkPathMgr) GetNginxDirPath() string {
	return filepath.Join(wpm.GetImageConfigPath(NginxManagerName), NginxDirName)
}

// GetComponentBinaryPath returns single component's binary path by component's name in mef-center softlink
func (wpm *WorkPathMgr) GetComponentBinaryPath(component string) string {
	return filepath.Join(wpm.GetImageConfigPath(component), component)
}

// GetComponentLibPath returns component's lib dir that would be deleted after docker build
func (wpm *WorkPathMgr) GetComponentLibPath(component string) string {
	return filepath.Join(wpm.GetImageConfigPath(component), ComponentLibDir)
}

// GetComponentYamlPath returns the relative component's yaml path in mef-center softlink
func (wpm *WorkPathMgr) GetComponentYamlPath(component string) string {
	return filepath.Join(wpm.GetImageConfigPath(component), fmt.Sprintf("%s.yaml", component))
}

// GetImageDirPath returns the relative component's image dir in mef-center softlink by single component's name
func (wpm *WorkPathMgr) GetImageDirPath(component string) string {
	return filepath.Join(wpm.GetImagesDirPath(), component, ImageDir)
}

// GetUpgradeFlagPath returns the relative upgrade-flag path
func (wpm *WorkPathMgr) GetUpgradeFlagPath() string {
	return filepath.Join(wpm.workPath, UpgradeFlagFile)
}

// WorkPathAMgr is a struct that controls all dir/file path in the mef-center-A dir
// all dir/file path in the mef-center-A dir should be got by specified func in this struct
type WorkPathAMgr struct {
	workAPath string
}

// GetWorkPath returns the mef-center-A dir path
func (wam *WorkPathAMgr) GetWorkPath() string {
	return wam.workAPath
}

// GetImagesDirPath returns the images dir path in work-A dir
func (wam *WorkPathAMgr) GetImagesDirPath() string {
	return filepath.Join(wam.workAPath, ImagesDirName)
}

// GetBinDirPath returns the bin dir path in work-A dir
func (wam *WorkPathAMgr) GetBinDirPath() string {
	return filepath.Join(wam.workAPath, MefBinDir)
}

// GetInstallParamJsonPath returns the install-param.json path in mef-center-A Dir
func (wam *WorkPathAMgr) GetInstallParamJsonPath() string {
	return filepath.Join(wam.workAPath, InstallParamJson)
}

// GetControllerBinPath returns the controller binary path in work-A dir
func (wam *WorkPathAMgr) GetControllerBinPath() string {
	return filepath.Join(wam.GetBinDirPath(), ControllerBin)
}

// GetRunShPath returns the run.sh path in work-A dir
func (wam *WorkPathAMgr) GetRunShPath() string {
	return filepath.Join(wam.workAPath, MefRunScript)
}

// GetVersionXmlPath returns the version.xml path in work-A dir
func (wam *WorkPathAMgr) GetVersionXmlPath() string {
	return filepath.Join(wam.workAPath, VersionXml)
}

// GetWorkLibDirPath returns lib path in work-A dir
func (wam *WorkPathAMgr) GetWorkLibDirPath() string {
	return filepath.Join(wam.workAPath, MefLibDir)
}

// GetWorkKmcLibDirPath returns kmc-lib path in work-A dir
func (wam *WorkPathAMgr) GetWorkKmcLibDirPath() string {
	return filepath.Join(wam.GetWorkLibDirPath(), MefKmcLibDir)
}

// GetImageConfigPath returns the image-config path in work-A dir
func (wam *WorkPathAMgr) GetImageConfigPath(component string) string {
	return filepath.Join(wam.GetImagesDirPath(), component, ImageConfigDir)
}

// GetDockerFilePath returns single component's Dockerfile path by component's name in work-A dir
func (wam *WorkPathAMgr) GetDockerFilePath(component string) string {
	return filepath.Join(wam.GetImageConfigPath(component), DockerFileName)
}

// GetNginxDirPath returns the nginx dir path in nginx module in work-A dir
func (wam *WorkPathAMgr) GetNginxDirPath() string {
	return filepath.Join(wam.GetImageConfigPath(NginxManagerName), NginxDirName)
}

// GetComponentBinaryPath returns single component's binary path by component's name in work-A dir
func (wam *WorkPathAMgr) GetComponentBinaryPath(component string) string {
	return filepath.Join(wam.GetImageConfigPath(component), component)
}

// GetImagePath returns single component's image dir path by component's name in work-A dir
func (wam *WorkPathAMgr) GetImagePath(component string) string {
	return filepath.Join(wam.GetImagesDirPath(), component, ImageDir)
}

// GetComponentLibPath returns component's lib dir that would be deleted after docker build
func (wam *WorkPathAMgr) GetComponentLibPath(component string) string {
	return filepath.Join(wam.GetImageConfigPath(component), ComponentLibDir)
}

// GetComponentYamlPath returns component's yaml file path by component's name in work-A dir
func (wam *WorkPathAMgr) GetComponentYamlPath(component string) string {
	return filepath.Join(wam.GetImageConfigPath(component), fmt.Sprintf("%s.yaml", component))
}

// TmpUpgradeMgr is a struct that controls all dir/file path in the temp-upgrade dir
// all dir/file path in the temp-upgrade dir should be got by specified func in this struct
type TmpUpgradeMgr struct {
	tempUpgradePath string
}

// GetWorkPath returns the temp-upgrade dir path
func (tum *TmpUpgradeMgr) GetWorkPath() string {
	return tum.tempUpgradePath
}

// GetWorkLibDirPath returns lib path in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetWorkLibDirPath() string {
	return filepath.Join(tum.tempUpgradePath, MefLibDir)
}

// GetRunShPath returns the run.sh path in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetRunShPath() string {
	return filepath.Join(tum.tempUpgradePath, MefRunScript)
}

// GetBinDirPath returns the bin dir path in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetBinDirPath() string {
	return filepath.Join(tum.tempUpgradePath, MefBinDir)
}

// GetInstallParamJsonPath returns the install-param.json path in temp-upgrade Dir
func (tum *TmpUpgradeMgr) GetInstallParamJsonPath() string {
	return filepath.Join(tum.tempUpgradePath, InstallParamJson)
}

// GetControllerBinPath returns the controller binary path in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetControllerBinPath() string {
	return filepath.Join(tum.GetBinDirPath(), ControllerBin)
}

// GetVersionXmlPath returns the version.xml path in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetVersionXmlPath() string {
	return filepath.Join(tum.tempUpgradePath, VersionXml)
}

// GetImagesDirPath returns the images dir path in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetImagesDirPath() string {
	return filepath.Join(tum.tempUpgradePath, ImagesDirName)
}

// GetImageConfigPath returns the image-config path in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetImageConfigPath(component string) string {
	return filepath.Join(tum.GetImagesDirPath(), component, ImageConfigDir)
}

// GetImagePath returns single component's image dir path by component's name in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetImagePath(component string) string {
	return filepath.Join(tum.GetImagesDirPath(), component, ImageDir)
}

// GetDockerFilePath returns single component's Dockerfile path by component's name in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetDockerFilePath(component string) string {
	return filepath.Join(tum.GetImageConfigPath(component), DockerFileName)
}

// GetNginxDirPath returns the nginx dir path in nginx module in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetNginxDirPath() string {
	return filepath.Join(tum.GetImageConfigPath(NginxManagerName), NginxDirName)
}

// GetComponentBinaryPath returns single component's binary path by component's name in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetComponentBinaryPath(component string) string {
	return filepath.Join(tum.GetImageConfigPath(component), component)
}

// GetComponentLibPath returns component's lib dir that would be deleted after docker build
func (tum *TmpUpgradeMgr) GetComponentLibPath(component string) string {
	return filepath.Join(tum.GetImageConfigPath(component), ComponentLibDir)
}

// GetComponentYamlPath returns component's yaml file path by component's name in temp-upgrade dir
func (tum *TmpUpgradeMgr) GetComponentYamlPath(component string) string {
	return filepath.Join(tum.GetImageConfigPath(component), fmt.Sprintf("%s.yaml", component))
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
	return filepath.Join(cpm.configPath, component)
}

// GetCompPubConfigPath returns public config path dir
func (cpm *ConfigPathMgr) GetCompPubConfigPath() string {
	return filepath.Join(cpm.configPath, PubCfgDir)
}

// GetMefCertsDirPath returns single component's certs dir path by component's name
func (cpm *ConfigPathMgr) GetMefCertsDirPath(component string) string {
	return filepath.Join(cpm.GetComponentConfigPath(component), CertsDir)
}

// GetComponentCertPath returns single component's certs file path by component's name
func (cpm *ConfigPathMgr) GetComponentCertPath(component string) string {
	return filepath.Join(cpm.GetMefCertsDirPath(component), component+CertSuffix)
}

// GetUserServerCrtPath returns user module's server cert file path
func (cpm *ConfigPathMgr) GetUserServerCrtPath() string {
	return filepath.Join(cpm.GetMefCertsDirPath(NginxManagerName), UserManagerName+CertSuffix)
}

// GetComponentKeyPath returns single component's certs key path by component's name
func (cpm *ConfigPathMgr) GetComponentKeyPath(component string) string {
	return filepath.Join(cpm.GetMefCertsDirPath(component), component+KeySuffix)
}

// GetUserServerKeyPath returns user module's server key file path
func (cpm *ConfigPathMgr) GetUserServerKeyPath() string {
	return filepath.Join(cpm.GetMefCertsDirPath(NginxManagerName), UserManagerName+KeySuffix)
}

// GetNginxNorthernCertPath returns the cert path of the 3rd party
func (cpm *ConfigPathMgr) GetNginxNorthernCertPath() string {
	return filepath.Join(cpm.GetMefCertsDirPath(NginxManagerName), NorthernRootName+CertSuffix)
}

// GetPublicConfigPath returns the public-config path
func (cpm *ConfigPathMgr) GetPublicConfigPath() string {
	return filepath.Join(cpm.configPath, PubConfigDir)
}

// GetRootCaDirPath returns the root ca dir path
func (cpm *ConfigPathMgr) GetRootCaDirPath() string {
	return filepath.Join(cpm.configPath, RootCaDir)
}

// GetRootCaCertDirPath returns the root ca cert dir path
func (cpm *ConfigPathMgr) GetRootCaCertDirPath() string {
	return filepath.Join(cpm.GetRootCaDirPath(), RootCaFileDir)
}

// GetRootCaKeyDirPath returns the root ca key dir path
func (cpm *ConfigPathMgr) GetRootCaKeyDirPath() string {
	return filepath.Join(cpm.GetRootCaDirPath(), RootCaKeyDir)
}

// GetRootCaCertPath returns the root ca cert file path
func (cpm *ConfigPathMgr) GetRootCaCertPath() string {
	return filepath.Join(cpm.GetRootCaCertDirPath(), RootCaFile)
}

// GetRootCaKeyPath returns the root ca key file path
func (cpm *ConfigPathMgr) GetRootCaKeyPath() string {
	return filepath.Join(cpm.GetRootCaKeyDirPath(), RootKeyFile)
}

// GetRootKmcDirPath returns the kmc dir path for root ca
func (cpm *ConfigPathMgr) GetRootKmcDirPath() string {
	return filepath.Join(cpm.GetRootCaDirPath(), KmcDir)
}

// GetRootMasterKmcPath returns the kmc master key file path for root ca
func (cpm *ConfigPathMgr) GetRootMasterKmcPath() string {
	return filepath.Join(cpm.GetRootKmcDirPath(), MasterKeyFile)
}

// GetRootBackKmcPath returns the kmc backup key file path for root ca
func (cpm *ConfigPathMgr) GetRootBackKmcPath() string {
	return filepath.Join(cpm.GetRootKmcDirPath(), BackUpKeyFile)
}

// GetApigRootPath returns the root crt file path in apig dir
func (cpm *ConfigPathMgr) GetApigRootPath() string {
	return filepath.Join(cpm.GetComponentConfigPath(CertManagerName), RootCaDir, ApigDirName, RootCrtName)
}

// GetComponentKmcDirPath returns the kmc dir path for single component by component's name
func (cpm *ConfigPathMgr) GetComponentKmcDirPath(component string) string {
	return filepath.Join(cpm.GetComponentConfigPath(component), KmcDir)
}

// GetComponentMasterKmcPath returns the kmc master key file path for single component by component's name
func (cpm *ConfigPathMgr) GetComponentMasterKmcPath(component string) string {
	return filepath.Join(cpm.GetComponentKmcDirPath(component), MasterKeyFile)
}

// GetComponentBackKmcPath returns the kmc backup key file path for single component by component's name
func (cpm *ConfigPathMgr) GetComponentBackKmcPath(component string) string {
	return filepath.Join(cpm.GetComponentKmcDirPath(component), BackUpKeyFile)
}

// InitInstallDirPathMgr returns the InstallDirPathMgr construct by the root path
// Each call to this func returns a distinct mgr value even if the root path is identical
func InitInstallDirPathMgr(rootPath string) *InstallDirPathMgr {
	mgrIns := InstallDirPathMgr{rootPath: rootPath}
	mgrIns.WorkPathMgr = &WorkPathMgr{workPath: mgrIns.GetWorkPath()}
	mgrIns.WorkPathAMgr = &WorkPathAMgr{workAPath: mgrIns.GetWorkAPath()}
	mgrIns.TmpPathMgr = &TmpUpgradeMgr{tempUpgradePath: mgrIns.GetTmpUpgradePath()}
	mgrIns.ConfigPathMgr = &ConfigPathMgr{configPath: mgrIns.GetConfigPath()}
	return &mgrIns
}
