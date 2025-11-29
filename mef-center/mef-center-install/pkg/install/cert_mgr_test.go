// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

func CertMgrTest() {
	convey.Convey("CertMgr DoInstallPrepare func", CertMgrDoPrepareTest)
	convey.Convey("prepareCertsDir func", PrepareCertsDirTest)
	convey.Convey("certMgrPrepareCert func", CertMgrPrepareCertTest)
}

func CertMgrDoPrepareTest() {
	var ins = &certPrepareCtl{}
	convey.Convey("test DoInstallPrepare func in certPrepareCtl struct success", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "prepareCertsDir",
			func(_ *certPrepareCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareCerts", func(_ *certPrepareCtl) error { return nil })
		defer p.Reset()
		convey.So(ins.doPrepare(), convey.ShouldBeNil)
	})

	convey.Convey("test DoInstallPrepare func in certPrepareCtl struct failed", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "prepareCertsDir",
			func(_ *certPrepareCtl) error { return ErrTest })
		defer p.Reset()
		convey.So(ins.doPrepare(), convey.ShouldResemble, ErrTest)
	})
}

func PrepareCertsDirTest() {
	convey.Convey("test prepareCertDir", func() {
		InstallDirPathMgrIns, err := util.InitInstallDirPathMgr()
		convey.So(err, convey.ShouldBeNil)
		var ins = &certPrepareCtl{
			certPathMgr: InstallDirPathMgrIns.ConfigPathMgr,
			components:  []string{"edge-manager"},
		}

		convey.Convey("test prepareCertDir func success", func() {
			defer ResetAndClearDir(nil, InstallDirPathMgrIns.GetMefPath())
			convey.So(ins.prepareCertsDir(), convey.ShouldBeNil)
		})

		convey.Convey("test prepareCertDir func makesure root cert path failed", func() {
			p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, ErrTest)
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			convey.So(ins.prepareCertsDir(), convey.ShouldResemble, errors.New("create root certs path failed"))
		})

		convey.Convey("test prepareCertDir func makesure root key path failed", func() {
			p := gomonkey.ApplyFuncSeq(fileutils.CreateDir,
				[]gomonkey.OutputCell{{Values: gomonkey.Params{nil}}, {Values: gomonkey.Params{ErrTest}}})
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			convey.So(ins.prepareCertsDir(), convey.ShouldResemble, errors.New("create root key path failed"))
		})

		convey.Convey("test prepareCertDir func makesure component's cert path failed", func() {
			p := gomonkey.ApplyFuncSeq(fileutils.CreateDir,
				[]gomonkey.OutputCell{
					{Values: gomonkey.Params{nil}},
					{Values: gomonkey.Params{nil}},
					{Values: gomonkey.Params{ErrTest}},
				})
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			convey.So(ins.prepareCertsDir(), convey.ShouldResemble,
				errors.New("prepare component [edge-manager]'s cert dir failed"))
		})
	})
}

func CertMgrPrepareCertTest() {
	convey.Convey("test prepareCertDir", func() {
		InstallDirPathMgrIns, err := util.InitInstallDirPathMgr()
		convey.So(err, convey.ShouldBeNil)
		var initCertMgrIns *certutils.RootCertMgr
		var selfSignCertIns *certutils.SelfSignCert
		var componentMgrIns *util.ComponentMgr
		var ins = &certPrepareCtl{
			certPathMgr: InstallDirPathMgrIns.ConfigPathMgr,
			components:  []string{"cert-manager"},
		}

		convey.Convey("test prepareCert func success", func() {
			p := gomonkey.ApplyMethodReturn(initCertMgrIns, "NewRootCa", nil, nil).
				ApplyMethodReturn(selfSignCertIns, "CreateSignCert", nil).
				ApplyFuncReturn(util.PrepareKubeConfigCert, nil)
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			convey.So(ins.prepareCerts(), convey.ShouldBeNil)
		})

		convey.Convey("test prepareCert func prepareCA certPathMgr is nil", func() {
			var speIns = &certPrepareCtl{
				certPathMgr: nil,
				components:  []string{"edge-manager"},
			}
			convey.So(speIns.prepareCerts(), convey.ShouldResemble, errors.New("pointer cpc.certPathMgr is nil"))
		})

		convey.Convey("test prepareCert func init root ca failed", func() {
			p := gomonkey.ApplyMethodReturn(initCertMgrIns, "NewRootCa", nil, ErrTest)
			defer p.Reset()
			convey.So(ins.prepareCerts(), convey.ShouldResemble, errors.New("init root ca info failed"))
		})

		convey.Convey("test prepareCert func prepare component cert failed", func() {
			p := gomonkey.ApplyMethodReturn(initCertMgrIns, "NewRootCa", nil, nil).
				ApplyMethodReturn(componentMgrIns, "PrepareComponentCert", ErrTest)
			defer p.Reset()
			convey.So(ins.prepareCerts(), convey.ShouldResemble,
				errors.New("prepare single component cert failed"))
		})

		convey.Convey("test prepareCert func set certs owner failed", func() {
			p := gomonkey.ApplyMethodReturn(initCertMgrIns, "NewRootCa", nil, nil).
				ApplyMethodReturn(componentMgrIns, "PrepareComponentCert", nil).
				ApplyFuncReturn(util.PrepareKubeConfigCert, ErrTest)
			defer p.Reset()
			convey.So(ins.prepareCerts(), convey.ShouldResemble, errors.New("prepare kube config cert failed"))
		})
	})
}
