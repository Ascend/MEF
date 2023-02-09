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

// ComponentMgr is the struct for a single component's installation
type ComponentMgr struct {
	name string
}

// GetComponentMgr is the func to init a ComponentMgr struct
func GetComponentMgr(name string) *ComponentMgr {
	return &ComponentMgr{name: name}
}

// GetCompulsorySlice returns a slice that contains all compulsory components
func GetCompulsorySlice() []string {
	return []string{
		EdgeManagerName,
		CertManagerName,
		NginxManagerName,
	}
}

func getComponentDns(component string) string {
	DnsMap := map[string]string{
		EdgeManagerName:     common.EdgeMgrDns,
		CertManagerName:     common.CertMgrDns,
		SoftwareManagerName: common.SoftwareMgrDns,
		NginxManagerName:    common.NginxMgrDns,
	}
	return DnsMap[component]
}

// LoadAndSaveImage is used to build a docker image and save it to component's image dir
func (c *ComponentMgr) LoadAndSaveImage(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	dockerDealerIns := GetDockerDealer(c.name, DockerTag)
	imageConfigPath := pathMgr.GetImageConfigPath(c.name)
	imagePath := pathMgr.GetImagePath(c.name)
	if err := c.loadImage(&dockerDealerIns, imageConfigPath); err != nil {
		return err
	}

	if err := c.saveImage(&dockerDealerIns, imagePath); err != nil {
		return err
	}

	return nil
}

func (c *ComponentMgr) loadImage(dealer *DockerDealer, imageConfigPath string) error {
	hwlog.RunLog.Infof("start to build [%s] component's docker", c.name)
	if dealer == nil {
		hwlog.RunLog.Error("pointer dealer is nil")
		return errors.New("pointer dealer is nil")
	}

	imageConfigAbsPath, err := filepath.EvalSymlinks(imageConfigPath)
	if err != nil {
		return fmt.Errorf("get absolute component path failed: %s", err.Error())
	}

	if !utils.IsExist(imageConfigAbsPath) {
		return fmt.Errorf("failed to build [%s] component's docker, the docker path does not exist", c.name)
	}

	if err = dealer.LoadImage(imageConfigAbsPath); err != nil {
		return err
	}

	return nil
}

func (c *ComponentMgr) saveImage(dealer *DockerDealer, savePath string) error {
	hwlog.RunLog.Infof("start to save [%s] component's image", c.name)
	if dealer == nil {
		hwlog.RunLog.Error("pointer dealer is nil")
		return errors.New("pointer dealer is nil")
	}

	if !utils.IsExist(savePath) {
		return fmt.Errorf("failed to save [%s] component's image, the save path does not exist", c.name)
	}

	if err := dealer.SaveImage(savePath); err != nil {
		return err
	}

	return nil
}

// ClearDockerFile is used to clear Dockerfile and its binary file after the docker has been build
func (c *ComponentMgr) ClearDockerFile(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	componentPath := pathMgr.GetDockerFilePath(c.name)
	componentAbsPath, err := filepath.EvalSymlinks(componentPath)
	if err != nil {
		hwlog.RunLog.Errorf("get [%s]'s Dockerfile's abs path failed: %s", c.name, err.Error())
		return fmt.Errorf("get [%s]'s Dockerfile's abs path failed", c.name)
	}

	if err = common.DeleteAllFile(componentAbsPath); err != nil {
		hwlog.RunLog.Errorf("delete component [%s]'s Dockerfile failed: %s", c.name, err.Error())
		return fmt.Errorf("delete component [%s]'s Dockerfile's failed", c.name)
	}

	if c.name == NginxManagerName {
		err = common.DeleteAllFile(pathMgr.GetNginxDirPath())
		if err != nil {
			hwlog.RunLog.Errorf("delete nginx's dir failed: %s", err.Error())
			return errors.New("delete nginx's dir failed")
		}
		return nil
	}

	err = common.DeleteAllFile(pathMgr.GetComponentBinaryPath(c.name))
	if err != nil {
		hwlog.RunLog.Errorf("delete component [%s]'s binary file failed: %s", c.name, err.Error())
		return fmt.Errorf("delete component [%s]'s binary file failed", c.name)
	}
	return nil
}

