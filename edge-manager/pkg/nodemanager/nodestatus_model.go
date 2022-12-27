// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"edge-manager/pkg/kubeclient"
	"huawei.com/mindxedge/base/common"
)

const (
	halfMin = time.Second * 30
)

var (
	nodeStatusService nodeStatusServiceImpl
)

// NodeStatusService provide node status from k8s
type NodeStatusService interface {
	// List lists all k8s node status
	List() *[]NodeInfoDynamic
	// Get gets specific node status by hostname
	Get(hostname string) (*NodeInfoDynamic, bool)
}

// NodeInfoDynamic dynamic node information from k8s
type NodeInfoDynamic struct {
	Hostname string `json:"-"`
	Status   string `json:"status"`
	Cpu      int64  `json:"cpu"`
	Npu      int64  `json:"npu"`
	Memory   int64  `json:"memory"`
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

func (s *nodeStatusServiceImpl) Get(nodeName string) (*NodeInfoDynamic, bool) {
	obj, ok, err := s.informer.GetStore().GetByKey(nodeName)
	if err != nil {
		hwlog.RunLog.Warnf("get node status failed: %s", err.Error())
		return nil, false
	}
	if !ok {
		hwlog.RunLog.Warnf("get node status failed: no such node %s", nodeName)
		return nil, false
	}
	node, ok := obj.(*v1.Node)
	if !ok {
		hwlog.RunLog.Warnf("get node status failed: type convert error %T", obj)
		return nil, false
	}
	return newDynamicInfo(node), true
}

func (s *nodeStatusServiceImpl) List() *[]NodeInfoDynamic {
	objects := s.informer.GetStore().List()
	var allNodes []NodeInfoDynamic
	for _, obj := range objects {
		node, ok := obj.(*v1.Node)
		if !ok {
			hwlog.RunLog.Warnf("list node status failed: failed to convert type %T", obj)
			continue
		}
		allNodes = append(allNodes, *newDynamicInfo(node))
	}
	return &allNodes
}

func newDynamicInfo(node *v1.Node) *NodeInfoDynamic {
	npuCapacity, ok := node.Status.Capacity[common.DeviceType]
	var npu int64
	if ok {
		npu = npuCapacity.Value()
	}
	return &NodeInfoDynamic{
		Hostname: node.Name,
		Status:   evalNodeStatus(node),
		Cpu:      node.Status.Capacity.Cpu().Value(),
		Memory:   node.Status.Capacity.Memory().Value(),
		Npu:      npu,
	}
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
