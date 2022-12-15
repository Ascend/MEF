// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud installation
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/mef-center-install/pkg/install"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

var (
	installAll             bool
	installImageManager    bool
	installResourceManager bool
	installSoftwareManager bool
	certRootPath           string
	logRootPath            string
	installLogPath         string
	rootCAPath             string
)

type component struct {
	name        string
	required    bool
	version     string
	libRequired bool
}

var compulsoryComponents = map[string]*component{
	install.EdgeManagerName: {
		install.EdgeManagerName,
		true,
		"",
		true,
	},
	install.CertManagerName: {
		install.CertManagerName,
		true,
		"",
		true,
	},
	install.NginxManagerName: {
		install.NginxManagerName,
		// todo:构建包中暂不包含nginx组件
		false,
		"",
		false,
	},
}

var optionalComponents = map[string]*component{
	install.ImageManagerName: {
		install.ImageManagerName,
		false,
		"",
		false,
	},
	install.ResourceManagerName: {
		install.ResourceManagerName,
		false,
		"",
		false,
	},
	install.SoftwareManagerName: {
		install.SoftwareManagerName,
		false,
		"",
		false,
	},
}

func init() {
	flag.BoolVar(&installAll, install.AllInstallFlag, false, "Install all optional components")
	flag.BoolVar(&installImageManager, install.ImageManagerFlag, false, "Install image manager")
	flag.BoolVar(&installResourceManager, install.ResourceManagerFlag, false, "Install resource manager")
	flag.BoolVar(&installSoftwareManager, install.SoftwareManagerFlag, false, "Install software manager")
	flag.StringVar(&certRootPath, install.CertPathFlag, "/etc", "The path used to save certs")
	flag.StringVar(&logRootPath, install.LogPathFlag, "/var", "The path used to save logs")
}

func checkParam() error {
	var err error
	if certRootPath == "" || !utils.IsExist(certRootPath) {
		fmt.Printf("cert dir [%s] dose not exist", certRootPath)
		return errors.New("cert root path does not exit")
	}

	if logRootPath == "" || !utils.IsExist(logRootPath) {
		fmt.Printf("log dir [%s] dose not exist", logRootPath)
		return errors.New("log root path does not exit")
	}

	if certRootPath, err = utils.RealDirChecker(certRootPath, true, false); err != nil {
		fmt.Printf("check cert dir failed, error: %s", err.Error())
		return errors.New("check cert directory failed")
	}

	if logRootPath, err = utils.RealDirChecker(logRootPath, true, false); err != nil {
		fmt.Printf("check log dir failed, error: %s", err.Error())
		return errors.New("check log directory failed")
	}

	return nil
}

func checkInstallUser() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	if usr.Username != "root" {
		return fmt.Errorf("install failed: the install user must be root, can not be %s", usr.Username)
	}

	return nil
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func loadImage(imageName string, version string, path string) error {
	absPath, err := utils.CheckPath(path)
	if err != nil {
		return err
	}
	// imageName is fixed name.
	// version is read from file or filename, the verification will be added in setVersion.
	// absPath has been verified
	cmdStr := "docker build -t" + imageName + ":" + version + " " + absPath + "/."
	cmd := exec.Command("sh", "-c", cmdStr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("load docker image failed:%s", err)
	}

	return nil
}

func (c *component) install() error {

	if !c.required {
		return nil
	}

	hwlog.RunLog.Infof("start to install [%s] component", c.name)
	componentPath := path.Join(install.MefComponentWorkPath, c.name)

	if !utils.IsExist(componentPath) {
		return fmt.Errorf("failed to to install %s component, the install package does not exist", c.name)
	}

	// todo: 规范中要求避免使用OS命令解析器,考虑改调用docker构建镜像。
	if err := loadImage(c.name, c.version, componentPath); err != nil {
		return fmt.Errorf("failed to to install %s component: %v", c.name, err)
	}

	hwlog.RunLog.Infof("install [%s] component successfully", c.name)
	return nil
}

func (c *component) setInstallOption(install bool) {
	c.required = install
}

func (c *component) setVersion() error {
	// todo: 从外部文件中或者目录名中获取组件版本号，这里需要添加白名单校验，后续会用到exec命令执行。
	version := "t2.0.4"
	c.version = version
	return nil
}

