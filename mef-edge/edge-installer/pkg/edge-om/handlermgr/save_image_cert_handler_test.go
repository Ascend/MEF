// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common test for saving image repository cert
package handlermgr

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

const testCaContent = `
-----BEGIN CERTIFICATE-----
MIIE/zCCA2egAwIBAgIVAO+ycnPRWNYNW9SSwr1oRM9ZlKFdMA0GCSqGSIb3DQEB
CwUAMGoxCzAJBgNVBAYTAkNOMQ8wDQYDVQQKEwZIdWF3ZWkxEzARBgNVBAsTCkNQ
TCBBc2NlbmQxNTAzBgNVBAMMLGh1Yl9zdnItMGIxYWM4YmEtZjM0Ny00ZjY1LWIw
MGMtMWE1ZDkwNDFlNmU4MB4XDTIzMDkwNDEyMDU1M1oXDTMzMDkwNDEyMDU1M1ow
ajELMAkGA1UEBhMCQ04xDzANBgNVBAoTBkh1YXdlaTETMBEGA1UECxMKQ1BMIEFz
Y2VuZDE1MDMGA1UEAwwsaHViX3N2ci0wYjFhYzhiYS1mMzQ3LTRmNjUtYjAwYy0x
YTVkOTA0MWU2ZTgwggGiMA0GCSqGSIb3DQEBAQUAA4IBjwAwggGKAoIBgQDMpC92
LjPgazpSeQXJy+S+CCgXC9S0D9BnOAZmVd70wr6tp4vAAEhC5bS3Bww7bDe6HxI4
D6DokLcOJgseh58kZbeh/kbT6dIVduV+yjahb0of5iopuGH9IDhrSE87KahphUXZ
JDcdlhjlohLw61ZttkqdT7VoNgQ6QfsAXnm9LEFcrYx9PPDUwFqp2laQzoRf+tzF
h1f2fy+RUHUrST3tUs8A9Kz+o1PNhBZY7XPW3f0ExC1w7DXv80ZnQcXDwqiks6NE
8CizVh6X6dfbEu9e6rqJbqavqNDko4YMfgqkBl0jVHBC5fwWxTj3IeL3AliVbCeS
6I6atQGec4bc3meVqpx7KFkMB/mtPV/ZDUm+xGBVdyepIQujp6baQba3lBTEmBtt
prSzHxr/dclMhrRBgBR3U8AkArhflB2cSfZrojhuptfaMnbV/ivNN3hdaPT18CzB
FBFK93G4M9+Tmmu0zP77a+eig/5gfIhWqlik2MsG9zHDVwuKD8S0z6O5PN0CAwEA
AaOBmzCBmDAOBgNVHQ8BAf8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsG
AQUFBwMBMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEIMDPz9+2vhim5rQTRaKA
I/R0fOoFMVLB1RxXpf1a+7DDMCsGA1UdIwQkMCKAIMDPz9+2vhim5rQTRaKAI/R0
fOoFMVLB1RxXpf1a+7DDMA0GCSqGSIb3DQEBCwUAA4IBgQB++eT4P8KdQ0MV6kzQ
eJjbkmD3vu5L3g0AEdSzNBuHe8E/WlIVz2KRolIrVqTW7jHXc8GpVywH3P+WbOaf
bewc8aNsGgAJPoERRWJ83lCcW3CNheFTEYWyNixIsyJMF6sXzbMPaLqsBHlnONUj
HNfc8efmusEd5PAGTcXYDM2IUOsL1F+B1iWqDpODMzetjELWlEKy870ZTVEB8qAl
MRP1cLRHeQNMZzaPX2loSjwmUO/nwpX4LSeLAdr3Qt5vUzi3HyBabLujJZXG8q6p
q2UhiKT7PCR7xO14ZMSVAxAcSXTUO9Z2iLAnWSsrzRqZwfQ1euGTxiYCfb8JQsUq
wyeKSbCGEDg/ulzDe3izHLTxu7gWKTjIsA7mWt7cBUcg/2VRTQ+8D7CZkOuUlnVa
GqsLV0X/FUy6mkUmhVn8yTc5LebY4zYmZR1KEzeZZkoa2Z/O0lSzqMoff9faflgq
QiVq+PC4ZIsiCEB6ETiAo4/h2anUCc7mPmyCTVpi4KSVAa0=
-----END CERTIFICATE-----`

