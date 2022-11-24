// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package kubeclient to init kubeclient
package kubeclient

import (
	"context"
	"edge-manager/pkg/appmanager"
	"fmt"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/k8stool"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

func (ki *Client) CreateDaemonSet(dm *appv1.DaemonSet) (*appv1.DaemonSet, error) {
	return ki.kubeClient.AppsV1().DaemonSets(v1.NamespaceAll).Create(context.Background(), dm, metav1.CreateOptions{})
}

func (ki *Client) InitDaemonSet(app *appmanager.AppInstanceInfo) (*appv1.DaemonSet, error) {

	tmpSpec := v1.PodSpec{}
	tmpSpec.Containers = getContainer(app.AppContainer)

	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"appManager": "1",
			},
		},
		Spec: tmpSpec,
	}
	return &appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: app.AppInfo.AppName,
			Labels: map[string]string{
				"test": "1",
			},
		},
		Spec: appv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"test": "test",
				},
			},
			Template: template,
		},
	}, nil
}

func getContainer(appContainer appmanager.AppContainer) []v1.Container {
	return []v1.Container{
		{
			Name:            appContainer.AppName,
			Image:           appContainer.ImageName,
			ImagePullPolicy: v1.PullIfNotPresent,
			Resources:       getResources(appContainer),
		},
	}
}

func getResources(appContainer appmanager.AppContainer) v1.ResourceRequirements {
	var Requests map[v1.ResourceName]resource.Quantity
	var limits map[v1.ResourceName]resource.Quantity

	cpuRequest, _ := resource.ParseQuantity(appContainer.CpuRequest)
	cpuLimit, _ := resource.ParseQuantity(appContainer.CpuLimit)
	memRequest, _ := resource.ParseQuantity(appContainer.MemoryRequest)
	memLimits, _ := resource.ParseQuantity(appContainer.MemoryLimit)
	Requests = map[v1.ResourceName]resource.Quantity{v1.ResourceCPU: cpuRequest, v1.ResourceMemory: memRequest}
	limits = map[v1.ResourceName]resource.Quantity{v1.ResourceCPU: cpuLimit, v1.ResourceMemory: memLimits}

	return v1.ResourceRequirements{
		Limits:   limits,
		Requests: Requests,
	}
}

// GetPodListWithDaemonSetName is to get pod list by daemonset name
func (ki *Client) GetPodListWithDaemonSetName(dmName string) (*v1.PodList, error) {
	label := "app" + dmName
	return ki.kubeClient.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: label,
	})
}
