// Copyright (c) 2024. Huawei Technologies Co., Ltd.  All rights reserved.

package certmanager

import (
	"context"
	cryptox509 "crypto/x509"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindxedge/base/common"
)

func newMockEdgeSvcCertUpdater() *EdgeSvcCertUpdater {
	var instance EdgeSvcCertUpdater
	instance.CaCertName = common.WsCltName
	instance.ctx, instance.cancel = context.WithCancel(context.Background())
	return &instance
}

func newMockEdgeCaCertUpdater() *EdgeCaCertUpdater {
	var instance EdgeCaCertUpdater
	instance.CaCertName = common.WsCltName
	instance.ctx, instance.cancel = context.WithCancel(context.Background())
	return &instance
}

func TestCertUpdater(t *testing.T) {
	convey.Convey("test new cert updater", t, testNewCertUpdater)

	convey.Convey("test testCaCheckAndSetUpdateFlag", t, testCaCheckAndSetUpdateFlag)
	convey.Convey("test testCaClearUpdateFlag", t, testCaClearUpdateFlag)
	convey.Convey("test testCaCertNeedUpdate", t, testCaCertNeedUpdate)
	convey.Convey("test testCaPrepareCertUpdate", t, testCaPrepareCertUpdate)
	convey.Convey("test testCaNotifyCertUpdate", t, testCaNotifyCertUpdate)
	convey.Convey("test testCaPostCertUpdate", t, testCaPostCertUpdate)
	convey.Convey("test testCaForceUpdateCheck", t, testCaForceUpdateCheck)
	convey.Convey("test testCaDoForceUpdate", t, testCaDoForceUpdate)

	convey.Convey("test testCaCheckAndSetUpdateFlag", t, testSvcCheckAndSetUpdateFlag)
	convey.Convey("test testCaClearUpdateFlag", t, testSvcClearUpdateFlag)
	convey.Convey("test testCaCertNeedUpdate", t, testSvcCertNeedUpdate)
	convey.Convey("test testSvcPrepareCertUpdate", t, testSvcPrepareCertUpdate)
	convey.Convey("test testCaNotifyCertUpdate", t, testSvcNotifyCertUpdate)
	convey.Convey("test testCaPostCertUpdate", t, testSvcPostCertUpdate)
	convey.Convey("test testSvcForceUpdateCheck", t, testSvcForceUpdateCheck)
	convey.Convey("test testSvcDoForceUpdate", t, testSvcDoForceUpdate)
}

func testNewCertUpdater() {
	convey.Convey("case: normal success", func() {
		caUpdater := NewCertUpdater(CertTypeEdgeCa)
		convey.So(caUpdater, convey.ShouldHaveSameTypeAs, &EdgeCaCertUpdater{})

		svcUpdater := NewCertUpdater(CertTypeEdgeSvc)
		convey.So(svcUpdater, convey.ShouldHaveSameTypeAs, &EdgeSvcCertUpdater{})
	})

	convey.Convey("case: wrong type", func() {
		certUpdater := NewCertUpdater("wrong type")
		convey.So(certUpdater, convey.ShouldEqual, nil)
	})
}

func testCaCheckAndSetUpdateFlag() {
	convey.Convey("case: update flag in or not in updating", func() {
		updater := newMockEdgeCaCertUpdater()
		convey.So(updater.CheckAndSetUpdateFlag(), convey.ShouldBeNil)
		convey.So(updater.CheckAndSetUpdateFlag(), convey.ShouldResemble,
			fmt.Errorf("edge root ca cert is in updating, try it later"))
		defer atomic.StoreInt64(&edgeCaUpdatingFlag, NotUpdating)
	})
}

func testCaClearUpdateFlag() {
	convey.Convey("case: normal success", func() {
		updater := newMockEdgeCaCertUpdater()
		atomic.StoreInt64(&edgeCaUpdatingFlag, InUpdating)
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			updater.cancel()
		}()
		updater.ClearUpdateFlag()
		convey.So(atomic.LoadInt64(&edgeCaUpdatingFlag), convey.ShouldEqual, NotUpdating)
	})
}

