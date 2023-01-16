// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
)

// InstallComponent is the struct for a single component's installation
type InstallComponent struct {
	Name     string
	Required bool
	version  string
}

// GetCompulsoryMap returns a map that contains all compulsory components
// the key is the component's name and value is a InstallComponent to control the component
func GetCompulsoryMap() map[string]*InstallComponent {
	return map[string]*InstallComponent{
		EdgeManagerName: {
			EdgeManagerName,
			true,
			"",
		},
		CertManagerName: {
			CertManagerName,
			true,
			"",
		},
		NginxManagerName: {
			NginxManagerName,
			true,
			"",
		},
	}
}

// GetOptionalMap returns a map that contains all optional components
// the key is the component's name and value is a InstallComponent to control the component
func GetOptionalMap() map[string]*InstallComponent {
	return map[string]*InstallComponent{
		SoftwareManagerName: {
			SoftwareManagerName,
			false,
			"",
		},
	}
}

// LoadAndSaveImage is used to build a docker image and save it to component's image dir
func (c *InstallComponent) LoadAndSaveImage(pathMgr *WorkPathAMgr) error {
	if !c.Required {
		return nil
	}

	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	dockerDealerIns := GetDockerDealer(c.Name, c.version)
	imageConfigPath := pathMgr.GetImageConfigPath(c.Name)
	imagePath := pathMgr.GetImagePath(c.Name)
	if err := c.loadImage(&dockerDealerIns, imageConfigPath); err != nil {
		return err
	}

	if err := c.saveImage(&dockerDealerIns, imagePath); err != nil {
		return err
	}

	return nil
}

func (c *InstallComponent) loadImage(dealer *DockerDealer, imageConfigPath string) error {
	hwlog.RunLog.Infof("start to build [%s] component's docker", c.Name)
	if dealer == nil {
		hwlog.RunLog.Error("pointer dealer is nil")
		return errors.New("pointer dealer is nil")
	}

	imageConfigAbsPath, err := filepath.EvalSymlinks(imageConfigPath)
	if err != nil {
		return fmt.Errorf("get absolute component path failed: %s", err.Error())
	}

	if !utils.IsExist(imageConfigAbsPath) {
		return fmt.Errorf("failed to build [%s] component's docker, the docker path does not exist", c.Name)
	}

	if err = dealer.LoadImage(imageConfigAbsPath); err != nil {
		return err
	}

	return nil
}

func (c *InstallComponent) saveImage(dealer *DockerDealer, savePath string) error {
	hwlog.RunLog.Infof("start to save [%s] component's image", c.Name)
	if dealer == nil {
		hwlog.RunLog.Error("pointer dealer is nil")
		return errors.New("pointer dealer is nil")
	}

	if !utils.IsExist(savePath) {
		return fmt.Errorf("failed to save [%s] component's image, the save path does not exist", c.Name)
	}

	if err := dealer.SaveImage(savePath); err != nil {
		return err
	}

	return nil
}

// ClearDockerFile is used to clear Dockerfile and its binary file after the docker has been build
func (c *InstallComponent) ClearDockerFile(pathMgr *WorkPathAMgr) error {
	if !c.Required {
		return nil
	}

	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	componentPath := pathMgr.GetDockerFilePath(c.Name)
	componentAbsPath, err := filepath.EvalSymlinks(componentPath)
	if err != nil {
		hwlog.RunLog.Errorf("get [%s]'s Dockerfile's abs path failed: %s", c.Name, err.Error())
		return fmt.Errorf("get [%s]'s Dockerfile's abs path failed", c.Name)
	}

	if err = common.DeleteAllFile(componentAbsPath); err != nil {
		hwlog.RunLog.Errorf("delete component [%s]'s Dockerfile failed: %s", c.Name, err.Error())
		return fmt.Errorf("delete component [%s]'s Dockerfile's failed", c.Name)
	}

	if c.Name == NginxManagerName {
		err = common.DeleteAllFile(pathMgr.GetNginxDirPath())
		if err != nil {
			hwlog.RunLog.Errorf("delete nginx's dir failed: %s", err.Error())
			return errors.New("delete nginx's dir failed")
		}
		return nil
	}

	err = common.DeleteAllFile(pathMgr.GetComponentBinaryPath(c.Name))
	if err != nil {
		hwlog.RunLog.Errorf("delete component [%s]'s binary file failed: %s", c.Name, err.Error())
		return fmt.Errorf("delete component [%s]'s binary file failed", c.Name)
	}
	return nil
}

