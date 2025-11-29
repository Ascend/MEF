// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr for deal every handler
package handlermgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
)

type updateModelFileOperateInfo struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
}

// OperateModelFile operate model file struct
type OperateModelFile struct {
	inactivePath   string
	activePath     string
	subDirUser     string
	rootDirUser    string
	operateContent types.OperateModelFileContent
}

// NewOperateModelFile create a new OperateModelFile struct
func NewOperateModelFile(operateContent types.OperateModelFileContent) *OperateModelFile {
	return &OperateModelFile{
		inactivePath:   constants.ModeFileDownloadDir,
		activePath:     constants.ModeFileActiveDir,
		subDirUser:     constants.HwHiAiUser,
		rootDirUser:    constants.RootUserName,
		operateContent: operateContent,
	}
}

// OperateModelFile operate model file
func (o *OperateModelFile) OperateModelFile() error {
	switch o.operateContent.Operate {
	case constants.OptCheck:
		return o.checkAndPreparePath()
	case constants.OptUpdate:
		return o.updateFile()
	case constants.OptDelete:
		return o.deleteAllModelFile()
	default:
		hwlog.RunLog.Error("operate model file failed, not support operate type")
		return errors.New("operate model file failed, not support operate type")
	}
}

func (o *OperateModelFile) checkAndPreparePath() error {
	rootPath := constants.ModelFileRootPath
	if _, err := fileutils.RealDirCheck(rootPath, true, false); err != nil {
		hwlog.RunLog.Errorf("check model file root path [%s] failed, error: %v", rootPath, err)
		return errors.New("check model file root path failed")
	}

	if err := fileutils.SetPathPermission(rootPath, constants.Mode711, false, true); err != nil {
		hwlog.RunLog.Errorf("set mode for model file root path [%s] failed, error: %v", rootPath, err)
		return errors.New("set mode for model file root path failed")
	}

	if err := o.checkAndPrepareSubPath(o.inactivePath, constants.EdgeUserName); err != nil {
		hwlog.RunLog.Errorf("check and prepare model file download path [%s] failed, error: %v",
			o.inactivePath, err)
		return fmt.Errorf("check and prepare model file download path [%s] failed", o.inactivePath)
	}

	if err := o.checkAndPrepareSubPath(o.activePath, constants.RootUserName); err != nil {
		hwlog.RunLog.Errorf("check and prepare model file path [%s] failed, error: %v", o.activePath, err)
		return fmt.Errorf("check and prepare model file path [%s] failed", o.activePath)
	}

	hwlog.RunLog.Info("check and prepare model file paths success")
	return nil
}

func (o *OperateModelFile) checkAndPrepareSubPath(path string, user string) error {
	userUid, err := envutils.GetUid(user)
	if err != nil {
		hwlog.RunLog.Errorf("get uid of user [%s] failed, error: %v", user, err)
		return errors.New("get uid of user failed")
	}
	userGid, err := envutils.GetGid(user)
	if err != nil {
		hwlog.RunLog.Errorf("get gid of user [%s] failed, error: %v", user, err)
		return errors.New("get gid of user failed")
	}

	if fileutils.IsExist(path) {
		if _, err = fileutils.CheckOriginPath(path); err != nil {
			hwlog.RunLog.Errorf("check path [%s] failed, error: %v", path, err)
			return fmt.Errorf("check path [%s] failed", path)
		}
		if _, err = fileutils.CheckOwnerAndPermission(path, constants.ModeUmask077, userUid); err != nil {
			hwlog.RunLog.Errorf("check path [%s] owner and permission failed, error: %v", path, err)
			return fmt.Errorf("check path [%s] owner and permission failed", path)
		}
		return nil
	}

	if err = fileutils.CreateDir(path, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("create model file path [%s] failed, error: %v", path, err)
		return errors.New("create model file path failed")
	}

	param := fileutils.SetOwnerParam{
		Path:       path,
		Uid:        userUid,
		Gid:        userGid,
		Recursive:  false,
		IgnoreFile: true,
	}
	if err = fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed, error: %v", path, err)
		return errors.New("set path owner and group failed")
	}
	return nil
}

