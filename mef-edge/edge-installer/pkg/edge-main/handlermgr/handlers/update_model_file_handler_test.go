// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlers for testing update model file handler
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/database"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
)

const (
	testUpdateModelFileMsg = `{
    "header":{
        "msg_id":"05843e95-9069-45a7-93f2-25302be579d9",
        "timestamp":1690278946,
        "sync":true
    },
    "route":{
        "source":"controller",
        "group":"hardware",
        "operation":"update",
        "resource":"websocket/modelfiles"
    },
    "content":{}
}`
	testUpdateContent = `{
    "operation": "update",
    "target": "all",
    "uuid": "bdf3242b-aec1-4100-af91-afa2b8fde88a",
    "modelfiles": [{
        "name": "module.om",
        "version": "1.0",
        "check_type": "sha256",
        "check_code": "XXXX",
        "size": "1024",
        "file_server": {
            "protocol": "https",
            "path": "GET https://FDAddr:port/models",
            "user_name": "userName",
            "password": "password"
        }
    }]
}`
	mockPodStatus = `{
          "kind":  "Pod",
          "spec":  {
				"volumes":  [{
						"hostPath":  {
							  "path":  "/var/lib/docker/modelfile/dd778847-a956-47d6-bc60-70b1af763c34/test.zip"
						}
				}]
          }
}`
)

var (
	updateModelFileMsg model.Message
	updateModelFile    = updateModelFileHandler{}
	testPodMetas       = []database.Meta{
		{
			Key:   constants.ResourceTypePod,
			Type:  constants.ResourceTypePod,
			Value: mockPodStatus,
		},
	}
	expectErr = errors.New("operate model file failed")
)

func setupUpdateModelFileHandler() error {
	if err := json.Unmarshal([]byte(testUpdateModelFileMsg), &updateModelFileMsg); err != nil {
		hwlog.RunLog.Errorf("unmarshal test update model file handler message failed, error: %v", err)
		return err
	}
	return nil
}

func TestUpdateModelFileHandler(t *testing.T) {
	if err := setupUpdateModelFileHandler(); err != nil {
		panic(err)
	}

	p := gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, nil).
		ApplyFuncReturn(modeltask.GetModelMgr, &modeltask.ModelMgr{})
	defer p.Reset()

	convey.Convey("test update model file handler, test update model file", t, testUpdateModelFile)
	convey.Convey("test update model file handler, test delete model file", t, testDeleteModelFile)
	convey.Convey("test update model file handler, parse model file content failed", t, parseModelFileContentFailed)
}

func testUpdateModelFile() {
	err := updateModelFileMsg.FillContent(testUpdateContent)
	convey.So(err, convey.ShouldBeNil)

	p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "LockGlobal", true).
		ApplyMethod(&modeltask.ModelMgr{}, "UnLockGlobal", func(*modeltask.ModelMgr) {}).
		ApplyMethod(&modeltask.ModelMgr{}, "CancelTasks", func(*modeltask.ModelMgr) {}).
		ApplyMethod(&modeltask.ModelMgr{}, "DelTaskByBriefs", func(*modeltask.ModelMgr, []types.ModelBrief) {}).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "GetFileList", []types.ModelBrief{{
			Uuid: "dd778847-a956-47d6-bc60-70b1af763c34",
			Name: "test.zip"},
		}).
		ApplyMethodReturn(database.GetMetaRepository(), "GetByType", testPodMetas, nil).
		ApplyFuncSeq(util.SendSyncMsg, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"{}", nil}, Times: 2},
			{Values: gomonkey.Params{constants.Success, nil}},
		}).
		ApplyFuncReturn(configpara.GetPodConfig, config.PodConfig{
			ContainerConfig: config.ContainerConfig{
				ContainerModelFileNumber: 20,
				TotalModelFileNumber:     40,
			}}).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
		ApplyMethod(&modeltask.ModelMgr{}, "UnLock", func(*modeltask.ModelMgr, string, string) {}).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "AddTask", nil).
		ApplyMethod(&modeltask.ModelMgr{}, "AddFailTask",
			func(*modeltask.ModelMgr, string, types.ModelFile, string) {})
	defer p.Reset()

	convey.Convey("update model file success", updateModelFileSuccess)
	convey.Convey("update model file failed, sync file failed", syncFileFailed)
	convey.Convey("update model file failed, check model file num failed", checkModelFileNumFailed)
	convey.Convey("update model file failed, get cert from edge om failed", getCertFailed)
	convey.Convey("update model file failed, check docker path from edge om failed", checkDockerPathFailed)
	convey.Convey("update model file failed, add download task failed", addTaskFailed)
}

func updateModelFileSuccess() {
	err := updateModelFile.Handle(&updateModelFileMsg)
	convey.So(err, convey.ShouldBeNil)
}

