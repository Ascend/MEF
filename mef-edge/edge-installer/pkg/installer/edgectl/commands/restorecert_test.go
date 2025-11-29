// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package commands
package commands

import (
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/edgectl/common"
)

const certToRestore = `-----BEGIN CERTIFICATE-----
MIIFATCCA2mgAwIBAgIVAIqzV5W+uLF6wE34dkjliDH70vBNMA0GCSqGSIb3DQEB
CwUAMGsxCzAJBgNVBAYTAkNOMQ8wDQYDVQQKEwZIdWF3ZWkxEzARBgNVBAsTCkNQ
TCBBc2NlbmQxNjA0BgNVBAMTLU1pbmRYTUVGLTc0OWUzMzgwLTQ2MWMtNDRlMS05
OGViLTRmMDg0MGIyMjY2MzAeFw0yNDA1MTUxMzExNDBaFw0zNDA1MTUxMzExNDBa
MGsxCzAJBgNVBAYTAkNOMQ8wDQYDVQQKEwZIdWF3ZWkxEzARBgNVBAsTCkNQTCBB
c2NlbmQxNjA0BgNVBAMTLU1pbmRYTUVGLTc0OWUzMzgwLTQ2MWMtNDRlMS05OGVi
LTRmMDg0MGIyMjY2MzCCAaIwDQYJKoZIhvcNAQEBBQADggGPADCCAYoCggGBAM51
FpqacP3wqdbJuEHwgON41uQCt8DOUEg0D9CQlXBucZAGVgc/DC71xC9+yInVzR8Q
WTDsevb0H0/oUZ1UtQ3GOpqpMzefvsk2LzZ+wQofwwNUEQ8kVJRKiLzG3D85u/sA
EPSaebtPWc4dDj7SG01kZO4Jln0PHimFwiInTSxJ4mG9Mz5zdXdH9uJIySJ5wHMO
Lfd8w1HnkAD+plZ7BlqXRse6LbBJpkVlAuLf74VlVatq99wefyro822eo4ncDYyA
anUMc74uoZx1rs1ia14RcZdcQH0lmFVjLSXUwl3rpjfc7ixVn/gURcxKDwGsmKWF
OvB//G0PbgC43cM2JW03Gr4mUqVnmmIZCMq5ZX/xD7qzV3rYNIJ0ZWY+k7j8cThC
P67QKrSkwV1DqL7+TB9MAI3Jx0+huPvHwu3iQeKsRRjfhiJ99lXemj1RIUkY07AO
WYw1MA3INEpbZ7KwcUrWjcfbmdQLGwFSu4cF1/DOmhRnzMFo4lbVIWozfUavPwID
AQABo4GbMIGYMA4GA1UdDwEB/wQEAwIBhjAdBgNVHSUEFjAUBggrBgEFBQcDAgYI
KwYBBQUHAwEwDwYDVR0TAQH/BAUwAwEB/zApBgNVHQ4EIgQg8O+12OUvt/LMIK35
YevPWuszQG6A3eUmYdBzmh9Sn4cwKwYDVR0jBCQwIoAg8O+12OUvt/LMIK35YevP
WuszQG6A3eUmYdBzmh9Sn4cwDQYJKoZIhvcNAQELBQADggGBAHOcNqATY17O4Yvr
bGg+4Dx9CUHCflUydgcbjZChbqRfTguNQ61UxwOJA9f2/COnggEJy+6O+wmwgCtz
e6Th+dGLOmn3NUmah122caZmUqCW6XJGs33oSfnCFz5yYVvbHJYQfrj07Fdl+JUk
4HTwUA2IoFTyl4RhbIu6RckdoqZeWxokobujV98Wn+NiKztuU7XjVVXxeEtBu1UR
IMmSws0aI5jV1B1tdj0p2iLl45mSbn8iotUVNmA0FfI0jP8z5dWDVmO5TA9s/IEt
CCF0HxrvPV0t9Fz0w/QxSnqKLGGDId7dLMCcH+V5H+n7WrNkfEUC9mw04POYriRP
ZePRHMJ8uBzgI19rcM0FTjRUd5T0lDmj7msKxf4kY5VK+oGEWf0z6ZK+Czh3Hl7O
lM4Xob9h8sys1ti7+QdF5+FIC1Hv3CDHC1HYATI06ngAnPayt3whJB/9JDMqOPKF
W5bBocccBwJI9GPq0jMxYX+w/O7qT74m532L6ccvgPEXP7KBtw==
-----END CERTIFICATE-----`

const (
	TempCertDir              = "/tmp/MEFEdge/config/edge_main/hub_certs_import"
	TempCertPath             = TempCertDir + "/cloud_root.crt"
	TempPreBackupPath        = TempCertDir + "/cloud_root.crt.pre"
	InvalidTempPreBackupPath = TempCertDir + "/invalid_cloud_root.crt.pre"
)

