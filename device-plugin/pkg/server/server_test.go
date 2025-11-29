/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"

	"Ascend-device-plugin/pkg/common"
)

// TestPluginServerGetRestartFlag Test PluginServer GetRestartFlag()
func TestPluginServerGetRestartFlag(t *testing.T) {
	convey.Convey("test GetRestartFlag", t, func() {
		ps := &PluginServer{restart: false}
		convey.So(ps.GetRestartFlag(), convey.ShouldBeFalse)
	})
}

// TestPluginServerSetRestartFlag Test PluginServer SetRestartFlag()
func TestPluginServerSetRestartFlag(t *testing.T) {
	convey.Convey("test SetRestartFlag", t, func() {
		ps := &PluginServer{restart: false}
		ps.SetRestartFlag(true)
		convey.So(ps.GetRestartFlag(), convey.ShouldBeTrue)
	})
}

// TestPluginServerStop Test PluginServer Stop()
func TestPluginServerStop(t *testing.T) {
	convey.Convey("test Stop", t, func() {
		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer()}

		ps.Stop()
		convey.So(ps.isRunning.Load(), convey.ShouldBeFalse)
	})
}
