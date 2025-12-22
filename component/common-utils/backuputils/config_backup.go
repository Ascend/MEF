// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
//	http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package backuputils do backup or recovery jobs for config or crt files.
// Attention, all updated backups' permission will be set to 0600, and .tmp file will not be effected.
package backuputils

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

// YamlFileType and the followings are files' type extension for backup
const (
	YamlFileType = ".yaml"
	JsonFileType = ".json"
	CrtFileType  = ".crt"
	CrlFileType  = ".crl"
	KeyFileType  = ".key"

	TmpFileType = ".tmp"
)

const (
	// BackupSuffix is the file extension of backup-type file
	BackupSuffix = ".backup"
	copyCmd      = "cp"
	copyTimeout  = 1
	splitFlag    = "\n</=--*^^||^^--*=/>"
)

var (
	lock             sync.Mutex
	mode400FileTypes = map[string]struct{}{
		CrtFileType: {},
		CrlFileType: {},
		KeyFileType: {},
	}
)

// BackupMgr [interface] for backup and recovery
type BackupMgr interface {
	BackUp() error
	Restore() error
}

// NewBackupFileMgr return BackupMgr implement for single file backup and recovery. Soft link is not supported
func NewBackupFileMgr(filePath string) BackupMgr {
	backupMgr := backupFileImpl{
		filePath:   filePath,
		backupPath: filePath + BackupSuffix,
	}

	return &backupMgr
}

type backupFileImpl struct {
	filePath   string
	backupPath string
}

// BackUp [method] back up single file from main path
func (bf *backupFileImpl) BackUp() error {
	lock.Lock()
	defer lock.Unlock()

	if filepath.Ext(bf.filePath) == TmpFileType {
		return nil
	}
	if !isExistFile(bf.filePath) {
		return errors.New("file path is not a file or do not exist")
	}

	fileData, err := fileutils.LoadFile(bf.filePath)
	if err != nil {
		return fmt.Errorf("load file error: %v", err)
	}
	checksumValue, err := fileutils.GetSha256Bytes(fileData)
	if err != nil {
		return fmt.Errorf("get file checksum error: %v", err)
	}
	backupBytes := append(append(fileData, []byte(splitFlag)...), checksumValue...)

	if err = makeSureFileWritePermission(bf.backupPath); err != nil {
		return fmt.Errorf("make sure backup path permission error: %v", err)
	}
	if err = fileutils.WriteData(bf.backupPath, backupBytes); err != nil {
		return fmt.Errorf("create file backup error: %v", err)
	}
	return nil
}

// Restore [method] restore single file from backup
func (bf *backupFileImpl) Restore() error {
	lock.Lock()
	defer lock.Unlock()

	if filepath.Ext(bf.filePath) == TmpFileType {
		return errors.New(".tmp file don't have backup")
	}
	if !isExistFile(bf.backupPath) {
		return errors.New("backup path is not a file or do not exist")
	}
	fullData, err := fileutils.LoadFile(bf.backupPath)
	if err != nil {
		return fmt.Errorf("load backup file error: %v", err)
	}
	index := bytes.LastIndex(fullData, []byte(splitFlag))
	if index == -1 {
		return fmt.Errorf("no checksum found in backup file")
	}
	fileData := fullData[0:index]
	recordChecksum := fullData[index+len(splitFlag):]
	checksum, err := fileutils.GetSha256Bytes(fileData)
	if err != nil {
		return fmt.Errorf("get file checksum error: %v", err)
	}
	if string(checksum) != string(recordChecksum) {
		return errors.New("check checksum of backup file failed")
	}

	if err = makeSureFileWritePermission(bf.filePath); err != nil {
		return fmt.Errorf("make sure main file permission error: %v", err)
	}
	if err = fileutils.WriteData(bf.filePath, fileData); err != nil {
		return fmt.Errorf("create file backup error: %v", err)
	}

	return bf.resetRestoredFilePerm()
}

func (bf *backupFileImpl) resetRestoredFilePerm() error {
	// reset [.crl .crt .key] file to mod 0400
	if len(mode400FileTypes) == 0 {
		return errors.New("mode map is nil")
	}
	_, ok := mode400FileTypes[filepath.Ext(bf.filePath)]
	if !ok {
		return nil
	}
	if err := fileutils.SetPathPermission(bf.filePath, fileutils.Mode400, false,
		false); err != nil {
		return fmt.Errorf("reset file mode error: %v", err)
	}
	return nil
}

