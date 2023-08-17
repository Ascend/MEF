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

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

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
		Phase:    taskschedule.Progressing,
		Message:  "start to create tar gz",
		Progress: progressStartCreateTarGz,
	}
	succeedStatus = taskschedule.TaskStatus{
		Phase:    taskschedule.Succeed,
		Message:  "task succeeded",
		Data:     map[string]interface{}{"fileName": constants.EdgeNodesTarGzFileName},
		Progress: 100,
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

	// 1. parse serial number
	var serialNumbers []string
	if err := ctx.Spec().Args.Get(paramNameNodeSerialNumbers, &serialNumbers); err != nil {
		return fmt.Errorf("failed to parse serial number, %v", err)
	}
	hwlog.RunLog.Info("start to dump logs of edge nodes")

	// 2. ensure temp dirs
	if err := prepareDirs(); err != nil {
		return err
	}

	// 3. check disk space
	if err := envutils.CheckDiskSpace(
		constants.LogDumpTempDir, uint64(len(serialNumbers))*constants.LogUploadMaxSize); err != nil {
		return fmt.Errorf("temp dir has no engouh disk space, %v", err)
	}
	// 4. dump edge logs
	succeedTasks, err := dumpEdgeLogs(ctx, serialNumbers)
	if err != nil {
		return err
	}

	// 5. create tar.gz
	if err := createTarGz(ctx, succeedTasks); err != nil {
		return err
	}

	// 6. rename file
	if err := fileutils.RenameFile(edgeNodesLogTempPath, edgeNodesLogPublicPath); err != nil {
		return fmt.Errorf("failed to rename file, %v", err)
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
		_, _ = utils.CleanTempFiles()
		hwlog.RunLog.Errorf("failed to update task status, %v", err)
		return
	}
	hwlog.RunLog.Info("dump log successful")
}

func createSubTasks(masterTaskCtx taskschedule.TaskContext, serialNumbers []string) {
	for _, serialNumber := range serialNumbers {
		subTask := taskschedule.TaskSpec{
			Name:             fmt.Sprintf("%s.%s", constants.DumpSingleNodeLogTaskName, serialNumber),
			ParentId:         masterTaskCtx.Spec().Id,
			GoroutinePool:    constants.DumpSingleNodeLogTaskName,
			Command:          constants.DumpSingleNodeLogTaskName,
			Args:             map[string]interface{}{paramNameNodeSerialNumber: serialNumber},
			HeartbeatTimeout: dumpSingleNodeLogTaskHeartbeatTimeout,
			ExecuteTimeout:   dumpSingleNodeLogTaskExecuteTimeout,
		}
		if err := taskschedule.DefaultScheduler().SubmitTask(&subTask); err != nil {
			hwlog.RunLog.Errorf("failed to create sub task for node(serialNumber=%s), %v", serialNumber, err)
			continue
		}
	}
}

func dumpEdgeLogs(masterTaskCtx taskschedule.TaskContext, serialNumbers []string) ([]taskschedule.Task, error) {
	createSubTasks(masterTaskCtx, serialNumbers)
	var (
		succeedTasks []taskschedule.Task
		err          error
		childCtx     taskschedule.TaskContext
		doneCount    int
	)
	taskIter := taskschedule.DefaultScheduler().NewSubTaskSelector(masterTaskCtx.Spec().Id)
	for {
		childCtx, err = taskIter.Select(masterTaskCtx.GracefulShutdown())
		if err != nil {
			break
		}

		doneCount++
		status := taskschedule.TaskStatus{
			Progress: (common.ProgressMax - progressStartCreateTarGz) *
				uint(float64(doneCount)/float64(len(serialNumbers))),
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
		}
	}
	if err != nil && err != taskschedule.ErrNoMoreChild {
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
	defer outputFile.Close()
	fileChecker := fileutils.NewFileLinkChecker(false)
	if err := fileChecker.Check(outputFile, edgeNodesLogTempPath); err != nil {
		return fmt.Errorf("failed to check log temp path, %v", err)
	}
	gzipWriter, err := gzip.NewWriterLevel(outputFile, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("failed to create gzip write, %v", err)
	}
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, task := range subTasks {
		if err := addSingleNodeTarGz(task, tarWriter); err != nil {
			return fmt.Errorf("failed to add single node tar gz, %v", err)
		}
	}
	hwlog.RunLog.Info("create tar.gz successful")
	return nil
}

func addSingleNodeTarGz(task taskschedule.Task, tarWriter *tar.Writer) error {
	taskName := task.Spec.Name
	if len(taskName) <= len(constants.DumpSingleNodeLogTaskName) {
		return errors.New("invalid task name")
	}
	serialNumber := taskName[len(constants.DumpSingleNodeLogTaskName)+1:]
	tarGzPath := filepath.Join(constants.LogDumpTempDir, task.Spec.Id+common.TarGzSuffix)
	if _, err := fileutils.CheckOriginPath(tarGzPath); err != nil {
		return fmt.Errorf("failed to check log temp path, %v", err)
	}

	defer func() {
		if err := fileutils.DeleteFile(tarGzPath); err != nil {
			hwlog.RunLog.Errorf("failed to delete temp file, %v", err)
		}
	}()

	tarGzFile, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to open temp file, %v", err)
	}
	defer tarGzFile.Close()

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
		if err := fileutils.CreateDir(dir, common.Mode700); err != nil {
			return fmt.Errorf("failed to creatre dir %s, %v", dir, err)
		}
	}
	return nil
}
