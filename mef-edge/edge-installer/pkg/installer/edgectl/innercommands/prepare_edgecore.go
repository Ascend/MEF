// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package innercommands

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
)

const (
	checkEdgecoreWaitTime = 5 * time.Second
	deletePipeInterval    = 3 * time.Second

	maxCoreStartTimes = 180
	writeTimes        = 2
	maxWriteTimes     = 10
	writePipeInterval = time.Millisecond * 500

	kubeletSocket = "/var/lib/kubelet/device-plugins/kubelet.sock"

	findEdgecorePortTimes = 5
)

// PrepareEdgecoreFlow is used to write edgecore info into pipe file
type PrepareEdgecoreFlow struct {
	pipePath        string
	pid             int
	establishedPort []int
	portExists      bool
	dbBackupCtx     context.Context
}

// NewPrepareEdgecore is the main func to init a WriteEdgecoreInfoTask struct
func NewPrepareEdgecore() *PrepareEdgecoreFlow {
	return &PrepareEdgecoreFlow{pipePath: constants.EdgeCorePipePath}
}

// Run starts to flow operation
func (pef *PrepareEdgecoreFlow) Run() error {
	var tasks = []func() error{
		pef.checkLocalHost,
		pef.checkDocker,
		pef.prepareConfig,
		pef.prepareDb,
		pef.createPipe,
		pef.startEdgecore,
		pef.writeInfo,
		pef.deletePipe,
		pef.addPortLimitRule,
		pef.minitorEdgecore,
	}

	defer func() {
		if err := common.RemoveLimitPortRule(); err != nil {
			hwlog.RunLog.Error("remove limit port rule error")
		}
	}()

	for _, function := range tasks {
		err := function()

		if err == nil {
			continue
		}

		if deleteErr := pef.deletePipeOnce(); deleteErr != nil {
			hwlog.RunLog.Errorf("delete pipe file failed: %s", deleteErr.Error())
		}

		return err
	}

	return nil
}

func (pef *PrepareEdgecoreFlow) checkLocalHost() error {
	localIp, err := net.ResolveIPAddr("ip", "localhost")
	if err != nil {
		hwlog.RunLog.Errorf("check localhost ip failed: %s", err.Error())
		return errors.New("check localhost ip failed")
	}

	if localIp.IP.String() != constants.LocalIp {
		hwlog.RunLog.Errorf("localhost ip is not %s", constants.LocalIp)
		return errors.New("localhost ip is incorrect")
	}

	return nil
}

func (pef *PrepareEdgecoreFlow) checkDocker() error {
	if _, err := envutils.RunCommand(constants.DockerCmd, envutils.DefCmdTimeoutSec, "ps"); err != nil {
		hwlog.RunLog.Error("docker status is abnormal, cannot start edgecore")
		return errors.New("docker status is abnormal")
	}

	return nil
}

func (pef *PrepareEdgecoreFlow) deletePipeOnce() error {
	if !fileutils.IsLexist(pef.pipePath) {
		return nil
	}

	if err := fileutils.DeleteFile(pef.pipePath); err != nil {
		return err
	}

	return nil
}

func (pef *PrepareEdgecoreFlow) prepareDb() error {
	ctx, err := util.StartBackupEdgeCoreDb(context.Background())
	if err != nil {
		return fmt.Errorf("prepare edgecore db failed, %v", err)
	}
	pef.dbBackupCtx = ctx
	return nil
}

func (pef *PrepareEdgecoreFlow) prepareConfig() error {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed, error: %v", err)
		return errors.New("get config path manager failed")
	}
	edgeCoreConfigPath := configPathMgr.GetEdgeCoreConfigPath()
	err = config.SmoothEdgeCoreConfigSystemReserve(configPathMgr.GetInstallRootDir(), false)
	if err == nil {
		if backupErr := backuputils.BackUpFiles(edgeCoreConfigPath); backupErr != nil {
			hwlog.RunLog.Warnf("create backup files for edgecore config failed, %v", err)
		}
		return nil
	}
	if restoreErr := backuputils.RestoreFiles(edgeCoreConfigPath); restoreErr != nil {
		hwlog.RunLog.Errorf("restoreErr %v", restoreErr)
		return fmt.Errorf("prepare edgecore config failed, %v, restore from backup error", err)
	}
	return config.SmoothEdgeCoreConfigSystemReserve(configPathMgr.GetInstallRootDir(), false)
}

