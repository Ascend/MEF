// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package k8stool offer the k8s client with support encoded kubeConfig file
package k8stool

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/utils"
	hwx509 "huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"
)

const (
	prefix           = "/etc/mindx-dl/"
	configPrefix     = "apiVersion:"
	maxLen           = 2048
	kubeCfgFile      = ".config/config6"
	handshakeTimeOut = 10 * time.Second
)

var (
	k8sClientOnce sync.Once
	kubeClientSet *kubernetes.Clientset
)

// K8sClient Get the internal k8s client of the cluster
func K8sClient(kubeconfig string) (*kubernetes.Clientset, error) {
	k8sClientOnce.Do(func() {
		if kubeconfig == "" {
			configPath := os.Getenv("KUBECONFIG")
			if len(configPath) > maxLen {
				hwlog.RunLog.Error("the path is too long")
				return
			}
			kubeconfig = configPath
		}
		var path string
		var err error
		if kubeconfig != "" {
			path, err = fileutils.CheckOriginPath(kubeconfig)
			if err != nil {
				hwlog.RunLog.Error(err)
				return
			}
		}
		config, err := BuildConfigFromFlags("", path)
		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
		// Create a new k8sClientSet based on the specified config using the current context
		kubeClientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
	})
	if kubeClientSet == nil {
		return nil, errors.New("get k8s client failed")
	}

	return kubeClientSet, nil
}

// K8sClientFor  component name is noded,task-manager,hccl-controller  etc.
func K8sClientFor(kubeConfig, component string) (*kubernetes.Clientset, error) {
	// if kubeConfig not set, check and use default path
	kubeConf := prefix + component + "/" + kubeCfgFile
	if kubeConfig == "" && component != "" && fileutils.IsExist(kubeConf) {
		return K8sClient(kubeConf)
	}
	// use custom path
	return K8sClient(kubeConfig)
}

// SelfClientConfigLoadingRules  extend   clientcmd.ClientConfigLoadingRules
type SelfClientConfigLoadingRules struct {
	clientcmd.ClientConfigLoadingRules
}

// Load  override the clientcmd.ClientConfigLoadingRules Load method
func (rules *SelfClientConfigLoadingRules) Load() (*api.Config, error) {
	var err error
	if len(rules.ExplicitPath) == 0 {
		return nil, errors.New("no ExplicitPath set")
	}
	config, err := loadFromFile(rules.ExplicitPath)
	if err != nil {
		err = fmt.Errorf(`error loading config file "%s": %#v`,
			utils.MaskPrefix(rules.ExplicitPath), err)
	}
	return config, err
}

// loadFromFile takes a filename and deserializes the contents into Config object
func loadFromFile(filename string) (*api.Config, error) {
	kubeconfigBytes, err := fileutils.ReadLimitBytes(filename, fileutils.Size10M)
	if err != nil {
		return nil, err
	}
	if bytes.Contains(kubeconfigBytes, []byte(configPrefix)) {
		return nil, errors.New("do not support non-encrypted kubeConfig")
	}
	if idx := bytes.LastIndex(kubeconfigBytes, []byte(utils.SplitFlag)); idx != -1 {
		kubeconfigBytes = kubeconfigBytes[:idx]
	}
	if err = kmc.Initialize(kmc.Aes256gcm, "", ""); err != nil {
		return nil, err
	}
	defer kmc.Finalize()
	hwlog.RunLog.Info("start to decrypt cfg")
	plainText, err := kmc.Decrypt(0, kubeconfigBytes)
	if err != nil {
		return nil, err
	}
	defer hwx509.PaddingAndCleanSlice(plainText)
	cfg, err := clientcmd.Load(plainText)
	if err != nil {
		return nil, err
	}
	hwlog.RunLog.Infof("Config loaded from file: %s", utils.MaskPrefix(filename))
	for key, val := range cfg.AuthInfos {
		val.LocationOfOrigin = filename
		cfg.AuthInfos[key] = val
	}
	for key, val := range cfg.Clusters {
		val.LocationOfOrigin = filename
		cfg.Clusters[key] = val
	}
	for key, val := range cfg.Contexts {
		val.LocationOfOrigin = filename
		cfg.Contexts[key] = val
	}
	if cfg.AuthInfos == nil {
		cfg.AuthInfos = map[string]*api.AuthInfo{}
	}
	if cfg.Clusters == nil {
		cfg.Clusters = map[string]*api.Cluster{}
	}
	if cfg.Contexts == nil {
		cfg.Contexts = map[string]*api.Context{}
	}
	return cfg, nil
}

