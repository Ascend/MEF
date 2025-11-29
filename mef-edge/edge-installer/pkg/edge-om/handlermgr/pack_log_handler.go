// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package handlermgr
package handlermgr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

const (
	logPkgMaxSizeMB  = 200
	logPkgExpireTime = 10 * time.Minute
)

var (
	edgeOmTempDir    = filepath.Join(constants.LogCollectTempDir, constants.EdgeOm)
	edgeOmTempFile   = filepath.Join(edgeOmTempDir, constants.LogCollectTempFileName)
	edgeMainTempDir  = filepath.Join(constants.LogCollectTempDir, constants.EdgeMain)
	edgeMainTempFile = filepath.Join(edgeMainTempDir, constants.LogCollectTempFileName)
)

type packLogHandler struct {
	running    int32
	packLock   sync.Mutex
	cleanTimer *time.Timer
}

// Handle packLogHandler compress log to tar.gz
func (h *packLogHandler) Handle(*model.Message) error {
	swapped := atomic.CompareAndSwapInt32(&h.running, 0, 1)
	if !swapped {
		err := errors.New("log pack handler is busy")
		h.feedbackResult(err)
		return err
	}

	go func() {
		defer atomic.StoreInt32(&h.running, 0)
		if h.cleanTimer != nil {
			h.cleanTimer.Stop()
		}
		err := h.handle()
		h.feedbackResult(err)
		h.doClean(err == nil)
	}()
	return nil
}

func (h *packLogHandler) handle() error {
	h.packLock.Lock()
	defer h.packLock.Unlock()

	if err := prepareDirs(); err != nil {
		return fmt.Errorf("failed to prepare temp dirs, %v", err)
	}

	if err := h.doCollect(); err != nil {
		return fmt.Errorf("failed to collect logs, %v", err)
	}

	if err := h.doChangePermission(); err != nil {
		return fmt.Errorf("failed to change temp dir's permission, %v", err)
	}
	return nil
}

func (h *packLogHandler) doCollect() error {
	installRootDir, err := path.GetInstallRootDir()
	if err != nil {
		return err
	}
	logRootDir, err := path.GetLogRootDir(installRootDir)
	if err != nil {
		return fmt.Errorf("failed to get logger's root dir %v", err)
	}
	logBackupRootDir, err := path.GetLogBackupRootDir(installRootDir)
	if err != nil {
		return fmt.Errorf("failed to get logger's backup root dir, %v", err)
	}

	collectPathWhiteList := []string{edgeOmTempFile}
	collector := util.GetLogCollector(
		edgeOmTempFile,
		filepath.Join(logRootDir, constants.MEFEdgeLogName),
		filepath.Join(logBackupRootDir, constants.MEFEdgeLogBackupName),
		collectPathWhiteList)

	if _, err = collector.Collect(); err != nil {
		return fmt.Errorf("failed to collect log, %v", err)
	}
	hwlog.RunLog.Info("collect log successful")
	return nil
}

func (h *packLogHandler) doClean(delay bool) {
	clean := func() {
		h.packLock.Lock()
		defer h.packLock.Unlock()

		if _, err := cleanTempFiles(); err != nil {
			hwlog.RunLog.Errorf("failed to clean temp files, %v", err)
		}
	}
	if !delay {
		clean()
		return
	}
	h.cleanTimer = time.AfterFunc(logPkgExpireTime, clean)
}

func (h *packLogHandler) doChangePermission() error {
	if err := utils.SafeChmod(edgeOmTempFile, logPkgMaxSizeMB, constants.Mode400); err != nil {
		return fmt.Errorf("failed to change permission of log, %v", err)
	}
	if err := util.SetPathOwnerGroupToMEFEdge(edgeOmTempFile, false, true); err != nil {
		return fmt.Errorf("failed to change ownership of log, %v", err)
	}
	if err := fileutils.RenameFile(edgeOmTempFile, edgeMainTempFile); err != nil {
		return fmt.Errorf("failed to rename log, %v", err)
	}
	hwlog.RunLog.Info("change file permission successful")
	return nil
}

func (h *packLogHandler) feedbackResult(packErr error) {
	result := constants.OK
	if packErr != nil {
		result = packErr.Error()
		hwlog.RunLog.Errorf("failed to pack log, feedback to cloud: %v", packErr)
	} else {
		hwlog.RunLog.Info("pack log successful")
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("failed to create feedback message")
		return
	}
	msg.SetRouter(constants.EdgeOm, constants.InnerClient, constants.OptResponse, constants.ResPackLogResponse)
	if err = msg.FillContent(result); err != nil {
		hwlog.RunLog.Errorf("fill result into content failed: %v", err)
		return
	}
	if err = modulemgr.SendMessage(msg); err != nil {
		hwlog.RunLog.Errorf("failed to send sync message: %v", err)
		return
	}
	hwlog.RunLog.Info("feedback result successful")
}

func cleanTempFiles() (bool, error) {
	if !fileutils.IsExist(constants.LogCollectTempDir) {
		return false, nil
	}
	if err := cleanTempFilesInternal(); err != nil {
		return false, err
	}
	return true, nil
}

