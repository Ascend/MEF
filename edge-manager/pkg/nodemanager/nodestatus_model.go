// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"errors"
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"edge-manager/pkg/kubeclient"
)

const (
	halfMin = time.Second * 30
)

var (
	nodeStatusService nodeStatusServiceImpl
)

// NodeStatusService provide node status from k8s
type NodeStatusService interface {
	// GetNodeStatus gets specific node status by hostname
	GetNodeStatus(hostname string) (string, error)
	// ListNodeStatus lists all k8s node status
	ListNodeStatus() map[string]string
	// GetAllocatableResource gets specific node resource(cpu & resource) by hostname
	GetAllocatableResource(hostname string) (*NodeResource, error)
	// GetAvailableResource gets available node resource(cpu & resource) by hostname
	GetAvailableResource(hostname string) (*NodeResource, error)
}

// NodeResource dynamic node information from k8s
type NodeResource struct {
	Cpu    resource.Quantity `json:"cpu"`
	Memory resource.Quantity `json:"memory"`
	Npu    resource.Quantity `json:"npu"`
}

type nodeStatusServiceImpl struct {
	informer cache.SharedIndexInformer
}

// NodeStatusServiceInstance get NodeStatusService singleton
func NodeStatusServiceInstance() NodeStatusService {
	return &nodeStatusService
}

// initNodeStatusService init k8s informer
func initNodeStatusService() error {
	nodeStatusService = nodeStatusServiceImpl{}
	client := kubeclient.GetKubeClient().GetClientSet()
	stopCh := make(chan struct{})
	nodeStatusService.initNodeInformer(stopCh, client)
	if err := nodeStatusService.run(stopCh); err != nil {
		return err
	}
	return nil
}

func (s *nodeStatusServiceImpl) run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	hwlog.RunLog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, s.informer.HasSynced); !ok {
		hwlog.RunLog.Error("failed to wait for caches to sync ")
		return errors.New("failed to wait for caches to sync")
	}
	return nil
}

func (s *nodeStatusServiceImpl) initNodeInformer(stopCh <-chan struct{}, clientSet *kubernetes.Clientset) {
	nodeInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientSet,
		halfMin,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {}),
	)
	s.informer = nodeInformerFactory.Core().V1().Nodes().Informer()
	nodeInformerFactory.Start(stopCh)
}

func (s *nodeStatusServiceImpl) GetAllocatableResource(hostname string) (*NodeResource, error) {
	node, err := s.getNode(hostname)
	if err != nil {
		return nil, err
	}
	nodeResource := &NodeResource{
		Cpu:    *node.Status.Allocatable.Cpu(),
		Memory: *node.Status.Allocatable.Memory(),
		Npu:    resource.Quantity{},
	}
	npu, ok := node.Status.Allocatable[common.DeviceType]
	if ok {
		hwlog.RunLog.Warnf("node [%s] do not have available NPU", hostname)
		nodeResource.Npu = npu
	}

	return nodeResource, nil
}

func (s *nodeStatusServiceImpl) GetNodeStatus(hostname string) (string, error) {
	node, err := s.getNode(hostname)
	if err != nil {
		return "", err
	}
	return evalNodeStatus(node), nil
}

func (s *nodeStatusServiceImpl) ListNodeStatus() map[string]string {
	objects := s.informer.GetStore().List()
	allNodeStatus := make(map[string]string)
	for _, obj := range objects {
		node, ok := obj.(*v1.Node)
		if !ok {
			hwlog.RunLog.Warnf("list node status failed: failed to convert type %T", obj)
			continue
		}
		allNodeStatus[node.Name] = evalNodeStatus(node)
	}
	return allNodeStatus
}

func (s *nodeStatusServiceImpl) GetAvailableResource(hostname string) (*NodeResource, error) {
	AllocatedRes, err := kubeclient.GetKubeClient().GetNodeAllocatedResource(hostname)
	if err != nil {
		return nil, fmt.Errorf("get node all allocated resource failed %s", err.Error())
	}
	AllocatableRes, err := s.GetAllocatableResource(hostname)
	if err != nil {
		return nil, fmt.Errorf("get node all allocatable resource failed: %s", err.Error())
	}
	allocatedCpu, ok := AllocatedRes[v1.ResourceCPU]
	if !ok {
		return nil, errors.New("get allocated resources cpu failed")
	}
	allocatedMemory, ok := AllocatedRes[v1.ResourceMemory]
	if !ok {
		return nil, errors.New("get allocated resources memory failed")
	}
	allocatedNpu, ok := AllocatedRes[common.DeviceType]
	if !ok {
		return nil, errors.New("get allocated resources npu failed")
	}
	AllocatableRes.Cpu.Sub(allocatedCpu)
	AllocatableRes.Memory.Sub(allocatedMemory)
	AllocatableRes.Npu.Sub(allocatedNpu)
	return AllocatableRes, nil
}

func (s *nodeStatusServiceImpl) getNode(hostname string) (*v1.Node, error) {
	obj, ok, err := s.informer.GetStore().GetByKey(hostname)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no such node %s", hostname)
	}
	node, ok := obj.(*v1.Node)
	if !ok {
		return nil, fmt.Errorf("type convert error %T", obj)
	}
	return node, nil
}

func evalNodeStatus(node *v1.Node) string {
	for _, cond := range node.Status.Conditions {
		if cond.Type != v1.NodeReady {
			continue
		}
		switch cond.Status {
		case v1.ConditionTrue:
			return statusReady
		case v1.ConditionFalse:
			return statusNotReady
		case v1.ConditionUnknown:
			return statusUnknown
		default:
			return statusOffline
		}
	}
	return statusOffline
}
