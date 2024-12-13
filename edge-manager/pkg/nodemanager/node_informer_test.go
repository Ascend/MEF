// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package nodemanager for node_informer test
package nodemanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestGetMEFNodeStatusForOffline(t *testing.T) {
	convey.Convey("test GetMEFNodeStatus For Get Lable Err", t, func() {
		hostname := "local"
		service := &nodeSyncImpl{}
		patch := gomonkey.ApplyFuncReturn(service.getNode, nil, nil)
		defer patch.Reset()
		str, err := service.GetMEFNodeStatus(hostname)
		convey.So(err, convey.ShouldBeNil)
		convey.So(str, convey.ShouldEqual, "offline")
	})
}

func TestGetMEFNodeStatusForGetNodeErr(t *testing.T) {
	convey.Convey("test GetMEFNodeStatus For Get Node Err", t, func() {
		hostname := "local"
		service := &nodeSyncImpl{}
		patch := gomonkey.ApplyFuncReturn(service.getNode, nil, "err")
		defer patch.Reset()
		str, err := service.GetMEFNodeStatus(hostname)
		convey.So(err, convey.ShouldBeNil)
		convey.So(str, convey.ShouldEqual, "offline")
	})
}

func TestGetK8sNodeStatus(t *testing.T) {
	convey.Convey("test GetK8sNodeStatus For offline", t, func() {
		hostname := "local"
		service := &nodeSyncImpl{}
		patch := gomonkey.ApplyFuncReturn(service.getNode, nil, nil)
		defer patch.Reset()
		str, err := service.GetK8sNodeStatus(hostname)
		convey.So(err, convey.ShouldBeNil)
		convey.So(str, convey.ShouldEqual, "offline")
	})
}

func TestGetAllocatableResource(t *testing.T) {
	convey.Convey("test GetAllocatableResource For err", t, func() {
		hostname := "local"
		service := &nodeSyncImpl{}
		patch := gomonkey.ApplyFuncReturn(service.getNode, nil, nil)
		defer patch.Reset()
		_, err := service.GetK8sNodeStatus(hostname)
		convey.So(err, convey.ShouldBeNil)
	})
}