func testCaCertNeedUpdate() {
	updater := newMockEdgeCaCertUpdater()
	convey.Convey("case: cert validation failed", func() {
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, false, false, test.ErrTest)
		defer patches.Reset()
		_, _, err := updater.IsCertNeedUpdate()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("case: cert need to be updated by force", func() {
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, false, true, nil)
		defer patches.Reset()
		needUpdate, needForceUpdate, err := updater.IsCertNeedUpdate()
		convey.So(err, convey.ShouldBeNil)
		convey.So(needUpdate, convey.ShouldBeTrue)
		convey.So(needForceUpdate, convey.ShouldBeTrue)
	})

	convey.Convey("case: cert need to updated", func() {
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, true, false, nil)
		defer patches.Reset()
		needUpdate, needForceUpdate, err := updater.IsCertNeedUpdate()
		convey.So(err, convey.ShouldBeNil)
		convey.So(needUpdate, convey.ShouldBeTrue)
		convey.So(needForceUpdate, convey.ShouldBeFalse)
	})

	convey.Convey("case: cert do not need update", func() {
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, false, false, nil)
		defer patches.Reset()
		needUpdate, needForceUpdate, err := updater.IsCertNeedUpdate()
		convey.So(err, convey.ShouldBeNil)
		convey.So(needUpdate, convey.ShouldBeFalse)
		convey.So(needForceUpdate, convey.ShouldBeFalse)
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})
}

func testCaPrepareCertUpdate() {
	updater := newMockEdgeCaCertUpdater()
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(CreateTempCaCert, "", nil)
		defer patches.Reset()
		convey.So(updater.PrepareCertUpdate(), convey.ShouldBeNil)
	})

	convey.Convey("case: case: create temp cert failed", func() {
		patches := gomonkey.ApplyFuncReturn(CreateTempCaCert, "", test.ErrTest)
		defer patches.Reset()
		convey.So(updater.PrepareCertUpdate(), convey.ShouldResemble, test.ErrTest)
	})
}

func testCaNotifyCertUpdate() {
	updater := newMockEdgeCaCertUpdater()
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(sendCertUpdateNotify, nil)
		defer patches.Reset()
		convey.So(updater.NotifyCertUpdate(), convey.ShouldBeNil)
	})

	convey.Convey("case: send cert update notify failed", func() {
		patches := gomonkey.ApplyFuncReturn(sendCertUpdateNotify, test.ErrTest)
		defer patches.Reset()
		convey.So(updater.NotifyCertUpdate(), convey.ShouldResemble, test.ErrTest)
	})
}

func testCaPostCertUpdate() {
	updater := newMockEdgeCaCertUpdater()
	convey.Convey("case: process canceled", func() {
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			updater.cancel()
		}()
		updater.PostCertUpdate()
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})

	convey.Convey("case: update failed", func() {
		updateResult := certUpdateResult{ResultCode: updateFailedCode}
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			if edgeCaResultChan != nil {
				edgeCaResultChan <- updateResult
			}
		}()
		updater.PostCertUpdate()
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})

	convey.Convey("case: force update failed", func() {
		updateResult := certUpdateResult{ResultCode: updateSuccessCode}
		patches := gomonkey.ApplyPrivateMethod(updater, "DoForceUpdate", func(*EdgeSvcCertUpdater) error {
			return test.ErrTest
		})
		defer patches.Reset()
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			if edgeCaResultChan != nil {
				edgeCaResultChan <- updateResult
			}
		}()
		updater.PostCertUpdate()
		convey.So(updater.ctx.Err(), convey.ShouldBeNil)
	})
}

func testCaForceUpdateCheck() {
	testCaForceUpdateCheckSucceedCases()
	testCaForceUpdateCheckFailedCases()
}

func testCaForceUpdateCheckSucceedCases() {
	convey.Convey("case: do force update succeed for the first call", func() {
		updater := newMockEdgeCaCertUpdater()
		var invokedFlag bool
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, true, true, nil).
			ApplyPrivateMethod(updater, "DoForceUpdate", func(*EdgeCaCertUpdater) error {
				invokedFlag = true
				return nil
			})
		defer patches.Reset()
		updater.ForceUpdateCheck()
		convey.So(invokedFlag, convey.ShouldEqual, true)
	})

	convey.Convey("case: do force update for later check", func() {
		updater := newMockEdgeCaCertUpdater()
		var invokedFlag bool
		outputCells := []gomonkey.OutputCell{
			{Values: gomonkey.Params{true, false, nil}, Times: 1},
			{Values: gomonkey.Params{false, false, test.ErrTest}, Times: 1},
			{Values: gomonkey.Params{false, false, nil}, Times: 1},
			{Values: gomonkey.Params{false, true, nil}, Times: 1},
		}
		patches := gomonkey.ApplyFuncReturn(time.NewTicker, time.NewTicker(defaultParallelExecWaitTime)).
			ApplyFuncSeq(checkCertValidity, outputCells).
			ApplyPrivateMethod(updater, "DoForceUpdate", func(*EdgeCaCertUpdater) error {
				invokedFlag = true
				return nil
			})
		defer patches.Reset()
		updater.ForceUpdateCheck()
		convey.So(invokedFlag, convey.ShouldEqual, true)
	})
}

