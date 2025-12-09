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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/apimachinery/pkg/util/intstr"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/msglistchecker"
)

func init() {
	convey.SetDefaultFailureMode(convey.FailureContinues)
}

func TestFdPodPara(t *testing.T) {
	patches := gomonkey.ApplyFunc(configpara.GetPodConfig, MockPodConfig).
		ApplyFuncReturn(configpara.GetNetType, constants.FDWithOM, nil).
		ApplyPrivateMethod(&MsgValidator{}, "checkSystemResources", func() error { return nil })
	defer patches.Reset()

	convey.Convey("test fd pod para", t, func() {
		convey.Convey("test container number", testFdContainerLimitNumber)
		convey.Convey("test container name changed", testFdContainerNameChanged)
		convey.Convey("test volume para", testVolumeValidate)
		convey.Convey("test probe para", testProbeValidate)
	})
}

func testFdContainerLimitNumber() {
	var basePod = getPodInfo()
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	// 构造数据库pod满规格
	for i := 0; i < constants.MaxContainerNumber; i++ {
		basePod.Name = fmt.Sprintf("test%d-eadea3ed-62ed-4b24-95a2-d2a4d98607e5", i)
		value, err := json.Marshal(basePod)
		if err != nil {
			hwlog.RunLog.Warnf("marsh pod failed %v", err)
		}
		meta := Meta{Type: "pod",
			Key:   fmt.Sprintf("websocket/pod/test%d-eadea3ed-62ed-4b24-95a2-d2a4d98607e5", i),
			Value: string(value)}
		database.GetDb().Create(&meta)
	}

	// 如果更新的pod已在数据库中，当前pod中的容器数没变化，容器数校验通过
	basePod.Name = "test19-eadea3ed-62ed-4b24-95a2-d2a4d98607e5"
	var msg model.Message
	setFdPodMsg(&msg, basePod)

	var err error
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldEqual, nil)

	// 如果更新的pod在数据库中，单当前pod中容器数已增加，容器数校验失败
	basePod.Name = "test21-eadea3ed-62ed-4b24-95a2-d2a4d98607e5"
	var c = basePod.Spec.Containers[0]
	basePod.Spec.Containers[0].Name = "container-00"
	basePod.Spec.Containers = append(basePod.Spec.Containers, c)
	setFdPodMsg(&msg, basePod)
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldNotEqual, nil)
	convey.So(err.Error(), convey.ShouldContainSubstring, "container num is out of limit")

	// 如果更新的pod不在数据库中，容器数校验失败
	basePod.Name = "test21-eadea3ed-62ed-4b24-95a2-d2a4d98607e5"
	setFdPodMsg(&msg, basePod)

	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldNotEqual, nil)
	convey.So(err.Error(), convey.ShouldContainSubstring, "container num is out of limit")

	// 清理数据库
	database.GetDb().Where(`type="pod"`).Delete(&Meta{})
}

func testFdContainerNameChanged() {
	var basePod = getPodInfo()
	var err error
	var msg model.Message

	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	value, err := json.Marshal(basePod)
	if err != nil {
		hwlog.RunLog.Warnf("marsh pod failed %v", err)
	}
	meta := Meta{Type: "pod",
		Key:   fmt.Sprintf("websocket/pod/%s", basePod.Name),
		Value: string(value)}
	database.GetDb().Create(&meta)

	// 优雅删除pod时，更新容器名，校验错误
	deletionTimestamp := "2023-03-04T06:50:13Z"
	basePod.DeletionTimestamp = &deletionTimestamp
	basePod.Spec.Containers[0].Name = "test"
	setFdPodMsg(&msg, basePod)

	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldNotEqual, nil)
	convey.So(err.Error(), convey.ShouldContainSubstring, "container name in pod has changed")

	// 清理数据库
	database.GetDb().Where(`type="pod"`).Delete(&Meta{})
}

type volumeTestCase struct {
	description string
	volume      types.Volume
	shouldErr   bool
	assert      convey.Assertion
	expected    interface{}
}

func getVolumeNameTestCase() []volumeTestCase {
	return []volumeTestCase{
		{
			description: "test volume name contain invalid char",
			volume:      types.Volume{Name: "a-b_c"},
			shouldErr:   true,
			assert:      convey.ShouldContainSubstring,
			expected:    "Pod.Spec.Volumes.Name",
		},
		{
			description: "test volume name contain invalid length 0",
			volume:      types.Volume{Name: ""},
			shouldErr:   true,
			assert:      convey.ShouldContainSubstring,
			expected:    "Pod.Spec.Volumes.Name",
		},
		{
			description: "test volume name contain invalid length 1",
			volume:      types.Volume{Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			shouldErr:   true,
			assert:      convey.ShouldContainSubstring,
			expected:    "Pod.Spec.Volumes.Name",
		},
	}
}
func getVolumeHostPathTestCase() []volumeTestCase {
	var hostPathType = ""
	return []volumeTestCase{
		{
			description: "test volume hostpath invalid length 0",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				HostPath: &types.HostPathVolumeSource{
					Path: "/",
					Type: &hostPathType,
				},
			}},
			shouldErr: true,
			assert:    convey.ShouldContainSubstring,
			expected:  "'Path' failed on the 'min' tag",
		}, {
			description: "test volume hostpath contain ..",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				HostPath: &types.HostPathVolumeSource{
					Path: "/..",
					Type: &hostPathType,
				},
			}},
			shouldErr: true,
			assert:    convey.ShouldContainSubstring,
			expected:  "'Path' failed on the 'excludes' tag",
		}, {
			description: "test volume hostpath not in white list",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				HostPath: &types.HostPathVolumeSource{
					Path: "/a",
					Type: &hostPathType,
				},
			}},
			shouldErr: true,
			assert:    convey.ShouldContainSubstring,
			expected:  "not in whitelist",
		}, {
			description: "test volume hostpath in white list",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				HostPath: &types.HostPathVolumeSource{
					Path: "/var/lib/docker/modelfile/123456",
					Type: &hostPathType,
				},
			}},
			shouldErr: false,
			assert:    convey.ShouldEqual,
			expected:  nil,
		},
	}
}