// SetInstallOption is used to set a component's required status, which indicates if a component should be installed
func (c *InstallComponent) SetInstallOption(install bool) {
	c.Required = install
}

// SetVersion is used to set the version of a component, which is used as the version tag of the docker
func (c *InstallComponent) SetVersion() error {
	version := DockerTag
	c.version = version
	return nil
}

// PrepareComponentCertDir is used to create the cert dir for a single component
func (c *InstallComponent) PrepareComponentCertDir(rootPath string) error {
	if !c.Required {
		return nil
	}

	certPath := path.Join(rootPath, c.Name, CertsDir)
	if err := common.MakeSurePath(certPath); err != nil {
		return fmt.Errorf("create cert path [%s] failed", certPath)
	}

	return nil
}

// PrepareSingleComponentDir is used to create a single component's dir and copy its files into it
func (c *InstallComponent) PrepareSingleComponentDir(pathMgr *WorkPathAMgr) error {
	if !c.Required {
		return nil
	}

	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	if err := c.prepareComponentDir(pathMgr); err != nil {
		return err
	}

	if err := c.copyComponentFiles(pathMgr); err != nil {
		hwlog.RunLog.Errorf("copy component [%s] files failed: %v", c.Name, err.Error())
		return errors.New("copy component files failed")
	}

	return nil
}

func (c *InstallComponent) prepareComponentDir(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	imageConfigPath := pathMgr.GetImageConfigPath(c.Name)
	if err := common.MakeSurePath(imageConfigPath); err != nil {
		hwlog.RunLog.Errorf(
			"create component [%s] cert path [%s] failed: %s", c.Name, imageConfigPath, err.Error())
		return errors.New("create component cert path failed")
	}

	imagePath := pathMgr.GetImagePath(c.Name)
	if err := common.MakeSurePath(imagePath); err != nil {
		hwlog.RunLog.Errorf(
			"create component [%s] image path [%s] failed: %s", c.Name, imageConfigPath, err.Error())
		return errors.New("create component image path failed")
	}

	return nil
}

func (c *InstallComponent) copyComponentFiles(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		hwlog.RunLog.Errorf("copy %s's file dir since cannot get current dir: %s", c.Name, err.Error())
		return errors.New("copy component dir failed")
	}

	installRootDir := path.Dir(path.Dir(currentDir))
	componentPath := path.Join(installRootDir, c.Name) + "/."

	// copy InstallComponent files to InstallComponent directory
	filesDst := pathMgr.GetImageConfigPath(c.Name)

	if err = common.CopyDir(componentPath, filesDst); err != nil {
		hwlog.RunLog.Errorf("copy %s's dir failed: %s", c.Name, err.Error())
		return fmt.Errorf("copy component dir failed")
	}

	return nil
}

// PrepareComponentCert is used to prepare certs for a single component
// the key file is encrypted by kmc
// for nginx module, an addition server certs needs to be created for northern channel
func (c *InstallComponent) PrepareComponentCert(certMng *certutils.RootCertMgr, certPathMgr *ConfigPathMgr) error {
	if !c.Required {
		return nil
	}

	if certMng == nil {
		hwlog.RunLog.Error("pointer certMng is nil")
		return errors.New("pointer certMng is nil")
	}
	if certPathMgr == nil {
		hwlog.RunLog.Errorf("pointer certPathMgr is nil")
		return errors.New("pointer certPathMgr is nil")
	}

	componentsCertPath := certPathMgr.GetComponentCertPath(c.Name)
	componentPrivPath := certPathMgr.GetComponentKeyPath(c.Name)

	componentCert := certutils.SelfSignCert{
		RootCertMgr: certMng,
		SvcCertPath: componentsCertPath,
		SvcKeyPath:  componentPrivPath,
		CommonName:  c.Name,
	}

	if err := componentCert.CreateSignCert(); err != nil {
		hwlog.RunLog.Errorf("create component [%s] cert failed: %v", c.Name, err)
		return fmt.Errorf("create component [%s] cert failed", c.Name)
	}

	if c.Name != NginxManagerName {
		return nil
	}

	if err := c.prepareNginxServerCert(certMng, certPathMgr); err != nil {
		return err
	}
	return nil
}

