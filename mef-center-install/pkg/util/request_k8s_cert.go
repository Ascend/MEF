// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
)

const (
	parceCertMinArrLen  = 2
	maxCertDataLen      = 4096
	clusterroleYamlName = "cluster-role.yaml"
	bindingName         = "edge-manager-clusterrolebinding"
	clusterroleName     = "edge-manager-role"
)

// PrepareKubeConfigCert prepares kubeconfig cert
func PrepareKubeConfigCert(certPathMgr *ConfigPathMgr) error {
	return newKubeConfig(certPathMgr).prepareKubeConfigCert()
}

type kubeConfig struct {
	certPathMgr *ConfigPathMgr
}

// kubeconfigYamlMgr use to modify kubeconfig data(csr and endpoint) in yaml
type kubeconfigYamlMgr struct {
	csr               string
	apiserverEndpoint string
}

type modifier struct {
	content        string
	mark           string
	modifiedString string
}

func newKubeConfig(certPathMgr *ConfigPathMgr) *kubeConfig {
	return &kubeConfig{
		certPathMgr: certPathMgr,
	}
}

func (k *kubeConfig) prepareKubeConfigCert() error {
	hwlog.RunLog.Info("start prepare kubeconfig cert")

	csr, err := k.prepareCsr()
	if err != nil {
		hwlog.RunLog.Errorf("prepare csr error: %v", err)
		return err
	}

	if err := k.signCertFormK8s(csr); err != nil {
		hwlog.RunLog.Errorf("sign cert from k8s error: %v", err)
		return err
	}

	if err := k.prepareApiServerInfo(); err != nil {
		hwlog.RunLog.Errorf("prepare kubeclient ca failed: %v", err)
		return err
	}

	if err := k.prepareAuth(); err != nil {
		hwlog.RunLog.Errorf("prepare kube auth failed: %v", err)
		return err
	}

	hwlog.RunLog.Info("prepare kubeconfig cert success")
	return nil
}

func (k *kubeConfig) prepareCsr() ([]byte, error) {
	kmcKeyPath := filepath.Join(k.certPathMgr.GetComponentKmcDirPath(EdgeManagerName), MasterKeyFile)
	kmcBackKeyPath := k.certPathMgr.GetComponentBackKmcPath(EdgeManagerName)
	kmcConfig := kmc.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)

	san := certutils.CertSan{DnsName: []string{common.EdgeMgrDns}}
	ips, err := common.GetHostIpV4()
	if err != nil {
		return nil, err
	}
	san.IpAddr = ips

	kubeConfigKeyPath := k.certPathMgr.GetKubeConfigKeyPath()
	certBytes, err := certutils.CreateKubeConfigCsr(kubeConfigKeyPath, common.MefCertCommonNamePrefix, kmcConfig, san)
	if err != nil {
		hwlog.RunLog.Errorf("create kubeconfig csr failed: %v", err)
		return nil, err
	}

	csr := certutils.PemWrapCsr(certBytes)
	return csr, nil
}

func (k *kubeConfig) signCertFormK8s(csr []byte) error {
	csrStr := base64.StdEncoding.EncodeToString(csr)
	csrPath := filepath.Join(k.certPathMgr.GetConfigPath(), "sign-cert.yaml")
	if err := modifyCsrYaml(csrStr, csrPath); err != nil {
		hwlog.RunLog.Errorf("modify csr data in sign request error: %v", err)
		return err
	}

	if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "apply", "-f", csrPath); err != nil {
		hwlog.RunLog.Errorf("apply kubeconfig csr failed: %v", err)
		return err
	}
	defer func() {
		if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "delete",
			"csr", EdgeManagerName); err != nil {
			hwlog.RunLog.Errorf("delete kubeconfig cert sign request failed: %v", err)
			return
		}
	}()

	if err := common.DeleteFile(csrPath); err != nil {
		hwlog.RunLog.Warnf("delete tmp yaml to sign cert from k8s error :%v", err)
	}

	if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "certificate",
		"approve", EdgeManagerName); err != nil {
		hwlog.RunLog.Errorf("approve kubeconfig cert failed: %v", err)
		return err
	}

	rawCertData, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "get", "csr",
		EdgeManagerName, "-o", "jsonpath='{.status.certificate}'")
	if err != nil {
		hwlog.RunLog.Errorf("get kubeconfig sign cert failed: %v", err)
		return err
	}
	certPath := k.certPathMgr.GetKubeConfigCertPath()
	if err := checkAndSaveCert(rawCertData, certPath); err != nil {
		hwlog.RunLog.Errorf("check and save cert error: %v", err)
		return err
	}
	return nil
}

