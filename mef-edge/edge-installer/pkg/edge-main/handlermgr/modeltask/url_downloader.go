// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package modeltask

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

const progressReportInterval = time.Second * 5

// UrlDownloader struct for download
type UrlDownloader struct {
	uuid           string
	name           string
	url            string
	savePath       string
	checkCode      string
	checkType      string
	userName       string
	passWord       string
	downloadedSize int
	filesize       int
	ca             []byte
	cancelCtx      context.Context
}

// NewUrlDownloader create a downloader
func NewUrlDownloader(uuid, savePath string, size int, m types.ModelFile, ctx context.Context) UrlDownloader {
	return UrlDownloader{
		uuid:      uuid,
		name:      m.Name,
		url:       strings.Split(m.FileServer.Path, " ")[1],
		savePath:  savePath,
		checkCode: m.CheckCode,
		checkType: m.CheckType,
		userName:  m.FileServer.UserName,
		passWord:  m.FileServer.PassWord,
		filesize:  size,
		cancelCtx: ctx,
	}
}

func (u *UrlDownloader) setCa(ca []byte) {
	u.ca = ca
}

func (u *UrlDownloader) download() {
	defer utils.ClearSliceByteMemory(u.ca)
	defer utils.ClearStringMemory(u.passWord)
	fileWriter, err := os.OpenFile(u.savePath, os.O_RDWR|os.O_CREATE, utils.FileMode)
	if err != nil {
		downFinishEvent := NewDownloadFinishEvent(u.uuid, u.name,
			"cannot create file when download", false)
		GetModelMgr().Notify(downFinishEvent)
		return
	}
	defer func() {
		err = fileWriter.Close()
		if err != nil {
			hwlog.RunLog.Error("close file error")
			return
		}
	}()
	pWriter := &ProgressWriter{writer: fileWriter, downloader: u, lastReportTime: time.Now()}
	tlsCfg := certutils.TlsCertInfo{
		RootCaContent: u.ca,
		RootCaOnly:    true,
		WithBackup:    true,
	}
	comStr := fmt.Sprintf("%s:%s", u.userName, u.passWord)
	defer utils.ClearStringMemory(comStr)
	comStrAfterEnc := base64.StdEncoding.EncodeToString([]byte(comStr))
	defer utils.ClearStringMemory(comStrAfterEnc)
	authStr := fmt.Sprintf("Basic %s", comStrAfterEnc)
	defer utils.ClearStringMemory(authStr)
	headers := make(map[string]interface{})
	headers["Authorization"] = authStr
	defer delete(headers, "Authorization")
	err = httpsmgr.GetHttpsReq(u.url, tlsCfg, headers).
		SetReadTimeout(time.Hour).GetRespToFileWithLimit(pWriter, constants.ModelFileMaxSize)
	if err != nil {
		hwlog.RunLog.Error("download from url failed")
		downFinishEvent := NewDownloadFinishEvent(u.uuid, u.name, "download from url failed", false)
		GetModelMgr().Notify(downFinishEvent)
		return
	}
	if !u.checkFileValid(u.savePath, u.checkCode, u.checkType, u.filesize) {
		downFinishEvent := NewDownloadFinishEvent(u.uuid, u.name, "check download file valid fail", false)
		GetModelMgr().Notify(downFinishEvent)
		return
	}
	downFinishEvent := NewDownloadFinishEvent(u.uuid, u.name, "", true)
	GetModelMgr().Notify(downFinishEvent)
	return
}

func (u *UrlDownloader) updateDownloadedSize(down int) {
	u.downloadedSize = down
}

func (u *UrlDownloader) isTaskCanceled() bool {
	select {
	case <-u.cancelCtx.Done():
		return true
	default:
		return false
	}
}

// ProgressWriter a writer which can report progress
type ProgressWriter struct {
	writer         io.Writer
	downloader     *UrlDownloader
	lastReportTime time.Time
}

func (l *ProgressWriter) Write(p []byte) (int, error) {
	if l.downloader.isTaskCanceled() {
		return 0, fmt.Errorf("task canceled by user")
	}
	n, err := l.writer.Write(p)
	l.downloader.updateDownloadedSize(l.downloader.downloadedSize + n)
	if l.downloader.downloadedSize > l.downloader.filesize {
		hwlog.RunLog.Errorf("downloaded file size %d, exceed wanted file size: %d",
			l.downloader.downloadedSize, l.downloader.filesize)
		return n, fmt.Errorf("downloaded file size %d, exceed wanted file size: %d",
			l.downloader.downloadedSize, l.downloader.filesize)
	}
	now := time.Now()
	if now.Sub(l.lastReportTime) >= progressReportInterval {
		l.lastReportTime = now
		progressEvent := NewProgressEvent(l.downloader.uuid, l.downloader.name, l.downloader.downloadedSize)
		GetModelMgr().Notify(progressEvent)
	}
	return n, err
}

func (u *UrlDownloader) checkFileValid(filepath, checkCode, checkType string, filesize int) bool {
	if filesize != u.downloadedSize {
		hwlog.RunLog.Errorf("filesize not right: want %d, downloaded %d", filesize, u.downloadedSize)
		return false
	}
	if !fileutils.IsExist(filepath) {
		hwlog.RunLog.Errorf("%s file not exist", filepath)
		return false
	}
	if checkType != constants.ModeFileCheckAgl {
		hwlog.RunLog.Errorf("check type not right: %s", checkType)
		return false
	}
	file, err := os.Open(filepath)
	if err != nil {
		hwlog.RunLog.Errorf("cannot open file %s to cal sha256", filepath)
		return false
	}
	defer func() {
		if err = file.Close(); err != nil {
			hwlog.RunLog.Errorf("failed to close  file, %v", err)
		}
	}()
	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		hwlog.RunLog.Errorf("failed to calculate sha256 checksum, %v", err)
		return false
	}
	if fmt.Sprintf("%x", hash.Sum(nil)) != checkCode {
		hwlog.RunLog.Error("file check code not right")
		return false
	}
	return true
}