// PrepareComponentCertDir is used to create the cert dir for a single component
func (c *ComponentMgr) PrepareComponentCertDir(rootPath string) error {
	certPath := path.Join(rootPath, c.name, CertsDir)
	if err := common.MakeSurePath(certPath); err != nil {
		return fmt.Errorf("create cert path [%s] failed", certPath)
	}

	return nil
}

// PrepareSingleComponentDir is used to create a single component's dir and copy its files into it
func (c *ComponentMgr) PrepareSingleComponentDir(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	if err := c.prepareComponentDir(pathMgr); err != nil {
		return err
	}

	if err := c.copyComponentFiles(pathMgr); err != nil {
		hwlog.RunLog.Errorf("copy component [%s] files failed: %v", c.name, err.Error())
		return errors.New("copy component files failed")
	}

	return nil
}

func (c *ComponentMgr) prepareComponentDir(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	imageConfigPath := pathMgr.GetImageConfigPath(c.name)
	if err := common.MakeSurePath(imageConfigPath); err != nil {
		hwlog.RunLog.Errorf(
			"create component [%s] cert path [%s] failed: %s", c.name, imageConfigPath, err.Error())
		return errors.New("create component cert path failed")
	}

	imagePath := pathMgr.GetImagePath(c.name)
	if err := common.MakeSurePath(imagePath); err != nil {
		hwlog.RunLog.Errorf(
			"create component [%s] image path [%s] failed: %s", c.name, imageConfigPath, err.Error())
		return errors.New("create component image path failed")
	}

	return nil
}

func (c *ComponentMgr) copyComponentFiles(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		hwlog.RunLog.Errorf("copy %s's file dir since cannot get current dir: %s", c.name, err.Error())
		return errors.New("copy component dir failed")
	}

	installRootDir := path.Dir(path.Dir(currentDir))
	componentPath := path.Join(installRootDir, c.name)

	// copy ComponentMgr files to ComponentMgr directory
	filesDst := pathMgr.GetImageConfigPath(c.name)

	if err = common.CopyDir(componentPath, filesDst, false); err != nil {
		hwlog.RunLog.Errorf("copy %s's dir failed: %s", c.name, err.Error())
		return fmt.Errorf("copy component dir failed")
	}

	return nil
}

// PrepareComponentCert is used to prepare certs for a single component
// the key file is encrypted by kmc
// for nginx module, an addition server certs needs to be created for northern channel
func (c *ComponentMgr) PrepareComponentCert(certMng *certutils.RootCertMgr, certPathMgr *ConfigPathMgr) error {
	if certMng == nil {
		hwlog.RunLog.Error("pointer certMng is nil")
		return errors.New("pointer certMng is nil")
	}
	if certPathMgr == nil {
		hwlog.RunLog.Errorf("pointer certPathMgr is nil")
		return errors.New("pointer certPathMgr is nil")
	}

	componentsCertPath := certPathMgr.GetComponentCertPath(c.name)
	componentPrivPath := certPathMgr.GetComponentKeyPath(c.name)

	componentCert := certutils.SelfSignCert{
		RootCertMgr: certMng,
		SvcCertPath: componentsCertPath,
		SvcKeyPath:  componentPrivPath,
		CommonName:  c.name,
		KmcCfg: &common.KmcCfg{
			SdpAlgID:       common.Aes256gcm,
			PrimaryKeyPath: certPathMgr.GetComponentMasterKmcPath(c.name),
			StandbyKeyPath: certPathMgr.GetComponentBackKmcPath(c.name),
			DoMainId:       common.DoMainId,
		},
		San: certutils.CertSan{
			DnsName: []string{getComponentDns(c.name)},
		},
	}
	if err := componentCert.CreateSignCert(); err != nil {
		hwlog.RunLog.Errorf("create component [%s] cert failed: %v", c.name, err)
		return fmt.Errorf("create component [%s] cert failed", c.name)
	}

	if c.name != NginxManagerName {
		return nil
	}

	if err := c.prepareUserMgrCert(certMng, certPathMgr); err != nil {
		return err
	}
	return nil
}

