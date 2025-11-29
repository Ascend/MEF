// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package tasks
package tasks

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/utils"
)

const (
	dumpSingleNodeLogTaskHeartbeatTimeout = 30 * time.Second
	dumpSingleNodeLogTaskExecuteTimeout   = 10 * time.Minute

	progressStartCreateTarGz = 50

	pkgExpireTime = 24 * time.Hour
)

var (
	edgeNodesLogTempPath   = filepath.Join(constants.LogDumpTempDir, constants.EdgeNodesTarGzFileName)
	edgeNodesLogPublicPath = filepath.Join(constants.LogDumpPublicDir, constants.EdgeNodesTarGzFileName)

	startCreateTarGzStatus = taskschedule.TaskStatus{
		Phase:    taskschedule.Processing,
		Message:  "start to create tar gz",
		Progress: progressStartCreateTarGz,
	}
	succeedStatus = taskschedule.TaskStatus{
		Phase:    taskschedule.Succeed,
		Message:  "task succeeded",
		Data:     map[string]interface{}{"fileName": constants.EdgeNodesTarGzFileName},
		Progress: common.ProgressMax,
	}
	partiallyFailedStatus = taskschedule.TaskStatus{
		Phase:   taskschedule.PartiallyFailed,
		Message: "task partially failed",
		Data:    map[string]interface{}{"fileName": constants.EdgeNodesTarGzFileName},
	}

	cleanTimer *time.Timer
	packLock   sync.Mutex
)

func doDumpMultiNodesLog(ctx taskschedule.TaskContext) {
	if cleanTimer != nil {
		cleanTimer.Stop()
	}
	if err := dumpMultiNodesLog(ctx); err != nil {
		utils.FeedbackTaskError(ctx, err)
		cleanTempFiles()
		return
	}

	cleanTimer = time.AfterFunc(pkgExpireTime, cleanTempFiles)
}

func dumpMultiNodesLog(ctx taskschedule.TaskContext) error {
	packLock.Lock()
	defer packLock.Unlock()

	// 1. parse serial number & node ids & ips
	serialNumbers, err := parseSerialNumbers(ctx)
	if err != nil {
		return errors.New("failed to parse serial number")
	}
	ips, err := parseIps(ctx)
	if err != nil {
		return errors.New("failed to parse ip")
	}
	nodeIds, err := parseNodeIds(ctx)
	if err != nil {
		return errors.New("failed to parse node id")
	}

	hwlog.RunLog.Info("start to dump logs of edge nodes")

	// 2. ensure temp dirs
	if err := prepareDirs(); err != nil {
		hwlog.RunLog.Errorf("failed to prepare temp dir, %v", err)
		return errors.New("failed to prepare temp dir")
	}

	// 3. check disk space
	if err := utils.CheckDiskSpace(
		constants.LogDumpTempDir, uint64(len(serialNumbers))*constants.LogUploadMaxSize); err != nil {
		hwlog.RunLog.Errorf("temp dir has no enough disk space, %v", err)
		return errors.New("temp dir has no enough disk space")
	}
	// 4. dump edge logs
	succeedTasks, err := dumpEdgeLogs(ctx, serialNumbers, ips, nodeIds)
	if err != nil {
		hwlog.RunLog.Errorf("failed to dump edge logs, %v", err)
		return errors.New("failed to dump edge logs")
	}

	// 5. create tar.gz
	if err := createTarGz(ctx, succeedTasks); err != nil {
		hwlog.RunLog.Errorf("failed to create tar gz, %v", err)
		return errors.New("failed to create tar gz")
	}

	// 6. rename file
	if err := fileutils.RenameFile(edgeNodesLogTempPath, edgeNodesLogPublicPath); err != nil {
		hwlog.RunLog.Errorf("failed to rename file, %v", err)
		return errors.New("failed to rename file")
	}
	hwlog.RunLog.Info("rename tar.gz successful")

	// 7. update task status
	updateMasterTaskStatus(ctx, len(succeedTasks) == len(serialNumbers))
	return nil
}

func cleanTempFiles() {
	packLock.Lock()
	defer packLock.Unlock()

	if _, err := utils.CleanTempFiles(); err != nil {
		hwlog.RunLog.Errorf("failed to clean temp dir, %v", err)
	}
}

func parseSerialNumbers(ctx taskschedule.TaskContext) ([]string, error) {
	var serialNumbers []string
	if err := ctx.Spec().Args.Get(paramNameNodeSerialNumbers, &serialNumbers); err != nil {
		hwlog.RunLog.Errorf("failed to parse serial number, %v", err)
		return nil, err
	}

	return serialNumbers, nil
}

