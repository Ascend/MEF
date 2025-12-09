// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package msgchecker

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/msglistchecker"
)

func TestMefPodPara(t *testing.T) {
	patches := gomonkey.ApplyFunc(configpara.GetPodConfig, MockPodConfig).
		ApplyFuncReturn(configpara.GetNetType, constants.MEF, nil).
		ApplyPrivateMethod(&MsgValidator{}, "checkSystemResources", func() error { return nil })
	defer patches.Reset()

	convey.Convey("test mef pod para", t, func() {
		convey.Convey("test mef volume para", testMefVolumeValidate)
		convey.Convey("test mef container resource", testMefContainerResourceValidate)
	})
}

func getVolumeTestCase() []volumeTestCase {
	var hostPathType = ""
	var defaultMode = int32(0644)

	return []volumeTestCase{
		// mef网管不支持 挂载卷不支持 /var/lib/docker/modelfile/
		{
			description: "test volume hostpath in white list",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				HostPath: &types.HostPathVolumeSource{
					Path: "/var/lib/docker/modelfile/", Type: &hostPathType},
			}},
			shouldErr: true, assert: convey.ShouldContainSubstring,
			expected: "model file path not permitted in host path",
		},
		// mef网管不支持 empty dir
		{
			description: "test volume empty dir success",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				EmptyDir: &types.EmptyDirVolumeSource{}},
			}, shouldErr: true, assert: convey.ShouldContainSubstring, expected: "cur config not support empty dir",
		},
		// mef网管不支持 config map
		{
			description: "test volume config map default mode success",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				ConfigMap: &types.ConfigMapVolumeSource{Name: "acac", DefaultMode: &defaultMode},
			},
			},
			shouldErr: true, assert: convey.ShouldContainSubstring, expected: "cur config not support config map",
		},
	}

}
func testMefVolumeValidate() {
	var basePod = getPodInfo()

	var err error
	msgValidator := NewMsgValidator(msglistchecker.NewCloudCoreMsgHeaderValidator(false))

	testCase := getVolumeTestCase()
	for _, tc := range testCase {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		var msg model.Message
		basePod.Spec.Volumes = []types.Volume{tc.volume}
		setMefPodMsg(&msg, basePod)

		if err = msgValidator.Check(&msg); err != nil {
			hwlog.RunLog.Errorf("check msg failed: %v", err)
		}

		if tc.shouldErr {
			convey.So(err.Error(), tc.assert, tc.expected)
		} else {
			convey.So(err, tc.assert, tc.expected)
		}
	}
}

type containerResourceTestCase struct {
	description string
	resource    map[string]resource.Quantity
	shouldErr   bool
	assert      convey.Assertion
	expected    interface{}
}

func getContainerResourceTestCase(resName v1.ResourceName) []containerResourceTestCase {
	quantity1 := resource.MustParse("1")
	quantity2 := resource.MustParse("2")

	return []containerResourceTestCase{
		{
			description: "test npu requests valid",
			resource: map[string]resource.Quantity{
				"Req": quantity1,
				"Lim": quantity2,
			},
			shouldErr: false,
			assert:    convey.ShouldEqual,
			expected:  nil,
		},
		{
			description: "test npu requests invalid",
			resource: map[string]resource.Quantity{
				"Req": quantity2,
				"Lim": quantity1,
			},
			shouldErr: true,
			assert:    convey.ShouldContainSubstring,
			expected: fmt.Errorf("resource [%s] request [%s] great than limit [%s]",
				resName, quantity2.String(), quantity1.String()).Error(),
		},
	}
}

func testMefContainerResourceValidate() {
	testPod := getPodInfo()
	resName := v1.ResourceName("huawei.com/Ascend310")
	msgValidator := NewMsgValidator(msglistchecker.NewCloudCoreMsgHeaderValidator(false))

	testCase := getContainerResourceTestCase(resName)
	for _, tc := range testCase {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		var msg model.Message
		testPod.Spec.Containers[0].Resources.Req[resName] = tc.resource["Req"]
		testPod.Spec.Containers[0].Resources.Lim[resName] = tc.resource["Lim"]
		setMefPodMsg(&msg, testPod)

		err := msgValidator.Check(&msg)
		if tc.shouldErr {
			convey.So(err.Error(), tc.assert, tc.expected)
		} else {
			convey.So(err, tc.assert, tc.expected)
		}
	}
}