// prepareUserMgrCert create service crt for user manager module
func (c *ComponentMgr) prepareUserMgrCert(certMng *certutils.RootCertMgr, certPathMgr *ConfigPathMgr) error {
	hwlog.RunLog.Info("start to prepare user manager server cert")
	if certMng == nil {
		hwlog.RunLog.Error("pointer certMng is nil")
		return errors.New("pointer certMng is nil")
	}
	if certPathMgr == nil {
		hwlog.RunLog.Error("pointer certPathMgr is nil")
		return errors.New("pointer certPathMgr is nil")
	}
	componentsCertPath := certPathMgr.GetUserServerCrtPath()
	componentPrivPath := certPathMgr.GetUserServerKeyPath()
	componentCert := certutils.SelfSignCert{
		RootCertMgr: certMng,
		SvcCertPath: componentsCertPath,
		SvcKeyPath:  componentPrivPath,
		CommonName:  c.name,
		KmcCfg: &common.KmcCfg{
			SdpAlgID:       common.Aes256gcm,
			PrimaryKeyPath: certPathMgr.GetComponentMasterKmcPath(c.name),
			StandbyKeyPath: certPathMgr.GetComponentBackKmcPath(c.name),
			DoMainId:       common.DoMainId,
		},
		San: certutils.CertSan{
			DnsName: []string{getComponentDns(c.name)},
		},
	}
	if err := componentCert.CreateSignCert(); err != nil {
		hwlog.RunLog.Errorf("create user-manager server cert failed: %v", err)
		return errors.New("create user-manager server cert failed")
	}

	hwlog.RunLog.Infof("prepare user-manager server cert successful")
	return nil
}

// PrepareLogDir creates log dir for a single component and change owner into MEFCenter:MEFCenter
func (c *ComponentMgr) PrepareLogDir(pathMgr *LogDirPathMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	logDir := pathMgr.GetComponentLogPath(c.name)
	if err := common.MakeSurePath(logDir); err != nil {
		hwlog.RunLog.Errorf("prepare component [%s] Log Dir failed: %s", c.name, err.Error())
		return fmt.Errorf("prepare component [%s] log dir failed", c.name)
	}

	if _, err := utils.CheckPath(logDir); err != nil {
		hwlog.RunLog.Errorf("check component [%s] Log Dir failed: %s", c.name, err.Error())
		return fmt.Errorf("check component [%s] log dir failed", c.name)
	}

	mefUid, mefGid, err := GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get mef uid or gid failed: %s", err.Error())
		return errors.New("get mef uid or gid failed")
	}

	if err = os.Chown(logDir, mefUid, mefGid); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner failed, error: %s", logDir, err.Error())
		return errors.New("set run script path owner failed")
	}
	return nil
}

// PrepareLibDir creates lib dir for a single component and copy libs into it
func (c *ComponentMgr) PrepareLibDir(libSrcPath string, pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	libDir := pathMgr.GetComponentLibPath(c.name)
	if err := common.MakeSurePath(libDir); err != nil {
		hwlog.RunLog.Errorf("prepare component [%s] lib Dir failed: %s", c.name, err.Error())
		return fmt.Errorf("prepare component [%s] lib dir failed", c.name)
	}

	mefUid, mefGid, err := GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get mef uid or gid failed: %s", err.Error())
		return errors.New("get mef uid or gid failed")
	}

	if err = os.Chown(libDir, mefUid, mefGid); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner failed, error: %s", libDir, err.Error())
		return fmt.Errorf("set path [%s] owner failed", libDir)
	}

	if err = common.CopyDir(libSrcPath, libDir, false); err != nil {
		hwlog.RunLog.Errorf("copy component [%s]'s lib dir failed, error: %v", c.name, err.Error())
		return fmt.Errorf("copy lib component [%s]'s dir failed", c.name)
	}

	return nil
}

// ClearLibDir deleted the lib dir for single component, which is used after docker has been build
func (c *ComponentMgr) ClearLibDir(pathMgr *WorkPathAMgr) error {
	if pathMgr == nil {
		hwlog.RunLog.Error("pointer pathMgr is nil")
		return errors.New("pointer pathMgr is nil")
	}

	libDir := pathMgr.GetComponentLibPath(c.name)
	absPath, err := filepath.EvalSymlinks(libDir)
	if err != nil {
		hwlog.RunLog.Errorf("get [%s]'s lib dir's abs path failed: %s", c.name, err.Error())
		return fmt.Errorf("get [%s]'s lib dir's abs path failed", c.name)
	}

	if err = common.DeleteAllFile(absPath); err != nil {
		hwlog.RunLog.Errorf("delete component [%s]'s lib dir failed: %s", c.name, err.Error())
		return fmt.Errorf("delete component [%s]'s lib dir failed", c.name)
	}

	return nil
}
