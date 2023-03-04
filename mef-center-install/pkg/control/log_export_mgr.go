// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package control
package control

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker/checker"
	"huawei.com/mindxedge/base/common/logmgmt/logcollect"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const (
	jobYamlName         = "log_export_tool.yaml"
	jobYamlTemplateName = "log_export_tool-template.yaml"

	groupOrOtherWrite = 0022
	toolsDir          = "tools"
	logExportToolDir  = "log-export-tool"
	logExportJob      = "ascend-mef-center-log-export-job"
	jobWaitTime       = 60 * 30

	envParamNodes   = "param_nodes"
	envLogDir       = "log_dir"
	envLogBackupDir = "log_backup_dir"
	envConfigDir    = "config_dir"
	envRootCaDir    = "root_ca_dir"
	envWorkDir      = "work_dir"
	envKmcLibDir    = "kmc_lib_dir"
	envMefCenterUid = "mef_center_uid"
	envMefCenterGid = "mef_center_gid"

	nodeSnRegexp        = `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`
	podNameRegexp       = `^[a-z0-9-]+$`
	podCreationWaitTime = 5 * time.Second
	maxSn               = 16
)

var (
	podNameReg = regexp.MustCompile(podNameRegexp)
)

// LogExportMgr handles log exporting
type LogExportMgr struct {
	module            string
	edgeNodes         []string
	logDirPathMgr     *util.LogDirPathMgr
	installDirPathMgr *util.InstallDirPathMgr
}

// GetLogExportMgrIns get log export manager instance
func GetLogExportMgrIns(
	module string, edgeNodes []string,
	logDirPathMgr *util.LogDirPathMgr, installDirPathMgr *util.InstallDirPathMgr) LogExportMgr {
	return LogExportMgr{
		module:            module,
		edgeNodes:         edgeNodes,
		logDirPathMgr:     logDirPathMgr,
		installDirPathMgr: installDirPathMgr,
	}
}

// DoExport is the main func to export logs
func (lem *LogExportMgr) DoExport() error {
	var controlTasks = []func() error{
		lem.checkParam,
		lem.clean,
		lem.deal,
	}
	for _, function := range controlTasks {
		if err := function(); err != nil {
			hwlog.RunLog.Errorf("failed to execute log collection: %v", err)
			return err
		}
	}
	return nil
}

func (lem *LogExportMgr) deal() error {
	switch lem.module {
	case logcollect.ModuleCenter:
		return lem.exportCenterLogs()
	case logcollect.ModuleEdge:
		return lem.exportEdgeLogs()
	default:
	}
	return fmt.Errorf("unknown module:%s", lem.module)
}

func (lem *LogExportMgr) clean() error {
	var logExportDir string
	switch lem.module {
	case logcollect.ModuleEdge:
		logExportDir = logcollect.EdgeLogExportDir
	case logcollect.ModuleCenter:
		logExportDir = logcollect.CenterLogExportDir
	default:
		return errors.New("unknown module")
	}
	entries, err := os.ReadDir(logExportDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		fileName := filepath.Join(logExportDir, e.Name())
		stat, err := os.Lstat(fileName)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			return common.DeleteAllFile(fileName)
		}
		return common.DeleteFile(fileName)
	}
	return nil
}

func (lem *LogExportMgr) checkParam() error {
	if lem.module == logcollect.ModuleCenter {
		if !(len(lem.edgeNodes) == 0 || (len(lem.edgeNodes) == 1 && lem.edgeNodes[0] == "")) {
			return errors.New("too much nodes")
		}
		return nil
	}
	if len(lem.edgeNodes) == 0 {
		return errors.New("too few nodes")
	}
	if len(lem.edgeNodes) > maxSn {
		return errors.New("too much nodes")
	}
	chker := checker.GetRegChecker("", nodeSnRegexp, true)
	for _, n := range lem.edgeNodes {
		checkResult := chker.Check(n)
		if !checkResult.Result {
			return errors.New(checkResult.Reason)
		}
	}
	return nil
}

func (lem *LogExportMgr) exportCenterLogs() error {
	packFileName, err := lem.getLogCollector().Collect()
	if err != nil {
		hwlog.RunLog.Errorf("failed to collect center logs: %v", err)
		return err
	}
	hwlog.RunLog.Infof("collect center log success, the name of pack is %s", packFileName)
	return nil
}

