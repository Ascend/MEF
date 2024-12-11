// Copyright (c) 2024. Huawei Technologies Co., Ltd.  All rights reserved.

package certmanager

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	"huawei.com/mindxedge/base/common"
)

const testContent = "testContent"
const defaultParallelExecWaitTime = time.Millisecond * 500

func TestMain(m *testing.M) {
	tcBaseWithDb := &test.TcBaseWithDb{}
	test.RunWithPatches(tcBaseWithDb, m, gomonkey.ApplyFuncReturn(fileutils.WriteData, nil))
}

func newMsgWithContentForUT(v interface{}) *model.Message {
	msg, err := model.NewMessage()
	if err != nil {
		panic(err)
	}
	err = msg.FillContent(v)
	if err != nil {
		panic(err)
	}
	return msg
}

func TestOnCertOrCrlChanged(t *testing.T) {
	convey.Convey("case: update client cert success", t, func() {
		var certAndCrlPaths = []string{
			getRootCaPath(common.ImageCertName),
			getRootCaPath(common.SoftwareCertName),
			getCrlPath(common.ImageCertName),
			getCrlPath(common.SoftwareCertName),
		}
		var invokeFlag bool
		patch := gomonkey.ApplyFunc(updateClientCert, func(certName, operation string) error {
			invokeFlag = true
			return nil
		})
		defer patch.Reset()
		onCertOrCrlChanged(certAndCrlPaths)
		convey.So(invokeFlag, convey.ShouldBeTrue)
	})
}

func TestDoCheckProcess(t *testing.T) {
	convey.Convey("case: update client cert success", t, func() {
		updater := newMockEdgeSvcCertUpdater()
		var invokeFlag bool
		patches := gomonkey.ApplyMethodReturn(updater, "CheckAndSetUpdateFlag", nil).
			ApplyMethod(updater, "ClearUpdateFlag", func(*EdgeSvcCertUpdater) {}).
			ApplyMethodReturn(updater, "IsCertNeedUpdate", true, true, nil).
			ApplyMethodReturn(updater, "DoForceUpdate", nil).
			ApplyMethodReturn(updater, "PrepareCertUpdate", nil).
			ApplyMethodReturn(updater, "NotifyCertUpdate", nil).
			ApplyMethod(updater, "PostCertUpdate", func(*EdgeSvcCertUpdater) {}).
			ApplyMethod(updater, "ForceUpdateCheck", func(*EdgeSvcCertUpdater) {
				invokeFlag = true
			})
		defer patches.Reset()
		doCheckProcess(updater)
		time.Sleep(defaultParallelExecWaitTime)
		convey.So(invokeFlag, convey.ShouldBeTrue)
	})
}