func (c *component) prepareComponentDir(rootPath string) error {
	if !c.required {
		return nil
	}

	certPath := path.Join(rootPath, c.name) + "/"
	if err := utils.MakeSureDir(certPath); err != nil {
		return fmt.Errorf("create cert path [%s] failed", certPath)
	}

	return nil
}

func (c *component) copyComponentFiles() error {
	if !c.required {
		return nil
	}

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("failed to to copy %s component, get current path failed", c.name)
	}

	installRootDir := path.Dir(currentDir)

	// todo: 组件安装包名称匹配,名称+版本+架构,后续根据构建出来的安装包修改。
	componentName := "Ascend-mindx_edge-" + c.name + "_3.0.0_linux-aarch64"
	componentPath := path.Join(installRootDir, componentName)
	libPath := path.Join(installRootDir, install.MefLibsDir)

	// copy component files to component directory
	filesDst := path.Join(install.MefComponentWorkPath, c.name)

	if err := utils.CopyDir(componentPath, filesDst); err != nil {
		return fmt.Errorf("copy component files failed, error: %v", err.Error())
	}

	if c.libRequired {
		// todo: 实现支持包含软链接的目录拷贝
		if _, err := util.RunCommand(util.CommandCopy, "-r", libPath, filesDst); err != nil {
			return fmt.Errorf("copy libs to component failed: %s", err.Error())
		}
		return nil

	}

	return nil
}

func (c *component) prepareComponentCert(certMng *certutils.RootCertMgr) error {
	if !c.required {
		return nil
	}
	certPath := path.Join(path.Join(certRootPath, install.CertsDir), c.name)
	componentsCertPath := path.Join(certPath, c.name+".crt")
	componentPrivPath := path.Join(certPath, c.name+".key")

	componentCert := certutils.SelfSignCert{
		RootCertMgr: certMng,
		SvcCertPath: componentsCertPath,
		SvcKeyPath:  componentPrivPath,
		CommonName:  c.name,
	}

	if err := componentCert.CreateSignCert(); err != nil {
		return fmt.Errorf("create component [%s] cert failed: %v", c.name, err)
	}
	return nil
}

func setComponentVersion() error {
	for _, component := range optionalComponents {
		if err := (*component).setVersion(); err != nil {
			return err
		}
	}
	for _, component := range compulsoryComponents {
		if err := (*component).setVersion(); err != nil {
			return err
		}
	}
	return nil
}

func paramParse() {
	if installAll {
		for _, component := range optionalComponents {
			component.setInstallOption(true)
		}
	}
	if isFlagSet(install.SoftwareManagerFlag) {
		optionalComponents["software-manager"].required = installSoftwareManager
	}
	if isFlagSet(install.ImageManagerFlag) {
		optionalComponents["image-manager"].required = installImageManager
	}
	if isFlagSet(install.ResourceManagerFlag) {
		optionalComponents["resource-manager"].required = installResourceManager
	}
}

func installComponents() error {
	hwlog.RunLog.Info("start to install compulsory components")
	for _, component := range compulsoryComponents {
		if err := (*component).install(); err != nil {
			return fmt.Errorf("install compulsory component [%s] failed: %v", component.name, err.Error())
		}
	}
	hwlog.RunLog.Info("install compulsory components successfully")

	hwlog.RunLog.Info("start to install optional components")
	for _, component := range optionalComponents {
		if err := (*component).install(); err != nil {
			return fmt.Errorf("install optional component [%s] failed: %v", component.name, err.Error())
		}
	}
	hwlog.RunLog.Info("install optional components successfully")
	return nil
}

func prepareCertsDir() error {
	hwlog.RunLog.Info("start to prepare component certs directories")
	certPath := path.Join(certRootPath, install.CertsDir) + "/"
	if err := utils.MakeSureDir(certPath); err != nil {
		return fmt.Errorf("create cert path [%s] failed: %v", certPath, err.Error())
	}
	// prepare RootCA directory
	rootCAPath = path.Join(certPath, install.RootCaDir) + "/"
	if err := utils.MakeSureDir(rootCAPath); err != nil {
		return fmt.Errorf("create root certs path [%s] failed: %v", certPath, err.Error())
	}
	// prepare compulsory component certs directory
	for _, component := range compulsoryComponents {
		if err := (*component).prepareComponentDir(certPath); err != nil {
			return fmt.Errorf("prepare compulsory component failed: %v", err.Error())
		}
	}
	// prepare optional component certs directory
	for _, component := range optionalComponents {
		if err := (*component).prepareComponentDir(certPath); err != nil {
			return fmt.Errorf("prepare optional component failed: %v", err.Error())
		}
	}
	hwlog.RunLog.Info("prepare component certs directories successfully")
	return nil
}