func syncFileFailed() {
	convey.Convey("global is locked", func() {
		p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "LockGlobal", false)
		defer p.Reset()
		err := updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("cannot marshal fileList", func() {
		p := gomonkey.ApplyFuncReturn(json.Marshal, nil, test.ErrTest)
		defer p.Reset()
		err := updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get used model files failed, get used pod id failed", func() {
		p := gomonkey.ApplyMethodReturn(database.GetMetaRepository(), "GetByType", testPodMetas, test.ErrTest)
		defer p.Reset()
		err := updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("send sync list failed", func() {
		p := gomonkey.ApplyFuncReturn(util.SendSyncMsg, "", test.ErrTest)
		defer p.Reset()
		err := updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("unmarshal syncDelList failed", func() {
		p := gomonkey.ApplyFuncReturn(util.SendSyncMsg, "", nil)
		defer p.Reset()
		err := updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func checkModelFileNumFailed() {
	convey.Convey("model file number of per pod up to limit", func() {
		p := gomonkey.ApplyFuncReturn(configpara.GetPodConfig, config.PodConfig{
			ContainerConfig: config.ContainerConfig{ContainerModelFileNumber: 0}})
		defer p.Reset()
		err := updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("total model file number up to limit", func() {
		p := gomonkey.ApplyFuncReturn(configpara.GetPodConfig, config.PodConfig{
			ContainerConfig: config.ContainerConfig{ContainerModelFileNumber: 20, TotalModelFileNumber: 0}})
		defer p.Reset()
		err := updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func getCertFailed() {
	p := gomonkey.ApplyFuncSeq(util.SendSyncMsg, []gomonkey.OutputCell{
		{Values: gomonkey.Params{"{}", nil}},
		{Values: gomonkey.Params{"", test.ErrTest}},

		{Values: gomonkey.Params{"{}", nil}},
		{Values: gomonkey.Params{constants.Failed, nil}},
	})
	defer p.Reset()

	err := updateModelFile.Handle(&updateModelFileMsg)
	convey.So(err, convey.ShouldResemble, expectErr)

	err = updateModelFile.Handle(&updateModelFileMsg)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func checkDockerPathFailed() {
	p := gomonkey.ApplyFuncSeq(util.SendSyncMsg, []gomonkey.OutputCell{
		{Values: gomonkey.Params{"{}", nil}, Times: 2},
		{Values: gomonkey.Params{"", test.ErrTest}},

		{Values: gomonkey.Params{"{}", nil}, Times: 2},
		{Values: gomonkey.Params{constants.Failed, nil}},
	})
	defer p.Reset()

	err := updateModelFile.Handle(&updateModelFileMsg)
	convey.So(err, convey.ShouldResemble, expectErr)

	err = updateModelFile.Handle(&updateModelFileMsg)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func addTaskFailed() {
	p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "AddTask", test.ErrTest)
	defer p.Reset()
	err := updateModelFile.Handle(&updateModelFileMsg)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func testDeleteModelFile() {
	p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "LockUuid", true).
		ApplyMethod(&modeltask.ModelMgr{}, "UnLockUuid", func(*modeltask.ModelMgr, string) {})
	defer p.Reset()

	convey.Convey("test deleteNotActive", testDeleteNotActive)
	convey.Convey("test deleteByUuid", testDeleteByUuid)
	convey.Convey("test deleteActiveAndNotActive", testDeleteActiveAndNotActive)
}

func getDeleteModelContent(typ string) types.ModelFileInfo {
	return types.ModelFileInfo{Operation: "delete", Target: typ}
}

func testDeleteNotActive() {
	err := updateModelFileMsg.FillContent(getDeleteModelContent(constants.TargetTypeTemp))
	convey.So(err, convey.ShouldBeNil)

	p := gomonkey.ApplyMethod(&modeltask.ModelMgr{}, "DelNotActiveTasks",
		func(*modeltask.ModelMgr, string, []types.ModelFile) {})
	defer p.Reset()

	convey.Convey("deleteNotActive success", func() {
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deleteNotActive failed, uuid is locked", func() {
		p1 := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "LockUuid", false)
		defer p1.Reset()
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("deleteNotActive failed, send msg to edge om failed", func() {
		p1 := gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, test.ErrTest)
		defer p1.Reset()
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func testDeleteByUuid() {
	err := updateModelFileMsg.FillContent(getDeleteModelContent(constants.TargetTypeAll))
	convey.So(err, convey.ShouldBeNil)

	p := gomonkey.ApplyMethod(&modeltask.ModelMgr{}, "DelTasksByUuid", func(*modeltask.ModelMgr, string) {})
	defer p.Reset()

	convey.Convey("deleteByUuid success", func() {
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deleteByUuid failed, uuid is locked", func() {
		p1 := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "LockUuid", false)
		defer p1.Reset()
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("deleteByUuid failed, send msg to edge om failed", func() {
		p1 := gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, test.ErrTest)
		defer p1.Reset()
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func testDeleteActiveAndNotActive() {
	err := updateModelFileMsg.FillContent(getDeleteModelContent(""))
	convey.So(err, convey.ShouldBeNil)

	p := gomonkey.ApplyMethod(&modeltask.ModelMgr{}, "DelActiveAndNotActiveTasks",
		func(*modeltask.ModelMgr, string, []types.ModelFile) {})
	defer p.Reset()

	convey.Convey("deleteActiveAndNotActive success", func() {
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deleteActiveAndNotActive failed, uuid is locked", func() {
		p1 := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "LockUuid", false)
		defer p1.Reset()
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("deleteActiveAndNotActive failed, send msg to edge om failed", func() {
		p1 := gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, test.ErrTest)
		defer p1.Reset()
		err = updateModelFile.Handle(&updateModelFileMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func parseModelFileContentFailed() {
	p := gomonkey.ApplyMethodReturn(&model.Message{}, "ParseContent", test.ErrTest)
	defer p.Reset()
	err := updateModelFile.Handle(&updateModelFileMsg)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("updateModelfileHandler failed,"+
		" model file update param error: %v", test.ErrTest))
}