func (o *OperateModelFile) syncFiles() modeltask.SyncList {
	hwlog.RunLog.Info("start synchronizing model files")
	toDelFileList := o.syncFile("fileList")
	delList := modeltask.SyncList{FileList: toDelFileList}
	return delList
}

func (o *OperateModelFile) syncFile(key string) []types.ModelBrief {
	var fileList []types.ModelBrief
	fileListStr, ok := o.operateContent.OperateInfo[key]
	if !ok {
		return fileList
	}
	err := json.Unmarshal([]byte(fileListStr), &fileList)
	if err != nil {
		hwlog.RunLog.Error("receive bad sync list, skip this sync")
		return []types.ModelBrief{}
	}

	if err = o.checkAndPreparePath(); err != nil {
		hwlog.RunLog.Errorf("check path failed before synchronizing model files, error: %v", err)
		return []types.ModelBrief{}
	}

	o.cleanAllUnusedFiles(o.operateContent.UsedFiles, o.operateContent.CurrentUuid)
	o.cleanRedundantFiles(fileList)
	o.cleanEmptyDirs(o.operateContent.CurrentUuid)
	edgeMainDeleteFiles := o.findMissFiles(fileList)

	return edgeMainDeleteFiles
}

func (o *OperateModelFile) findMissFiles(fileList []types.ModelBrief) []types.ModelBrief {
	var edgeMainDeleteFiles []types.ModelBrief
	for _, brief := range fileList {
		if brief.Status == types.StatusActive.String() &&
			!fileutils.IsExist(filepath.Join(o.activePath, brief.Uuid, brief.Name, brief.Name)) {
			edgeMainDeleteFiles = append(edgeMainDeleteFiles, brief)
		} else if brief.Status == types.StatusInactive.String() &&
			!fileutils.IsExist(filepath.Join(o.inactivePath, brief.Uuid, brief.Name)) {
			edgeMainDeleteFiles = append(edgeMainDeleteFiles, brief)
		}
	}
	return edgeMainDeleteFiles
}

func queryOneLevelDirList(dir string) ([]string, error) {
	if !fileutils.IsExist(dir) {
		return []string{}, nil
	}
	reader, dirEntries, err := fileutils.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	defer func() {
		fileutils.CloseFile(reader)
	}()

	if len(dirEntries) > maxFileCount {
		return nil, errors.New("dir num up to limit")
	}

	var pathList []string
	for _, dirEntry := range dirEntries {
		pathList = append(pathList, filepath.Join(dir, dirEntry.Name()))
	}
	return pathList, nil
}

func (o *OperateModelFile) queryModelFile(dir string) (*utils.Set, error) {
	var fileSet = utils.NewSet()
	pathList, err := queryOneLevelDirList(dir)
	if err != nil {
		return fileSet, err
	}
	for _, path := range pathList {
		if fileutils.IsFile(path) {
			fileSet.Add(path)
			continue
		}

		if !fileutils.IsDir(path) {
			continue
		}

		tmpPaths, err := queryOneLevelDirList(path)
		if err != nil {
			return fileSet, err
		}

		for _, tmpPath := range tmpPaths {
			fileSet.Add(tmpPath)
		}
	}

	return fileSet, nil
}

func (o *OperateModelFile) cleanEmptyDirs(currentUuid string) {
	var allDirs []string
	activeDirs, err := queryOneLevelDirList(o.activePath)
	if err != nil {
		hwlog.RunLog.Error("query active dirs failed")
		return
	}
	allDirs = append(allDirs, activeDirs...)
	notActiveDirs, err := queryOneLevelDirList(o.inactivePath)
	if err != nil {
		hwlog.RunLog.Error("query not active dirs failed")
		return
	}
	allDirs = append(allDirs, notActiveDirs...)
	for _, dir := range allDirs {
		if strings.Contains(dir, currentUuid) {
			continue
		}
		if ok, err := fileutils.IsEmptyDir(dir); err != nil || !ok {
			hwlog.RunLog.Infof("dir %s not empty, skip it", dir)
			continue
		}
		hwlog.RunLog.Infof("delete empty dir : %s", dir)
		if err = fileutils.DeleteAllFileWithConfusion(dir); err != nil {
			hwlog.RunLog.Warnf("delete empty dir [%v] failed: %v", dir, err)
		}
	}
}