func (k *kubeConfig) prepareApiServerInfo() error {
	podName, err := getApiserverPodName()
	if err != nil {
		hwlog.RunLog.Errorf("get apiserver pod name failed: %v", err)
		return err
	}
	podCommand, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "get", "pod", "-n", "kube-system",
		podName, "-o", "jsonpath='{.spec.containers[0].command}'")
	if err != nil {
		return fmt.Errorf("get apiserver command failed: %v", err)
	}

	srcCaPath, err := getKubeClientCA(podCommand)
	if err != nil {
		hwlog.RunLog.Errorf("get kubeclient ca failed: %v", err)
		return err
	}
	caPath := k.certPathMgr.GetKubeConfigCa()
	if err := utils.CopyFile(srcCaPath, caPath); err != nil {
		hwlog.RunLog.Errorf("copy kubeclient ca cert failed: %v", err)
		return err
	}

	if err := getApiserverEndpoint(podCommand); err != nil {
		hwlog.RunLog.Warnf("get apiserver endpoint failed: %v", err)
	}
	return nil
}

func (k *kubeConfig) prepareAuth() error {
	if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "delete",
		"clusterrole", clusterroleName); err != nil {
		hwlog.RunLog.Debugf("delete clusterrole failed: %v", err)
	}

	configPath := filepath.Join(k.certPathMgr.GetConfigPath(), clusterroleYamlName)
	if err := getClusterroleYaml(configPath); err != nil {
		return err
	}
	if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "apply",
		"-f", configPath); err != nil {
		return fmt.Errorf("delete clusterrole failed: %v", err)
	}
	if err := utils.DeleteFile(configPath); err != nil {
		return err
	}

	if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "delete",
		"clusterrolebinding", bindingName); err != nil {
		hwlog.RunLog.Debugf("delete clusterrolebinding failed: %v", err)
	}
	if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "create",
		"clusterrolebinding", bindingName, fmt.Sprintf("--clusterrole=%s", clusterroleName),
		"--user=edge-manager"); err != nil {
		return fmt.Errorf("set clusterrolebinding failed: %v", err)
	}
	return nil
}

func checkAndSaveCert(rawCertData, certPath string) error {
	if len(rawCertData) > maxCertDataLen {
		return errors.New("invalid cert data length")
	}
	// certdata result is like 'xxx', so need split char '
	res := strings.Split(rawCertData, "'")
	if len(res) < parceCertMinArrLen {
		return errors.New("parse cert data error")
	}
	certData := res[1]

	cert := make([]byte, base64.StdEncoding.DecodedLen(len([]byte(certData))))
	if _, err := base64.StdEncoding.Decode(cert, []byte(certData)); err != nil {
		return fmt.Errorf("decode kubeconfig cert failed: %v", err)
	}

	if err := utils.WriteData(certPath, cert); err != nil {
		return fmt.Errorf("save kubeconfig cert failed: %v", err)
	}
	if err := utils.SetPathPermission(certPath, utils.Mode400, false, false); err != nil {
		return err
	}
	return nil
}

func getCsrYaml() string {
	return `
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: edge-manager
spec:
  request: ${csr}
  signerName: kubernetes.io/kube-apiserver-client
  usages:
  - client auth
`
}

func modifyCsrYaml(csr string, yamlPath string) error {
	yamlDealer := &kubeconfigYamlMgr{
		csr: csr,
	}

	content, err := yamlDealer.csrModifier("${csr}").modifyMntDir()
	if err != nil {
		return err
	}

	if err = common.WriteData(yamlPath, []byte(content)); err != nil {
		return fmt.Errorf("cannot save yaml content: %v", err)
	}
	if err := utils.SetPathPermission(yamlPath, utils.Mode400, false, false); err != nil {
		return fmt.Errorf("set yaml permission error: %v", err)
	}
	return nil
}

func (yd *kubeconfigYamlMgr) csrModifier(mark string) *modifier {
	return &modifier{
		content:        getCsrYaml(),
		mark:           mark,
		modifiedString: yd.csr,
	}
}

func (yd *kubeconfigYamlMgr) endpointModifier(content, mark string) *modifier {
	return &modifier{
		content:        content,
		mark:           mark,
		modifiedString: yd.apiserverEndpoint,
	}
}

func (md *modifier) modifyMntDir() (string, error) {
	var retString string
	subStrings := strings.SplitN(md.content, md.mark, SplitCount)
	if len(subStrings) < SplitCount {
		hwlog.RunLog.Errorf("split yaml by %s failed, not enough substrings", md.mark)
		return "", fmt.Errorf("modify yaml failed")
	}
	retString = subStrings[0] + md.modifiedString

	subStrings = strings.SplitN(subStrings[1], LineSplitter, SplitCount)
	if len(subStrings) < SplitCount {
		hwlog.RunLog.Errorf("split yaml by %s failed, not enough substrings", LineSplitter)
		return "", fmt.Errorf("modify yaml failed")
	}
	retString = retString + LineSplitter + subStrings[1]

	return retString, nil
}