func (lem *LogExportMgr) getLogCollector() logcollect.Collector {
	logFiles := logcollect.LogGroup{
		RootDir:   filepath.Join(lem.logDirPathMgr.GetLogRootPath(), util.ModuleLogName),
		BaseDir:   logcollect.ModuleCenter,
		CheckFunc: lem.checkLogFile,
	}
	logBackupFiles := logcollect.LogGroup{
		RootDir:   filepath.Join(lem.logDirPathMgr.GetLogBackupRootPath(), util.ModuleLogBackupName),
		BaseDir:   logcollect.ModuleCenter,
		CheckFunc: lem.checkLogFile,
	}
	packFileName := logcollect.GetLogPackFileName(logcollect.ModuleCenter, "")
	packFilePath := filepath.Join(logcollect.EdgeLogExportDir, packFileName)
	return logcollect.NewCollector(
		packFilePath, []logcollect.LogGroup{logFiles, logBackupFiles}, logcollect.CenterMaxPackSize)
}

func (lem *LogExportMgr) exportEdgeLogs() error {
	if _, err := common.RunCommand(util.CommandKubectl, true, common.DefCmdTimeoutSec,
		"delete", "job", "-n", util.MefNamespace, logExportJob); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			hwlog.RunLog.Errorf("delete job failed: %s", err.Error())
			fmt.Println("delete job failed")
			return err
		}
	}
	hwlog.RunLog.Info("delete job success")
	yamlPath, err := lem.createJobYaml()
	if err != nil {
		hwlog.RunLog.Errorf("create yaml failed: %s", err.Error())
		fmt.Println("create yaml failed")
		return err
	}
	hwlog.RunLog.Info("create yaml success")
	if _, err := common.RunCommand(util.CommandKubectl, true, common.DefCmdTimeoutSec,
		"apply", "-f", yamlPath); err != nil {
		hwlog.RunLog.Errorf("create job failed: %s", err.Error())
		fmt.Println("create job failed")
		return err
	}
	hwlog.RunLog.Info("create job success")

	lem.waitForJob()

	exitCode, err := lem.getExitCode()
	if err != nil {
		hwlog.RunLog.Errorf("get exit code failed: %s", err.Error())
		fmt.Println("get exit code failed")
		return err
	}
	if exitCode != 0 {
		hwlog.RunLog.Errorf("job exit with code %d", exitCode)
		fmt.Println("job exit with non-zero code")
		return errors.New("job exit with non-zero code")
	}
	hwlog.RunLog.Info("execute job success")
	return nil
}

func (lem *LogExportMgr) waitForJob() {
	completeCh := make(chan struct{})
	failedCh := make(chan struct{})
	ctx, cancelFunc := context.WithCancel(context.Background())
	go lem.waitCondition("complete", completeCh)
	go lem.waitCondition("failed", failedCh)
	go lem.redirectStdout(ctx)
	select {
	case _, ok := <-completeCh:
		if !ok {
			hwlog.RunLog.Error("channel closed")
		}
	case _, ok := <-failedCh:
		if !ok {
			hwlog.RunLog.Error("channel closed")
		}
	}
	cancelFunc()
}

func (lem *LogExportMgr) redirectStdout(ctx context.Context) {
	time.Sleep(podCreationWaitTime)
	podName, err := common.RunCommand(util.CommandKubectl, true, jobWaitTime,
		"get", "pods", "-n", util.MefNamespace, "--selector", "job-name="+logExportJob,
		"-o=jsonpath={.items[0].metadata.name}")
	if err != nil {
		hwlog.RunLog.Error("failed to get pod name of job")
		return
	}
	if !podNameReg.MatchString(podName) || strings.ContainsAny(podName, common.IllegalChars) {
		hwlog.RunLog.Error("invalid pod name")
		return
	}
	cmd := getCmd(util.CommandKubectl, "logs", "-f", "-n", util.MefNamespace, podName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		hwlog.RunLog.Error("failed to watch output of job")
		return
	}
	<-ctx.Done()
	if err := cmd.Process.Kill(); err != nil {
		hwlog.RunLog.Error("failed to kill the watch process")
	}
}

