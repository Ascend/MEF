// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package netconfig this file for config mef net manager
package netconfig

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/terminal"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/certmgr"
	"edge-installer/pkg/common/checker"
	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

// MefConfigFlow mef config flow
type MefConfigFlow struct {
	param    Param
	postFunc []func(*Param) error
}

// Param the parameters for net config
type Param struct {
	NetType       string
	Ip            string
	Port          int
	AuthPort      int
	RootCa        string
	TestConnect   bool
	ConfigPathMgr *pathmgr.ConfigPathMgr
}

type checkParamTask struct {
	ip            string
	port          int
	authPort      int
	token         string
	rootCa        string
	configPathMgr *pathmgr.ConfigPathMgr
}

type importRootCaTask struct {
	rootCa        string
	configPathMgr *pathmgr.ConfigPathMgr
}

type setConfigToDbTask struct {
	token         []byte
	netType       string
	ip            string
	port          int
	authPort      int
	rootCa        string
	testConnect   bool
	configPathMgr *pathmgr.ConfigPathMgr
}

type optUserId struct {
	uid uint32
	gid uint32
}

type importCaPath struct {
	tempCaPath        string
	tempCaBackupPath  string
	cloudCaCrlPath    string
	cloudCaPath       string
	cloudCaBackupPath string
}

// previous certs need to be deleted in netconfig mef mode
type mefDeleteCertsPath struct {
	edgeHubSvcCertPath string
	edgeHubSvcKeyPath  string
	mindXOMCaPath      string
}

// NewMefConfigFlow create mef config flow instance
func NewMefConfigFlow(param Param) *MefConfigFlow {
	mcf := &MefConfigFlow{param: param}
	mcf.postFunc = append(mcf.postFunc, func(pm *Param) error {
		return cleanUpTempCA(pm.ConfigPathMgr)
	})
	return mcf
}

func (cmf *MefConfigFlow) doPostProcess() {
	for _, f := range cmf.postFunc {
		if err := f(&cmf.param); err != nil {
			hwlog.RunLog.Errorf("do post process failed, error: %v", err)
		}
	}
}

// RunTasks run mef config task
func (cmf *MefConfigFlow) RunTasks() error {
	defer cmf.doPostProcess()

	checkParam := checkParamTask{
		ip:            cmf.param.Ip,
		port:          cmf.param.Port,
		authPort:      cmf.param.AuthPort,
		rootCa:        cmf.param.RootCa,
		configPathMgr: cmf.param.ConfigPathMgr,
	}
	if err := checkParam.runTask(); err != nil {
		hwlog.RunLog.Errorf("check MEF net config param failed, error: %v", err)
		return err
	}

	setConfig := setConfigToDbTask{
		netType:       cmf.param.NetType,
		ip:            cmf.param.Ip,
		port:          cmf.param.Port,
		authPort:      cmf.param.AuthPort,
		rootCa:        cmf.param.RootCa,
		testConnect:   cmf.param.TestConnect,
		configPathMgr: cmf.param.ConfigPathMgr,
	}
	if err := setConfig.runTask(); err != nil {
		hwlog.RunLog.Errorf("set MEF net config to database failed, error: %v", err)
		return err
	}

	importRootCa := importRootCaTask{
		configPathMgr: cmf.param.ConfigPathMgr,
	}
	if err := importRootCa.runTask(); err != nil {
		hwlog.RunLog.Errorf("import root ca failed, error: %v", err)
		return err
	}

	return nil
}