func parseIps(ctx taskschedule.TaskContext) ([]string, error) {
	var ips []string
	if err := ctx.Spec().Args.Get(paramNameNodeIps, &ips); err != nil {
		hwlog.RunLog.Errorf("failed to get node ips, %v", err)
		return nil, err
	}

	return ips, nil
}

func parseNodeIds(ctx taskschedule.TaskContext) ([]uint64, error) {
	var nodeIds []uint64
	if err := ctx.Spec().Args.Get(paramNameNodeIDs, &nodeIds); err != nil {
		hwlog.RunLog.Errorf("failed to parse node id, %v", err)
		return nil, err
	}

	return nodeIds, nil
}

func prepareDirs() error {
	exists, err := utils.CleanTempFiles()
	if err != nil {
		return fmt.Errorf("failed to clean temp dirs, %v", err)
	}
	if !exists {
		if err := createTempDirs(); err != nil {
			return fmt.Errorf("failed to create temp dirs, %v", err)
		}
	}
	hwlog.RunLog.Info("clean temp files successful")
	return nil
}

func updateMasterTaskStatus(ctx taskschedule.TaskContext, success bool) {
	var status taskschedule.TaskStatus
	if success {
		status = succeedStatus
	} else {
		status = partiallyFailedStatus
	}
	if err := ctx.UpdateStatus(status); err != nil {
		_, cleanErr := utils.CleanTempFiles()
		if cleanErr != nil {
			hwlog.RunLog.Warnf("clean temp files failed, error: %v", err)
		}
		hwlog.RunLog.Errorf("failed to update task status, %v", err)
		return
	}
	hwlog.RunLog.Info("dump log successful")
}

func createSubTasks(masterTaskCtx taskschedule.TaskContext, serialNumbers, ips []string, nodeIDs []uint64) {
	for idx := range serialNumbers {
		if len(nodeIDs) <= idx {
			continue
		}
		serialNumber := serialNumbers[idx]
		ip := ips[idx]
		nodeID := nodeIDs[idx]
		subTask := taskschedule.TaskSpec{
			Name:          fmt.Sprintf("%s.%s", constants.DumpSingleNodeLogTaskName, serialNumber),
			ParentId:      masterTaskCtx.Spec().Id,
			GoroutinePool: constants.DumpSingleNodeLogTaskName,
			Command:       constants.DumpSingleNodeLogTaskName,
			Args: map[string]interface{}{
				constants.NodeSnAndIp: serialNumber,
				constants.NodeID:      nodeID,
				constants.PeerInfo: model.MsgPeerInfo{
					Sn: serialNumber,
					Ip: ip,
				},
			},
			HeartbeatTimeout: dumpSingleNodeLogTaskHeartbeatTimeout,
			ExecuteTimeout:   dumpSingleNodeLogTaskExecuteTimeout,
		}
		if err := taskschedule.DefaultScheduler().SubmitTask(&subTask); err != nil {
			hwlog.RunLog.Errorf("failed to create sub task for node(serialNumber=%s), %v", serialNumber, err)
			continue
		}
	}
}

func dumpEdgeLogs(masterTaskCtx taskschedule.TaskContext, serialNumbers,
	ips []string, nodeIDs []uint64) ([]taskschedule.Task, error) {
	if len(serialNumbers) == 0 {
		return nil, errors.New("no edge node to dump logs")
	}
	createSubTasks(masterTaskCtx, serialNumbers, ips, nodeIDs)
	var (
		succeedTasks []taskschedule.Task
		err          error
		childCtx     taskschedule.TaskContext
		doneCount    int
	)
	taskIter := taskschedule.DefaultScheduler().NewSubTaskSelector(masterTaskCtx.Spec().Id)
	const maxNodes = 100
	for i := 0; i < maxNodes; i++ {
		childCtx, err = taskIter.Select(masterTaskCtx.GracefulShutdown())
		if err != nil {
			break
		}
		doneCount++
		status := taskschedule.TaskStatus{
			Progress: uint(float64(common.ProgressMax-progressStartCreateTarGz) *
				float64(doneCount) / float64(len(serialNumbers))),
			Message: fmt.Sprintf("receving (%d/%d) files", doneCount, len(serialNumbers)),
		}
		if err = masterTaskCtx.UpdateStatus(status); err != nil {
			break
		}

		status, err = childCtx.GetStatus()
		if err != nil {
			break
		}
		if status.Phase == taskschedule.Succeed {
			succeedTasks = append(succeedTasks, taskschedule.Task{Spec: childCtx.Spec(), Status: status})
			continue
		}
		var serialNumber string
		if err := childCtx.Spec().Args.Get(constants.NodeSnAndIp, &serialNumber); err != nil {
			continue
		}
		hwlog.RunLog.Errorf("sub task for node(%s) failed, %s", serialNumber, status.Message)
	}
	if err != nil && err != taskschedule.ErrNoRunningSubTask {
		return nil, fmt.Errorf("failed to dump edge logs, %v", err)
	}
	if len(succeedTasks) == 0 {
		return nil, errors.New("none of task succeed")
	}
	hwlog.RunLog.Info("dump logs of edge node successful")
	return succeedTasks, nil
}

