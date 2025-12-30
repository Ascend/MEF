// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package containerinfochecker for container info checker
package containerinfochecker

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/types"
)

const (
	containerInfoCount = 6
	testContent        = `{
    "container":[
        {
            "image":"fd.fusiondirector.huawei.com/library/image:1.0",
            "mailbox_path":"/run/docker/ha-mailbox/container-0_test-modelfile-77b9a1fa-27b6-4e23-bfb4-5d3073ee552b",
            "modelfile":[
                {
                    "active_type":"hot_update",
                    "name":"Ascend-mindxedge-mefedge_5.0.RC1.1_linux-aarch64.zip",
                    "version":"2.0"
                }
            ],
            "name":"container-0"
        },
        {
            "image":"fd.fusiondirector.huawei.com/library/image:1.0",
            "mailbox_path":"/run/docker/ha-mailbox/container-1_test-modelfile-77b9a1fa-27b6-4e23-bfb4-5d3073ee552b",
            "modelfile":[
                {
                    "active_type":"hot_update",
                    "name":"Ascend-mindxedge-mefedge_5.0.RC1.1_linux-aarch64.zip",
                    "version":"2.0"
                }
            ],
            "name":"container-1"
        }
    ],
    "hot_standby":false,
    "key_container":true,
    "operation":"update",
    "pod_name":"test-modelfile-77b9a1fa-27b6-4e23-bfb4-5d3073ee552b",
    "pod_uid":"77b9a1fa-27b6-4e23-bfb4-5d3073ee552b",
    "source":"all",
    "uuid":"5c3ef203-28b6-4f63-be8d-a38f3f59197f"
}`
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func getContainerInfoTemplate() *types.UpdateContainerInfo {
	var containerInfo types.UpdateContainerInfo
	if err := json.Unmarshal([]byte(testContent), &containerInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal content to containerInfo failed, error: %v", err)
		return nil
	}
	return &containerInfo
}

func getValidContainerInfo() ([]byte, error) {
	containerInfo := getContainerInfoTemplate()
	containerInfo.Container[0].ModelFile = []types.ModelFileEffectInfo{}
	containerInfo.Container[1].ModelFile = []types.ModelFileEffectInfo{}
	bytes, err := json.Marshal(containerInfo)
	if err != nil {
		hwlog.RunLog.Error("marshal content failed")
		return nil, err
	}
	return bytes, nil
}

func getInvalidContainerInfo(key string) ([]byte, error) {
	containerInfo := getContainerInfoTemplate()
	switch key {
	case "InvalidOperation":
		containerInfo.Operation = "delete"
	case "InvalidSource":
		containerInfo.Source = "modelfiles"
	case "InvalidPodName":
		containerInfo.PodName = "test_modelfile-77b9a1fa-27b6-4e23-bfb4-5d3073ee552b"
	case "InvalidPodUid":
		containerInfo.PodUid = "77b9A1FA-27B6-4E23-bFB4-5D3073EE552B"
	case "InvalidUuid":
		containerInfo.Uuid = "5C3EF203-28B6-4F63-BE8D-A38F3F59197F"
	case "InvalidContainersMinLen":
		containerInfo.Container = []types.ContainerInfo{}
	case "InvalidContainersMaxLen":
		containers := make([]types.ContainerInfo, len(containerInfo.Container)*containerInfoCount)
		for i := 0; i < containerInfoCount; i++ {
			copy(containers[len(containers):], containerInfo.Container)
		}
		containerInfo.Container = containers
	case "InvalidActiveType":
		containerInfo.Container[0].ModelFile[0].ActiveType = "update"
	case "InvalidModelFileName1":
		containerInfo.Container[0].ModelFile[0].Name = "~test_invalid.zip"
	case "InvalidModelFileName2":
		containerInfo.Container[0].ModelFile[0].Name = "..test_invalid.zip"
	case "InvalidModelFileName3":
		containerInfo.Container[0].ModelFile[0].Name = "test_invalid.tar"
	case "InvalidVersion1":
		containerInfo.Container[0].ModelFile[0].Version = "1V.0"
	case "InvalidVersion2":
		containerInfo.Container[0].ModelFile[0].Version = ".1.0"
	case "InvalidVersion3":
		containerInfo.Container[0].ModelFile[0].Version = "1.0."
	default:
		return nil, errors.New("key is not found")
	}
	bytes, err := json.Marshal(containerInfo)
	if err != nil {
		hwlog.RunLog.Error("marshal content failed")
		return nil, err
	}
	return bytes, nil
}

func TestContainerInfoChecker(t *testing.T) {
	convey.Convey("test container info checker", t, func() {
		convey.Convey("valid container info case", testCheckValidCase)
		convey.Convey("invalid container info case", func() {
			convey.Convey("invalid operation case", testCheckInvalidOperation)
			convey.Convey("invalid source case", testCheckInvalidSource)
			convey.Convey("invalid pod_name case", testCheckInvalidPodName)
			convey.Convey("invalid pod_uid case", testCheckInvalidPodUid)
			convey.Convey("invalid uuid case", testCheckInvalidUuid)
			convey.Convey("invalid containers len case", testCheckInvalidContainersLen)
			convey.Convey("invalid active_type case", testCheckInvalidActiveType)
			convey.Convey("invalid model file name case", testCheckInvalidModelFileName)
			convey.Convey("invalid model file version case", testCheckInvalidVersion)
		})
	})
}

func testCheckValidCase() {
	convey.Convey("valid container info with model files case", func() {
		err := CheckContainerInfo([]byte(testContent))
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("valid container info without model files case", func() {
		validContainerInfo, err := getValidContainerInfo()
		if err != nil {
			return
		}
		err = CheckContainerInfo(validContainerInfo)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testCheckInvalidOperation() {
	invalidContainerInfo, err := getInvalidContainerInfo("InvalidOperation")
	if err != nil {
		return
	}
	err = CheckContainerInfo(invalidContainerInfo)
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckInvalidSource() {
	invalidContainerInfo, err := getInvalidContainerInfo("InvalidSource")
	if err != nil {
		return
	}
	err = CheckContainerInfo(invalidContainerInfo)
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckInvalidPodName() {
	invalidContainerInfo, err := getInvalidContainerInfo("InvalidPodName")
	if err != nil {
		return
	}
	err = CheckContainerInfo(invalidContainerInfo)
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckInvalidPodUid() {
	invalidContainerInfo, err := getInvalidContainerInfo("InvalidPodUid")
	if err != nil {
		return
	}
	err = CheckContainerInfo(invalidContainerInfo)
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckInvalidUuid() {
	invalidContainerInfo, err := getInvalidContainerInfo("InvalidUuid")
	if err != nil {
		return
	}
	err = CheckContainerInfo(invalidContainerInfo)
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckInvalidContainersLen() {
	convey.Convey("invalid containers min len", func() {
		invalidContainerInfo, err := getInvalidContainerInfo("InvalidContainersMinLen")
		if err != nil {
			return
		}
		err = CheckContainerInfo(invalidContainerInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("invalid containers max len", func() {
		invalidContainerInfo, err := getInvalidContainerInfo("InvalidContainersMaxLen")
		if err != nil {
			return
		}
		err = CheckContainerInfo(invalidContainerInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testCheckInvalidActiveType() {
	invalidContainerInfo, err := getInvalidContainerInfo("InvalidActiveType")
	if err != nil {
		return
	}
	err = CheckContainerInfo(invalidContainerInfo)
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckInvalidModelFileName() {
	convey.Convey("invalid name, not match reg", func() {
		invalidContainerInfo, err := getInvalidContainerInfo("InvalidModelFileName1")
		if err != nil {
			return
		}
		err = CheckContainerInfo(invalidContainerInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("invalid name, include invalid words", func() {
		invalidContainerInfo, err := getInvalidContainerInfo("InvalidModelFileName2")
		if err != nil {
			return
		}
		err = CheckContainerInfo(invalidContainerInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("invalid name, with invalid suffix", func() {
		invalidContainerInfo, err := getInvalidContainerInfo("InvalidModelFileName3")
		if err != nil {
			return
		}
		err = CheckContainerInfo(invalidContainerInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testCheckInvalidVersion() {
	convey.Convey("invalid version, include invalid words", func() {
		invalidContainerInfo, err := getInvalidContainerInfo("InvalidVersion1")
		if err != nil {
			return
		}
		err = CheckContainerInfo(invalidContainerInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("invalid version, start with invalid character", func() {
		invalidContainerInfo, err := getInvalidContainerInfo("InvalidVersion2")
		if err != nil {
			return
		}
		err = CheckContainerInfo(invalidContainerInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("invalid version, end with invalid character", func() {
		invalidContainerInfo, err := getInvalidContainerInfo("InvalidVersion3")
		if err != nil {
			return
		}
		err = CheckContainerInfo(invalidContainerInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