func (o *OperateModelFile) cleanUnusedFiles(filesInMemory *utils.Set, filesInDevice *utils.Set,
	currentUuid string) {
	notUsedFiles := filesInDevice.Difference(filesInMemory)
	for _, file := range notUsedFiles.List() {
		if len(currentUuid) > 0 && strings.Contains(file, currentUuid) {
			continue
		}
		hwlog.RunLog.Infof("delete unused model file : %s", file)
		if err := fileutils.DeleteAllFileWithConfusion(file); err != nil {
			hwlog.RunLog.Warnf("delete unused model file [%v] failed: %v", file, err)
		}
	}
}

func (o *OperateModelFile) cleanAllUnusedFiles(usedFiles []string, currentUuid string) {
	activeFilesInMemory := utils.NewSet()
	notActiveFilesInMemory := utils.NewSet()
	for _, file := range usedFiles {
		activeFilesInMemory.Add(file)
		notActivePath := strings.Replace(file, o.activePath, o.inactivePath, 1)
		notActiveFilesInMemory.Add(notActivePath)
	}

	actives, err := o.queryModelFile(o.activePath)
	if err != nil {
		hwlog.RunLog.Warnf("query active dir %s failed", o.activePath)
		return
	}
	activeFilesInDevice := utils.NewSet(actives.List()...)

	notActives, err := o.queryModelFile(o.inactivePath)
	if err != nil {
		hwlog.RunLog.Warnf("query not active dir %s failed", o.inactivePath)
		return
	}
	notActiveFilesInDevice := utils.NewSet(notActives.List()...)

	o.cleanUnusedFiles(activeFilesInMemory, activeFilesInDevice, "")
	o.cleanUnusedFiles(notActiveFilesInMemory, notActiveFilesInDevice, currentUuid)
}

func (o *OperateModelFile) cleanRedundantFiles(fileList []types.ModelBrief) {
	modelFilesInCache := utils.NewSet()
	for _, file := range fileList {
		if file.Status == types.StatusActive.String() {
			modelFilesInCache.Add(filepath.Join(o.activePath, file.Uuid, file.Name))
		} else {
			modelFilesInCache.Add(filepath.Join(o.inactivePath, file.Uuid, file.Name))
		}
	}

	activeModelFiles, err := o.queryModelFile(o.activePath)
	if err != nil {
		hwlog.RunLog.Warnf("query active model file failed: %v", err)
		return
	}

	inactiveModelFiles, err := o.queryModelFile(o.inactivePath)
	if err != nil {
		hwlog.RunLog.Warnf("query not active model file failed: %v", err)
		return
	}

	modelFilesOnDevice := utils.NewSet()
	modelFilesOnDevice.Add(activeModelFiles.List()...)
	modelFilesOnDevice.Add(inactiveModelFiles.List()...)

	for _, file := range modelFilesOnDevice.List() {
		if modelFilesInCache.Find(file) {
			continue
		}
		hwlog.RunLog.Infof("delete redundant file [%s]", file)
		if err = fileutils.DeleteAllFileWithConfusion(file); err != nil {
			hwlog.RunLog.Warnf("delete file [%v] failed: %v", file, err)
		}
	}
}

func (o *OperateModelFile) deleteAllModelFile() error {
	hwlog.RunLog.Info("start deleting all model files")
	if err := o.checkAndPreparePath(); err != nil {
		hwlog.RunLog.Errorf("check path failed before deleting all model files, error: %v", err)
		return errors.New("check model file path failed")
	}

	activeModelFiles, err := queryOneLevelDirList(o.activePath)
	if err != nil {
		hwlog.RunLog.Errorf("query active model file failed: %v", err)
		return fmt.Errorf("query active model file failed")
	}
	inactiveModelFiles, err := queryOneLevelDirList(o.inactivePath)
	if err != nil {
		hwlog.RunLog.Errorf("query not active model file failed: %v", err)
		return fmt.Errorf("query not active model file failed")
	}
	modelFilesOnDevice := utils.NewSet()
	modelFilesOnDevice.Add(activeModelFiles...)
	modelFilesOnDevice.Add(inactiveModelFiles...)
	for _, modelFile := range modelFilesOnDevice.List() {
		if err = fileutils.DeleteAllFileWithConfusion(modelFile); err != nil {
			hwlog.RunLog.Warnf("delete file [%v] failed: %v", modelFile, err)
		}
	}
	return nil
}