// the basic constraints of the second layer of the cert Chain is CA:FALSE
const testWrongCaContent = `
-----BEGIN CERTIFICATE-----
MIIFBTCCA22gAwIBAgIUYUksqFe7MMNCCilapohGaf5sZRQwDQYJKoZIhvcNAQEL
BQAwfDELMAkGA1UEBhMCQ04xEDAOBgNVBAgMB1NpQ2h1YW4xEDAOBgNVBAcMB0No
ZW5nRHUxFTATBgNVBAoMDFRlc3QgQ29tcGFueTEYMBYGA1UECwwPVGVzdCBEZXBh
cnRtZW50MRgwFgYDVQQDDA93d3cuZXhhbXBsZS5jb20wHhcNMjMxMDIwMDIzNjU4
WhcNMzMxMDE3MDIzNjU4WjB8MQswCQYDVQQGEwJDTjEQMA4GA1UECAwHU2lDaHVh
bjEQMA4GA1UEBwwHQ2hlbmdEdTEVMBMGA1UECgwMVGVzdCBDb21wYW55MRgwFgYD
VQQLDA9UZXN0IERlcGFydG1lbnQxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTCC
AaIwDQYJKoZIhvcNAQEBBQADggGPADCCAYoCggGBAL68lDZRs0I7kbJsAucj/EjK
PLe+t7bzcRGvQNE2BXN0d+pTuOwX5bsAuucwMIZRkYB1TXGbFwtxLY2Dj6rhHrHc
4aq20Hzu3cEAY2pdryMGLQGu7ODqrjj2tCj7pVnEzNfadXDZ4XQtw/y6FGaTuGwm
4BZhSK9GrZb/wSMsFNnJnA78aFLUQF7bKyWWGdlISxcy1sq/IaLKqIEfFuKmzspw
ybmUdiDag/Kz+4va2PSYxec9Kf65iUwmVeNtyPSRTuO8sjCQRwpusYlg50YWK3hZ
hNL1Ewjg1aOy3V0c5dYvphQazgXL7rIduSIX3uyQbi/FNENTeAsr2crpyVtuSu0k
A1ZI5a8r9knkmfRe3IyEdPt9l5C90IgXD7pwORrZ9fmp8Q9nLKlTMYy2EZaMXlP0
GHjrwIfkDOQEaeeEOu170Cy0gFHcwvqABdfAJTw+k5eq28GuSpTmMOt3uAlzeabe
VBYhpHReO+GKJnevf7CdtaQNYDeYL402WDs1po2kIwIDAQABo38wfTAdBgNVHQ4E
FgQUsEd5oXg4mcApvRB8R2BbhIprhfswHwYDVR0jBBgwFoAUsEd5oXg4mcApvRB8
R2BbhIprhfswDwYDVR0TAQH/BAUwAwEB/zALBgNVHQ8EBAMCAQYwHQYDVR0lBBYw
FAYIKwYBBQUHAwEGCCsGAQUFBwMCMA0GCSqGSIb3DQEBCwUAA4IBgQAVCiEmqWKQ
SLgID8NIg96NnOQ53TG0iYlnLEuCVrzXDd9D0rKfZ0BuI8s9teHgIfRQpIp7OUbw
AICW9PIg3KfPzpmG72LeNLtqEl1P9VFgb2MAxjes+HuGU0iIfOfmft8P3TZAXk8a
Jqn3ajkdMLXzIOWfSJYG+EhRBo4r0hVT1cDlcH3KGEIN/DXK8wnFi3XEvJGFJgfT
46h34cxYl/WQF5clz5jHUdpBi5U6x6zqX/WOqGLBDLQtpzrWr4B5M6Mw/JsYESPZ
xZjm9CUCeSLM/cO8uu8T1uVSXHey95F4Mff5zmffLDbUXF7/esSzZID2KrMZ1b1q
vvuz9ZdBuCEIRLDXbX3hp8frvO1ypgRKMQAr6/8s81PTX8etOSncff0UtwO6HW1x
N0iiYB5Z9H5yz3Y4A/Xxp0/ArYO/jxXqSOuQP0g0ScDVfGQoqqCyUnt7pjKgpNL2
4ca9QSD5+nEXPCFGdpbYscRqJRbUoRw4rMGUm7qf5yYgy8R13DO8RIQ=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIFHjCCA4agAwIBAgIUPLHV2e5dwwZUqmg/xDbYvjMv65MwDQYJKoZIhvcNAQEL
BQAwfDELMAkGA1UEBhMCQ04xEDAOBgNVBAgMB1NpQ2h1YW4xEDAOBgNVBAcMB0No
ZW5nRHUxFTATBgNVBAoMDFRlc3QgQ29tcGFueTEYMBYGA1UECwwPVGVzdCBEZXBh
cnRtZW50MRgwFgYDVQQDDA93d3cuZXhhbXBsZS5jb20wHhcNMjMxMDIwMDIzNjU4
WhcNNDMxMDE1MDIzNjU4WjCBlzELMAkGA1UEBhMCQ04xEDAOBgNVBAgMB1NpQ2h1
YW4xEDAOBgNVBAcMB0NoZW5nRHUxIzAhBgNVBAoMGlRlc3QgSW50ZXJtZWRpYXRl
IENvbXBvYW55MSUwIwYDVQQLDBxUZXN0IEludGVybWVkaWF0ZSBEZXBhcnRtZW50
MRgwFgYDVQQDDA93d3cuZXhhbXBsZS5jb20wggGiMA0GCSqGSIb3DQEBAQUAA4IB
jwAwggGKAoIBgQDgdd2SRIWcerquX+ptz1TR9m0vxHM58STGrH4ToskJRiQKnItr
27yqplIB3f9txjzGcNBtM8PpdzI6RtZ0vyo4o4snaQ4JqGRzoIeVNx728kfIgdd2
iYfhhJ+mQE58TQspdy/D7kJozD74P7vBSw9dQfSyX4OZg+ne/kPhoJoNICPuuV51
Wp/+zOl25p6ZgLuUBww0lhSax7JZ7ZjqfHUyvYG+RbPYW5DuivZxxmTw2Rxz0GsK
nnB2FHSU5PBm3orrUGo3LLPuXBlLc1TWic2JCPGs6Rxwv0wuZBks8DqhrIGAP9WF
Ay42z+MFE6UopWjrcjmiBfe+A/OPteYUC/e/w+dtCf/J67TRrMd1HHaJToiTAkiP
Y0vw9qfa2cn/sCsPoy1Dx2pyKt/IPl/zxVMYUhpRgcbDLyk5MsmnoLX2GcKahJND
tTfSyRMlufV7EYmQLcxb0Dyceho8OrGNrvoqbNMYg9A9dtIOqYZipBeQvXNJshvf
N4QPv5+FwIWBmj0CAwEAAaN8MHowHQYDVR0OBBYEFJOHVryrN61WKsqPzPJsezse
7rygMB8GA1UdIwQYMBaAFLBHeaF4OJnAKb0QfEdgW4SKa4X7MAwGA1UdEwEB/wQC
MAAwCwYDVR0PBAQDAgEGMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAN
BgkqhkiG9w0BAQsFAAOCAYEAN2GZNHJOjUs2FrzT4EITO+W4dsj6bc0EEgb0URV9
KvVTtNVlHj7CuBDEhpjulJLfWKmgbqIkin2H18pmCEuNKswhtZgMAA+svsZirIfo
T0RFBJxwOvIgOHHu7Q+750mJDaFGXRUWqP+t9Hkp7QAZdXg7mj3PjME47UksYSbX
cTCcfdqjplOVN18LUP7+Z85xTbN6xrDq6JmNFCFj3oyhnqR44ZM5UrElj39rgkVh
APpmuusCaHVPTP1aQj+8u+4VWtcMTVhLkgIqDqPmgIhdWeoqDwrJl9v8zPE/PiON
N1zwh71fZYuqSIQO97MKyjjMQ2Z4hg6eAxbNoTTdukFMoOE1V/wXTzUSb0Q+giJq
VnwjjTpAaOnDu3jSSL3MRL48WrHF4vMbXzcZT508dWeidKMZovJWqidW/s/3MwsF
l9zPc+Z1LHsN/pzQAh82g8/JxXANdOeUej4ZZie5vs6wDostvaD/S60KpYMeEzaX
gkM1nzsAyYOb27aLMMpHew5S
-----END CERTIFICATE-----`

