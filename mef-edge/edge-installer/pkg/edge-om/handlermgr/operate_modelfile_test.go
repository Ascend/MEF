// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr for deal every handler
package handlermgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
)

var (
	testName                            = "test.om"
	testUuid                            = "447c3c43-4bbe-4dd1-b493-d6dc442b4c72"
	testUsedUuid                        = "4d7ca424-6624-4b92-892e-f4c45539affd"
	testModeFileDir                     = "/tmp/var_lib_docker"
	testModeFileActiveDir               = filepath.Join(testModeFileDir, "modelfile")
	testModeFileDownloadDir             = filepath.Join(testModeFileDir, "model_file_download")
	testSymlinkDir                      = filepath.Join(testModeFileDir, "model_file_symlink")
	testDownloadFile                    = filepath.Join(testModeFileDownloadDir, testUuid, testName)
	testUsedFile                        = filepath.Join(testModeFileActiveDir, testUsedUuid, testName)
	testInvalidMode         os.FileMode = 0777
)

var mockOperateModelFile = OperateModelFile{
	inactivePath: testModeFileDownloadDir,
	activePath:   testModeFileActiveDir,
	subDirUser:   constants.HwHiAiUser,
	rootDirUser:  constants.RootUserName,
}

func setupOperateModelFile() error {
	if err := fileutils.CreateDir(testModeFileDir, constants.Mode711); err != nil {
		hwlog.RunLog.Errorf("create test model file dir failed, error: %v", err)
		return err
	}
	return nil
}

func teardownOperateModelFile() {
	if err := os.RemoveAll(testModeFileDir); err != nil {
		hwlog.RunLog.Errorf("clear test model file dir failed, error: %v", err)
	}
}

func TestOperateModelFile(t *testing.T) {
	if err := setupOperateModelFile(); err != nil {
		hwlog.RunLog.Errorf("setup test operate model file environment failed, error: %v", err)
		return
	}
	defer teardownOperateModelFile()
	convey.Convey("test check model file", t, testCheckModelFile)
	convey.Convey("test update model file", t, testUpdateModelFile)
	convey.Convey("test delete model file", t, testDeleteModelFile)
	convey.Convey("test sync model file", t, testSyncFiles)
}

func testCheckModelFile() {
	convey.Convey("test check model file successful", checkModelFileSuccess)
	convey.Convey("test check model file failed", func() {
		convey.Convey("check and prepare path failed", checkAndPreparePathFailed)
		convey.Convey("check and prepare sub path failed", checkAndPrepareSubPathFailed)
	})
}

func testUpdateModelFile() {
	convey.Convey("test update model file successful", updateModelFileSuccess)
	convey.Convey("test update model file failed", func() {
		convey.Convey("check update operate info failed", checkUpdateOperateInfoFailed)
		convey.Convey("move file failed", moveFileFailed)
	})
}

func testDeleteModelFile() {
	convey.Convey("test delete model file successful", deleteModelFileSuccess)
}

func testSyncFiles() {
	convey.Convey("test sync model file successful", syncFilesSuccess)
}

func getOperateContent(operate string) (types.OperateModelFileContent, error) {
	operateContent := types.OperateModelFileContent{
		Operate:     "",
		OperateInfo: nil,
	}
	switch operate {
	case optInvalid:
		operateContent.Operate = optInvalid
	case constants.OptCheck:
		operateContent.Operate = constants.OptCheck
	case constants.OptUpdate:
		operateContent.Operate = constants.OptUpdate
		operateContent.OperateInfo = map[string]string{
			"uuid": testUuid,
			"name": testName,
		}
	case constants.OptDelete:
		operateContent.Operate = constants.OptDelete
	case constants.OptSync:
		fileList := []types.ModelBrief{
			{
				Uuid:   testUuid,
				Name:   testName,
				Status: "inactive",
			},
		}
		fileListBytes, err := json.Marshal(fileList)
		if err != nil {
			hwlog.RunLog.Errorf("marshal fileList failed, error: %v", err)
			return types.OperateModelFileContent{}, err
		}
		opInfo := map[string]string{"fileList": string(fileListBytes)}
		operateContent.Operate = constants.OptSync
		operateContent.OperateInfo = opInfo
		operateContent.UsedFiles = []string{testUsedFile}
		operateContent.CurrentUuid = testUsedUuid
	default:
		return types.OperateModelFileContent{}, errors.New("not support operate type")
	}
	return operateContent, nil
}

func checkModelFileSuccess() {
	operateContent, err := getOperateContent(constants.OptCheck)
	if err != nil {
		return
	}
	mockOperateModelFile.operateContent = operateContent
	p := gomonkey.ApplyFuncReturn(NewOperateModelFile, &mockOperateModelFile).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()
	defer clearTestDownloadDir()
	err = NewOperateModelFile(operateContent).OperateModelFile()
	convey.So(err, convey.ShouldBeNil)
}

func checkAndPreparePathFailed() {
	operateContent, err := getOperateContent(constants.OptCheck)
	if err != nil {
		return
	}
	convey.Convey("check model file root path failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", testErr)
		defer p.Reset()
		err = NewOperateModelFile(operateContent).OperateModelFile()
		convey.So(err, convey.ShouldResemble, errors.New("check model file root path failed"))
	})

	convey.Convey("set mode for model file root path failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
			ApplyFuncReturn(fileutils.SetPathPermission, testErr)
		defer p.Reset()
		err = NewOperateModelFile(operateContent).OperateModelFile()
		convey.So(err, convey.ShouldResemble, errors.New("set mode for model file root path failed"))
	})
}