func (pef *PrepareEdgecoreFlow) createPipe() error {
	if err := pef.deletePipeOnce(); err != nil {
		hwlog.RunLog.Errorf("pipe file exists and delete it failed: %s", err.Error())
		return errors.New("pipe file exists and delete it failed")
	}

	if err := syscall.Mkfifo(pef.pipePath, constants.Mode600); err != nil {
		hwlog.RunLog.Errorf("create pipe failed: %s", err.Error())
		return errors.New("create pipe failed")
	}

	return nil
}

func (pef *PrepareEdgecoreFlow) startEdgecore() error {
	workPathMgr, err := path.GetWorkPathMgr()
	if err != nil {
		return fmt.Errorf("get work path manager failed: %v", err)
	}

	coreCmd, err := filepath.EvalSymlinks(workPathMgr.GetCompBinaryPath(constants.EdgeCore, constants.EdgeCoreFileName))
	if err != nil {
		return fmt.Errorf("eval edgecore file symlink failed: %v", err)
	}
	coreJsonPath, err := filepath.EvalSymlinks(workPathMgr.GetCompJsonPath(constants.EdgeCore, constants.EdgeCoreJsonName))
	if err != nil {
		return fmt.Errorf("eval edgecore.json symlink failed: %v", err)
	}
	coreLogPath, err := filepath.EvalSymlinks(workPathMgr.
		GetCompLogLinkPath(constants.EdgeCore, constants.EdgeCoreLogFile))
	if err != nil {
		return fmt.Errorf("eval edge_core_run.log symlink failed: %v", err)
	}

	if err = fileutils.DeleteFile(kubeletSocket); err != nil {
		hwlog.RunLog.Errorf("kubelet.sock file exists and delete it failed: %s", err.Error())
		return err
	}
	pef.pid, err = envutils.RunResidentCmd(coreCmd, fmt.Sprintf("--config=%s", coreJsonPath),
		fmt.Sprintf("--log_file=%s", coreLogPath), "--logtostderr=false")
	if err != nil {
		return err
	}
	// to makesure the edgecore is reading the pipe file
	return pef.makesureEdgeCoreStart()
}

func (pef *PrepareEdgecoreFlow) makesureEdgeCoreStart() error {
	var startCoreTimes int
	for {
		startCoreTimes++
		if fileutils.IsExist(kubeletSocket) {
			return nil
		}
		if err := checkProcessExists(pef.pid); err != nil {
			hwlog.RunLog.Error("edgecore process dead, release prepare_edgecore process now")
			return err
		}
		time.Sleep(time.Second)
		if startCoreTimes > maxCoreStartTimes {
			startCoreTimes = 0
			hwlog.RunLog.Error("startup edgecore takes too long, release prepare_edgecore process now")
			return errors.New("startup edgecore takes too long")
		}
	}
}

func (pef *PrepareEdgecoreFlow) writeInfo() error {
	task := NewWriteEdgecoreInfoTask(pef.pipePath, pef.pid)
	if err := task.Run(); err != nil {
		if deleteErr := fileutils.DeleteFile(pef.pipePath); deleteErr != nil {
			hwlog.RunLog.Warnf("delete edgecore pipePath failed: %s", deleteErr.Error())
		}
		return err
	}

	return nil
}

func (pef *PrepareEdgecoreFlow) deletePipe() error {
	go func() {
		timer := time.NewTimer(deletePipeInterval)
		defer timer.Stop()
		select {
		case <-timer.C:
			if err := fileutils.DeleteFile(pef.pipePath); err != nil {
				hwlog.RunLog.Errorf("delete pipe file failed: %s", err.Error())
			}
		}
	}()

	return nil
}

func (pef *PrepareEdgecoreFlow) minitorEdgecore() error {
	for {
		if err := checkProcessExists(pef.pid); err != nil {
			hwlog.RunLog.Warn("edgecore process dead, release prepare_edgecore process now")
			return nil
		}
		if !pef.checkEdgecoreConn() {
			hwlog.RunLog.Warn("edgecore websocket disconnected, restart edgecore process now")
			return nil
		}
		select {
		case <-pef.dbBackupCtx.Done():
			hwlog.RunLog.Warn("edgecore database was restored, restart edgecore process now")
			if err := util.RestartService(constants.EdgeMainServiceFile); err != nil {
				hwlog.RunLog.Errorf("failed to restart edge-main process, %v", err)
			}
			return nil
		default:
		}

		time.Sleep(checkEdgecoreWaitTime)
	}
}