// BuildConfigFromFlags local implement of k8s client buildConfig
func BuildConfigFromFlags(masterURL, confPath string) (*rest.Config, error) {
	if confPath == "" && masterURL == "" {
		hwlog.RunLog.Warn("Neither --kubeconfig nor --master was specified." +
			"Using the inClusterConfig.  This might not work.")
		kubeconf, err := rest.InClusterConfig()
		if err == nil {
			return kubeconf, nil
		}
		hwlog.RunLog.Warn("error creating inClusterConfig")
	}

	cfg, err := loadCreateRestConfig(masterURL, confPath)
	defer func() {
		if cfg != nil && cfg.KeyData != nil {
			hwx509.PaddingAndCleanSlice(cfg.KeyData)
		}
	}()
	if err != nil {
		return nil, err
	}
	return makeSafeConfig(cfg, masterURL, confPath)
}

func loadCreateRestConfig(masterURL, confPath string) (*rest.Config, error) {
	cliRule := clientcmd.ClientConfigLoadingRules{ExplicitPath: confPath}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&SelfClientConfigLoadingRules{ClientConfigLoadingRules: cliRule},
		&clientcmd.ConfigOverrides{ClusterInfo: api.Cluster{Server: masterURL}}).ClientConfig()
}

func makeSafeConfig(cfg *rest.Config, masterURL, confPath string) (*rest.Config, error) {
	var tlsCertInfo = &certutils.TlsCertInfo{
		CertContent:   cfg.CertData,
		RootCaContent: cfg.CAData,
		KeyContent:    cfg.KeyData,
	}
	defer hwx509.PaddingAndCleanSlice(tlsCertInfo.KeyContent)
	tlsCfg, err := getSafeTlsConfigForK8s(tlsCertInfo, masterURL, confPath)
	if err != nil {
		return nil, err
	}

	safeConfig := &rest.Config{
		Host: cfg.Host,
		Transport: &http.Transport{
			TLSClientConfig:     tlsCfg,
			TLSHandshakeTimeout: handshakeTimeOut,
		},
		TLSClientConfig: rest.TLSClientConfig{
			Insecure:   false,
			ServerName: "",
			NextProtos: []string(nil),
		},
	}

	return safeConfig, nil
}

// getSafeTlsConfigForK8s return a tls config with safe CipherSuites
func getSafeTlsConfigForK8s(tlsCertInfo *certutils.TlsCertInfo, masterURL, confPath string) (*tls.Config, error) {
	if tlsCertInfo == nil {
		return nil, errors.New("get tls config failed, tls cert is nil")
	}
	if tlsCertInfo.RootCaContent == nil {
		return nil, errors.New("get tls config failed, root ca is empty")
	}
	if tlsCertInfo.CertContent == nil {
		return nil, errors.New("get tls config failed, client certifaction is empty")
	}
	if tlsCertInfo.KeyContent == nil {
		return nil, errors.New("get tls config failed, client key is empty")
	}

	tlsCfg := &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: false,
		CipherSuites:       certutils.DefaultSafeCipherSuites,
		MinVersion:         tls.VersionTLS13,
	}
	rootCaPool := x509.NewCertPool()
	if ok := rootCaPool.AppendCertsFromPEM(tlsCertInfo.RootCaContent); !ok {
		return nil, errors.New("append root ca to cert pool failed")
	}
	tlsCfg.RootCAs = rootCaPool
	tlsCfg.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
		cfg, err := loadCreateRestConfig(masterURL, confPath)
		defer func() {
			if cfg != nil && cfg.KeyData != nil {
				hwx509.PaddingAndCleanSlice(cfg.KeyData)
			}
		}()
		if err != nil {
			hwlog.RunLog.Errorf("get certification from encrypt kubeconfig file error:%v", err)
			return nil, err
		}
		pair, err := tls.X509KeyPair(cfg.CertData, cfg.KeyData)
		hwlog.RunLog.Debugf("get certification X509KeyPair error:%v", err)
		return &pair, err
	}
	return tlsCfg, nil
}