func testCaForceUpdateCheckFailedCases() {
	convey.Convey("case: cert invalid", func() {
		updater := newMockEdgeCaCertUpdater()
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, false, false, test.ErrTest)
		defer patches.Reset()
		updater.ForceUpdateCheck()
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})

	convey.Convey("case: do force update failed for the first call", func() {
		updater := newMockEdgeCaCertUpdater()
		var invokedFlag bool
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, true, true, nil).
			ApplyPrivateMethod(updater, "DoForceUpdate", func(*EdgeCaCertUpdater) error {
				invokedFlag = true
				return test.ErrTest
			})
		defer patches.Reset()
		updater.ForceUpdateCheck()
		convey.So(invokedFlag, convey.ShouldEqual, true)
	})

	convey.Convey("case: do force update cancel by other func", func() {
		updater := newMockEdgeCaCertUpdater()
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, true, false, nil)
		defer patches.Reset()
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			updater.cancel()
		}()
		updater.ForceUpdateCheck()
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})
}

func testCaDoForceUpdate() {
	updater := newMockEdgeCaCertUpdater()
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(execForceUpdate, testContent, nil)
		defer patches.Reset()
		convey.So(updater.DoForceUpdate(), convey.ShouldBeNil)
		convey.So(updater.TempCaCertContent, convey.ShouldEqual, testContent)
	})

	convey.Convey("case: exec force update failed", func() {
		patches := gomonkey.ApplyFuncReturn(execForceUpdate, nil, test.ErrTest)
		defer patches.Reset()
		convey.So(updater.DoForceUpdate(), convey.ShouldResemble,
			fmt.Errorf("do force update process for ca cert [%v] failed: %v", updater.CaCertName, test.ErrTest))
	})
}

func testSvcCheckAndSetUpdateFlag() {
	convey.Convey("case: update flag in or not in updating", func() {
		updater := newMockEdgeSvcCertUpdater()
		convey.So(updater.CheckAndSetUpdateFlag(), convey.ShouldBeNil)
		convey.So(updater.CheckAndSetUpdateFlag(), convey.ShouldResemble,
			fmt.Errorf("edge service cert is in updating, try it later"))
		defer atomic.StoreInt64(&edgeSvcUpdatingFlag, NotUpdating)
	})
}

func testSvcClearUpdateFlag() {
	convey.Convey("case: normal success", func() {
		updater := newMockEdgeSvcCertUpdater()
		atomic.StoreInt64(&edgeSvcUpdatingFlag, InUpdating)
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			updater.cancel()
		}()
		updater.ClearUpdateFlag()
		convey.So(atomic.LoadInt64(&edgeSvcUpdatingFlag), convey.ShouldEqual, NotUpdating)
	})
}

func testSvcCertNeedUpdate() {
	updater := newMockEdgeSvcCertUpdater()
	convey.Convey("case: cert validation failed", func() {
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, false, false, test.ErrTest)
		defer patches.Reset()
		_, _, err := updater.IsCertNeedUpdate()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("case: cert need to be updated by force", func() {
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, false, true, nil)
		defer patches.Reset()
		needUpdate, needForceUpdate, err := updater.IsCertNeedUpdate()
		convey.So(err, convey.ShouldBeNil)
		convey.So(needUpdate, convey.ShouldBeTrue)
		convey.So(needForceUpdate, convey.ShouldBeTrue)
	})

	convey.Convey("case: cert need to updated", func() {
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, true, false, nil)
		defer patches.Reset()
		needUpdate, needForceUpdate, err := updater.IsCertNeedUpdate()
		convey.So(err, convey.ShouldBeNil)
		convey.So(needUpdate, convey.ShouldBeTrue)
		convey.So(needForceUpdate, convey.ShouldBeFalse)
	})

	convey.Convey("case: cert do not need update", func() {
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, false, false, nil)
		defer patches.Reset()
		needUpdate, needForceUpdate, err := updater.IsCertNeedUpdate()
		convey.So(err, convey.ShouldBeNil)
		convey.So(needUpdate, convey.ShouldBeFalse)
		convey.So(needForceUpdate, convey.ShouldBeFalse)
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})
}

