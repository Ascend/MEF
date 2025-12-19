// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config test for capability manager
package config

import (
	"fmt"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"huawei.com/mindx/common/envutils"

	"edge-installer/pkg/common/constants"
)

func TestGetScaledNpu(t *testing.T) {
	convey.Convey("test func GetScaledNpu", t, func() {
		// clear the capabilities to avoid affecting subsequent test cases
		GetCapabilityCache().capabilities = sync.Map{}

		originVal := resource.Quantity{}
		npuVal := GetScaledNpu("other npu resource name", originVal)
		convey.So(npuVal, convey.ShouldResemble, originVal)
		npuVal = GetScaledNpu(SharedNpuResName, originVal)
		convey.So(npuVal, convey.ShouldResemble, originVal)

		var p1 = gomonkey.ApplyMethodReturn(&CapabilityCache{}, "HasCapability", true)
		defer p1.Reset()
		GetScaledNpu(SharedNpuResName, originVal)
	})
}

func TestModifyNpuRes(t *testing.T) {
	convey.Convey("test func ModifyNpuRes", t, func() {
		// sharing npu config exists, no modification is required
		var p1 = gomonkey.ApplyMethodReturn(&CapabilityCache{}, "HasCapability", false)
		defer p1.Reset()
		resourceList := map[v1.ResourceName]resource.Quantity{SharedNpuResName: {}}
		ModifyNpuRes(resourceList, true)

		// sharing npu config doesn't exist, but "huawei.com/Ascend310" is not exist
		var p2 = gomonkey.ApplyMethodReturn(&CapabilityCache{}, "HasCapability", true)
		defer p2.Reset()
		ModifyNpuRes(resourceList, true)

		// "huawei.com/Ascend310" exists
		resourceList[constants.CenterNpuName] = resource.Quantity{}
		ModifyNpuRes(resourceList, true)
		ModifyNpuRes(resourceList, false)
	})
}

func TestCapabilityMgr(t *testing.T) {
	convey.Convey("test CapabilityMgr methods", t, func() {
		capabilityMgr := GetCapabilityMgr()
		capabilityMgr.GetCaps()
		err := capabilityMgr.Switch(npuSharingConfigKey, true)
		convey.So(err, convey.ShouldBeNil)
	})
}

var (
	npuSharingConfigCapability = NpuSharingConfigCapability{
		CapabilityItem: CapabilityItem{
			name:   npuSharingConfigKey,
			enable: false,
			mgr:    &CapabilityMgr{},
		}}

	capabilities = map[string]CapabilityIntf{
		npuSharingConfigKey: &NpuSharingConfigCapability{
			CapabilityItem: CapabilityItem{
				name:   npuSharingConfigKey,
				enable: true,
				mgr:    capabilityInstance,
			}},
	}
	npuSharingCapability = NpuSharingCapability{
		CapabilityItem: CapabilityItem{
			name:   npuSharingKey,
			enable: false,
			mgr:    &CapabilityMgr{capabilities: capabilities},
		}}
)

func TestNpuSharingConfigCapability(t *testing.T) {
	convey.Convey("load should be success", t, testConfigLoad)
	convey.Convey("get name should be success", t, testConfigGetName)
	convey.Convey("open and close should be success", t, testConfigOpenAndClose)
}

func testConfigLoad() {
	output1 := `NPU ID                         : 0
        Chip Count                     : 1
        Chip ID                        : 0
        Product Type                   : Atlas 500 A2`
	output2 := `+--------------------------------------------------------------------------------------------------------+
| npu-smi 23.0.rc1.b060                            Version: 23.0.rc1.b060                                |
+-------------------------------+-----------------+------------------------------------------------------+
| NPU     Name                  | Health          | Power(W)     Temp(C)           Hugepages-Usage(page) |
| Chip    Device                | Bus-Id          | AICore(%)    Memory-Usage(MB)                        |
+===============================+=================+======================================================+
| 0       310B                  | OK              | 0.0          0                 15    / 1400          |
| 0       0                     | NA              | 0            2376 / 11578                            |
+===============================+=================+======================================================+`
	output3 := `NPU ID                         : 0
        Chip Count                     : 1
        Chip ID                        : 0
        Product Type                   : Atlas 500 A1`
	output4 := `+--------------------------------------------------------------------------------------------------------+
| npu-smi 23.0.rc1.b060                            Version: 23.0.rc1.b060                                |
+-------------------------------+-----------------+------------------------------------------------------+
| NPU     Name                  | Health          | Power(W)     Temp(C)           Hugepages-Usage(page) |
| Chip    Device                | Bus-Id          | AICore(%)    Memory-Usage(MB)                        |
+===============================+=================+======================================================+
| 0       31B                  | OK              | 0.0          0                 15    / 1400          |
| 0       0                     | NA              | 0            2376 / 11578                            |
+===============================+=================+======================================================+`

	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{output1, nil}},
		{Values: gomonkey.Params{output2, nil}},
		{Values: gomonkey.Params{output1, testErr}},
		{Values: gomonkey.Params{output1, nil}},
		{Values: gomonkey.Params{output2, testErr}},
		{Values: gomonkey.Params{output3, nil}},
		{Values: gomonkey.Params{output1, nil}},
		{Values: gomonkey.Params{output4, nil}},
	}
	var p1 = gomonkey.ApplyFuncSeq(envutils.RunCommand, outputs)
	defer p1.Reset()

	npuSharingConfigCapability.load()
	convey.So(npuSharingConfigCapability.enable, convey.ShouldEqual, true)
	npuSharingConfigCapability.load()
	convey.So(npuSharingConfigCapability.enable, convey.ShouldEqual, false)
	npuSharingConfigCapability.load()
	convey.So(npuSharingConfigCapability.enable, convey.ShouldEqual, false)
	npuSharingConfigCapability.load()
	convey.So(npuSharingConfigCapability.enable, convey.ShouldEqual, false)
	npuSharingConfigCapability.load()
	convey.So(npuSharingConfigCapability.enable, convey.ShouldEqual, false)
}