// NewBackupDirMgr return BackupMgr implement for dir backup and recovery. Soft link is not supported.
// fileTypes decides which types of file will be effected, and if it's empty, all files will be effected.
func NewBackupDirMgr(fileDir string, fileTypes ...string) BackupMgr {
	fileTypesTable := make(map[string]struct{}, len(fileTypes))
	for _, fileType := range fileTypes {
		fileTypesTable[fileType] = struct{}{}
	}
	backupMgr := backupDirImpl{
		fileDir:   fileDir,
		fileTypes: fileTypesTable,
	}
	return &backupMgr
}

type backupDirImpl struct {
	fileDir   string
	fileTypes map[string]struct{}
}

// BackUp [method] back up dir from main dir
func (bd *backupDirImpl) BackUp() error {
	if !fileutils.IsDir(bd.fileDir) {
		return errors.New("backup dir failed, path of is not a dir")
	}
	return filepath.Walk(bd.fileDir, bd.walkForBackup)
}

func (bd *backupDirImpl) walkForBackup(subPath string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() || !bd.checkTypes(filepath.Ext(subPath)) {
		return nil
	}
	// ignore single failed case
	if err = NewBackupFileMgr(subPath).BackUp(); err != nil {
		hwlog.RunLog.Errorf(err.Error())
		return filepath.SkipDir
	}
	return nil
}

// Restore [method] restore dir from backup
func (bd *backupDirImpl) Restore() error {
	if !fileutils.IsDir(bd.fileDir) {
		return errors.New("restore dir failed, path is not a dir")
	}
	return filepath.Walk(bd.fileDir, bd.walkForRestore)
}

func (bd *backupDirImpl) walkForRestore(subPath string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() || filepath.Ext(subPath) != BackupSuffix {
		return nil
	}

	fileName := strings.TrimSuffix(filepath.Base(subPath), BackupSuffix)
	if !bd.checkTypes(filepath.Ext(fileName)) {
		return nil
	}

	// ignore single failed case
	if err = NewBackupFileMgr(fileName).Restore(); err != nil {
		hwlog.RunLog.Errorf(err.Error())
		return filepath.SkipDir
	}
	return nil
}

func (bd *backupDirImpl) checkTypes(ext string) bool {
	if len(bd.fileTypes) == 0 {
		return true
	}
	_, exist := bd.fileTypes[ext]
	return exist
}

func isExistFile(filePath string) bool {
	return fileutils.IsLexist(filePath) && fileutils.IsFile(filePath)
}

func makeSureFileWritePermission(filePath string) error {
	if !fileutils.IsLexist(filePath) {
		return nil
	}
	return fileutils.SetPathPermission(filePath, fileutils.Mode600, false, false)
}

type tryInitFunc func(filepath string) error

// InitConfig [method] try init config. It will exec tryInitFuc one or two times.
// If the first execution is successful, it will save file and checksum to backup-path.
// If the first execution is failed, it will restore file from backup and try again.
func InitConfig(filepath string, tryInit tryInitFunc) error {
	var initErr error
	if initErr = tryInit(filepath); initErr == nil {
		if backupErr := NewBackupFileMgr(filepath).BackUp(); backupErr != nil {
			hwlog.RunLog.Warnf("backup config [%s] failed, error:[%v], please check backup files manually",
				filepath, backupErr)
		}
		return nil
	}

	hwlog.RunLog.Warnf("init config [%s] failed, %v, try restore from backup", filepath, initErr)
	if err := NewBackupFileMgr(filepath).Restore(); err != nil {
		hwlog.RunLog.Errorf("restore config [%s] from backup failed, error: %v", filepath, err)
		return err
	}

	if initErr = tryInit(filepath); initErr != nil {
		hwlog.RunLog.Errorf("init config [%s] failed after recovery, %v", filepath, initErr)
		return fmt.Errorf("init config [%s] failed after recovery, %v", filepath, initErr)
	}
	return nil
}

// BackUpFiles provides function which can be directly invoked for files' back up.
func BackUpFiles(filePaths ...string) error {
	var errFiles []string
	for _, file := range filePaths {
		backErr := NewBackupFileMgr(file).BackUp()
		if backErr != nil {
			hwlog.RunLog.Errorf("back up file failed, %v", backErr)
			errFiles = append(errFiles, file)
		}
	}
	if len(errFiles) != 0 {
		return fmt.Errorf("back up [%v] failed", errFiles)
	}
	return nil
}

// RestoreFiles provides function which can be directly invoked for files' recovery.
func RestoreFiles(filePaths ...string) error {
	var errFiles []string
	for _, file := range filePaths {
		restoreErr := NewBackupFileMgr(file).Restore()
		if restoreErr != nil {
			hwlog.RunLog.Errorf("restore file failed, %v", restoreErr)
			errFiles = append(errFiles, file)
		}
	}
	if len(errFiles) != 0 {
		return fmt.Errorf("restore [%v] failed", errFiles)
	}
	return nil
}