func testSvcPrepareCertUpdate() {
	updater := newMockEdgeSvcCertUpdater()
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(CreateTempCaCert, "", nil)
		defer patches.Reset()
		convey.So(updater.PrepareCertUpdate(), convey.ShouldBeNil)
	})

	convey.Convey("case: create temp cert failed", func() {
		patches := gomonkey.ApplyFuncReturn(CreateTempCaCert, "", test.ErrTest)
		defer patches.Reset()
		convey.So(updater.PrepareCertUpdate(), convey.ShouldResemble, test.ErrTest)
	})
}

func testSvcNotifyCertUpdate() {
	updater := newMockEdgeSvcCertUpdater()
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(sendCertUpdateNotify, nil)
		defer patches.Reset()
		convey.So(updater.NotifyCertUpdate(), convey.ShouldBeNil)
	})

	convey.Convey("case: send cert update notify failed", func() {
		patches := gomonkey.ApplyFuncReturn(sendCertUpdateNotify, test.ErrTest)
		defer patches.Reset()
		convey.So(updater.NotifyCertUpdate(), convey.ShouldResemble, test.ErrTest)
	})
}

func testSvcPostCertUpdate() {
	updater := newMockEdgeSvcCertUpdater()
	convey.Convey("case: process canceled", func() {
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			updater.cancel()
		}()
		updater.PostCertUpdate()
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})

	convey.Convey("case: update failed", func() {
		updateResult := certUpdateResult{ResultCode: updateFailedCode}
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			if edgeSvcResultChan != nil {
				edgeSvcResultChan <- updateResult
			}
		}()
		updater.PostCertUpdate()
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})

	convey.Convey("case: force update failed", func() {
		updateResult := certUpdateResult{ResultCode: updateSuccessCode}
		patches := gomonkey.ApplyPrivateMethod(updater, "DoForceUpdate", func(*EdgeSvcCertUpdater) error {
			return test.ErrTest
		})
		defer patches.Reset()
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			if edgeSvcResultChan != nil {
				edgeSvcResultChan <- updateResult
			}
		}()
		updater.PostCertUpdate()
		convey.So(updater.ctx.Err(), convey.ShouldBeNil)
	})
}

func testSvcForceUpdateCheck() {
	testSvcForceUpdateCheckSucceedCases()
	testSvcForceUpdateCheckFailedCases()
}

func testSvcForceUpdateCheckSucceedCases() {
	convey.Convey("case: do force update succeed for the first call", func() {
		updater := newMockEdgeSvcCertUpdater()
		var invokedFlag bool
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, true, true, nil).
			ApplyPrivateMethod(updater, "DoForceUpdate", func(*EdgeSvcCertUpdater) error {
				invokedFlag = true
				return nil
			})
		defer patches.Reset()
		updater.ForceUpdateCheck()
		convey.So(invokedFlag, convey.ShouldEqual, true)
	})

	convey.Convey("case: do force update for later check", func() {
		updater := newMockEdgeSvcCertUpdater()
		var invokedFlag bool
		outputCells := []gomonkey.OutputCell{
			{Values: gomonkey.Params{true, false, nil}, Times: 1},
			{Values: gomonkey.Params{false, false, test.ErrTest}, Times: 1},
			{Values: gomonkey.Params{false, false, nil}, Times: 1},
			{Values: gomonkey.Params{false, true, nil}, Times: 1},
		}
		patches := gomonkey.ApplyFuncReturn(time.NewTicker, time.NewTicker(defaultParallelExecWaitTime)).
			ApplyFuncSeq(checkCertValidity, outputCells).
			ApplyPrivateMethod(updater, "DoForceUpdate", func(*EdgeSvcCertUpdater) error {
				invokedFlag = true
				return nil
			})
		defer patches.Reset()
		updater.ForceUpdateCheck()

		convey.So(invokedFlag, convey.ShouldEqual, true)
	})
}