func (c *InstallComponent) prepareNginxServerCert(certMng *certutils.RootCertMgr, certPathMgr *ConfigPathMgr) error {
	hwlog.RunLog.Infof("start to prepare nginx server cert")

	if certMng == nil {
		hwlog.RunLog.Error("pointer certMng is nil")
		return errors.New("pointer certMng is nil")
	}
	if certPathMgr == nil {
		hwlog.RunLog.Error("pointer certPathMgr is nil")
		return errors.New("pointer certPathMgr is nil")
	}

	componentsCertPath := certPathMgr.GetNginxServerCrtPath()
	componentPrivPath := certPathMgr.GetNginxServerKeyPath()

	componentCert := certutils.SelfSignCert{
		RootCertMgr: certMng,
		SvcCertPath: componentsCertPath,
		SvcKeyPath:  componentPrivPath,
		CommonName:  c.Name + NginxServerSuffix,
	}

	if err := componentCert.CreateSignCert(); err != nil {
		hwlog.RunLog.Errorf("create nginx server cert failed: %v", err)
		return errors.New("create nginx server cert failed")
	}

	hwlog.RunLog.Infof("prepare nginx server cert successful")
	return nil
}

// PrepareLogDir creates log dir for a single component and change owner into MEFCenter:MEFCenter
func (c *InstallComponent) PrepareLogDir(pathMgr *LogDirPathMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	logDir := pathMgr.GetComponentLogPath(c.Name)
	if err := common.MakeSurePath(logDir); err != nil {
		hwlog.RunLog.Errorf("prepare component [%s] Log Dir failed: %s", c.Name, err.Error())
		return fmt.Errorf("prepare component [%s] log dir failed", c.Name)
	}

	if err := os.Chown(logDir, MefCenterUid, MefCenterGid); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner failed, error: %s", logDir, err.Error())
		return errors.New("set run script path owner failed")
	}
	return nil
}

// PrepareLibDir creates lib dir for a single component and copy libs into it
func (c *InstallComponent) PrepareLibDir(libSrcPath string, pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	libDir := pathMgr.GetComponentLibPath(c.Name)
	if err := common.MakeSurePath(libDir); err != nil {
		hwlog.RunLog.Errorf("prepare component [%s] lib Dir failed: %s", c.Name, err.Error())
		return fmt.Errorf("prepare component [%s] lib dir failed", c.Name)
	}

	if err := os.Chown(libDir, MefCenterUid, MefCenterGid); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner failed, error: %s", libDir, err.Error())
		return fmt.Errorf("set path [%s] owner failed", libDir)
	}

	if err := common.CopyDir(libSrcPath, libDir); err != nil {
		hwlog.RunLog.Errorf("copy component [%s]'s lib dir failed, error: %v", c.Name, err.Error())
		return fmt.Errorf("copy lib component [%s]'s dir failed", c.Name)
	}

	return nil
}

// ClearLibDir deleted the lib dir for single component, which is used after docker has been build
func (c *InstallComponent) ClearLibDir(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	libDir := pathMgr.GetComponentLibPath(c.Name)
	absPath, err := filepath.EvalSymlinks(libDir)
	if err != nil {
		hwlog.RunLog.Errorf("get [%s]'s lib dir's abs path failed: %s", c.Name, err.Error())
		return fmt.Errorf("get [%s]'s lib dir's abs path failed", c.Name)
	}

	if err = common.DeleteAllFile(absPath); err != nil {
		hwlog.RunLog.Errorf("delete component [%s]'s lib dir failed: %s", c.Name, err.Error())
		return fmt.Errorf("delete component [%s]'s lib dir failed", c.Name)
	}

	return nil
}
