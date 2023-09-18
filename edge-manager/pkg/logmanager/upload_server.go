// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logmanager enables collecting logs
package logmanager

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/utils"
)

const (
	headerPackageSize    = "Package-Size"
	headerTaskId         = "Task-Id"
	headerSha256Checksum = "Sha256-Checksum"

	gzipHeader = "\x1F\x8B\x08"

	ioBufferSize            = common.MB
	reportSizeThreshold     = common.MB
	reportDurationThreshold = 20 * time.Second

	progressBeginReceive = 50
)

var (
	statusStartReceive = taskschedule.TaskStatus{
		Message:  "start to receive file from edge",
		Progress: progressBeginReceive,
		Phase:    taskschedule.Processing,
	}
	statusFinishReceive = taskschedule.TaskStatus{
		Phase:    taskschedule.Succeed,
		Message:  "receive file from edge successful",
		Progress: common.ProgressMax,
	}
)

type uploadProcess struct {
	httpRequest    *http.Request
	responseWriter http.ResponseWriter
	clientIP       string
	taskId         string
	serialNumber   string
	packageSize    int64
	sha256Checksum string
}

func (p uploadProcess) processUpload() error {
	// 1. start receiving
	hwlog.RunLog.Infof("start to receive file from edge(%s)", p.serialNumber)
	taskCtx, err := taskschedule.DefaultScheduler().GetTaskContext(p.taskId)
	if err != nil {
		return fmt.Errorf("failed to get task for edge(%s)", p.serialNumber)
	}
	if err := taskCtx.UpdateStatus(statusStartReceive); err != nil {
		return fmt.Errorf("failed to update task status for edge(%s)", p.serialNumber)
	}

	// 2. receiving file
	if err := p.receiveFile(taskCtx); err != nil {
		p.deleteFile()
		utils.FeedbackTaskError(taskCtx, errors.New("failed to receive file from edge"))
		return fmt.Errorf("failed to receive file from edge, %v", err)
	}

	// 3. verifying file
	if err := p.verifyFile(taskCtx); err != nil {
		p.deleteFile()
		utils.FeedbackTaskError(taskCtx, errors.New("failed to verify file from edge"))
		return fmt.Errorf("failed to verify file from edge, %v", err)
	}

	// 4. finish receiving
	if err := taskCtx.UpdateStatus(statusFinishReceive); err != nil {
		p.deleteFile()
		return fmt.Errorf("failed to update task status for edge(%s)", p.serialNumber)
	}
	hwlog.RunLog.Infof("handle uploading for edge(%s) successful", p.serialNumber)

	// 5. send OK to edge
	if _, err = p.responseWriter.Write([]byte(common.OK)); err != nil {
		hwlog.RunLog.Errorf("failed to send successful response to edge(%s): %v", p.serialNumber, err)
	}
	return nil
}

func (p uploadProcess) deleteFile() {
	filePath := filepath.Join(constants.LogDumpTempDir, p.taskId+common.TarGzSuffix)
	if err := fileutils.DeleteFile(filePath); err != nil {
		hwlog.RunLog.Errorf("failed to delete local file, node=%s, err=%s", p.taskId, err)
	}
}

func (p uploadProcess) receiveFile(taskCtx taskschedule.TaskContext) error {
	localPath := filepath.Join(constants.LogDumpTempDir, p.taskId+common.TarGzSuffix)
	if _, err := fileutils.CheckOriginPath(localPath); err != nil {
		return fmt.Errorf("failed to check local file for writing, %v", err)
	}
	localFile, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, common.Mode600)
	if err != nil {
		return fmt.Errorf("failed to open local file for writing, %v", err)
	}
	defer func() {
		if err := localFile.Close(); err != nil {
			hwlog.RunLog.Errorf("failed to close local file, %v", err)
		}
	}()

	if err := doReceiveFile(taskCtx, p.httpRequest.Body,
		utils.WithDiskPressureProtect(localFile, localPath), p.packageSize); err != nil {
		if err := fileutils.DeleteFile(localFile.Name()); err != nil {
			hwlog.RunLog.Errorf("failed to delete local file, %v", err)
		}
		return fmt.Errorf("failed to receive file from edge, %v", err)
	}
	hwlog.RunLog.Infof("receive file from edge(%s) successful", p.serialNumber)
	return nil
}

