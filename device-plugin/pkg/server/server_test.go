/* Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
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
