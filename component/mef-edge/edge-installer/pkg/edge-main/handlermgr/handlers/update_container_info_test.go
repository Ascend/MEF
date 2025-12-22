// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlers
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
)

const containerInfoMsg = `{
    "container":[
        {
            "modelfile":[
                {
                    "active_type":"hot_update",
                    "name":"Ascend-mindxedge-mefedge_5.0.RC2_linux-aarch64.zip",
                    "version":"v1"
                }
            ]
        }
    ],
    "operation":"update",
    "pod_name":"test-modelfile-6a1fb6a5-8c4c-4242-b4d9-89365248c146",
    "pod_uid":"6a1fb6a5-8c4c-4242-b4d9-89365248c146",
    "source":"all",
    "uuid":"5c3ef203-28b6-4f63-be8d-a38f3f59197f"
}`

var (
	testHotUpdateMsg  types.UpdateContainerInfo
	testColdUpdateMsg types.UpdateContainerInfo
	mockModelMgr      modeltask.ModelMgr
)

func setupUpdateContainerInfo() error {
	if err := json.Unmarshal([]byte(containerInfoMsg), &testHotUpdateMsg); err != nil {
		hwlog.RunLog.Errorf("unmarshal test update container info message failed, error: %v", err)
		return err
	}

	coldUpdateContainer := []types.ContainerInfo{
		{ModelFile: []types.ModelFileEffectInfo{
			{
				Name:       "Ascend-mindxedge-mefedge_5.0.RC2_linux-aarch64.zip",
				Version:    "v1",
				ActiveType: "cold_update",
			},
		}},
	}
	testColdUpdateMsg = testHotUpdateMsg
	testColdUpdateMsg.Container = coldUpdateContainer
	return nil
}

func TestUpdateContainerInfo(t *testing.T) {
	if err := setupUpdateContainerInfo(); err != nil {
		fmt.Printf("setup test update container info environment failed: %v\n", err)
		return
	}

	p := gomonkey.ApplyFuncReturn(modeltask.GetModelMgr, &mockModelMgr)
	defer p.Reset()
	convey.Convey("test update container info successful", t, func() {
		convey.Convey("no model file need to update", noModelFileNeedToUpdate)
		convey.Convey("hot update success", hotUpdateSuccess)
		convey.Convey("cold update success", coldUpdateSuccess)
	})

	convey.Convey("test update container info failed", t, func() {
		convey.Convey("hot update failed", func() {
			convey.Convey("check model file info failed", checkModelFileInfoFailed)
			convey.Convey("effect model file by edge om failed", effectModelFileByEdgeOmFailed)
		})
		convey.Convey("cold update failed", func() {
			convey.Convey("check pod restart policy failed", checkRestartPolicyFailed)
			convey.Convey("restart pod by edge om failed", restartPodFailed)
		})
	})
}

func noModelFileNeedToUpdate() {
	var noModelFileMsg types.UpdateContainerInfo
	noModelFileMsg.Container = []types.ContainerInfo{
		{ModelFile: []types.ModelFileEffectInfo{}},
	}
	updateContainerInfo := UpdateContainerInfo{containerInfo: noModelFileMsg}
	err := updateContainerInfo.EffectModelFile()
	convey.So(err, convey.ShouldBeNil)
}

func hotUpdateSuccess() {
	convey.Convey("model file has already taken effect", func() {
		p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
			ApplyMethodReturn(&modeltask.ModelMgr{}, "GetActiveTask", &modeltask.ModelFileTask{
				ModelFile: types.ModelFile{Version: "v1"},
			})
		defer p.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("model file takes effect success", func() {
		p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
			ApplyMethodReturn(&modeltask.ModelMgr{}, "GetActiveTask", nil).
			ApplyMethodReturn(&modeltask.ModelMgr{}, "GetNotActiveTask", &modeltask.ModelFileTask{
				ModelFile: types.ModelFile{Version: "v1"},
			}).
			ApplyMethodReturn(&modeltask.ModelFileTask{}, "GetStatusType", types.StatusInactive).
			ApplyFuncReturn(util.SendSyncMsg, constants.Success, nil).
			ApplyMethodReturn(&modeltask.ModelMgr{}, "Active", nil)
		defer p.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldBeNil)
	})
}

func coldUpdateSuccess() {
	p := gomonkey.ApplyFuncReturn(CheckPodRestartPolicy, nil).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "GetActiveTask", &modeltask.ModelFileTask{
			ModelFile: types.ModelFile{Version: "v1"},
		}).
		ApplyFuncReturn(util.SendSyncMsg, constants.Success, nil)
	defer p.Reset()
	updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
	err := updateContainerInfo.EffectModelFile()
	convey.So(err, convey.ShouldBeNil)
}

