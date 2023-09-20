// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package kubeclient to init kubeclient
package kubeclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"huawei.com/mindx/common/checker/valid"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
)

const (
	keyOp               = "op"
	keyPath             = "path"
	keyValue            = "value"
	opRemove            = "remove"
	opAdd               = "add"
	labelResourcePrefix = "/metadata/labels/"
	fieldSelectorPrefix = "spec.nodeName="
	arrLen              = 2

	mefEdgeNodeLabel    = "mef-edge-node"
	kubeSystemNamespace = "kube-system"
	systemNamespace     = "kubeedge"
	tokenSecretName     = "tokensecret"
	tokenDataName       = "tokendata"
	caSecretName        = "casecret"
	caDataName          = "cadata"
	caKeyDataName       = "cakeydata"
	// K8sNotFoundErrorFragment for check if the error is found type
	K8sNotFoundErrorFragment = "not found"
	// DefaultImagePullSecretKey for getting image pull secret
	DefaultImagePullSecretKey = "image-pull-secret"
	// DefaultImagePullSecretValue for initialization of app manager to create a default image pull secret value
	DefaultImagePullSecretValue = "{}"

	configKeyQps       = "KUBE_CLIENT_QPS"
	configKeyBurst     = "KUBE_CLIENT_BURST"
	configEndpoint     = "API_SERVER_ENDPOINT"
	kubeConfigCertPath = "/home/data/config/kube-config/server.crt"
	kubeConfigKeyPath  = "/home/data/config/kube-config/server.key"
	kubeConfigCaPath   = "/home/data/config/kube-config/root.crt"
)

var k8sClient *Client

// Client k8s client
type Client struct {
	kubeClient *kubernetes.Clientset
}

// NewClientK8s create ClientK8s
func NewClientK8s() (*Client, error) {
	pemPair, err := certutils.GetCertPairForPemWithBackup(kubeConfigCertPath, kubeConfigKeyPath, nil)
	if err != nil {
		return nil, fmt.Errorf("get kubeconfig cert pair failed: %v", err)
	}
	rootCaPemBytes, err := certutils.GetCertContentWithBackup(kubeConfigCaPath)
	if err != nil || rootCaPemBytes == nil {
		return nil, fmt.Errorf("load kube client ca failed")
	}

	selfCreateConfig := rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			Insecure:   false,
			ServerName: "",
			CertData:   pemPair.CertPem,
			KeyData:    pemPair.KeyPem,
			CAData:     rootCaPemBytes,
			NextProtos: []string(nil),
		},
	}
	if err := setupClientConfig(&selfCreateConfig); err != nil {
		return nil, fmt.Errorf("set kube client config failed: %v", err)
	}
	client, err := kubernetes.NewForConfig(&selfCreateConfig)
	if err != nil || client == nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}
	hwlog.RunLog.Info("init k8s success")
	k8sClient = &Client{
		kubeClient: client,
	}
	k8sClient.systemComponentsModify()
	return k8sClient, nil
}

func (ki *Client) systemComponentsModify() {
	dsList, err := ki.kubeClient.AppsV1().DaemonSets(kubeSystemNamespace).
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		hwlog.RunLog.Warnf("modify kube-system components failed, get system daemonset error: %v", err)
		return
	}
	for _, ds := range dsList.Items {
		if ds.Spec.Template.Spec.Affinity == nil {
			ds.Spec.Template.Spec.Affinity = &v1.Affinity{}
		}
		if ds.Spec.Template.Spec.Affinity.NodeAffinity == nil {
			ds.Spec.Template.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{}
		}
		if ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution =
				&v1.NodeSelector{}
		}
		terms := ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.
			NodeSelectorTerms
		if isModified(terms) {
			return
		}
		term := v1.NodeSelectorTerm{
			MatchExpressions: []v1.NodeSelectorRequirement{
				{
					Key:      mefEdgeNodeLabel,
					Operator: v1.NodeSelectorOpDoesNotExist,
				},
			},
		}
		ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.
			NodeSelectorTerms = append(terms, term)
		if _, err := ki.kubeClient.AppsV1().DaemonSets(kubeSystemNamespace).
			Update(context.Background(), &ds, metav1.UpdateOptions{}); err != nil {
			hwlog.RunLog.Warnf("modify kube-system daemonset failed, %v", err)
		}
	}
}

func isModified(terms []v1.NodeSelectorTerm) bool {
	for _, term := range terms {
		for _, expression := range term.MatchExpressions {
			if expression.Key == mefEdgeNodeLabel && expression.Operator == v1.NodeSelectorOpDoesNotExist {
				return true
			}
		}
	}
	return false
}