const testUnmatchedCrlContent = `-----BEGIN X509 CRL-----
MIICezCB5AIBATANBgkqhkiG9w0BAQsFADBrMQswCQYDVQQGEwJDTjEPMA0GA1UE
ChMGSHVhd2VpMRMwEQYDVQQLEwpDUEwgQXNjZW5kMTYwNAYDVQQDEy1NaW5kWE1F
Ri1mYWEzMGM3My04NzcyLTRlNDQtYTgyMS0wYjBkODQ3ODc3YWIXDTI0MDYyNTAy
MTYxN1oXDTQ0MDYyNjAyMTYxN1owFDASAgEBFw0yNTA2MjYwMjE2MTdaoC8wLTAf
BgNVHSMEGDAWgBR3sZKpxVgcRF5d57h7z9+jNqvYhjAKBgNVHRQEAwIBATANBgkq
hkiG9w0BAQsFAAOCAYEAmD7Fg3nnS0RgJ3Rs+aT6mQmeaAwCb3ebPBZx8vKJjFXP
ipOADk/7wIYWpfTWD5uH1Atp9qLSSa5EplAUNIKJUX2LcCwOc4sii4Efwn3asRTv
V19hz28P3JF8mgmIAvY96QsDBxB+BJnRe51gU5RMvbdp+ICLa+plU7YKQO+OWJcB
dWDSeEpJLk8uxHyjBcKPJ+NRba6QWUubXM0jfOKKWHoBByEOgscwf4dij7eyPhcM
CReO0BsNWi7UwQPl4i46EbtQxF4QYXQ8QFoOXS0X3TPjGdelJw+ba9PsblBngnYR
69b4XDvMmRYQYMSsCOBvq1U8dDSD37mb/JzEDt3dduvDyIColpZG7udZadCLIf60
1I4JZQTZY8UEtLdF5SyxZ8ZJnRb+X3P7JOUqX5ZXIWyMlNO/++9uZLnGs5S1RWHE
D508DinhMy0jsgQNhBUvtKw1khdpwSjYSFSSr4OGX3k8fFZhVMlSgPruyGlEQ4d4
RLOeYJ/bFL+bAOX25H6/
-----END X509 CRL-----`

