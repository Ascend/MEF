// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build !MEFEdge_A500

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
)

const (
	dmidecodeRes = "Serial Number: eb63bbc9 dec9 40d2 9d91 81e769d5ae19"
	elabelRes    = "[elabel_data] 2102312NSF10K8000130"
	testSnName   = "serialNumber"
	testSnValue  = "eb63bbc9-dec9-40d2-9d91-81e769d5ae19"
)

func TestGetSerialNumber(t *testing.T) {
	convey.Convey("TestGetSerialNumber", t, func() {
		convey.Convey("read File\n", func() {
			patches2 := gomonkey.ApplyFunc(fileutils.LoadFile, mockReadFileForGetSnFromFile)
			defer patches2.Reset()
			_, err := GetSerialNumber("NULL")
			convey.So(err, convey.ShouldBeNil)
		})
	})

	convey.Convey("test func GetSerialNumber success, is 500", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, elabelRes, nil)
		defer p1.Reset()
		_, err := GetSerialNumber("")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func GetSerialNumber success, get sn from file", t, func() {
		data, err := json.Marshal(map[string]string{testSnName: testSnValue})
		convey.So(err, convey.ShouldBeNil)
		patches := gomonkey.ApplyFuncReturn(getA500Sn, "", test.ErrTest).
			ApplyFuncReturn(getA500ProSn, "", test.ErrTest).
			ApplyFuncReturn(fileutils.LoadFile, data, nil)
		defer patches.Reset()
		_, err = GetSerialNumber("")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetA500ProSn(t *testing.T) {
	convey.Convey("test func getA500ProSn success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, dmidecodeRes, nil)
		defer p1.Reset()
		_, err := getA500ProSn()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func getA500ProSn failed, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		_, err := getA500ProSn()
		expErr := fmt.Errorf("get a500pro serial number failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func getA500ProSn failed, sn not found", t, func() {
		const errDmidecodeRes = "Serial Number"
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, errDmidecodeRes, nil)
		defer p1.Reset()
		_, err := getA500ProSn()
		expErr := errors.New("get a500pro serial number failed, error: serial number not found")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func mockReadFileForGetSnFromFile(_ string, _ ...fileutils.FileChecker) ([]byte, error) {
	date := []byte(`{"serialNumber":"eb63bbc9-dec9-40d2-9d91-81e769d5ae19"}`)
	return date, nil
}

func TestGetSnFromFile(t *testing.T) {
	const testErrSnName = "errSnName"
	const testErrSnValue = "?err-sn-value"
	data, err := json.Marshal(map[string]string{testSnName: testSnValue})
	if err != nil {
		panic(err)
	}
	patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, data, nil)
	defer patches.Reset()
	convey.Convey("test func getSnFromFile success", t, func() {
		sn := getSnFromFile("")
		convey.So(sn, convey.ShouldResemble, testSnValue)
	})

	convey.Convey("test func getSnFromFile failed, load sn file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer p1.Reset()
		sn := getSnFromFile("")
		convey.So(sn, convey.ShouldResemble, "")
	})

	convey.Convey("test func getSnFromFile failed, unmarshal failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(json.Unmarshal, test.ErrTest)
		defer p1.Reset()
		sn := getSnFromFile("")
		convey.So(sn, convey.ShouldResemble, "")
	})

	convey.Convey("test func getSnFromFile failed, file map failed", t, func() {
		data, err = json.Marshal(map[string]string{testErrSnName: testSnValue})
		convey.So(err, convey.ShouldBeNil)
		var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, data, nil)
		defer p1.Reset()
		sn := getSnFromFile("")
		convey.So(sn, convey.ShouldResemble, "")
	})

	convey.Convey("test func getSnFromFile failed, file map failed", t, func() {
		data, err = json.Marshal(map[string]string{testSnName: testErrSnValue})
		convey.So(err, convey.ShouldBeNil)
		var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, data, nil)
		defer p1.Reset()
		sn := getSnFromFile("")
		convey.So(sn, convey.ShouldResemble, "")
	})
}