func setupClientConfig(clientConfig *rest.Config) error {
	var (
		qps   float32
		burst int
	)
	decoder := json.NewDecoder(strings.NewReader(os.Getenv(configKeyQps)))
	if err := decoder.Decode(&qps); err != nil {
		return fmt.Errorf("decode %s failed", configKeyQps)
	}
	decoder = json.NewDecoder(strings.NewReader(os.Getenv(configKeyBurst)))
	if err := decoder.Decode(&burst); err != nil {
		return fmt.Errorf("decode %s failed", configKeyBurst)
	}
	clientConfig.QPS = qps
	clientConfig.Burst = burst

	endpointStr := os.Getenv(configEndpoint)
	if err := checkEndpoint(endpointStr); err != nil {
		return err
	}
	clientConfig.Host = fmt.Sprintf("https://%s", endpointStr)
	return nil
}

func checkEndpoint(endpointStr string) error {
	if endpointStr == "" {
		return errors.New("apiserver endpoint is nil, please modify " +
			"edge-manager.yaml in env API_SERVER_ENDPOINT with real endpoint")
	}
	endpoint := strings.Split(endpointStr, ":")
	if len(endpoint) != arrLen {
		return errors.New("endpoint parse failed")
	}

	parsedIp := net.ParseIP(endpoint[0])
	if parsedIp == nil {
		return errors.New("apiserver advertise address is invalid")
	}

	port, err := strconv.Atoi(endpoint[1])
	if err != nil {
		return fmt.Errorf("convert port to int value error:%v", err)
	}
	if !valid.IsPortInRange(common.MinPort, common.MaxPort, port) {
		return fmt.Errorf("apiserver secure port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
	}
	return nil
}

// GetKubeClient get k8s client
func GetKubeClient() *Client {
	return k8sClient
}

// GetClientSet get k8s client set
func (ki *Client) GetClientSet() *kubernetes.Clientset {
	return ki.kubeClient
}

// ListNode list nodes
func (ki *Client) ListNode() (*v1.NodeList, error) {
	return ki.kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{FieldSelector: ""})
}

// GetNodeAllocatedResource [method] for calculating all allocated resources of one node
func (ki *Client) GetNodeAllocatedResource(nodeName string) (v1.ResourceList, error) {
	podInterface := ki.GetClientSet().CoreV1().Pods(v1.NamespaceAll)
	fieldSelector, err := fields.ParseSelector(fieldSelectorPrefix + nodeName)
	if err != nil {
		return nil, errors.New("parse field selector error")
	}
	listOptions := metav1.ListOptions{FieldSelector: fieldSelector.String()}
	podList, err := podInterface.List(context.Background(), listOptions)
	if err != nil {
		return nil, errors.New("get pod allocated pod list error")
	}
	AllocatedRes := map[v1.ResourceName]resource.Quantity{
		v1.ResourceCPU:    {},
		v1.ResourceMemory: {},
		common.DeviceType: {},
	}
	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			for name, quantity := range container.Resources.Limits {
				tmp := AllocatedRes[name]
				tmp.Add(quantity)
				AllocatedRes[name] = tmp
			}
		}
	}
	return AllocatedRes, nil
}

// DeletePodByForce compulsorily delete pod by namespace and name
func (ki *Client) DeletePodByForce(pod *v1.Pod) error {
	gracePeriodSeconds := int64(0)
	return ki.kubeClient.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	})
}

// GetPodList is to get pod list
func (ki *Client) GetPodList() (*v1.PodList, error) {
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": ""})
	return ki.kubeClient.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: selector.String(),
	})
}

// DeleteNode is to remove node
func (ki *Client) DeleteNode(nodeName string) error {
	return ki.kubeClient.CoreV1().Nodes().Delete(context.Background(), nodeName, metav1.DeleteOptions{})
}

// DeleteNodeLabels is to remove node labels, failing if any label not exists
func (ki *Client) DeleteNodeLabels(nodeName string, labelNames []string) (*v1.Node, error) {
	if len(labelNames) == 0 {
		return nil, errors.New("labelNames can't be empty")
	}
	patch := make([]map[string]interface{}, 0, len(labelNames))
	for _, labelName := range labelNames {
		op := map[string]interface{}{
			keyOp:   opRemove,
			keyPath: ki.makeLabelPath(labelName),
		}
		patch = append(patch, op)
	}
	return ki.patchNode(nodeName, patch)
}

// AddNodeLabels is to add node labels, overwriting label value if exists
func (ki *Client) AddNodeLabels(nodeName string, labels map[string]string) (*v1.Node, error) {
	if len(labels) == 0 {
		return nil, errors.New("labels can't be empty")
	}
	patch := make([]map[string]interface{}, 0, len(labels))
	for name, value := range labels {
		op := map[string]interface{}{
			keyOp:    opAdd,
			keyPath:  ki.makeLabelPath(name),
			keyValue: value,
		}
		patch = append(patch, op)
	}
	return ki.patchNode(nodeName, patch)
}