func cleanTempFilesInternal() error {
	rootDirFile, err := os.Open(constants.LogCollectTempDir)
	if err != nil {
		return err
	}
	defer rootDirFile.Close()
	rootDirChecker := fileutils.NewFileOwnerChecker(true, false, 0, 0)
	rootDirChecker.SetNext(fileutils.NewFileLinkChecker(false))
	rootDirChecker.SetNext(fileutils.NewFileModeChecker(true, constants.ModeUmask022, true, true))
	if err := rootDirChecker.Check(rootDirFile, constants.LogCollectTempDir); err != nil {
		return fmt.Errorf("failed to check root dir %s, %v", constants.LogCollectTempDir, err)
	}
	edgeOmDirFile, err := os.Open(edgeOmTempDir)
	if err != nil {
		return err
	}
	defer edgeOmDirFile.Close()
	edgeOmDirChecker := fileutils.NewFileOwnerChecker(false, false, 0, 0)
	edgeOmDirChecker.SetNext(fileutils.NewFileLinkChecker(false))
	edgeOmDirChecker.SetNext(fileutils.NewFileModeChecker(false, constants.ModeUmask077, true, true))
	if err := edgeOmDirChecker.Check(edgeOmDirFile, edgeOmTempDir); err != nil {
		return fmt.Errorf("failed to check edge om dir %s, %v", edgeOmTempDir, err)
	}
	uid, gid, err := getMefId()
	if err != nil {
		return err
	}
	edgeMainDirFile, err := os.Open(edgeMainTempDir)
	if err != nil {
		return err
	}
	defer edgeMainDirFile.Close()
	edgeMainDirChecker := fileutils.NewFileOwnerChecker(false, false, uid, gid)
	edgeMainDirChecker.SetNext(fileutils.NewFileLinkChecker(false))
	edgeMainDirChecker.SetNext(fileutils.NewFileModeChecker(false, constants.ModeUmask277, true, true))
	if err := edgeMainDirChecker.Check(edgeMainDirFile, edgeMainTempDir); err != nil {
		return fmt.Errorf("failed to check edge main dir %s, %v", edgeMainTempDir, err)
	}
	if err := fileutils.DeleteFile(filepath.Join(fileutils.GetFdPath(edgeOmDirFile),
		constants.LogCollectTempFileName), &fileutils.FileBaseChecker{}); err != nil {
		return fmt.Errorf("failed to delete edge om file %s, %v", edgeOmTempFile, err)
	}
	if err := fileutils.DeleteFile(filepath.Join(fileutils.GetFdPath(edgeMainDirFile),
		constants.LogCollectTempFileName), &fileutils.FileBaseChecker{}); err != nil {
		return fmt.Errorf("failed to delete edge main file %s, %v", edgeMainTempFile, err)
	}
	return nil
}

func getMefId() (uint32, uint32, error) {
	uid, err := envutils.GetUid(constants.EdgeUserName)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get MEFEdge uid, %v", err)
	}
	gid, err := envutils.GetUid(constants.EdgeUserGroup)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get MEFEdge gid, %v", err)
	}
	return uid, gid, err
}

func createTempDirs() error {
	if err := fileutils.CreateDir(constants.LogCollectTempDir, constants.Mode755); err != nil {
		return fmt.Errorf("failed to create root dir %s, %v", constants.LogCollectTempDir, err)
	}
	if _, err := fileutils.RealDirCheck(filepath.Dir(filepath.Dir(constants.LogCollectTempDir)), true, false); err != nil {
		return fmt.Errorf("failed to check dir %s, %v", filepath.Dir(filepath.Dir(constants.LogCollectTempDir)), err)
	}
	// /home/data/mef_logcollect
	if err := fileutils.SetPathPermission(
		constants.LogCollectTempDir, constants.Mode755, false, true); err != nil {
		return fmt.Errorf("failed to set permission of root dir %s, %v", constants.LogCollectTempDir, err)
	}
	// /home/data, we assume that the directory `/home` exists and has proper permission.
	if err := fileutils.SetPathPermission(
		filepath.Dir(constants.LogCollectTempDir), constants.Mode755, false, true); err != nil {
		return fmt.Errorf(
			"failed to set permission of root dir %s, %v", filepath.Dir(constants.LogCollectTempDir), err)
	}
	if err := fileutils.CreateDir(edgeOmTempDir, constants.Mode700); err != nil {
		return fmt.Errorf("failed to create edge om dir %s, %v", edgeOmTempDir, err)
	}
	if err := fileutils.CreateDir(edgeMainTempDir, constants.Mode500); err != nil {
		return fmt.Errorf("failed to create edge main dir %s, %v", edgeMainTempDir, err)
	}
	if err := util.SetPathOwnerGroupToMEFEdge(edgeMainTempDir, false, true); err != nil {
		return fmt.Errorf("failed to set ownership of edge main dir %s, %v", edgeMainTempDir, err)
	}
	return nil
}

func prepareDirs() error {
	exists, err := cleanTempFiles()
	if err != nil {
		return fmt.Errorf("failed to clean temp dir: %v", err)
	}
	if exists {
		return nil
	}
	if err := createTempDirs(); err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	hwlog.RunLog.Info("prepare upload dir successful")
	return nil
}

func initLogDumpDirs() error {
	if err := prepareDirs(); err != nil {
		hwlog.RunLog.Errorf("failed to init log dump dirs, %v", err)
	}
	// do not stop edge-om
	return nil
}
