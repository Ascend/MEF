// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"huawei.com/mindx/common/checker/valid"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
)

const (
	podMaxLen     = 100
	podNameMaxLen = 64
	arrLen        = 2
)

var endpoint string

func getClusterroleYaml(path string) error {
	content := `kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: edge-manager-role
  namespace: mef-center
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get","list","watch","delete","patch"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["create","get","list","watch","update","delete"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["create","get","update"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["create","get"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["create","update","delete","get"]
  - apiGroups: ["apps"]
    resources: ["daemonsets"]
    verbs: ["create","update","watch","list","delete","get","patch"]
`
	if err := utils.WriteData(path, []byte(content)); err != nil {
		hwlog.RunLog.Errorf("write yaml meets error: %v", err)
		return err
	}
	if err := utils.SetPathPermission(path, utils.Mode400, false, false); err != nil {
		return err
	}
	return nil
}

func getApiserverPodName() (string, error) {
	pods, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "get", "pod", "-n", "kube-system")
	if err != nil {
		return "", fmt.Errorf("get kube-system pod list failed: %v", err)
	}

	var podName string
	lines := strings.Split(pods, "\n")
	if len(lines) > podMaxLen {
		return "", errors.New("the number of pod exceed max size")
	}

	for _, line := range lines {
		found, err := regexp.MatchString(`^kube-apiserver-`, line)
		if err != nil {
			return "", fmt.Errorf("pod name reg match failed: %s", err)
		}
		if found {
			podName = regexp.MustCompile(`^kube-apiserver-(.*?) `).FindString(line)
			if podName == "" || len(podName) > podNameMaxLen {
				return "", errors.New("apiserver pod name invalid")
			}
		}
	}
	podName = strings.TrimSpace(podName)
	return podName, nil
}

func getKubeClientCA(podCommand string) (string, error) {
	kubeclientCaPathRes := regexp.MustCompile(`client-ca-file=(.*?).crt`).FindString(podCommand)
	if kubeclientCaPathRes == "" {
		return "", errors.New("no found apiserver client ca path")
	}
	kubeclientCaPathArr := strings.Split(kubeclientCaPathRes, "=")
	if len(kubeclientCaPathArr) != arrLen {
		return "", errors.New("ca path parse failed")
	}

	if _, err := utils.RealFileChecker(kubeclientCaPathArr[1], false, false, 1); err != nil {
		return "", fmt.Errorf("ca path [%s] check failed: %v", kubeclientCaPathArr[1], err)
	}
	caContent, err := utils.LoadFile(kubeclientCaPathArr[1])
	if err != nil {
		return "", fmt.Errorf("load file failed: %s", err.Error())
	}
	caMgr, err := x509.NewCaChainMgr(caContent)
	if err != nil {
		return "", fmt.Errorf("new ca chain mgr failed: %v", err)
	}

	if err = caMgr.CheckCertChain(); err != nil {
		found, err := regexp.MatchString(`the length of RSA public key 2048 less than`, err.Error())
		if err != nil || !found {
			return "", fmt.Errorf("check k8s apiserver client ca failed: %s", err)
		}
		hwlog.RunLog.Warn("k8s apiserver client ca public key length not enough")
	}
	return kubeclientCaPathArr[1], nil
}

func getApiserverEndpoint(podCommand string) error {
	addrStr := regexp.MustCompile(`advertise-address=(.*?)"`).FindString(podCommand)
	if addrStr == "" {
		return errors.New("no found apiserver advertise address")
	}

	// len(addrStr)-1: because addrStr result is like "advertise-address=xx.xx.xx.xx", so need split end char "
	addr := strings.Split(addrStr[:len(addrStr)-1], "=")
	if len(addr) != arrLen {
		return errors.New("advertise address parse failed")
	}
	parsedIp := net.ParseIP(addr[1])
	if parsedIp == nil {
		return errors.New("apiserver advertise address is invalid")
	}

	portStr := regexp.MustCompile(`--secure-port=(.*?)"`).FindString(podCommand)
	if portStr == "" {
		return errors.New("no found apiserver secure port")
	}
	portArr := strings.Split(portStr[:len(portStr)-1], "=")
	if len(portArr) != arrLen {
		return errors.New("apiserver secure port parse failed")
	}
	port, err := strconv.Atoi(portArr[1])
	if err != nil {
		return fmt.Errorf("convert port to int value error:%v", err)
	}
	if !valid.IsPortInRange(common.MinPort, common.MaxPort, port) {
		return fmt.Errorf("apiserver secure port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
	}

	endpoint = fmt.Sprintf(`"%s:%s"`, addr[1], portArr[1])
	return nil
}

// GetApiserverEndpoint get apiserver endpoint
func GetApiserverEndpoint() string {
	if endpoint == "" {
		hwlog.RunLog.Warnf("cannot get apiserver endpoint,please check and modify" +
			"edge-manager.yaml in env API_SERVER_ENDPOINT with real endpoint")
		fmt.Println("cannot get apiserver endpoint,please check and modify" +
			"edge-manager.yaml in env API_SERVER_ENDPOINT with real endpoint")
	}
	return endpoint
}

// ModifyEndpointYaml modify apiserver endpoint in edgemanager yaml
func ModifyEndpointYaml(endpoint string, yamlPath string) error {
	yamlDealer := &kubeconfigYamlMgr{
		apiserverEndpoint: endpoint,
	}
	ret, err := utils.LoadFile(yamlPath)
	if err != nil {
		hwlog.RunLog.Errorf("reading yaml [%s] meets error: %v", yamlPath, err)
		return err
	}
	content := string(ret)
	content, err = yamlDealer.endpointModifier(content, "${apiserver_endpoint}").modifyMntDir()
	if err != nil {
		return err
	}

	if err = common.WriteData(yamlPath, []byte(content)); err != nil {
		return fmt.Errorf("cannot save yaml content: %v", err)
	}
	return nil
}