func createTarGz(ctx taskschedule.TaskContext, subTasks []taskschedule.Task) error {
	if err := ctx.UpdateStatus(startCreateTarGzStatus); err != nil {
		return fmt.Errorf("failed to update task status, %v", err)
	}

	outputFile, err := os.OpenFile(edgeNodesLogTempPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, common.Mode400)
	if err != nil {
		return fmt.Errorf("failed to create output file")
	}
	defer func() {
		if err = outputFile.Close(); err != nil {
			hwlog.RunLog.Errorf("close output file handle failed, error: %v", err)
		}
	}()

	fileChecker := fileutils.NewFileLinkChecker(false)
	if err = fileChecker.Check(outputFile, edgeNodesLogTempPath); err != nil {
		return fmt.Errorf("failed to check log temp path, %v", err)
	}
	gzipWriter, err := gzip.NewWriterLevel(
		utils.WithDiskPressureProtect(outputFile, edgeNodesLogTempPath), gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("failed to create gzip write, %v", err)
	}
	defer func() {
		if err = gzipWriter.Close(); err != nil {
			hwlog.RunLog.Errorf("close gzip writer handle failed, error: %v", err)
		}
	}()

	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if err = tarWriter.Close(); err != nil {
			hwlog.RunLog.Errorf("close tar writer handle failed, error: %v", err)
		}
	}()

	for _, task := range subTasks {
		if err = addSingleNodeTarGz(task, tarWriter); err != nil {
			return fmt.Errorf("failed to add single node tar gz, %v", err)
		}
	}
	hwlog.RunLog.Info("create tar.gz successful")
	return nil
}

func addSingleNodeTarGz(task taskschedule.Task, tarWriter *tar.Writer) error {
	var serialNumber string
	if err := task.Spec.Args.Get(constants.NodeSnAndIp, &serialNumber); err != nil {
		return errors.New("can't get serial number")
	}
	tarGzPath := filepath.Join(constants.LogDumpTempDir, task.Spec.Id+common.TarGzSuffix)
	defer func() {
		if err := fileutils.DeleteFile(tarGzPath); err != nil {
			hwlog.RunLog.Errorf("failed to delete temp file, %v", err)
		}
	}()
	if _, err := fileutils.CheckOriginPath(tarGzPath); err != nil {
		return fmt.Errorf("failed to check log temp path, %v", err)
	}

	tarGzFile, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to open temp file, %v", err)
	}
	defer func() {
		if err = tarGzFile.Close(); err != nil {
			hwlog.RunLog.Errorf("close file handle failed, error: %v", err)
		}
	}()

	stat, err := tarGzFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get size of temp file, %v", err)
	}
	hdr := tar.Header{
		Name: serialNumber + common.TarGzSuffix,
		Size: stat.Size(),
	}
	if err := tarWriter.WriteHeader(&hdr); err != nil {
		return fmt.Errorf("failed to write tar header, %v", err)
	}
	if _, err := io.Copy(tarWriter, tarGzFile); err != nil {
		return fmt.Errorf("failed to write tar content, %v", err)
	}
	return nil
}

func createTempDirs() error {
	dirs := []string{constants.LogDumpTempDir, constants.LogDumpPublicDir}
	for _, dir := range dirs {
		if err := fileutils.CreateDir(dir, fileutils.Mode700); err != nil {
			return fmt.Errorf("failed to creatre dir %s, %v", dir, err)
		}

		if _, err := fileutils.RealDirCheck(dir, true, false); err != nil {
			return fmt.Errorf("failed to check temp dir %s after creation, %v", dir, err)
		}
	}
	return nil
}