func testSvcForceUpdateCheckFailedCases() {
	convey.Convey("case: cert invalid", func() {
		updater := newMockEdgeSvcCertUpdater()
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, false, false, test.ErrTest)
		defer patches.Reset()
		updater.ForceUpdateCheck()
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})

	convey.Convey("case: do force update failed for the first call", func() {
		updater := newMockEdgeSvcCertUpdater()
		var invokedFlag bool
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, true, true, nil).
			ApplyPrivateMethod(updater, "DoForceUpdate", func(*EdgeSvcCertUpdater) error {
				invokedFlag = true
				return test.ErrTest
			})
		defer patches.Reset()
		updater.ForceUpdateCheck()
		convey.So(invokedFlag, convey.ShouldEqual, true)
	})

	convey.Convey("case: do force update cancel by other func", func() {
		updater := newMockEdgeSvcCertUpdater()
		patches := gomonkey.ApplyFuncReturn(checkCertValidity, true, false, nil)
		defer patches.Reset()
		go func() {
			time.Sleep(defaultParallelExecWaitTime)
			updater.cancel()
		}()
		updater.ForceUpdateCheck()
		convey.So(updater.ctx.Err(), convey.ShouldNotBeNil)
	})
}

func testSvcDoForceUpdate() {
	updater := newMockEdgeSvcCertUpdater()
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(execForceUpdate, testContent, nil)
		defer patches.Reset()
		convey.So(updater.DoForceUpdate(), convey.ShouldBeNil)
		convey.So(updater.TempCaCertContent, convey.ShouldEqual, testContent)
	})

	convey.Convey("case: exec force update failed", func() {
		patches := gomonkey.ApplyFuncReturn(execForceUpdate, nil, test.ErrTest)
		defer patches.Reset()
		convey.So(updater.DoForceUpdate(), convey.ShouldResemble,
			fmt.Errorf("do force update process for ca cert [%v] failed: %v", updater.CaCertName, test.ErrTest))
	})
}

func TestExecForceUpdate(t *testing.T) {
	updater := newMockEdgeSvcCertUpdater()
	convey.Convey("case: normal success", t, func() {
		patches := gomonkey.ApplyFuncReturn(CreateTempCaCert, testContent, nil).
			ApplyFuncReturn(sendCertUpdateNotify, nil).
			ApplyFuncReturn(UpdateCaCertWithTemp, nil).
			ApplyFuncReturn(RemoveTempCaCert, nil)
		defer patches.Reset()
		tempContent, err := execForceUpdate(updater.CaCertName, CertTypeEdgeSvc)
		convey.So(err, convey.ShouldBeNil)
		convey.So(tempContent, convey.ShouldEqual, testContent)
	})

	convey.Convey("case: create temp cert failed", t, func() {
		patches := gomonkey.ApplyFuncReturn(CreateTempCaCert, nil, test.ErrTest)
		defer patches.Reset()
		_, err := execForceUpdate(updater.CaCertName, CertTypeEdgeSvc)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("create or load temp ca cert [%v] failed: %v", updater.CaCertName, test.ErrTest))
	})

	convey.Convey("case: send cert update notify failed", t, func() {
		patches := gomonkey.ApplyFuncReturn(CreateTempCaCert, testContent, nil).
			ApplyFuncReturn(sendCertUpdateNotify, test.ErrTest)
		defer patches.Reset()
		_, err := execForceUpdate(updater.CaCertName, CertTypeEdgeSvc)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("send cert [%v] force update notify failed: %v", updater.CaCertName, test.ErrTest))
	})

	convey.Convey("case: update cert with temp failed", t, func() {
		patches := gomonkey.ApplyFuncReturn(CreateTempCaCert, testContent, nil).
			ApplyFuncReturn(sendCertUpdateNotify, nil).
			ApplyFuncReturn(UpdateCaCertWithTemp, test.ErrTest)
		defer patches.Reset()
		_, err := execForceUpdate(updater.CaCertName, CertTypeEdgeSvc)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("update local ca cert failed: %v", test.ErrTest))
	})
}

func TestCheckCertValidity(t *testing.T) {
	convey.Convey("successful cases", t, testCheckCertValiditySucceedCases)
	convey.Convey("failed cases", t, testCheckCertValidityFailedCases)
}

func testCheckCertValiditySucceedCases() {
	convey.Convey("case: no cert filed found", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, nil)
		defer patches.Reset()
		needUpdate, needForceUpdate, err := checkCertValidity(CertTypeEdgeCa)
		convey.So(needUpdate, convey.ShouldBeFalse)
		convey.So(needForceUpdate, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("case: no need to update", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, []byte(testContent), nil).
			ApplyFuncReturn(x509.LoadCertsFromPEM, &cryptox509.Certificate{}, nil).
			ApplyFuncReturn(x509.GetValidityPeriod, float64(1), nil)
		defer patches.Reset()
		needUpdate, needForceUpdate, err := checkCertValidity(CertTypeEdgeCa)
		convey.So(needUpdate, convey.ShouldBeFalse)
		convey.So(needForceUpdate, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldBeNil)
	})

}