func getConfigMapAndEmptyDirTestCase() []volumeTestCase {
	var defaultMode = int32(0644)
	return []volumeTestCase{
		{
			description: "test volume empty dir success",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				EmptyDir: &types.EmptyDirVolumeSource{},
			},
			},
			shouldErr: false,
			assert:    convey.ShouldEqual,
			expected:  nil,
		},
		{
			description: "test volume config map name failed",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				ConfigMap: &types.ConfigMapVolumeSource{
					Name: "0az",
				},
			},
			},
			shouldErr: true,
			assert:    convey.ShouldContainSubstring,
			expected:  "'DefaultMode' failed on the 'eq' tag",
		},
		{
			description: "test volume config map default mode success",
			volume: types.Volume{Name: "test", VolumeSource: types.VolumeSource{
				ConfigMap: &types.ConfigMapVolumeSource{
					Name:        "acac",
					DefaultMode: &defaultMode,
				},
			},
			},
			shouldErr: false,
			assert:    convey.ShouldEqual,
			expected:  nil,
		},
	}
}

func testVolumeValidate() {
	var basePod = getPodInfo()

	var testCase = getVolumeNameTestCase()
	testCase = append(testCase, getVolumeHostPathTestCase()...)
	testCase = append(testCase, getConfigMapAndEmptyDirTestCase()...)

	var err error
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	for _, tc := range testCase {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		var msg model.Message
		basePod.Spec.Volumes = []types.Volume{tc.volume}
		setFdPodMsg(&msg, basePod)

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

type probeTestCase struct {
	description    string
	livenessProbe  *types.Probe
	readinessProbe *types.Probe
	startupProbe   *types.Probe
	shouldErr      bool
	assert         convey.Assertion
	expected       interface{}
}

func getProbeTestCases() []probeTestCase {
	const defaultProbePort = 8080
	return []probeTestCase{
		{description: "nil probe should success",
			shouldErr: false, assert: convey.ShouldEqual, expected: nil},
		{description: "startupProbe should be not supported",
			startupProbe: &types.Probe{ProbeHandler: types.ProbeHandler{
				HTTPGet: &types.HTTPGetAction{Path: "/a", Port: intstr.FromInt(defaultProbePort)}}},
			shouldErr: true, assert: convey.ShouldContainSubstring, expected: `StartupProbe`},
		{description: "livenessProbe should be supported",
			livenessProbe: &types.Probe{ProbeHandler: types.ProbeHandler{
				HTTPGet: &types.HTTPGetAction{Path: "/a", Port: intstr.FromInt(defaultProbePort)}}},
			shouldErr: false, assert: convey.ShouldEqual, expected: nil},
		{description: "readinessProbe should be supported",
			readinessProbe: &types.Probe{ProbeHandler: types.ProbeHandler{
				HTTPGet: &types.HTTPGetAction{Path: "/a", Port: intstr.FromInt(defaultProbePort)}}},
			shouldErr: false, assert: convey.ShouldEqual, expected: nil},
		{description: "execAction should be supported",
			livenessProbe: &types.Probe{ProbeHandler: types.ProbeHandler{
				Exec: &types.ExecAction{Command: []string{"/a"}}}},
			shouldErr: false, assert: convey.ShouldEqual, expected: nil},
		{description: "relative path for execAction should not be allowed",
			livenessProbe: &types.Probe{ProbeHandler: types.ProbeHandler{
				Exec: &types.ExecAction{Command: []string{"../a"}}}},
			shouldErr: true, assert: convey.ShouldContainSubstring, expected: "LivenessProbe.ProbeHandler.Exec.Command"},
		{description: "relative path for httpGetAction should not be allowed",
			livenessProbe: &types.Probe{ProbeHandler: types.ProbeHandler{
				HTTPGet: &types.HTTPGetAction{Path: "../a", Port: intstr.FromInt(defaultProbePort)}}},
			shouldErr: true, assert: convey.ShouldContainSubstring, expected: ""},
		{description: "non-ip host should not be allowed",
			livenessProbe: &types.Probe{ProbeHandler: types.ProbeHandler{
				HTTPGet: &types.HTTPGetAction{Path: "/a", Host: "localhost", Port: intstr.FromInt(defaultProbePort)}}},
			shouldErr: true, assert: convey.ShouldContainSubstring, expected: ""},
		{description: "only HTTP and HTTPS protocols should be allowed",
			livenessProbe: &types.Probe{ProbeHandler: types.ProbeHandler{
				HTTPGet: &types.HTTPGetAction{Path: "/a", Scheme: "WEBSOCKET", Port: intstr.FromInt(defaultProbePort)}}},
			shouldErr: true, assert: convey.ShouldContainSubstring, expected: "LivenessProbe.ProbeHandler.HTTPGet.Scheme"},
	}
}

func testProbeValidate() {
	var basePod = getPodInfo()

	var testCase = getProbeTestCases()

	var err error
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	for _, tc := range testCase {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		var msg model.Message
		basePod.Spec.Containers[0].LivenessProbe = tc.livenessProbe
		basePod.Spec.Containers[0].ReadinessProbe = tc.readinessProbe
		basePod.Spec.Containers[0].StartupProbe = tc.startupProbe
		setFdPodMsg(&msg, basePod)

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