func (lem *LogExportMgr) getExitCode() (int, error) {
	output, err := common.RunCommand(util.CommandKubectl, true, jobWaitTime,
		"get", "pods", "-n", util.MefNamespace, "--selector", "job-name="+logExportJob,
		"-o=jsonpath={.items[0].status.containerStatuses[0].state.terminated.exitCode}")
	if err != nil {
		return 0, err
	}
	exitCode, err := strconv.Atoi(output)
	if err != nil {
		return 0, err
	}
	return exitCode, nil
}

func (lem *LogExportMgr) waitCondition(condition string, ch chan<- struct{}) {
	if output, err := common.RunCommand(util.CommandKubectl, true, jobWaitTime,
		"wait", "--for=condition="+condition, "-n", util.MefNamespace, "job/"+logExportJob, "--timeout=30m"); err != nil {
		hwlog.RunLog.Errorf("wait job failed: %s", err.Error())
		fmt.Println(output)
	}
	if ch != nil {
		ch <- struct{}{}
	}
}

func (lem *LogExportMgr) getYamlReplacements() (map[string]string, error) {
	uid, gid, err := util.GetMefId()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		envParamNodes:   strings.Join(lem.edgeNodes, ","),
		envLogDir:       lem.logDirPathMgr.GetComponentLogPath(util.EdgeManagerName),
		envLogBackupDir: lem.logDirPathMgr.GetComponentBackupLogPath(util.EdgeManagerName),
		envConfigDir:    lem.installDirPathMgr.ConfigPathMgr.GetComponentConfigPath(util.EdgeManagerName),
		envRootCaDir:    lem.installDirPathMgr.ConfigPathMgr.GetRootCaCertDirPath(),
		envWorkDir:      filepath.Join(lem.installDirPathMgr.GetWorkPath(), toolsDir, logExportToolDir),
		envKmcLibDir: filepath.Join(lem.installDirPathMgr.GetWorkPath(), toolsDir, logExportToolDir,
			util.MefKmcLibDir),
		envMefCenterUid: strconv.Itoa(uid),
		envMefCenterGid: strconv.Itoa(gid),
	}, nil
}

func (lem *LogExportMgr) createJobYaml() (string, error) {
	yamlTemplatePath := filepath.Join(
		lem.installDirPathMgr.GetWorkPath(), toolsDir, logExportToolDir, jobYamlTemplateName)
	yamlTemplateFile, err := os.Open(yamlTemplatePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := yamlTemplateFile.Close(); err != nil {
			hwlog.RunLog.Errorf("failed to close yaml template file, %v", err)
		}
	}()
	yamlPath := filepath.Join(lem.installDirPathMgr.GetWorkPath(), toolsDir, logExportToolDir, jobYamlName)
	yamlFile, err := os.OpenFile(yamlPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, common.Mode600)
	if err != nil {
		return "", nil
	}
	defer func() {
		if err := yamlFile.Close(); err != nil {
			hwlog.RunLog.Errorf("failed to close yaml file, %v", err)
		}
	}()
	yamlTemplateBytes, err := io.ReadAll(yamlTemplateFile)
	if err != nil {
		return "", err
	}
	yamlOutput := string(yamlTemplateBytes)
	replacements, err := lem.getYamlReplacements()
	if err != nil {
		return "", err
	}
	for key, value := range replacements {
		yamlOutput = strings.ReplaceAll(yamlOutput, fmt.Sprintf("${%s}", key), value)
	}
	if _, err = yamlFile.WriteString(yamlOutput); err != nil {
		return "", err
	}
	return yamlPath, nil
}

func (lem *LogExportMgr) checkLogFile(filePath string) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if stat.Size() > logcollect.CenterMaxFileSize {
		return errors.New("log file is too large")
	}
	uid, gid, err := util.GetMefId()
	if err != nil {
		return err
	}
	syscallStat, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("unsupported operate system")
	}
	if !((syscallStat.Uid == 0 && syscallStat.Gid == 0) ||
		(syscallStat.Uid == uint32(uid)) && (syscallStat.Gid == uint32(gid))) {
		return errors.New("bad file owner")
	}
	if (stat.Mode() & groupOrOtherWrite) != 0 {
		return errors.New("bad file permission")
	}
	realPath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		return err
	}
	if realPath != filePath {
		return errors.New("symlink is not allowed")
	}
	return nil
}

func getCmd(cmd string, args ...string) *exec.Cmd {
	return exec.Command(cmd, args...)
}