var (
	saveImageCertMsg *model.Message
	saveImageCert    = saveCertHandler{}
)

func TestImageSaveCertHandler(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(x509.NewCaChainMgr, nil, nil).
		ApplyMethodReturn(&x509.CaChainMgr{}, "CheckCertChain", nil).
		ApplyMethodReturn(net.DefaultResolver, "LookupIP", []net.IP{}, nil)
	defer p.Reset()

	var err error
	saveImageCertMsg, err = newSaveImageCertMsg()
	if err != nil {
		fmt.Printf("new save image cert msg failed, error: %v\n", err)
		return
	}
	convey.Convey("save image repo cert should be success", t, testSaveCertHandler)
	convey.Convey("save image repo cert should be failed, param convert failed", t, testSaveCertHandlerErrParamConv)
	convey.Convey("save image repo cert should be failed, check failed", t, testSaveCertHandlerErrCheckField)
	convey.Convey("test func saveCert should be failed, get install root dir failed", t, testSaveCertErrGetInstallRootDir)
	convey.Convey("test func saveCert should be failed, write data failed", t, testSaveCertErrWriteData)
	convey.Convey("test func saveCert should be failed, set path permission failed", t, testSaveCertErrSetPerm)
	convey.Convey("test func saveCert should be failed, create dir failed", t, testSaveCertErrCreateDir)
	convey.Convey("test func creatDockerCertLink should be failed, delete failed", t, testCopyCertToDockerErrDelete)
	convey.Convey("test func creatDockerCertLink should be failed, cp file failed", t, testCopyCertToDockerErrCpFile)
}

func testSaveCertHandler() {
	var p1 = gomonkey.ApplyFuncReturn(utils.IsLocalIp, false)
	defer p1.Reset()
	err := saveImageCert.Handle(saveImageCertMsg)
	convey.So(err, convey.ShouldBeNil)
}