func (p uploadProcess) verifyFile(taskCtx taskschedule.TaskContext) error {
	taskId := taskCtx.Spec().Id
	localPath := filepath.Join(constants.LogDumpTempDir, taskId+common.TarGzSuffix)
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file for reading, %v", err)
	}
	defer func() {
		if err := localFile.Close(); err != nil {
			hwlog.RunLog.Errorf("failed to close local file, %v", err)
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, localFile); err != nil {
		return fmt.Errorf("failed to calculate sha256 checksum, %v", err)
	}
	if fmt.Sprintf("%x", hash.Sum(nil)) != p.sha256Checksum {
		return errors.New("sha256 checksum error")
	}

	if _, err := localFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	const headerBytesLen = 3
	fileHeader := make([]byte, headerBytesLen)
	if _, err := localFile.Read(fileHeader); err != nil {
		return err
	}
	if bytes.Compare(fileHeader, []byte(gzipHeader)) != 0 {
		return errors.New("format error")
	}
	hwlog.RunLog.Infof("verify file from edge(%s) successful", p.serialNumber)
	return nil
}

func doReceiveFile(
	taskCtx taskschedule.TaskContext, src io.Reader, dst io.Writer, totalSize int64) error {
	buffer := make([]byte, ioBufferSize)
	reader := io.LimitReader(src, totalSize)
	var (
		currentRead     int64
		lastReportCount int64
		lastReportTime  time.Time
	)
	for currentRead < totalSize {
		select {
		case <-taskCtx.GracefulShutdown():
			return errors.New("cancel")
		default:
		}

		nRead, err := reader.Read(buffer)
		if err != nil && !(err == io.EOF && int64(nRead) == totalSize-currentRead) {
			return err
		}
		currentRead += int64(nRead)
		nWrite, err := dst.Write(buffer[:nRead])
		if err != nil {
			return err
		}
		if nRead != nWrite {
			return errors.New("insufficient write")
		}
		if err := taskCtx.UpdateLiveness(); err != nil {
			return err
		}
		if currentRead-lastReportCount < reportSizeThreshold &&
			time.Now().Sub(lastReportTime) < reportDurationThreshold {
			continue
		}
		lastReportTime = time.Now()
		lastReportCount = currentRead
		progress := taskschedule.TaskStatus{
			Phase:   taskschedule.Processing,
			Message: "receiving the file from edge",
			Progress: uint(progressBeginReceive) +
				uint(float64(common.ProgressMax-progressBeginReceive)*(float64(currentRead)/float64(totalSize))),
		}
		if err := taskCtx.UpdateStatus(progress); err != nil {
			return err
		}
	}
	return nil
}

func createUploadProcess(w http.ResponseWriter, r *http.Request) (*uploadProcess, error) {
	taskId := r.Header.Get(headerTaskId)
	pkgSizeStr := r.Header.Get(headerPackageSize)
	sha256Checksum := r.Header.Get(headerSha256Checksum)
	if matched, err := regexp.MatchString(constants.SingleNodeTaskIdRegexpStr, taskId); err != nil || !matched {
		return nil, errors.New("invalid task id")
	}
	pkgSize, err := strconv.ParseInt(pkgSizeStr, common.BaseHex, common.BitSize64)
	if err != nil {
		return nil, err
	}
	if pkgSize > constants.LogUploadMaxSize || pkgSize <= 0 {
		return nil, errors.New("invalid package size")
	}
	const tokensLen = 3
	tokens := strings.Split(taskId, ".")
	if len(tokens) != tokensLen {
		return nil, errors.New("node serial number not found")
	}
	clientIps := r.Header["X-Forwarded-For"]
	if len(clientIps) < 1 {
		return nil, errors.New("client ip not found")
	}
	return &uploadProcess{
		httpRequest:    r,
		responseWriter: w,
		taskId:         taskId,
		packageSize:    pkgSize,
		sha256Checksum: sha256Checksum,
		serialNumber:   tokens[1],
		clientIP:       clientIps[0],
	}, nil
}

// HandleUpload handles the uploading of log files
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	p, err := createUploadProcess(w, r)
	if err != nil {
		hwlog.RunLog.Errorf("abort the upload request, check args failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := p.processUpload(); err != nil {
		hwlog.RunLog.Errorf("abort the upload request, reason: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(template.HTMLEscapeString(err.Error()))); err != nil {
			hwlog.RunLog.Errorf("failed to respond error to edge: %v", err)
		}
		hwlog.OpLog.Errorf("[edge(%s)@%s] %s %s failed", p.serialNumber, p.clientIP, r.Method, r.URL.Path)
		return
	}
	hwlog.OpLog.Infof("[edge(%s)@%s] %s %s success", p.serialNumber, p.clientIP, r.Method, r.URL.Path)
}
