package websocket

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"huawei.com/mindx/common/hwlog"
)

type ProxyConfig struct {
	name       string
	tlsConfig  *tls.Config
	hosts      string
	headers    http.Header
	handlerMgr WsMsgHandler
	ctx        context.Context
	cancel     context.CancelFunc
}

func (pc *ProxyConfig) RegModInfos(regHandlers []RegisterModuleInfo) {
	for _, reg := range regHandlers {
		pc.handlerMgr.Register(reg)
	}
}

func InitProxyConfig(name string, ip string, port int, certInfo CertPathInfo) (*ProxyConfig, error) {
	netConfig := &ProxyConfig{}
	netConfig.name = name
	netConfig.hosts = fmt.Sprintf("%s:%d", ip, port)
	netConfig.handlerMgr = WsMsgHandler{}
	tlsConfig, err := getTlsConfig(certInfo)
	if err != nil {
		return nil, fmt.Errorf("get tls config failed: %v\n", err)
	}
	netConfig.tlsConfig = tlsConfig
	netConfig.headers = http.Header{}
	netConfig.headers.Set(clientNameKey, name)
	netConfig.ctx, netConfig.cancel = context.WithCancel(context.Background())
	return netConfig, nil
}

func getTlsConfig(certInfo CertPathInfo) (*tls.Config, error) {
	pubBytes, err := ioutil.ReadFile(certInfo.SvrCertPath)
	if err != nil {
		return nil, err
	}
	priBytes, err := ioutil.ReadFile(certInfo.SvrKeyPath)
	if err != nil {
		return nil, err
	}
	certBytes, err := ioutil.ReadFile(certInfo.RootCaPath)
	if err != nil {
		return nil, err
	}
	pair, err := tls.X509KeyPair(pubBytes, priBytes)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(certBytes)
	if !ok {
		return nil, fmt.Errorf("append certs failed")
	}
	cfg := tls.Config{
		Rand:               rand.Reader,
		Certificates:       []tls.Certificate{pair},
		RootCAs:            certPool,
		InsecureSkipVerify: true, // todo 设置为false
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		MinVersion: tls.VersionTLS13,
	}
	if certInfo.ServerFlag {
		cfg.ClientAuth = tls.RequireAnyClientCert
		cfg.ClientCAs = certPool
	} else {
		cfg.RootCAs = certPool
	}
	hwlog.RunLog.Info("get tls config success")
	return &cfg, nil
}
