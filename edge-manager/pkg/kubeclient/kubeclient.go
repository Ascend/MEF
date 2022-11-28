// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package kubeclient to init kubeclient
package kubeclient

import (
	"context"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/k8stool"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

var k8sClient *Client

// Client k8s client
type Client struct {
	kubeClient *kubernetes.Clientset
}

// NewClientK8s create ClientK8s
func NewClientK8s() (*Client, error) {
	client, err := k8stool.K8sClientFor("", "")
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

// GetNode get node
func (ki *Client) GetNode(nodeName string) (*v1.Node, error) {
	return ki.kubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
}

// ListNode list nodes
func (ki *Client) ListNode() (*v1.NodeList, error) {
	return ki.kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{FieldSelector: ""})
}

// GetPod get pod by namespace and name
func (ki *Client) GetPod(pod *v1.Pod) (*v1.Pod, error) {
	return ki.kubeClient.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
}

// GetPodList is to get pod list
func (ki *Client) GetPodList() (*v1.PodList, error) {
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": ""})
	return ki.kubeClient.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: selector.String(),
	})
}

// CreateDaemonSet create daemonset
func (ki *Client) CreateDaemonSet(dm *appv1.DaemonSet) (*appv1.DaemonSet, error) {
	return ki.kubeClient.AppsV1().DaemonSets("default").Create(context.Background(), dm, metav1.CreateOptions{})
}

// UpdateDaemonSet Update daemonset
func (ki *Client) UpdateDaemonSet(dm *appv1.DaemonSet) (*appv1.DaemonSet, error) {
	return ki.kubeClient.AppsV1().DaemonSets("default").Update(context.Background(), dm, metav1.UpdateOptions{})
}
