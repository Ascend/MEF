//  Copyright(C) 2024. Huawei Technologies Co.,Ltd.  All rights reserved.

package websocketmgr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/x509/certutils"
)

// TestCertStatus test cases for InitProxyConfig
func TestProxyConfig(t *testing.T) {
	convey.Convey("init normal proxy config", t, testInitProxyConfig)
	convey.Convey("init invalid proxy config", t, testInitProxyConfigErrPath)
	convey.Convey("register message route", t, testRegModInfos)
	convey.Convey("update tls ca cert", t, testUpdateTlsCa)
	convey.Convey("update empty tls ca cert", t, testUpdateTlsCaErrEmpty)
	convey.Convey("update invalid tls ca cert", t, testUpdateTlsCaErrInvalid)
	convey.Convey("set rps limiter config", t, testSetRpsLimiterCfg)
	convey.Convey("set invalid rps limiter config", t, testSetRpsLimiterCfgErrInvalid)
	convey.Convey("set bandwidth limiter config", t, testSetBandwidthLimiterCfg)
	convey.Convey("set bandwidth limiter config with reserve rate", t, testSetBandwidthLimiterCfgWithReserveRate)
	convey.Convey("set invalid bandwidth limiter config", t, testSetBandwidthLimiterCfgErrInvalid)
	convey.Convey("set http timeout with default value", t, testSetTimeoutDefault)
	convey.Convey("set http timeout", t, testSetTimeout)
	convey.Convey("set http body size limit with default value", t, testSetSizeLimitDefault)
	convey.Convey("set http body size limit", t, testSetSizeLimit)
}

func testInitProxyConfig() {
	tlsInfo := prepareTlsInfo(true)
	_, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
}

func testInitProxyConfigErrPath() {
	tlsInfo := prepareTlsInfo(true)
	tlsInfo.CertPath = "not exists path"
	_, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldNotBeNil)
}

func testRegModInfos() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	testRegInfo := []modulemgr.MessageHandlerIntf{
		&modulemgr.RegisterModuleInfo{MsgOpt: "GET", MsgRes: "DT_TEST", ModuleName: "DT_TEST"},
	}
	proxyCfg.RegModInfos(testRegInfo)
}

func testUpdateTlsCa() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	clientTlsCertInfo := prepareTlsInfo(false)
	data, err := os.ReadFile(clientTlsCertInfo.RootCaPath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(proxyCfg.UpdateTlsCa(data), convey.ShouldBeNil)

}

func testUpdateTlsCaErrEmpty() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	convey.So(proxyCfg.UpdateTlsCa([]byte{}), convey.ShouldNotBeNil)
}

func testUpdateTlsCaErrInvalid() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	invalidCaCert := []byte("invalid ca cert content")
	convey.So(proxyCfg.UpdateTlsCa(invalidCaCert), convey.ShouldNotBeNil)
}

func testSetRpsLimiterCfg() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	convey.So(proxyCfg.SetRpsLimiterCfg(1, 1), convey.ShouldBeNil)
}

func testSetRpsLimiterCfgErrInvalid() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	convey.So(proxyCfg.SetRpsLimiterCfg(0, 0), convey.ShouldNotBeNil)
}

func testSetBandwidthLimiterCfg() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	convey.So(proxyCfg.SetBandwidthLimiterCfg(1, 1), convey.ShouldBeNil)
}

func testSetBandwidthLimiterCfgWithReserveRate() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	convey.So(proxyCfg.SetBandwidthLimiterCfg(1, 1, 1), convey.ShouldBeNil)
}

func testSetBandwidthLimiterCfgErrInvalid() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	convey.So(proxyCfg.SetBandwidthLimiterCfg(0, 0), convey.ShouldNotBeNil)
}

func testSetTimeoutDefault() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	proxyCfg.SetTimeout(0, 0, 0)
}

func testSetTimeout() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	proxyCfg.SetTimeout(1, 1, 1)
}

func testSetSizeLimitDefault() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	proxyCfg.SetSizeLimit(0)
}

func testSetSizeLimit() {
	tlsInfo := prepareTlsInfo(true)
	proxyCfg, err := InitProxyConfig(endpointName, localHostIp, dtTestPort, tlsInfo, serviceURL)
	convey.So(err, convey.ShouldBeNil)
	proxyCfg.SetSizeLimit(1)
}

// get server or client tls certs
func prepareTlsInfo(isServer bool) certutils.TlsCertInfo {
	testDataDir := "./testdata"
	var certInfo certutils.TlsCertInfo
	if isServer {
		certInfo.RootCaPath = filepath.Join(testDataDir, "server_ca.crt")
		certInfo.CertPath = filepath.Join(testDataDir, "server.crt")
		certInfo.KeyPath = filepath.Join(testDataDir, "server.key")
		certInfo.SvrFlag = true
	} else {
		certInfo.RootCaPath = filepath.Join(testDataDir, "client_ca.crt")
		certInfo.CertPath = filepath.Join(testDataDir, "client.crt")
		certInfo.KeyPath = filepath.Join(testDataDir, "client.key")
	}
	return certInfo
}