func prepareCA() (*certutils.RootCertMgr, error) {
	rootCaFilePath := path.Join(rootCAPath, install.RootCaFile)
	rootPrivFilePath := path.Join(rootCAPath, install.RootKeyFile)
	initCertMgr := certutils.InitRootCertMgr(rootCaFilePath, rootPrivFilePath, install.CaCommonName, nil)
	if _, err := initCertMgr.NewRootCa(); err != nil {
		return nil, fmt.Errorf("init root ca info failed: %v", err)
	}
	return initCertMgr, nil
}

func prepareCerts() error {
	hwlog.RunLog.Info("start to prepare component certs")
	// prepare root ca
	certMng, err := prepareCA()
	if err != nil {
		return err
	}
	// prepare compulsory component certs
	for _, component := range compulsoryComponents {
		if err := (*component).prepareComponentCert(certMng); err != nil {
			return err
		}
	}
	// prepare optional component certs
	for _, component := range optionalComponents {
		if err := (*component).prepareComponentCert(certMng); err != nil {
			return err
		}
	}
	if err := util.SetPathOwnerGroup(path.Join(certRootPath, install.CertsDir), install.HwMindXUserUid,
		install.HwMindXUserGid, true, false); err != nil {
		return err
	}

	hwlog.RunLog.Info("prepare component certs successfully")

	return nil
}

func prepareComponentWorkDir() error {
	hwlog.RunLog.Info("start to prepare component work directories")
	workPath := install.MefComponentWorkPath + "/"
	if err := utils.MakeSureDir(workPath); err != nil {
		return fmt.Errorf("create component root work path [%s] failed: %v", workPath, err.Error())
	}
	// prepare compulsory component working directory
	for _, component := range compulsoryComponents {
		if err := (*component).prepareComponentDir(install.MefComponentWorkPath); err != nil {
			return fmt.Errorf("create component [%s] work path failed: %v", workPath, err.Error())
		}
		if err := (*component).copyComponentFiles(); err != nil {
			return fmt.Errorf("copy component [%s] files failed: %v", workPath, err.Error())
		}
	}
	// prepare optional component working directory
	for _, component := range optionalComponents {
		if err := (*component).prepareComponentDir(install.MefComponentWorkPath); err != nil {
			return fmt.Errorf("create component [%s] work path failed: %v", workPath, err.Error())
		}
		if err := (*component).copyComponentFiles(); err != nil {
			return fmt.Errorf("copy component [%s] files failed: %v", workPath, err.Error())
		}
	}
	return nil

}

func prepareRootWorkDir() error {
	hwlog.RunLog.Info("start to prepare root work directories")
	// create mef working directory
	mefWorkPath := install.MefWorkPath + "/"
	if err := utils.MakeSureDir(mefWorkPath); err != nil {
		return fmt.Errorf("create mef root work path failed: %v", err.Error())
	}
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("prepare root work path [%s] failed, get current path failed", mefWorkPath)
	}
	// copy run.sh to working directory
	currentPath := path.Dir(currentDir)
	scriptSrc := path.Join(currentPath, install.MefScriptsDir)
	if err := utils.CopyDir(scriptSrc, mefWorkPath); err != nil {
		return fmt.Errorf("copy mef scripts dir failed, error: %v", err.Error())
	}
	runScripPath := path.Join(mefWorkPath, install.MefRunScript)
	if err := os.Chmod(runScripPath, install.ScriptMode); err != nil {
		hwlog.RunLog.Errorf("set path [%s] mode failed, error: %s", runScripPath, err.Error())
		return err
	}
	// create sbin directory and copy binaries to it
	sbinDst := path.Join(install.MefWorkPath, install.MefSbinDir) + "/"
	if err := utils.MakeSureDir(sbinDst); err != nil {
		return fmt.Errorf("create sbin work path failed: %v", err.Error())
	}
	sbinSrc := path.Join(currentPath, install.MefSbinDir)
	if err := utils.CopyDir(sbinSrc, sbinDst); err != nil {
		return fmt.Errorf("copy mef scripts dir failed, error: %v", err.Error())
	}

	return nil
}