func testConfigGetName() {
	name := npuSharingConfigCapability.getName()
	convey.So(name, convey.ShouldEqual, npuSharingConfigKey)
}

func testConfigOpenAndClose() {
	err := npuSharingConfigCapability.open()
	convey.So(err, convey.ShouldBeNil)
	err = npuSharingConfigCapability.close()
	convey.So(err, convey.ShouldBeNil)
}

func TestNpuSharingCapability(t *testing.T) {
	convey.Convey("load should be success", t, testLoad)
	convey.Convey("get name should be success", t, testGetName)
	convey.Convey("open should be success", t, testOpen)
	convey.Convey("close should be success", t, testClose)
}

func testLoad() {
	output1 := `Device-share Status            : False`
	output2 := `Device-share Status            : True`
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{output1, nil}},
		{Values: gomonkey.Params{output2, nil}},
		{Values: gomonkey.Params{output2, testErr}},
	}
	var p1 = gomonkey.ApplyFuncSeq(envutils.RunCommand, outputs)
	defer p1.Reset()

	npuSharingCapability.load()
	convey.So(npuSharingCapability.enable, convey.ShouldEqual, false)
	npuSharingCapability.load()
	convey.So(npuSharingCapability.enable, convey.ShouldEqual, true)
	npuSharingCapability.load()
	convey.So(npuSharingCapability.enable, convey.ShouldEqual, true)
}

func testGetName() {
	name := npuSharingCapability.getName()
	convey.So(name, convey.ShouldEqual, npuSharingKey)
}

var (
	outputOk = `Status                         : OK
        Message                        : The device-share is set successfully.`
	outputErr = `Status                         : Error
        Message                        : The device-share is set failed.`
	output = []gomonkey.OutputCell{
		{Values: gomonkey.Params{outputOk, nil}},
		{Values: gomonkey.Params{outputErr, nil}},
		{Values: gomonkey.Params{outputOk, testErr}},
		{Values: gomonkey.Params{outputOk, nil}},
		{Values: gomonkey.Params{outputErr, nil}},
		{Values: gomonkey.Params{outputOk, testErr}},
	}
)

func testOpen() {
	npuSharingCapability.enable = false
	var p1 = gomonkey.ApplyFuncSeq(envutils.RunInteractCommand, output)
	defer p1.Reset()

	err := npuSharingCapability.open()
	convey.So(err, convey.ShouldBeNil)
	npuSharingCapability.enable = false
	err = npuSharingCapability.open()
	convey.So(err, convey.ShouldBeNil)
	err = npuSharingCapability.open()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("open npu sharing failed: %s", testErr.Error()))
	npuSharingCapability.enable = true
	err = npuSharingCapability.open()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("already open, do nothing"))
	npuSharingCapability.CapabilityItem.mgr = &CapabilityMgr{}
	err = npuSharingCapability.open()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("cannot open at this device"))
}

func testClose() {
	npuSharingCapability.CapabilityItem.mgr = &CapabilityMgr{capabilities: capabilities}
	var p1 = gomonkey.ApplyFuncSeq(envutils.RunInteractCommand, output)
	defer p1.Reset()

	err := npuSharingCapability.close()
	convey.So(err, convey.ShouldBeNil)
	npuSharingCapability.enable = true
	err = npuSharingCapability.close()
	convey.So(err, convey.ShouldBeNil)
	err = npuSharingCapability.close()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("close npu sharing failed: %s", testErr.Error()))
	npuSharingCapability.enable = false
	err = npuSharingCapability.close()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("already close, no need to do so"))
	npuSharingCapability.CapabilityItem.mgr = &CapabilityMgr{}
	err = npuSharingCapability.close()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("cannot turn on at this type of machine"))
}