func TestNewRestoreCertInfoCmd(t *testing.T) {
	convey.Convey("prepare test dir and files", t, func() {
		if err := os.MkdirAll(TempCertDir, os.ModeDir); err != nil {
			t.Errorf("create test dir failed: %v", err)
		}
		if err := os.WriteFile(TempPreBackupPath, []byte(certToRestore), os.ModeType); err != nil {
			t.Errorf("create test cert file faild: %v", err)
		}
	})
	convey.Convey("test restore unused cert cmd methods", t, restoreCertInfoCmdMethods)
	convey.Convey("test restore unused cert cmd successful", t, restoreCertInfoCmdSuccess)
	convey.Convey("test restore unused cert cmd failed", t, restoreCertInfoCmdFailed)
}

func restoreCertInfoCmdMethods() {
	convey.So(NewRestoreCertInfoCmd().Name(), convey.ShouldEqual, common.RestoreCert)
	convey.So(NewRestoreCertInfoCmd().Description(), convey.ShouldEqual, common.RestoreCertDesc)
	convey.So(NewRestoreCertInfoCmd().LockFlag(), convey.ShouldBeTrue)
}

func restoreCertInfoCmdSuccess() {
	p := gomonkey.ApplyFuncReturn(NewRestoreCertInfoCmd, &restoreCertInfoCmd{certName: "cloud_root"}).
		ApplyGlobalVar(&ctx, &common.Context{
			WorkPathMgr:   pathmgr.NewWorkPathMgr("/tmp"),
			ConfigPathMgr: pathmgr.NewConfigPathMgr("/tmp"),
		}).
		ApplyFuncReturn(ctx.ConfigPathMgr.GetHubSvrRootCertPrevBackupPath, TempPreBackupPath).
		ApplyFuncReturn(ctx.ConfigPathMgr.GetHubSvrRootCertPath, TempCertPath).
		ApplyFuncReturn(envutils.GetUid, uint32(0), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(0), nil)
	defer p.Reset()
	convey.So(NewRestoreCertInfoCmd().Execute(ctx), convey.ShouldBeNil)
	NewRestoreCertInfoCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func restoreCertInfoCmdFailed() {
	convey.Convey("ctx is nil failed", func() {
		convey.So(NewRestoreCertInfoCmd().Execute(nil), convey.ShouldResemble, fmt.Errorf("parameter ctx is invalid"))
		NewRestoreCertInfoCmd().PrintOpLogFail(userRoot, ipLocalhost)
	})

	convey.Convey("invalid cert name parameter", func() {
		p := gomonkey.ApplyFuncReturn(NewRestoreCertInfoCmd, &deleteCertInfoCmd{certName: "not_exists"}).
			ApplyGlobalVar(&ctx, &common.Context{
				WorkPathMgr:   pathmgr.NewWorkPathMgr("/tmp"),
				ConfigPathMgr: pathmgr.NewConfigPathMgr("/tmp"),
			}).
			ApplyFuncReturn(ctx.ConfigPathMgr.GetHubSvrRootCertPrevBackupPath, TempPreBackupPath).
			ApplyFuncReturn(ctx.ConfigPathMgr.GetHubSvrRootCertPath, TempCertPath)
		defer p.Reset()
		convey.So(NewRestoreCertInfoCmd().Execute(ctx), convey.ShouldResemble,
			fmt.Errorf("invalid certificate name, please check"))
		NewRestoreCertInfoCmd().PrintOpLogFail(userRoot, ipLocalhost)
	})

	convey.Convey("previous backup cert not exists", func() {
		p := gomonkey.ApplyFuncReturn(NewRestoreCertInfoCmd, &restoreCertInfoCmd{certName: "cloud_root"}).
			ApplyGlobalVar(&ctx, &common.Context{
				WorkPathMgr:   pathmgr.NewWorkPathMgr("/tmp"),
				ConfigPathMgr: pathmgr.NewConfigPathMgr("/tmp"),
			}).
			ApplyFuncReturn(ctx.ConfigPathMgr.GetHubSvrRootCertPrevBackupPath, TempPreBackupPath).
			ApplyFuncReturn(ctx.ConfigPathMgr.GetHubSvrRootCertPath, TempCertPath).
			ApplyFuncReturn(envutils.GetUid, uint32(0), nil).
			ApplyFuncReturn(envutils.GetGid, uint32(0), nil)
		defer p.Reset()
		convey.So(NewRestoreCertInfoCmd().Execute(ctx), convey.ShouldResemble,
			fmt.Errorf("previous backup certificate not found"))
		NewRestoreCertInfoCmd().PrintOpLogFail(userRoot, ipLocalhost)
	})
}