func prepareSymlinks() error {
	// create cert symlink
	configSrc := path.Join(certRootPath, install.CertsDir)
	configDst := path.Join(install.MefWorkPath, install.MefWorkCertDir)
	if err := os.Symlink(configSrc, configDst); err != nil {
		return fmt.Errorf("create cert symlink failed, error: %s", err.Error())
	}
	// create log symlink
	configSrc = path.Join(logRootPath, install.LogDir)
	configDst = path.Join(install.MefWorkPath, install.MefWorkLogDir)
	if err := os.Symlink(configSrc, configDst); err != nil {
		return fmt.Errorf("create log symlink failed, error: %s", err.Error())
	}
	return nil
}

func prepareWorkingDir() error {

	if err := prepareRootWorkDir(); err != nil {
		return fmt.Errorf("failed to prepare root working directory: %v", err.Error())
	}
	if err := prepareComponentWorkDir(); err != nil {
		return fmt.Errorf("failed to prepare component working directoies: %v", err.Error())
	}
	if err := prepareSymlinks(); err != nil {
		return fmt.Errorf("failed to prepare cert and log symlink: %v", err.Error())
	}

	return nil
}

func prepareCertsAndDirs() error {
	// 6.1 create cert root path
	if err := prepareCertsDir(); err != nil {
		hwlog.RunLog.Errorf("failed to to prepare component certs directories: %v", err.Error())
		return err
	}
	// 6.2 create certs
	if err := prepareCerts(); err != nil {
		hwlog.RunLog.Errorf("failed to to prepare component certs: %v", err.Error())
		return err
	}
	return nil
}

func prepareBeforeInstall() error {
	// 5.1 set component version
	if err := setComponentVersion(); err != nil {
		hwlog.RunLog.Errorf("failed to set component version: %v", err.Error())
		return err
	}
	// 6. prepare certs
	hwlog.RunLog.Info("start to prepare component certs")
	if err := prepareCertsAndDirs(); err != nil {
		hwlog.RunLog.Errorf("failed to prepare certs: %v", err.Error())
		return err
	}
	// 7. prepare working directory
	if err := prepareWorkingDir(); err != nil {
		hwlog.RunLog.Errorf("failed to  working directory: %v", err.Error())
		return err
	}

	return nil
}

func doInstall() error {
	flag.Parse()
	// 1. verify the paths
	if err := checkParam(); err != nil {
		// install log has not initialized yet
		fmt.Println(err.Error())
		return err
	}
	// 2. create install log directory
	installLogPath = path.Join(logRootPath, install.LogDir) + "/"
	if err := utils.MakeSureDir(installLogPath); err != nil {
		// install log has not initialized yet
		fmt.Printf("create log path [%s] failed\n", installLogPath)
		return err
	}
	// 3. initialize the installation running and operating log
	if err := util.InitLogPath(installLogPath); err != nil {
		// install log has not initialized yet
		fmt.Println(err.Error())
		return err
	}
	hwlog.RunLog.Info("initialize install log successfully")
	// 4. user check
	if err := checkInstallUser(); err != nil {
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("check user success, start to install")
	// 5. parse the params
	paramParse()

	// 5 6 7 prepare before installing components
	if err := prepareBeforeInstall(); err != nil {
		hwlog.RunLog.Errorf("prepare before installing components failed: %v", err.Error())
		return err
	}
	// 8. start to install components
	if err := installComponents(); err != nil {
		hwlog.RunLog.Errorf("install components failed: %v", err.Error())
		return err
	}
	hwlog.RunLog.Info("install MEF Center successfully")
	return nil
}

func main() {
	if err := doInstall(); err != nil {
		hwlog.OpLog.Error("install MEF Center failed")
		os.Exit(1)
	}
	hwlog.OpLog.Info("install MEF Center successfully")
}