func (pef *PrepareEdgecoreFlow) addPortLimitRule() error {
	processPortMgr := envutils.ProcessPortMgr{
		Pid: pef.pid,
	}
	var ports []int
	var err error
	var i int
	for i = 0; i < findEdgecorePortTimes; i++ {
		ports, err = processPortMgr.GetPortByPid(envutils.TcpProtocol, envutils.ListenState)
		if err != nil {
			hwlog.RunLog.Errorf("get edgecore's port by its pid failed: %s", err.Error())
			return errors.New("get edgecore's port by its pid failed")
		}
		if len(ports) != 0 {
			break
		}
		hwlog.RunLog.Info("cannot find edgecore rand port, try again")
		time.Sleep(time.Second)
	}
	if i == findEdgecorePortTimes {
		hwlog.RunLog.Error("cannot find edgecore port to limit")
		return errors.New("cannot find edgecore port to limit")
	}

	if len(ports) > 1 {
		hwlog.RunLog.Warn("there has some unknown edgecore port")
	}
	for _, port := range ports {
		res, err := filepath.EvalSymlinks(constants.IptablesPath)
		if err != nil {
			hwlog.RunLog.Error("cannot get iptables command")
			return err
		}
		if _, err = envutils.RunCommand(res, envutils.DefCmdTimeoutSec, constants.Iptables, "-t", "filter", "-I",
			constants.PortLimitIptablesRuleName, "-p", "tcp", "--dport", strconv.Itoa(port), "-j", "DROP"); err != nil {
			hwlog.RunLog.Errorf("limit port error: %v", err)
			return err
		}
	}

	hwlog.RunLog.Info("limit port success")
	return nil
}

func (pef *PrepareEdgecoreFlow) checkEdgecoreConn() bool {
	ports, err := pef.getPortByPid()
	if err != nil {
		return false
	}

	if !pef.portExists {
		if ports == nil || len(ports) == 0 {
			return false
		}

		pef.establishedPort = ports
		pef.portExists = true
	}

	if !reflect.DeepEqual(pef.establishedPort, ports) {
		return false
	}

	return true
}

func (pef *PrepareEdgecoreFlow) getPortByPid() ([]int, error) {
	processPortMgr := envutils.ProcessPortMgr{
		Pid: pef.pid,
	}

	ports, err := processPortMgr.GetPortByPid(envutils.TcpProtocol, envutils.EstablishedState)
	if err != nil {
		hwlog.RunLog.Errorf("get edgecore's port by its pid failed: %s", err.Error())
		return nil, errors.New("get edgecore's port by its pid failed")
	}

	return ports, nil
}

// WriteEdgecoreInfoTask is used to write edgecore key into pipe file
type WriteEdgecoreInfoTask struct {
	pipePath string
	corePid  int
}

// NewWriteEdgecoreInfoTask is the main func to init a WriteEdgecoreInfoTask struct
func NewWriteEdgecoreInfoTask(pipePath string, corePid int) *WriteEdgecoreInfoTask {
	return &WriteEdgecoreInfoTask{pipePath: pipePath, corePid: corePid}
}

// Run starts to flow operation
func (wek WriteEdgecoreInfoTask) Run() error {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed: %v", err)
		return errors.New("get config path manager failed")
	}

	cipherKeyPath := configPathMgr.GetCompInnerSvrKeyPath(constants.EdgeCore)
	if !fileutils.IsExist(cipherKeyPath) {
		return fmt.Errorf("edgecore tls key file not exist")
	}

	kmcCfgDir := configPathMgr.GetCompKmcConfigPath(constants.EdgeCore)
	if err = backuputils.InitConfig(kmcCfgDir, kmc.InitKmcCfg); err != nil {
		hwlog.RunLog.Warnf("init edge core kmc config from json failed: %v, use default kmc config", err)
	}

	kmcDir := configPathMgr.GetCompKmcDir(constants.EdgeCore)
	kmcCfg, err := util.GetKmcConfig(kmcDir)
	if err != nil {
		return err
	}

	if err := checkCert(configPathMgr); err != nil {
		return fmt.Errorf("check cert failed, %v", err)
	}

	plainKeyBytes, err := certutils.GetKeyContentWithBackup(cipherKeyPath, kmcCfg)
	if err != nil {
		return err
	}
	defer utils.ClearSliceByteMemory(plainKeyBytes)

	if err = wek.writeInfoDataToPipe(plainKeyBytes, writeTimes); err != nil {
		hwlog.RunLog.Errorf("write key into pipe file failed: %s", err.Error())
		return errors.New("write key into pipe file failed")
	}
	hwlog.RunLog.Infof("write edgecore tls key data %v times success", writeTimes)
	return nil
}

