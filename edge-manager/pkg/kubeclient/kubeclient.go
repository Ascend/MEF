// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package kubeclient to init kubeclient
package kubeclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/k8stool"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

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

	systemNamespace = "kubeedge"
	tokenSecretName = "tokensecret"
	tokenDataName   = "tokendata"

	// K8sNotFoundErrorFragment for check if the error is found type
	K8sNotFoundErrorFragment = "not found"
	// DefaultImagePullSecretKey for getting image pull secret
	DefaultImagePullSecretKey = "image-pull-secret"
	// DefaultImagePullSecretValue for initialization of app manager to create a default image pull secret value
	DefaultImagePullSecretValue = "{}"
)

var k8sClient *Client

// Client k8s client
type Client struct {
	kubeClient *kubernetes.Clientset
}

// NewClientK8s create ClientK8s
func NewClientK8s(kubeConfig string) (*Client, error) {
	client, err := k8stool.K8sClientFor(kubeConfig, "")
	if err != nil || client == nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}
	hwlog.RunLog.Info("init k8s success")
	k8sClient = &Client{
		kubeClient: client,
	}
	return k8sClient, nil
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
	return secret.Data[tokenDataName], nil
}

// CreateConfigMap create configmap
func (ki *Client) CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).Create(context.Background(), cm, metav1.CreateOptions{})
}

// DeleteConfigMap delete configmap
func (ki *Client) DeleteConfigMap(name string) error {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// UpdateConfigMap update configmap
func (ki *Client) UpdateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).Update(context.Background(), cm, metav1.UpdateOptions{})
}

// GetConfigMap get configmap
func (ki *Client) GetConfigMap(name string) (*v1.ConfigMap, error) {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).Get(context.Background(), name, metav1.GetOptions{})
}

// ListConfigMapList list configmap list
func (ki *Client) ListConfigMapList() (*v1.ConfigMapList, error) {
	return ki.kubeClient.CoreV1().ConfigMaps(common.MefUserNs).List(context.Background(), metav1.ListOptions{})
}

// GetSecret [method] for creating secret
func (ki *Client) GetSecret(name string) (*v1.Secret, error) {
	return ki.GetClientSet().CoreV1().Secrets(common.MefUserNs).Get(context.Background(), name, metav1.GetOptions{})
}

// CreateOrUpdateSecret [method] for updating  secret or creating secret if it is not exist
func (ki *Client) CreateOrUpdateSecret(secret *v1.Secret) (*v1.Secret, error) {
	_, err := ki.GetSecret(secret.Name)
	if err == nil {
		return ki.kubeClient.CoreV1().Secrets(common.MefUserNs).Update(context.Background(), secret, metav1.UpdateOptions{})
	}
	if strings.Contains(err.Error(), K8sNotFoundErrorFragment) {
		return ki.kubeClient.CoreV1().Secrets(common.MefUserNs).Create(context.Background(), secret, metav1.CreateOptions{})
	}
	return nil, err
}
