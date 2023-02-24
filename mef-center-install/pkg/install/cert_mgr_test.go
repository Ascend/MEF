// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

func CertMgrTest() {
	Convey("CertMgr DoInstallPrepare func", CertMgrDoPrepareTest)
	Convey("prepareCertsDir func", PrepareCertsDirTest)
	Convey("certMgrPrepareCert func", CertMgrPrepareCertTest)
	Convey("serCertsOwner func", SetCertsOwnerTest)
	Convey("setCertsOwner func set right test", SetCertsOwnerSetRightTest)
}

func CertMgrDoPrepareTest() {
	var ins = &certPrepareCtl{}
	Convey("test DoInstallPrepare func in certPrepareCtl struct success", func() {
		p := ApplyPrivateMethod(ins, "prepareCertsDir", func(_ *certPrepareCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareCerts", func(_ *certPrepareCtl) error { return nil }).
			ApplyPrivateMethod(ins, "deleteRootKey", func(_ *certPrepareCtl) error { return nil })
		defer p.Reset()
		So(ins.doPrepare(), ShouldBeNil)
	})

	Convey("test DoInstallPrepare func in certPrepareCtl struct failed", func() {
		p := ApplyPrivateMethod(ins, "prepareCertsDir",
			func(_ *certPrepareCtl) error { return ErrTest })
		defer p.Reset()
		So(ins.doPrepare(), ShouldResemble, ErrTest)
	})
}

func PrepareCertsDirTest() {
	Convey("test prepareCertDir", func() {
		currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		So(err, ShouldBeNil)
		InstallDirPathMgrIns := util.InitInstallDirPathMgr(currentPath)
		var ins = &certPrepareCtl{
			certPathMgr: InstallDirPathMgrIns.ConfigPathMgr,
			components:  []string{"edge-manager"},
		}

		Convey("test prepareCertDir func success", func() {
			defer ResetAndClearDir(nil, InstallDirPathMgrIns.GetMefPath())
			So(ins.prepareCertsDir(), ShouldBeNil)
		})

		Convey("test prepareCertDir func makesure cert path failed", func() {
			p := ApplyFuncReturn(common.MakeSurePath, ErrTest)
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			So(ins.prepareCertsDir(), ShouldResemble, errors.New("create cert path failed"))
		})

		Convey("test prepareCertDir func makesure root ca path failed", func() {
			p := ApplyFuncSeq(common.MakeSurePath,
				[]OutputCell{{Values: Params{nil}}, {Values: Params{ErrTest}}})
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			So(ins.prepareCertsDir(), ShouldResemble, errors.New("create root certs path failed"))
		})

		Convey("test prepareCertDir func makesure root ca key path failed", func() {
			p := ApplyFuncSeq(common.MakeSurePath,
				[]OutputCell{{Values: Params{nil}}, {Values: Params{nil}}, {Values: Params{ErrTest}}})
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			So(ins.prepareCertsDir(), ShouldResemble, errors.New("create root key path failed"))
		})

		Convey("test prepareCertDir func makesure component's cert path failed", func() {
			p := ApplyFuncSeq(common.MakeSurePath,
				[]OutputCell{
					{Values: Params{nil}},
					{Values: Params{nil}},
					{Values: Params{nil}},
					{Values: Params{ErrTest}},
				})
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			So(ins.prepareCertsDir(), ShouldResemble,
				errors.New("prepare component [edge-manager]'s cert dir failed"))
		})
	})
}

func CertMgrPrepareCertTest() {
	Convey("test prepareCertDir", func() {
		currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		So(err, ShouldBeNil)
		InstallDirPathMgrIns := util.InitInstallDirPathMgr(currentPath)
		var initCertMgrIns *certutils.RootCertMgr
		var selfSignCertIns *certutils.SelfSignCert
		var componentMgrIns *util.ComponentMgr
		var ins = &certPrepareCtl{
			certPathMgr: InstallDirPathMgrIns.ConfigPathMgr,
			components:  []string{"edge-manager"},
		}

		Convey("test prepareCert func success", func() {
			p := ApplyMethodReturn(initCertMgrIns, "NewRootCa", nil, nil).
				ApplyMethodReturn(selfSignCertIns, "CreateSignCert", nil).
				ApplyPrivateMethod(ins, "setCertsOwner", func(_ *certPrepareCtl) error { return nil })
			defer ResetAndClearDir(p, InstallDirPathMgrIns.GetMefPath())
			So(ins.prepareCerts(), ShouldBeNil)
		})

		Convey("test prepareCert func prepareCA certPathMgr is nil", func() {
			var speIns = &certPrepareCtl{
				certPathMgr: nil,
				components:  []string{"edge-manager"},
			}
			So(speIns.prepareCerts(), ShouldResemble, errors.New("pointer cpc.certPathMgr is nil"))
		})

		Convey("test prepareCert func init root ca failed", func() {
			p := ApplyMethodReturn(initCertMgrIns, "NewRootCa", nil, ErrTest)
			defer p.Reset()
			So(ins.prepareCerts(), ShouldResemble, errors.New("init root ca info failed"))
		})

		Convey("test prepareCert func prepare component cert failed", func() {
			p := ApplyMethodReturn(initCertMgrIns, "NewRootCa", nil, nil).
				ApplyMethodReturn(componentMgrIns, "PrepareComponentCert", ErrTest)
			defer p.Reset()
			So(ins.prepareCerts(), ShouldResemble, errors.New("prepare single component cert failed"))
		})

		Convey("test prepareCert func set certs owner failed", func() {
			p := ApplyMethodReturn(initCertMgrIns, "NewRootCa", nil, nil).
				ApplyMethodReturn(componentMgrIns, "PrepareComponentCert", nil).
				ApplyPrivateMethod(ins, "setCertsOwner",
					func(_ *certPrepareCtl) error { return ErrTest })
			defer p.Reset()
			So(ins.prepareCerts(), ShouldResemble, ErrTest)
		})
	})
}

func SetCertsOwnerTest() {
	var ins = &certPrepareCtl{
		certPathMgr: &util.ConfigPathMgr{},
		components:  []string{"edge-manager"},
	}

	Convey("test setCertsOwner func success", func() {
		p := ApplyFuncReturn(util.GetMefId, 0, 0, nil).ApplyFuncReturn(util.SetPathOwnerGroup, nil)
		defer p.Reset()
		So(ins.setCertsOwner(), ShouldBeNil)
	})

	Convey("test setCertsOwner func getMefId failed", func() {
		p := ApplyFuncReturn(util.GetMefId, 0, 0, ErrTest)
		defer p.Reset()
		So(ins.setCertsOwner(), ShouldResemble, errors.New("get mef uid or gid failed"))
	})

	Convey("test setCertsOwner func set config path right failed", func() {
		p := ApplyFuncReturn(util.GetMefId, 0, 0, nil).ApplyFuncReturn(util.SetPathOwnerGroup, ErrTest)
		defer p.Reset()
		So(ins.setCertsOwner(), ShouldResemble, errors.New("set cert root path owner and group failed"))
	})

}

func SetCertsOwnerSetRightTest() {
	var ins = &certPrepareCtl{certPathMgr: &util.ConfigPathMgr{}, components: []string{"edge-manager"}}

	Convey("test setCertsOwner func set config dir right failed", func() {
		p := ApplyFuncReturn(util.GetMefId, 0, 0, ErrTest).
			ApplyFuncSeq(util.SetPathOwnerGroup, []OutputCell{{Values: Params{nil}}, {Values: Params{ErrTest}}})
		defer p.Reset()
		So(ins.setCertsOwner(), ShouldResemble, errors.New("get mef uid or gid failed"))
	})

	Convey("test setCertsOwner func set kmc dir right failed", func() {
		p := ApplyFuncReturn(util.GetMefId, 0, 0, ErrTest).
			ApplyFuncSeq(util.SetPathOwnerGroup,
				[]OutputCell{
					{Values: Params{nil}},
					{Values: Params{nil}},
					{Values: Params{ErrTest}},
				})
		defer p.Reset()
		So(ins.setCertsOwner(), ShouldResemble, errors.New("get mef uid or gid failed"))
	})

	Convey("test setCertsOwner func set root-ca key failed", func() {
		p := ApplyFuncReturn(util.GetMefId, 0, 0, ErrTest).
			ApplyFuncSeq(util.SetPathOwnerGroup,
				[]OutputCell{
					{Values: Params{nil}},
					{Values: Params{nil}},
					{Values: Params{nil}},
					{Values: Params{ErrTest}},
				})
		defer p.Reset()
		So(ins.setCertsOwner(), ShouldResemble, errors.New("get mef uid or gid failed"))
	})

	Convey("test setCertsOwner func set root-ca cert right failed", func() {
		p := ApplyFuncReturn(util.GetMefId, 0, 0, ErrTest).
			ApplyFuncSeq(util.SetPathOwnerGroup,
				[]OutputCell{
					{Values: Params{nil}},
					{Values: Params{nil}},
					{Values: Params{nil}},
					{Values: Params{nil}},
					{Values: Params{ErrTest}},
				})
		defer p.Reset()
		So(ins.setCertsOwner(), ShouldResemble, errors.New("get mef uid or gid failed"))
	})
}