func (cpt *checkParamTask) runTask() error {
	var checkFunc = []func() error{
		cpt.checkParamIp,
		cpt.checkParamPort,
		cpt.checkParamRootCa,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (cpt *checkParamTask) checkParamIp() error {
	ip := net.ParseIP(cpt.ip)
	if cpt.ip == constants.IpZero || cpt.ip == constants.IpBroadcast || ip.To4() == nil {
		hwlog.RunLog.Error("check param ip failed, ip is invalid")
		return errors.New("param ip is invalid")
	}

	if err := cpt.checkLocalIp(); err != nil {
		hwlog.RunLog.Errorf("check param ip failed, error: %v", err)
		return err
	}

	hwlog.RunLog.Info("check param ip success")
	return nil
}

func (cpt *checkParamTask) checkLocalIp() error {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		hwlog.RunLog.Errorf("get local ip failed, error: %v", err)
		return errors.New("get local ip failed")
	}

	for _, address := range addresses {
		ipNet, ok := address.(*net.IPNet)
		if !ok {
			continue
		}
		if cpt.ip == ipNet.IP.String() {
			return errors.New("param ip is the same as local ip")
		}
	}

	return nil
}

func (cpt *checkParamTask) checkParamPort() error {
	if !checker.IntChecker(cpt.port, constants.MinPort, constants.MaxPort) {
		hwlog.RunLog.Errorf("check param port failed, port is out of range [%d, %d]",
			constants.MinPort, constants.MaxPort)
		return fmt.Errorf("param port is out of range [%d, %d]", constants.MinPort, constants.MaxPort)
	}
	if !checker.IntChecker(cpt.authPort, constants.MinPort, constants.MaxPort) {
		hwlog.RunLog.Errorf("check param auth_port failed, auth_port is out of range [%d, %d]",
			constants.MinPort, constants.MaxPort)
		return fmt.Errorf("param auth_port is out of range [%d, %d]", constants.MinPort, constants.MaxPort)
	}
	if cpt.port == cpt.authPort {
		hwlog.RunLog.Error("auth_port cannot equal to port")
		return errors.New("auth_port cannot equal to port")
	}
	hwlog.RunLog.Info("check param port and auth_port success")
	return nil
}

func (cpt *checkParamTask) checkParamRootCa() error {
	var checkFunc = []func() error{
		cpt.checkRootCaFile,
		cpt.saveRootCaToTmp,
		cpt.checkCaContent,
		cpt.showRootCaFingerprint,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			return err
		}
	}

	hwlog.RunLog.Info("check param root_ca success")
	return nil
}

func (cpt *checkParamTask) checkRootCaFile() error {
	if _, err := fileutils.RealFileCheck(cpt.rootCa, false, false, constants.MaxCertSize); err != nil {
		hwlog.RunLog.Errorf("check param root_ca failed, error: %v", err)
		return errors.New("param root_ca is invalid")
	}

	hwlog.RunLog.Info("check root ca file success")
	return nil
}

func (cpt *checkParamTask) saveRootCaToTmp() error {
	configDir := cpt.configPathMgr.GetConfigDir()
	realCfgDir, err := filepath.EvalSymlinks(configDir)
	if err != nil {
		hwlog.RunLog.Errorf("evaluate symlinks %s failed, error: %v", configDir, err)
		return errors.New("evaluate symlinks failed")
	}
	tempCertsDir := filepath.Join(realCfgDir, constants.NetCfgTempDirName)
	if err = fileutils.CreateDir(tempCertsDir, constants.Mode755); err != nil {
		hwlog.RunLog.Errorf("create temp cert dir failed, error: %v", err)
		return errors.New("create temp cert dir failed")
	}
	certMgr := certmgr.NewCertMgr(tempCertsDir, constants.RootCertName, constants.RootCertBackUpName)
	if err = certMgr.SaveCertByFile(cpt.rootCa, constants.Mode444); err != nil {
		hwlog.RunLog.Errorf("save root ca by file to %s failed, error: %v", constants.TmpCerts, err)
		return fmt.Errorf("save root ca to %s failed", constants.TmpCerts)
	}

	hwlog.RunLog.Infof("save root ca to %s success", constants.TmpCerts)
	return nil
}

func (cpt *checkParamTask) checkCaContent() error {
	if _, err := x509.CheckCertsChainReturnContent(cpt.rootCa); err != nil {
		hwlog.RunLog.Errorf("check importing cert failed, error: %s", err.Error())
		return errors.New("check importing cert failed")
	}

	hwlog.RunLog.Info("check ca content success")
	return nil
}