func checkModelFileInfoFailed() {
	expectErr := errors.New("not all model files take effect successfully")
	convey.Convey("lock model file database failed", func() {
		p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", false)
		defer p.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("no model file to be activated", func() {
		convey.Convey("not active task is nil", func() {
			p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
				ApplyMethodReturn(&modeltask.ModelMgr{}, "GetActiveTask", nil).
				ApplyMethodReturn(&modeltask.ModelMgr{}, "GetNotActiveTask", nil)
			defer p.Reset()
			updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
			err := updateContainerInfo.EffectModelFile()
			convey.So(err, convey.ShouldResemble, expectErr)
		})

		convey.Convey("status type is not inactive", func() {
			p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
				ApplyMethodReturn(&modeltask.ModelMgr{}, "GetActiveTask", &modeltask.ModelFileTask{
					ModelFile: types.ModelFile{Version: "v2"},
				}).
				ApplyMethodReturn(&modeltask.ModelMgr{}, "GetNotActiveTask", &modeltask.ModelFileTask{}).
				ApplyMethodReturn(&modeltask.ModelFileTask{}, "GetStatusType", types.StatusDownloading)
			defer p.Reset()
			updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
			err := updateContainerInfo.EffectModelFile()
			convey.So(err, convey.ShouldResemble, expectErr)
		})
	})

	convey.Convey("the model file version is not correct", func() {
		p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
			ApplyMethodReturn(&modeltask.ModelMgr{}, "GetActiveTask", nil).
			ApplyMethodReturn(&modeltask.ModelMgr{}, "GetNotActiveTask", &modeltask.ModelFileTask{
				ModelFile: types.ModelFile{Version: "v2"},
			}).
			ApplyMethodReturn(&modeltask.ModelFileTask{}, "GetStatusType", types.StatusInactive)
		defer p.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func effectModelFileByEdgeOmFailed() {
	expectErr := errors.New("not all model files take effect successfully")
	p := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "GetActiveTask", nil).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "GetNotActiveTask", &modeltask.ModelFileTask{
			ModelFile: types.ModelFile{Version: "v1"},
		}).
		ApplyMethodReturn(&modeltask.ModelFileTask{}, "GetStatusType", types.StatusInactive)
	defer p.Reset()

	convey.Convey("send effect model file message to edge om failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, "", test.ErrTest)
		defer p1.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("model file takes effect failed by edge om", func() {
		p2 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, constants.Failed, nil)
		defer p2.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("update model file status to active failed", func() {
		p3 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, constants.Success, nil).
			ApplyMethodReturn(&modeltask.ModelMgr{}, "Active", test.ErrTest)
		defer p3.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testHotUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func checkRestartPolicyFailed() {
	p := gomonkey.ApplyFuncReturn(CheckPodRestartPolicy, test.ErrTest)
	defer p.Reset()
	updateContainerInfo := UpdateContainerInfo{containerInfo: testColdUpdateMsg}
	err := updateContainerInfo.EffectModelFile()
	expectErr := fmt.Errorf("check pod restart policy failed, %v", test.ErrTest)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func restartPodFailed() {
	expectErr := errors.New("restart pod for model file effect failed")
	p := gomonkey.ApplyFuncReturn(CheckPodRestartPolicy, nil).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "Lock", true).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "GetActiveTask", &modeltask.ModelFileTask{
			ModelFile: types.ModelFile{Version: "v1"},
		})
	defer p.Reset()
	convey.Convey("send restart pod message to edge om failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, "", test.ErrTest)
		defer p1.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testColdUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("restart pod failed by edge om", func() {
		p2 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, constants.Failed, nil)
		defer p2.Reset()
		updateContainerInfo := UpdateContainerInfo{containerInfo: testColdUpdateMsg}
		err := updateContainerInfo.EffectModelFile()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