// patchNode use "json" patch type
func (ki *Client) patchNode(nodeName string, patch []map[string]interface{}) (*v1.Node, error) {
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}
	return ki.kubeClient.CoreV1().Nodes().
		Patch(context.Background(), nodeName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
}

// makeLabelPath make valid JSON Pointer: rfc6901
func (ki *Client) makeLabelPath(name string) string {
	name = strings.ReplaceAll(name, "~", "~0")
	name = strings.ReplaceAll(name, "/", "~1")
	return labelResourcePrefix + name
}

// CreateDaemonSet create daemonSet
func (ki *Client) CreateDaemonSet(dm *appv1.DaemonSet) (*appv1.DaemonSet, error) {
	return ki.kubeClient.AppsV1().DaemonSets(common.MefUserNs).Create(context.Background(), dm, metav1.CreateOptions{})
}

// UpdateDaemonSet Update daemonSet
func (ki *Client) UpdateDaemonSet(dm *appv1.DaemonSet) (*appv1.DaemonSet, error) {
	return ki.kubeClient.AppsV1().DaemonSets(common.MefUserNs).Update(context.Background(), dm, metav1.UpdateOptions{})
}

// DeleteDaemonSet Delete daemonSet
func (ki *Client) DeleteDaemonSet(name string) error {
	return ki.kubeClient.AppsV1().DaemonSets(common.MefUserNs).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// GetToken get token
func (ki *Client) GetToken() ([]byte, error) {
	secret, err := ki.kubeClient.CoreV1().Secrets(systemNamespace).
		Get(context.Background(), tokenSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	token, ok := secret.Data[tokenDataName]
	if !ok {
		return nil, errors.New("token obtained from secret is not found")
	}
	return token, nil
}

// CreateConfigMap create configmap
func (ki *Client) CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).Create(context.Background(), cm, metav1.CreateOptions{})
}

// GetConfigMap get configmap
func (ki *Client) GetConfigMap(name string) (*v1.ConfigMap, error) {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).Get(context.Background(), name, metav1.GetOptions{})
}

// DeleteConfigMap delete configmap
func (ki *Client) DeleteConfigMap(name string) error {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// UpdateConfigMap update configmap
func (ki *Client) UpdateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).Update(context.Background(), cm, metav1.UpdateOptions{})
}

// GetSecret [method] for creating secret
func (ki *Client) GetSecret(name string) (*v1.Secret, error) {
	return ki.GetClientSet().CoreV1().Secrets(common.MefUserNs).Get(context.Background(), name, metav1.GetOptions{})
}

// GetCloudCoreCa [method] for get cloud core ca
func (ki *Client) GetCloudCoreCa() ([]byte, error) {
	caSecret, err := ki.GetClientSet().CoreV1().Secrets(systemNamespace).Get(context.Background(),
		caSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	caData, ok := caSecret.Data[caDataName]
	if !ok {
		return nil, errors.New("cloud ca data not exist")
	}

	if _, ok := caSecret.Data[caKeyDataName]; !ok {
		return nil, errors.New("cloud ca key data not exist")
	}

	utils.ClearSliceByteMemory(caSecret.Data[caKeyDataName])

	if err = x509.CheckDerCertChain(caData); err != nil {
		return nil, fmt.Errorf("parse cloudcore cert failed: %s", err.Error())
	}

	return caData, nil
}

// CreateOrUpdateSecret [method] for updating secret or creating secret if it is not exist
func (ki *Client) CreateOrUpdateSecret(secret *v1.Secret) (*v1.Secret, error) {
	_, err := ki.GetSecret(secret.Name)
	if err == nil {
		return ki.kubeClient.CoreV1().Secrets(common.MefUserNs).Update(context.Background(),
			secret, metav1.UpdateOptions{})
	}
	if strings.Contains(err.Error(), K8sNotFoundErrorFragment) {
		return ki.kubeClient.CoreV1().Secrets(common.MefUserNs).Create(context.Background(),
			secret, metav1.CreateOptions{})
	}
	return nil, err
}

// CreateNamespace [method] for creating namespace if it is node exist
func (ki *Client) CreateNamespace(ns *v1.Namespace) (*v1.Namespace, error) {
	_, err := ki.GetClientSet().CoreV1().Namespaces().Get(context.Background(), ns.Namespace, metav1.GetOptions{})
	if err == nil {
		return nil, nil
	}
	if strings.Contains(err.Error(), K8sNotFoundErrorFragment) {
		return ki.GetClientSet().CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	}
	return nil, err
}
