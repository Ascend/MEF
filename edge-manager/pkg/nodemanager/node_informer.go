// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	nodeSyncService      nodeSyncImpl
	nodeGroupLabelRegexp = regexp.MustCompile(fmt.Sprintf("^%s(\\d+)$", common.NodeGroupLabelPrefix))
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
	Prepare() error
	Handlers() cache.ResourceEventHandlerFuncs
}

// NodeResource dynamic node information from k8s
type NodeResource struct {
	Cpu    resource.Quantity `json:"cpu"`
	Memory resource.Quantity `json:"memory"`
	Npu    resource.Quantity `json:"npu"`
}

type nodeSyncImpl struct {
	informer cache.SharedIndexInformer
}

// NodeSyncInstance get nodeSyncImpl singleton
func NodeSyncInstance() *nodeSyncImpl {
	return &nodeSyncService
}

// initNodeSyncService init k8s informer
func initNodeSyncService() error {
	nodeSyncService = nodeSyncImpl{}
	if err := nodeSyncService.Prepare(); err != nil {
		return err
	}
	client := kubeclient.GetKubeClient().GetClientSet()
	stopCh := make(chan struct{})
	nodeSyncService.initNodeInformer(stopCh, client)
	if err := nodeSyncService.run(stopCh); err != nil {
		return err
	}
	return nil
}

func (s *nodeSyncImpl) run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	hwlog.RunLog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, s.informer.HasSynced); !ok {
		hwlog.RunLog.Error("failed to wait for caches to sync ")
		return errors.New("failed to wait for caches to sync")
	}
	return nil
}

func (s *nodeSyncImpl) initNodeInformer(stopCh <-chan struct{}, clientSet *kubernetes.Clientset) {
	nodeInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientSet,
		halfMin,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {}),
	)
	s.informer = nodeInformerFactory.Core().V1().Nodes().Informer()
	s.informer.AddEventHandler(s.Handlers())
	nodeInformerFactory.Start(stopCh)
}

func (s *nodeSyncImpl) Prepare() error {
	return NodeServiceInstance().deleteAllUnManagedNodes()
}

func (s *nodeSyncImpl) Handlers() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    s.nodeAdded,
		UpdateFunc: s.nodeUpdated,
		DeleteFunc: s.nodeDeleted,
	}
}

func (s *nodeSyncImpl) nodeAdded(obj interface{}) {
	node := transferNodeType(obj)
	if node == nil {
		return
	}
	s.handleAddNode(node)
}

func (s *nodeSyncImpl) nodeUpdated(oldObj, newObj interface{}) {
	oldNode := transferNodeType(oldObj)
	if oldNode == nil {
		return
	}
	newNode := transferNodeType(newObj)
	if newNode == nil {
		return
	}
	s.handleUpdateNode(newNode)
}

func (s *nodeSyncImpl) nodeDeleted(Obj interface{}) {
}

func transferNodeType(Obj interface{}) *v1.Node {
	var node *v1.Node
	var ok bool
	if Obj == nil {
		return nil
	}
	node, ok = Obj.(*v1.Node)
	if !ok || node == nil {
		hwlog.RunLog.Errorf("invalid node type %T", Obj)
		return nil
	}
	if _, ok = node.Labels[masterNodeLabelKey]; ok {
		node = nil
	}

	return node
}

func (s *nodeSyncImpl) handleAddNode(node *v1.Node) {
	nodeInfo := &NodeInfo{
		NodeName:     node.Name,
		UniqueName:   node.Name,
		IP:           evalIpAddress(node),
		SerialNumber: node.Labels[snNodeLabelKey],
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	if checkResult := newNodeInfoChecker().Check(*nodeInfo); !checkResult.Result {
		hwlog.RunLog.Errorf("node info check failed: %s", checkResult.Reason)
		return
	}
	err := NodeServiceInstance().createNode(nodeInfo)
	if err != nil && !strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
		hwlog.RunLog.Errorf("automatically adding node(%s) failed, db add node error", node.Name)
		return
	}

	if label := getLabelAndGroupIDsFromNode(node); len(label) != 0 {
		s.autoUpdateLabel(node, label)
	}
	hwlog.RunLog.Infof("automatically adding node(%s) success", node.Name)
}

func (s *nodeSyncImpl) handleUpdateNode(newNode *v1.Node) {
	if label := getLabelAndGroupIDsFromNode(newNode); len(label) != 0 {
		s.autoUpdateLabel(newNode, label)
	}
}

func (s *nodeSyncImpl) autoUpdateLabel(nodeK8s *v1.Node, label []string) {
	nodeDb, err := NodeServiceInstance().getNodeByUniqueName(nodeK8s.Name)
	if err != nil {
		hwlog.RunLog.Errorf("automatically updating node(%s) failed, db query error", nodeK8s.Name)
		return
	}
	if nodeDb.IsManaged {
		return
	}
	if _, err = kubeclient.GetKubeClient().DeleteNodeLabels(nodeK8s.Name, label); err != nil {
		hwlog.RunLog.Errorf("automatically delete unmanaged node(%s) label error", nodeK8s.Name)
		return
	}
	hwlog.RunLog.Infof("automatically delete unmanaged node(%s) label", nodeK8s.Name)
}

func getLabelAndGroupIDsFromNode(node *v1.Node) []string {
	var labels []string
	for label := range node.Labels {
		matches := nodeGroupLabelRegexp.FindStringSubmatch(label)
		if len(matches) > 1 {
			labels = append(labels, matches[0])
		}
	}
	return labels
}

func (s *nodeSyncImpl) GetAllocatableResource(hostname string) (*NodeResource, error) {
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

func (s *nodeSyncImpl) GetNodeStatus(hostname string) (string, error) {
	node, err := s.getNode(hostname)
	if err != nil {
		return "", err
	}
	return evalNodeStatus(node), nil
}

func (s *nodeSyncImpl) ListNodeStatus() map[string]string {
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

func (s *nodeSyncImpl) GetAvailableResource(hostname string) (*NodeResource, error) {
	AllocatedRes, err := kubeclient.GetKubeClient().GetNodeAllocatedResource(hostname)
	if err != nil {
		return nil, fmt.Errorf("get node all allocated resource failed: %s", err.Error())
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

func (s *nodeSyncImpl) getNode(hostname string) (*v1.Node, error) {
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
