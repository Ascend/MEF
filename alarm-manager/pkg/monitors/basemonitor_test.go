// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package monitors test for basemonitor.go
package monitors

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

func TestAlarmMonitor(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&common.DbMgr{}, "GetAlarmConfig", 1, nil).
		ApplyMethodReturn(&requests.ReqCertParams{}, "GetImportedCertsInfo",
			"test cert info", nil).
		ApplyFuncReturn(x509.CheckCertsOverdue, nil)
	defer patches.Reset()
	convey.Convey("test func Monitoring, stop by cancel", t, testMonitoring)
	convey.Convey("test func CollectOnce", t, testCollectOnce)
	convey.Convey("test func CollectOnce, alarmIdFuncMap is nil", t, testCollectOnceErrFuncMap)
	convey.Convey("test func CollectOnce, alarmIdFuncMap is nil and length is zero", t, testCollectOnceErrFuncMap)
	convey.Convey("test func CollectOnce, create alarm failed", t, testCollectOnceErrCreateAlarm)
	convey.Convey("test func CollectOnce, send alarms failed", t, testCollectOnceErrSendAlarms)
	convey.Convey("test func CollectOnce, cert is nil", t, testCollectOnceCertNil)
	convey.Convey("test func CollectOnce, check cert failed", t, testCollectOnceErrCheckCert)
	convey.Convey("test func certReset", t, testCertReset)
}

func testMonitoring() {
	certTask.alarmIdFuncMap = make(map[string]func() error, importedCertsNum)
	certTask.alarmIdFuncMap[alarms.NorthCertAbnormal] = isNorthCertOverdue
	certTask.alarmIdFuncMap[alarms.SoftwareCertAbnormal] = isSoftwareCertOverdue
	certTask.alarmIdFuncMap[alarms.ImageCertAbnormal] = isImageCertOverdue
	certTask.interval = 1 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	go certTask.Monitoring(ctx)
	cancel()
}

func testCollectOnce() {
	certTask.CollectOnce()
}

func testCollectOnceErrFuncMap() {
	certTask.alarmIdFuncMap = nil
	certTask.CollectOnce()

	certTask.alarmIdFuncMap = make(map[string]func() error)
	certTask.CollectOnce()
}

func testCollectOnceErrCreateAlarm() {
	var p1 = gomonkey.ApplyFuncReturn(alarms.CreateAlarm, nil, test.ErrTest)
	defer p1.Reset()

	certTask.alarmIdFuncMap = make(map[string]func() error, importedCertsNum)
	certTask.alarmIdFuncMap[alarms.NorthCertAbnormal] = isNorthCertOverdue
	certTask.alarmIdFuncMap[alarms.SoftwareCertAbnormal] = isSoftwareCertOverdue
	certTask.alarmIdFuncMap[alarms.ImageCertAbnormal] = isImageCertOverdue
	certTask.CollectOnce()
}

func testCollectOnceErrSendAlarms() {
	var p1 = gomonkey.ApplyFuncReturn(SendAlarms, test.ErrTest)
	defer p1.Reset()
	certTask.CollectOnce()
}

func testCollectOnceCertNil() {
	getCertsInfoFlag = true
	certTask.CollectOnce()
}

func testCollectOnceErrCheckCert() {
	importedCertsInfo.NorthCert = []byte("north cert")
	importedCertsInfo.SoftwareCert = []byte("software cert")
	importedCertsInfo.ImageCert = []byte("image cert")
	certTask.CollectOnce()

	var p1 = gomonkey.ApplyFuncReturn(x509.CheckCertsOverdue, test.ErrTest)
	defer p1.Reset()
	certTask.CollectOnce()
}

func testCertReset() {
	certReset()
}
