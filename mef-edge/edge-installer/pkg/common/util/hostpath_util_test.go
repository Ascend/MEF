// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
)

func TestGetHostNameMapPath(t *testing.T) {
	// 创建一个名为my-volume的空目录卷
	volume1 := v1.Volume{
		Name: "my-volume-1",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}

	// 创建一个名为my-volume的hostPath卷
	volume2 := v1.Volume{
		Name: "my-volume-2",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/my-host-path",
			},
		},
	}

	// 创建一个包含两个卷的卷列表
	volumes := []v1.Volume{volume1, volume2}
	convey.Convey("GetHostNameMapPath", t, func() {
		hostNameMapPath := GetHostNameMapPath(volumes)
		convey.So(hostNameMapPath, convey.ShouldResemble, []string{"/my-host-path"})
	})
}

func TestInWhiteList(t *testing.T) {
	hostPath := "home"
	whiteList := []string{"home", "date"}
	convey.Convey("GetHostNameMapPath", t, func() {
		hostNameMapPath := InFdWhiteList(hostPath, whiteList)
		convey.So(hostNameMapPath, convey.ShouldResemble, true)
	})

	convey.Convey("test func InFdWhiteList", t, func() {
		const hostPath = "/var/lib/docker/modelfile/hosts"
		hostNameMapPath := InFdWhiteList(hostPath, whiteList)
		convey.So(hostNameMapPath, convey.ShouldBeFalse)
	})
}