func testSaveCertHandlerErrParamConv() {
	errMsg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}
	err = errMsg.FillContent(model.RawMessage{})
	convey.So(err, convey.ShouldBeNil)
	err = saveImageCert.Handle(errMsg)
	convey.So(err, convey.ShouldResemble, errors.New("parse image repository cert info para failed"))
}

func testSaveCertHandlerErrCheckField() {
	var p1 = gomonkey.ApplyFuncSeq(x509.NewCaChainMgr, []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, testErr}},
		{Values: gomonkey.Params{nil, nil}},
	}).
		ApplyMethodReturn(&x509.CaChainMgr{}, "CheckCertChain", testErr)
	defer p1.Reset()

	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}

	certInfos := getTestCertInfos()
	for _, certInfo := range certInfos {
		err = msg.FillContent(certInfo, true)
		convey.So(err, convey.ShouldBeNil)
		err = saveImageCert.Handle(msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "check image repository cert info failed")
	}
}

func getTestCertInfos() []CertInfo {
	return []CertInfo{
		{
			"",
			"443",
			"fd.fusiondirector",
			testCaContent,
		},
		{
			"127.0.0.1",
			"errPort",
			"fd.fusiondirector",
			testCaContent,
		},
		{
			"127.0.0.1",
			"443",
			".fd.fusion.director",
			testCaContent,
		},
		{
			"127.0.0.1",
			"443",
			"-fd.fusion.director",
			testCaContent,
		},
		{
			"127.0.0.1",
			"443",
			"fd.fusiondirector",
			`errCaContent`,
		},
		{
			"127.0.0.1",
			"443",
			"fd.fusiondirector",
			`errCaContent2`,
		},
	}
}

func testSaveCertErrGetInstallRootDir() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, testErr).
		ApplyFuncReturn(utils.IsLocalIp, false)
	defer p1.Reset()

	err := saveCert(nil)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get config path manager, error: %v", testErr))
	err = saveImageCert.Handle(saveImageCertMsg)
	convey.So(err, convey.ShouldResemble, errors.New("save image repository cert info failed"))
}

func testSaveCertErrWriteData() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.WriteData, testErr)
	defer p1.Reset()

	certInfo := &CertInfo{
		Port:      "443",
		Domain:    "fd.fusiondirector",
		CaContent: testCaContent,
	}
	err := saveCert(certInfo)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("write cert to file failed, error: %v", testErr))
}

func testSaveCertErrSetPerm() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.SetPathPermission, testErr)
	defer p1.Reset()

	certInfo := &CertInfo{
		Port:      "443",
		Domain:    "fd.fusiondirector",
		CaContent: testCaContent,
	}
	err := saveCert(certInfo)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("set cert mode failed, error: %v", testErr))
}

func testSaveCertErrCreateDir() {
	var p1 = gomonkey.ApplyFunc(fileutils.CreateDir,
		func(tgtPath string, mode os.FileMode, checkerParam ...fileutils.FileChecker) error {
			return testErr
		})
	defer p1.Reset()

	certInfo := &CertInfo{
		Port:      "443",
		Domain:    "fd.fusiondirector",
		CaContent: testCaContent,
	}
	certDir := filepath.Join(constants.DockerCertDir, certInfo.Domain)
	err := saveCert(certInfo)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("create docker cert %s failed, error: %v", certDir, testErr))
}

func testCopyCertToDockerErrDelete() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.DeleteFile, testErr)
	defer p1.Reset()
	err := copyCertToDocker("", "")
	convey.So(err, convey.ShouldResemble, testErr)
}

func testCopyCertToDockerErrCpFile() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.DeleteFile, nil).
		ApplyFuncReturn(fileutils.CopyFile, testErr)
	defer p1.Reset()
	err := copyCertToDocker("", "")
	convey.So(err, convey.ShouldNotBeNil)
}

func newSaveImageCertMsg() (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return nil, errors.New("new message failed")
	}
	msg.SetRouter(constants.InnerClient, constants.ConfigMgr, constants.OptUpdate, constants.ResImageCertInfo)

	certInfo := CertInfo{
		Ip:        "127.0.0.1",
		Port:      "443",
		Domain:    "fd.fusiondirector",
		CaContent: testCaContent,
	}
	if err = msg.FillContent(certInfo, true); err != nil {
		fmt.Printf("fill content failed: %v", err)
		return nil, errors.New("fill content failed")
	}

	return msg, nil
}