func (o *OperateModelFile) updateFile() error {
	hwlog.RunLog.Info("start updating model file")
	var operateInfo updateModelFileOperateInfo
	opInfo, err := json.Marshal(o.operateContent.OperateInfo)
	if err != nil {
		hwlog.RunLog.Errorf("marshal operate info failed: %v", err)
		return errors.New("marshal operate info failed")
	}
	if err = json.Unmarshal(opInfo, &operateInfo); err != nil {
		hwlog.RunLog.Errorf("parse update operate info failed, error: %v", err)
		return errors.New("parse update operate info failed")
	}

	operateInfoChecker := checker.GetAndChecker(
		checker.GetRegChecker("Uuid", "^"+constants.UUIDRegex+"$", true),
		checker.GetRegChecker("Name", constants.ModelFileNameReg, true),
	)
	if checkResult := operateInfoChecker.Check(operateInfo); !checkResult.Result {
		hwlog.RunLog.Errorf("check update operate info failed, error: %s", checkResult.Reason)
		return errors.New("check update operate info failed")
	}

	if err := o.checkAndPreparePath(); err != nil {
		hwlog.RunLog.Errorf("check and prepare path failed before updating model file, error: %v", err)
		return errors.New("check and prepare model file path failed")
	}

	srcFile := filepath.Join(o.inactivePath, operateInfo.Uuid, operateInfo.Name)
	dstFile := filepath.Join(o.activePath, operateInfo.Uuid, operateInfo.Name, operateInfo.Name)
	if err := o.moveFile(srcFile, dstFile); err != nil {
		hwlog.RunLog.Errorf("move model file failed, error: %v", err)
		return errors.New("move model file failed")
	}

	uuidPath := filepath.Join(o.activePath, operateInfo.Uuid)
	if err := fileutils.SetPathPermission(uuidPath, constants.Mode600, true, false); err != nil {
		hwlog.RunLog.Errorf("set mode for files in path [%s] failed, error: %v", uuidPath, err)
		return errors.New("set mode for files failed")
	}
	if err := fileutils.SetPathPermission(uuidPath, constants.Mode700, true, true); err != nil {
		hwlog.RunLog.Errorf("set mode for directories in path [%s] failed, error: %v", uuidPath, err)
		return errors.New("set mode for directories failed")
	}

	hwlog.RunLog.Info("update model file success")
	return nil
}

func (o *OperateModelFile) moveFile(srcFile, dstFile string) error {
	if err := fileutils.DeleteFile(dstFile); err != nil {
		hwlog.RunLog.Errorf("remove existed model file [%s] failed, error: %v", dstFile, err)
		return err
	}
	if err := fileutils.MakeSureDir(dstFile); err != nil {
		hwlog.RunLog.Errorf("create directory [%s] failed, error: %v", filepath.Dir(dstFile), err)
		return err
	}
	if err := fileutils.RenameFile(srcFile, dstFile); err != nil {
		hwlog.RunLog.Errorf("move model file from [%s] to [%s] failed, error: %v", srcFile, dstFile, err)
		return err
	}

	hwHiAiUserUid, err := envutils.GetUid(constants.HwHiAiUser)
	if err != nil {
		hwlog.RunLog.Errorf("get uid of user [%s] failed, error: %v", constants.HwHiAiUser, err)
		return err
	}
	hwHiAiUserGid, err := envutils.GetGid(constants.HwHiAiUser)
	if err != nil {
		hwlog.RunLog.Errorf("get gid of user [%s] failed, error: %v", constants.HwHiAiUser, err)
		return err
	}

	uuidDir := filepath.Dir(filepath.Dir(dstFile))
	param := fileutils.SetOwnerParam{
		Path:       uuidDir,
		Uid:        hwHiAiUserUid,
		Gid:        hwHiAiUserGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	if err = fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set owner and group for directory [%s] failed, error: %v", uuidDir, err)
		return err
	}
	return nil
}