func checkAndPrepareSubPathFailed() {
	operateContent, err := getOperateContent(constants.OptCheck)
	if err != nil {
		return
	}
	mockOperateModelFile.operateContent = operateContent
	expectErr := fmt.Errorf("check and prepare model file download path [%s] failed", testModeFileDownloadDir)
	p := gomonkey.ApplyFuncReturn(NewOperateModelFile, &mockOperateModelFile).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()

	convey.Convey("create model file download path failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CreateDir, testErr)
		defer p1.Reset()
		err = NewOperateModelFile(operateContent).OperateModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set download path owner and group failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.SetPathOwnerGroup, testErr)
		defer p2.Reset()
		defer clearTestDownloadDir()
		err = NewOperateModelFile(operateContent).OperateModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check existed download path symlink failed", func() {
		if err = createTestSymlinkDir(); err != nil {
			return
		}
		defer clearTestDownloadDir()
		err = NewOperateModelFile(operateContent).OperateModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check existed download path mode failed", func() {
		if err = createInvalidModeDir(); err != nil {
			return
		}
		defer clearTestDownloadDir()
		err = NewOperateModelFile(operateContent).OperateModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func updateModelFileSuccess() {
	operateContent, err := getOperateContent(constants.OptUpdate)
	if err != nil {
		return
	}
	if err = prepareDownloadFile(); err != nil {
		return
	}
	mockOperateModelFile.operateContent = operateContent
	p := gomonkey.ApplyFuncReturn(NewOperateModelFile, &mockOperateModelFile).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()
	err = NewOperateModelFile(operateContent).OperateModelFile()
	convey.So(err, convey.ShouldBeNil)
}

func checkUpdateOperateInfoFailed() {
	operateContent, err := getOperateContent(constants.OptUpdate)
	if err != nil {
		return
	}
	operateContent.OperateInfo["name"] = "~test.zip"
	err = NewOperateModelFile(operateContent).OperateModelFile()
	convey.So(err, convey.ShouldResemble, errors.New("check update operate info failed"))
}

func moveFileFailed() {
	operateContent, err := getOperateContent(constants.OptUpdate)
	if err != nil {
		return
	}
	mockOperateModelFile.operateContent = operateContent
	p := gomonkey.ApplyFuncReturn(NewOperateModelFile, &mockOperateModelFile).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()
	err = NewOperateModelFile(operateContent).OperateModelFile()
	convey.So(err, convey.ShouldResemble, errors.New("move model file failed"))
}

func deleteModelFileSuccess() {
	operateContent, err := getOperateContent(constants.OptDelete)
	if err != nil {
		return
	}
	if err = prepareDownloadFile(); err != nil {
		return
	}
	mockOperateModelFile.operateContent = operateContent
	p := gomonkey.ApplyFuncReturn(NewOperateModelFile, &mockOperateModelFile).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()
	err = NewOperateModelFile(operateContent).OperateModelFile()
	convey.So(err, convey.ShouldBeNil)
}

func syncFilesSuccess() {
	operateContent, err := getOperateContent(constants.OptSync)
	if err != nil {
		return
	}
	if err = prepareDownloadFile(); err != nil {
		return
	}
	mockOperateModelFile.operateContent = operateContent
	p := gomonkey.ApplyFuncReturn(NewOperateModelFile, &mockOperateModelFile).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()
	toDelList := NewOperateModelFile(operateContent).syncFiles()
	convey.So(toDelList, convey.ShouldResemble, modeltask.SyncList{FileList: []types.ModelBrief{{
		Uuid:   testUuid,
		Name:   testName,
		Status: "inactive",
	}}})
}

func createTestSymlinkDir() error {
	if err := fileutils.DeleteFile(testModeFileDownloadDir); err != nil {
		hwlog.RunLog.Errorf("clear existed download dir failed, error: %v", err)
		return err
	}
	if err := fileutils.CreateDir(testSymlinkDir, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("create test symlink dir failed, error: %v", err)
		return err
	}
	if err := os.Symlink(testSymlinkDir, testModeFileDownloadDir); err != nil {
		hwlog.RunLog.Errorf("create symlink failed, error: %v", err)
		return err
	}
	return nil
}

func createInvalidModeDir() error {
	mask := syscall.Umask(0)
	defer syscall.Umask(mask)
	if err := fileutils.CreateDir(testModeFileDownloadDir, testInvalidMode); err != nil {
		hwlog.RunLog.Errorf("create test invalid mode dir failed, error: %v", err)
		return err
	}
	return nil
}

func clearTestDownloadDir() {
	if err := fileutils.DeleteFile(testModeFileDownloadDir); err != nil {
		hwlog.RunLog.Errorf("clear test download dir failed, error: %v", err)
		return
	}
}

func prepareDownloadFile() error {
	modelFiles := []string{testDownloadFile, testUsedFile}
	for _, file := range modelFiles {
		if err := fileutils.MakeSureDir(file); err != nil {
			hwlog.RunLog.Errorf("create dir failed, error: %v", err)
			return err
		}
		if err := fileutils.CreateFile(file, constants.Mode600); err != nil {
			hwlog.RunLog.Errorf("create file failed, error: %v", err)
			return err
		}
	}
	return nil
}