func (cpt *checkParamTask) showRootCaFingerprint() error {
	hash, err := fileutils.GetFileSha256(cpt.rootCa)
	if err != nil {
		hwlog.RunLog.Errorf("get file sha256 sum failed, error: %s", err.Error())
		return errors.New("get file sha256 sum failed")
	}
	fmt.Printf("the sha256sum of the importing cert file is: %s\n", hash)
	hwlog.RunLog.Infof("the sha256sum of the importing cert file is: %s", hash)

	hwlog.RunLog.Info("show root ca finger print success")
	return nil
}

func (sct *setConfigToDbTask) runTask() error {
	var setFunc = []func() error{
		sct.getToken,
		sct.testConnection,
		sct.setNetManagerToDb,
	}
	for _, function := range setFunc {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (sct *setConfigToDbTask) getToken() error {
	fmt.Println("Please enter token: ")

	token, err := terminal.ReadPasswordWithTimeout(common.StandardInput, common.MaxTokenLen, common.EnterTokenWaitTime)
	if err != nil {
		hwlog.RunLog.Error("get token failed")
		return errors.New("get token failed")
	}

	defer utils.ClearSliceByteMemory(token)
	if !checker.IntChecker(len(token), common.MinTokenLen, common.MaxTokenLen) {
		return errors.New("input token length invalid")
	}

	if err := utils.CheckPassWordComplexity(token); err != nil {
		return errors.New("token complex does not meet the requirement")
	}

	kmcDir := sct.configPathMgr.GetCompKmcDir(constants.EdgeOm)
	kmcCfg, err := util.GetKmcConfig(kmcDir)
	if err != nil {
		hwlog.RunLog.Errorf("get kmc config failed when encrypt token, error: %v", err)
		return err
	}

	encryptToken, err := kmc.EncryptContent(token, kmcCfg)
	if err != nil {
		hwlog.RunLog.Errorf("encrypt token failed, error: %v", err)
		return err
	}
	sct.token = encryptToken
	fmt.Println("get token success")
	hwlog.RunLog.Info("get token success")
	return nil
}

func (sct *setConfigToDbTask) testConnection() error {
	if !sct.testConnect {
		hwlog.RunLog.Info("skip MEF Center connection test")
		return nil
	}

	hwlog.RunLog.Info("start test connection to MEF center")
	url := fmt.Sprintf("https://%s:%d%s", sct.ip, sct.authPort, constants.MefCenterConnTestUrl)
	tlsCfg := certutils.TlsCertInfo{
		RootCaPath: sct.rootCa,
		RootCaOnly: true,
		WithBackup: false,
	}
	kmcDir := sct.configPathMgr.GetCompKmcDir(constants.EdgeOm)
	kmcCfg, err := util.GetKmcConfig(kmcDir)
	if err != nil {
		hwlog.RunLog.Errorf("get kmc config failed when decrypt token, error: %v", err)
		return err
	}

	decryptToken, err := kmc.DecryptContent(sct.token, kmcCfg)
	if err != nil {
		hwlog.RunLog.Errorf("decrypt token failed, error: %v", err)
		return err
	}
	defer utils.ClearSliceByteMemory(decryptToken)

	reqHeaders := map[string]interface{}{
		constants.Token: string(decryptToken),
	}
	defer func() {
		if str, ok := reqHeaders[constants.Token].(string); ok {
			utils.ClearStringMemory(str)
		}
	}()
	if _, err = httpsmgr.GetHttpsReq(url, tlsCfg, reqHeaders).Get(nil); err == nil {
		hwlog.RunLog.Info("test connection between MEF Edge and Center success")
		return nil
	}
	if strings.Contains(err.Error(), strconv.Itoa(http.StatusUnauthorized)) {
		hwlog.RunLog.Error("auth failed by center: token is incorrect")
	}
	if strings.Contains(err.Error(), strconv.Itoa(http.StatusLocked)) {
		hwlog.RunLog.Error("auth failed by center: ip is lock")
	}
	hwlog.RunLog.Errorf("auth failed by center: %v", err)
	return errors.New("test connection between MEF Edge and Center failed")
}

func (sct *setConfigToDbTask) setNetManagerToDb() error {
	netConfig := config.NetManager{
		NetType:  sct.netType,
		IP:       sct.ip,
		Port:     sct.port,
		AuthPort: sct.authPort,
		Token:    sct.token,
	}
	defer utils.ClearSliceByteMemory(netConfig.Token)
	defer utils.ClearSliceByteMemory(sct.token)

	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed, error: %v", err)
		return errors.New("get config path manager failed")
	}
	edgeOmCfgDir := configPathMgr.GetCompConfigDir(constants.EdgeOm)
	dbMgr := config.NewDbMgr(edgeOmCfgDir, constants.DbEdgeOmPath)
	if err = config.SetNetManager(dbMgr, &netConfig); err != nil {
		hwlog.RunLog.Errorf("set net manager to database failed, error: %v", err)
		return errors.New("set net manager to database failed")
	}

	hwlog.RunLog.Info("set net manager to database success")
	return nil
}

func (ict *importRootCaTask) runTask() error {
	var importFunction = []func() error{
		ict.backupPrevRootCa,
		ict.importRootCa,
		ict.removeMefInvalidCerts,
	}
	for _, function := range importFunction {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (ict *importRootCaTask) importRootCa() error {
	edgeUserId, err := envutils.GetUid(constants.EdgeUserName)
	if err != nil {
		hwlog.RunLog.Errorf("get edge user id failed, error: %v", err)
		return errors.New("get edge user id failed")
	}
	edgeGroupId, err := envutils.GetGid(constants.EdgeUserGroup)
	if err != nil {
		hwlog.RunLog.Errorf("get edge group id failed, error: %v", err)
		return errors.New("get edge group id failed")
	}

	optUser := &optUserId{uid: edgeUserId, gid: edgeGroupId}
	caPath, err := ict.getCaPath(optUser)
	if err != nil {
		hwlog.RunLog.Errorf("get root ca cert paths failed, error: %v", err)
		return errors.New("get root ca cert paths failed")
	}

	if err = ict.processRootCa(caPath, optUser); err != nil {
		hwlog.RunLog.Errorf("import root ca failed, error: %v", err)
		return errors.New("import root ca failed")
	}

	hwlog.RunLog.Info("import root ca success")
	return nil
}

// backup current root ca to xxx.pre
func (ict *importRootCaTask) backupPrevRootCa() error {
	currentCaPath := ict.configPathMgr.GetHubSvrRootCertPath()
	if !fileutils.IsExist(currentCaPath) {
		return nil
	}
	if err := fileutils.IsSoftLink(currentCaPath); err != nil {
		return errors.New("current root ca path is soft link")
	}
	operatorIdMgr := util.NewEdgeUGidMgr()
	if err := operatorIdMgr.SetEUGidToEdge(); err != nil {
		hwlog.RunLog.Errorf("set euid/egid to mef-edge failed: %v", err)
		return errors.New("set euid/egid to mef-edge failed")
	}
	defer func() {
		if err := operatorIdMgr.ResetEUGid(); err != nil {
			hwlog.RunLog.Errorf("reset euid/egid failed: %v", err)
		}
	}()
	prevBackupPath := ict.configPathMgr.GetHubSvrRootCertPrevBackupPath()
	// delete old backup file before copying
	if fileutils.IsExist(prevBackupPath) {
		if err := fileutils.IsSoftLink(prevBackupPath); err != nil {
			return errors.New("previous backup root ca path is soft link")
		}
		if err := fileutils.DeleteFile(prevBackupPath); err != nil {
			return errors.New("delete backup root ca failed")
		}
	}

	if err := fileutils.CopyFile(currentCaPath, prevBackupPath); err != nil {
		return fmt.Errorf("backup current root ca failed: %v", err)
	}

	if err := fileutils.SetPathPermission(prevBackupPath, constants.CertFileMode, false, false); err != nil {
		return fmt.Errorf("set prev backup root ca permission failed: %v", err)
	}
	hwlog.RunLog.Info("backup root ca success")
	return nil
}

func (ict *importRootCaTask) createEdgeCertsDir(edgeCertsDir string, uid, gid uint32) error {
	if err := fileutils.CreateDir(edgeCertsDir, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("create dest cert dir failed, error: %v", err)
		return errors.New("create dest cert dir failed")
	}
	param := fileutils.SetOwnerParam{
		Path:       edgeCertsDir,
		Uid:        uid,
		Gid:        gid,
		Recursive:  false,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set dest cert dir owner failed, error: %v", err)
		return errors.New("set dest cert dir owner failed")
	}
	return nil
}

func (ict *importRootCaTask) getCaPath(user *optUserId) (*importCaPath, error) {
	tmpRootCaPath := ict.configPathMgr.GetNetCfgTempRootCertPath()
	tmpRootCaBckPath := ict.configPathMgr.GetNetCfgTempRootCertBackupPath()
	errCaPath := fileutils.IsSoftLink(tmpRootCaPath)
	errCaPathBck := fileutils.IsSoftLink(tmpRootCaBckPath)
	if errCaPath != nil || errCaPathBck != nil {
		hwlog.RunLog.Errorf("temp root ca or backup root ca path is soft link: %v %v", errCaPath, errCaPathBck)
		return nil, errors.New("temp root ca or backup root ca path is soft link")
	}
	edgeCertsDir := ict.configPathMgr.GetHubSvrCertDir()
	if !fileutils.IsExist(edgeCertsDir) {
		if err := ict.createEdgeCertsDir(edgeCertsDir, user.uid, user.gid); err != nil {
			return nil, err
		}
	}
	destRootCaPath := ict.configPathMgr.GetHubSvrRootCertPath()
	destRootCaBckPath := ict.configPathMgr.GetHubSvrRootCertBackupPath()
	rootCaCrlPath := ict.configPathMgr.GetHubSvrCrlPath()
	if fileutils.IsExist(destRootCaPath) {
		if err := fileutils.IsSoftLink(destRootCaPath); err != nil {
			hwlog.RunLog.Error("dest cert path is soft link")
			return nil, errors.New("dest cert path is soft link")
		}
	}
	if fileutils.IsExist(destRootCaBckPath) {
		if err := fileutils.IsSoftLink(destRootCaBckPath); err != nil {
			hwlog.RunLog.Error("dest backup cert path is soft link")
			return nil, errors.New("dest backup cert path is soft link")
		}
	}
	if fileutils.IsExist(rootCaCrlPath) {
		if err := fileutils.IsSoftLink(rootCaCrlPath); err != nil {
			hwlog.RunLog.Error("dest cert crl path is soft link")
			return nil, errors.New("dest cert crl path is soft link")
		}
	}
	return &importCaPath{
		tempCaPath:        tmpRootCaPath,
		tempCaBackupPath:  tmpRootCaBckPath,
		cloudCaPath:       destRootCaPath,
		cloudCaBackupPath: destRootCaBckPath,
		cloudCaCrlPath:    rootCaCrlPath,
	}, nil
}

func (ict *importRootCaTask) processRootCa(certsPath *importCaPath, user *optUserId) error {
	if _, err := envutils.RunCommandWithUser(constants.RmCmd, envutils.DefCmdTimeoutSec, user.uid, user.gid,
		constants.ForceFlag, certsPath.cloudCaPath, certsPath.cloudCaBackupPath, certsPath.cloudCaCrlPath); err != nil {
		return fmt.Errorf("remove old root ca certs failed: %v", err)
	}

	if _, err := envutils.RunCommandWithUser(constants.CpCmd, envutils.DefCmdTimeoutSec, user.uid, user.gid,
		certsPath.tempCaPath, certsPath.cloudCaPath); err != nil {
		return fmt.Errorf("copy temp root ca to edge-main failed: %v", err)
	}

	if _, err := envutils.RunCommandWithUser(constants.CpCmd, envutils.DefCmdTimeoutSec, user.uid, user.gid,
		certsPath.tempCaBackupPath, certsPath.cloudCaBackupPath); err != nil {
		return fmt.Errorf("copy temp backup root ca to edge-main failed: %v", err)
	}

	if err := fileutils.SetPathPermission(certsPath.cloudCaPath, constants.CertFileMode, false, false); err != nil {
		return fmt.Errorf("set root ca permission failed: %v", err)
	}
	if err := fileutils.SetPathPermission(certsPath.cloudCaBackupPath, constants.Mode600, false, false); err != nil {
		return fmt.Errorf("set backup root ca permission failed: %v", err)
	}
	return nil
}

func getMefDeleteCertsPath(configPathMgr *pathmgr.ConfigPathMgr) *mefDeleteCertsPath {
	return &mefDeleteCertsPath{
		edgeHubSvcCertPath: configPathMgr.GetHubSvrCertPath(),
		edgeHubSvcKeyPath:  configPathMgr.GetHubSvrKeyPath(),
		mindXOMCaPath:      configPathMgr.GetOMRootCertPath(),
	}
}

func (ict *importRootCaTask) removeMefInvalidCerts() error {
	paths := getMefDeleteCertsPath(ict.configPathMgr)
	deletePath := []string{paths.edgeHubSvcCertPath, paths.edgeHubSvcKeyPath, paths.mindXOMCaPath,
		paths.mindXOMCaPath + backuputils.BackupSuffix}
	if err := removeFilesByMEFEdgeUser(deletePath); err != nil {
		hwlog.RunLog.Errorf("remove invalid certs failed, error: %v", err)
		return errors.New("remove invalid certs failed")
	}
	hwlog.RunLog.Info("remove invalid certs and key success")
	return nil
}

func cleanUpTempCA(configPathMgr *pathmgr.ConfigPathMgr) error {
	rawTempCertsDir := configPathMgr.GetNetConfigTempDir()
	tempCertsDir, err := filepath.EvalSymlinks(rawTempCertsDir)
	if err != nil {
		hwlog.RunLog.Errorf("evaluate symlinks %s failed: %v", tempCertsDir, err)
		return errors.New("evaluate symlinks failed")
	}
	if err = fileutils.DeleteAllFileWithConfusion(tempCertsDir); err != nil {
		hwlog.RunLog.Errorf("remove temp certs dir failed: %v", err)
		return errors.New("remove temp certs dir failed")
	}
	hwlog.RunLog.Info("clean up temp files success")
	return nil
}

func removeFilesByMEFEdgeUser(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	MefUid, errUid := envutils.GetUid(constants.EdgeUserName)
	MefGid, errGid := envutils.GetGid(constants.EdgeUserGroup)
	if errUid != nil || errGid != nil {
		hwlog.RunLog.Errorf("get MEFEdge uid or gid failed: %v %v", errUid, errGid)
		return errors.New("get MEFEdge uid or gid failed")
	}
	var failedPath []string
	for _, path := range paths {
		if !fileutils.IsExist(path) {
			continue
		}
		if err := fileutils.IsSoftLink(path); err != nil {
			return err
		}
		if _, err := envutils.RunCommandWithUser(constants.RmCmd, envutils.DefCmdTimeoutSec, MefUid, MefGid,
			constants.ForceFlag, path); err != nil {
			failedPath = append(failedPath, path)
			hwlog.RunLog.Errorf("path [%s] is deleted failed: %v", path, err)
		}
	}
	if len(failedPath) > 0 {
		return fmt.Errorf("the following paths are deleted failed: %v", failedPath)
	}
	return nil
}