func checkCert(configPathMgr *pathmgr.ConfigPathMgr) error {
	certPaths := []string{configPathMgr.GetCompInnerRootCertPath(constants.EdgeCore),
		configPathMgr.GetCompInnerSvrCertPath(constants.EdgeCore)}
	for _, certPath := range certPaths {
		if _, err := certutils.GetCertContentWithBackup(certPath); err != nil {
			hwlog.RunLog.Errorf("check edgecore inner cert failed %v", err)
			return fmt.Errorf("check cert failed, %v", err)
		}
	}
	return nil
}

func (wek WriteEdgecoreInfoTask) writeInfoDataToPipe(data []byte, writeTimes int) error {
	if writeTimes > maxWriteTimes {
		return fmt.Errorf("write key time exceed")
	}

	waitReadInterval := time.Second
	timer := time.NewTimer(waitReadInterval)
	defer timer.Stop()
	for i := 0; i < writeTimes; i++ {
		done := make(chan error)
		go func(data []byte, done chan error) { done <- wek.doPipeWriteOperation(data) }(data, done)

		select {
		case <-timer.C:
			hwlog.RunLog.Error("no process is reading pipe file, delete it now")
			if err := wek.deletePipe(); err != nil {
				return errors.New("no process is reading pipe file, and delete pipe file failed")
			}
			return errors.New("no process is reading pipe file")
		case err := <-done:
			if err != nil {
				return err
			}
		}
		time.Sleep(writePipeInterval)
		timer.Reset(waitReadInterval)
	}
	return nil
}

func (wek WriteEdgecoreInfoTask) doPipeWriteOperation(data []byte) error {
	pipe, err := os.OpenFile(wek.pipePath, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return fmt.Errorf("open edgecore pipe file error: %v", err)
	}
	defer func() {
		if err := pipe.Close(); err != nil {
			hwlog.RunLog.Errorf("close edgecore pipe file error:%v", err)
		}
	}()

	LinkChecker := fileutils.NewFileLinkChecker(false)
	ownerChecker := fileutils.NewFileOwnerChecker(false, false, constants.RootUserUid, constants.RootUserGid)
	modeChecker := fileutils.NewFileModeChecker(false, fileutils.DefaultWriteFileMode, false, false)
	LinkChecker.SetNext(ownerChecker)
	LinkChecker.SetNext(modeChecker)
	if err = LinkChecker.Check(pipe, wek.pipePath); err != nil {
		hwlog.RunLog.Errorf("check edgecore pipe failed: %s", err.Error())
		return errors.New("check edgecore pipe failed")
	}

	if err = checkProcessExists(wek.corePid); err != nil {
		hwlog.RunLog.Errorf("%s, cannot write into pipe file", err.Error())
		return err
	}

	writtenLen, err := pipe.Write(data)
	if err != nil {
		return fmt.Errorf("write edgecore pipe file error: %v", err)
	}
	if writtenLen != len(data) {
		return fmt.Errorf("write edgecore pipe data not correct")
	}
	return nil
}

func (wek WriteEdgecoreInfoTask) deletePipe() error {
	if err := fileutils.DeleteFile(wek.pipePath); err != nil {
		hwlog.RunLog.Errorf("delete pipe file failed: %s", err.Error())
		return errors.New("delete pipe file failed")
	}

	return nil
}

func checkProcessExists(pid int) error {
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("edge core process does not exist")
		}

		hwlog.RunLog.Error("get edge core process status failed")
		return errors.New("get edge core process status failed")
	}

	return nil
}