func testCheckCertValidityFailedCases() {
	convey.Convey("case: load file failed", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer patches.Reset()
		_, _, err := checkCertValidity(CertTypeEdgeCa)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("case: load certs failed", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, []byte(testContent), nil).
			ApplyFuncReturn(x509.LoadCertsFromPEM, nil, test.ErrTest)
		defer patches.Reset()
		_, _, err := checkCertValidity(CertTypeEdgeCa)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("case: get validity period failed", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, []byte(testContent), nil).
			ApplyFuncReturn(x509.LoadCertsFromPEM, &cryptox509.Certificate{}, nil).
			ApplyFuncReturn(x509.GetValidityPeriod, nil, test.ErrTest)
		defer patches.Reset()
		_, _, err := checkCertValidity(CertTypeEdgeCa)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestSendCertUpdateNotify(t *testing.T) {
	convey.Convey("successful cases", t, testSendCertUpdateNotifySucceedCases)
	convey.Convey("failed cases", t, testSendCertUpdateNotifyFailedCases)
}

func testSendCertUpdateNotifySucceedCases() {
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(json.Marshal, nil, nil).
			ApplyFunc(json.Unmarshal, func(data []byte, v any) error {
				resp, _ := v.(*common.RespMsg)
				resp.Status = common.Success
				return nil
			}).
			ApplyPrivateMethod(&httpsmgr.HttpsRequest{}, "PostJson",
				func(*httpsmgr.HttpsRequest, []byte) ([]byte, error) {
					return nil, nil
				})
		defer patches.Reset()
		convey.So(sendCertUpdateNotify(CertUpdatePayload{}), convey.ShouldBeNil)
	})

	convey.Convey("case: only resp status failed", func() {
		patches := gomonkey.ApplyFuncReturn(json.Marshal, nil, nil).
			ApplyFunc(json.Unmarshal, func(data []byte, v any) error {
				resp, _ := v.(*common.RespMsg)
				resp.Status = common.ErrorGetResponse
				resp.Msg = test.ErrTest.Error()
				return nil
			}).
			ApplyPrivateMethod(&httpsmgr.HttpsRequest{}, "PostJson",
				func(*httpsmgr.HttpsRequest, []byte) ([]byte, error) {
					return nil, nil
				})
		defer patches.Reset()
		convey.So(sendCertUpdateNotify(CertUpdatePayload{}), convey.ShouldResemble, fmt.Errorf(
			"cert update operation failed, result status:%s, msg:%s", common.ErrorGetResponse, test.ErrTest))
	})
}

func testSendCertUpdateNotifyFailedCases() {
	convey.Convey("case: marshal failed", func() {
		patches := gomonkey.ApplyFuncReturn(json.Marshal, nil, test.ErrTest)
		defer patches.Reset()
		convey.So(sendCertUpdateNotify(CertUpdatePayload{}), convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("case: https post json failed", func() {
		patches := gomonkey.ApplyFuncReturn(json.Marshal, nil, nil).
			ApplyPrivateMethod(&httpsmgr.HttpsRequest{}, "PostJson",
				func(*httpsmgr.HttpsRequest, []byte) ([]byte, error) {
					return nil, test.ErrTest
				}).
			ApplyFunc(json.Unmarshal, func(data []byte, v any) error {
				resp, _ := v.(*common.RespMsg)
				resp.Status = common.Success
				return nil
			})
		defer patches.Reset()
		convey.So(sendCertUpdateNotify(CertUpdatePayload{}), convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("case: unmarshal failed", func() {
		patches := gomonkey.ApplyFuncReturn(json.Marshal, nil, nil).
			ApplyPrivateMethod(&httpsmgr.HttpsRequest{}, "PostJson",
				func(*httpsmgr.HttpsRequest, []byte) ([]byte, error) {
					return nil, nil
				}).
			ApplyFuncReturn(json.Unmarshal, test.ErrTest)
		defer patches.Reset()
		convey.So(sendCertUpdateNotify(CertUpdatePayload{}), convey.ShouldResemble, test.ErrTest)
	})
}
