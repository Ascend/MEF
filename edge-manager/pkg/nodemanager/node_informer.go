// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/websocketmgr"

	"huawei.com/mindxedge/base/common"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"edge-manager/pkg/kubeclient"
)

const (
	halfMin          = time.Second * 30
	nodeActionAdd    = "add"
	nodeActionDelete = "delete"
	maxNodeSize      = 2048
)

type simpleNodeInfo struct {
	Sn string `json:"sn"`
	Ip string `json:"ip"`
}

type changedNodeInfo struct {
	AddedNodeInfo   []simpleNodeInfo
	DeletedNodeInfo []simpleNodeInfo
}

var (
	nodeSyncService      nodeSyncImpl
	nodeGroupLabelRegexp = regexp.MustCompile(fmt.Sprintf("^%s(\\d+)$", common.NodeGroupLabelPrefix))
)

// NodeStatusService provide node status from k8s
type NodeStatusService interface {
	// GetK8sNodeStatus gets specific k8s node status by hostname
	GetK8sNodeStatus(hostname string) (string, error)
	// GetMEFNodeStatus gets specific mef node status by hostname
	GetMEFNodeStatus(hostname string) (string, error)
	// ListMEFNodeStatus lists all k8s node status
	ListMEFNodeStatus() map[string]string
	// GetAllocatableResource gets specific node resource(cpu & resource) by hostname
	GetAllocatableResource(hostname string) (*NodeResource, error)
	// GetAvailableResource gets available node resource(cpu & resource) by hostname
	GetAvailableResource(nodeID uint64, hostname string) (*NodeResource, error)
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
	reportChangedNodeInfo(nodeActionAdd, node)
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

func (s *nodeSyncImpl) nodeDeleted(obj interface{}) {
	node := transferNodeType(obj)
	if node == nil {
		return
	}
	reportChangedNodeInfo(nodeActionDelete, node)
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
	count, err := GetTableCount(NodeInfo{})
	if err != nil {
		hwlog.RunLog.Error("get node info table count error")
		return
	}
	if count >= maxNodeSize {
		hwlog.RunLog.Errorf("node count cannot exceed %d, please delete no need node", maxNodeSize)
		return
	}
	err = NodeServiceInstance().createNode(nodeInfo)
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

func (s *nodeSyncImpl) GetK8sNodeStatus(hostname string) (string, error) {
	node, err := s.getNode(hostname)
	if err != nil {
		return "", err
	}
	return evalNodeStatus(node), nil
}

func (s *nodeSyncImpl) GetMEFNodeStatus(hostname string) (string, error) {
	node, err := s.getNode(hostname)
	if err != nil {
		return "", err
	}
	serialNumber, ok := node.Labels[snNodeLabelKey]
	if !ok {
		return statusAbnormal, nil
	}
	status := evalNodeStatus(node)
	states := getEdgeConnStatus(serialNumber)
	connected, ok := states[serialNumber]
	if !ok || !connected {
		status = statusAbnormal
	}
	return status, nil
}

func (s *nodeSyncImpl) ListMEFNodeStatus() map[string]string {
	objects := s.informer.GetStore().List()
	nodeName2NodeStatus := make(map[string]string)
	serialNumber2NodeName := make(map[string]string)
	var serialNumbers []string
	for _, obj := range objects {
		node, ok := obj.(*v1.Node)
		if !ok {
			hwlog.RunLog.Warnf("list node status failed: failed to convert type %T", obj)
			continue
		}
		serialNumber, ok := node.Labels[snNodeLabelKey]
		if ok {
			serialNumber2NodeName[serialNumber] = node.Name
			serialNumbers = append(serialNumbers, serialNumber)
		}
		nodeName2NodeStatus[node.Name] = evalNodeStatus(node)
	}

	if len(serialNumbers) == 0 {
		return serialNumber2NodeName
	}
	for serialNumber, connected := range getEdgeConnStatus(serialNumbers...) {
		if connected {
			continue
		}
		if name, ok := serialNumber2NodeName[serialNumber]; ok {
			nodeName2NodeStatus[name] = statusAbnormal
		}
	}
	return nodeName2NodeStatus
}

func getEdgeConnStatus(snList ...string) map[string]bool {
	connectedMap := make(map[string]bool, len(snList))
	for _, sn := range snList {
		connectedMap[sn] = false
	}
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Warnf("get node connection status failed: failed to create message, %v", err)
		return connectedMap
	}
	msg.SetRouter(common.NodeManagerName, common.CloudHubName, common.OptGet, common.ResEdgeConnStatus)
	if err = msg.FillContent(snList); err != nil {
		hwlog.RunLog.Warnf("fill content failed: %v", err)
		return connectedMap
	}
	resp, err := modulemgr.SendSyncMessage(msg, common.ResponseTimeout)
	if err != nil {
		hwlog.RunLog.Warnf("get node connection status failed: failed to send sync message, %v", err)
		return connectedMap
	}
	var response []websocketmgr.WebsocketPeerInfo
	if err = resp.ParseContent(&response); err != nil {
		hwlog.RunLog.Warnf("get node connection status failed: parse content failed %v", err)
		return connectedMap
	}
	for _, peerInfo := range response {
		connectedMap[peerInfo.Sn] = true
	}
	return connectedMap
}

func (s *nodeSyncImpl) GetAvailableResource(nodeID uint64, hostname string) (*NodeResource, error) {
	var (
		allocatedCpu, allocatedMemory, allocatedNpu resource.Quantity
		err                                         error
	)
	groups, err := NodeServiceInstance().getGroupsByNodeID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("get node all groups failed: %s", err.Error())
	}
	for _, group := range *groups {
		var res NodeResource
		if err := json.Unmarshal([]byte(group.ResourcesRequest), &res); err != nil {
			continue
		}
		allocatedCpu.Add(res.Cpu)
		allocatedMemory.Add(res.Memory)
		allocatedNpu.Add(res.Npu)
	}
	allocatableRes, err := s.GetAllocatableResource(hostname)
	if err != nil {
		return nil, fmt.Errorf("get node all allocatable resource failed: %s", err.Error())
	}
	allocatableRes.Cpu.Sub(allocatedCpu)
	allocatableRes.Memory.Sub(allocatedMemory)
	allocatableRes.Npu.Sub(allocatedNpu)
	return allocatableRes, nil
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

// report node change info to cert-updater
func reportChangedNodeInfo(action string, node *v1.Node) {
	nodeSn := node.Labels[snNodeLabelKey]
	// empty sn label means it's not a MEF managed node, skip report
	if nodeSn == "" {
		hwlog.RunLog.Infof("node [%v] is not managed by MEF, skip report change info", node.Name)
		return
	}
	nodeIp := evalIpAddress(node)
	var changedInfo changedNodeInfo
	nodeInfo := simpleNodeInfo{
		Sn: nodeSn,
		Ip: nodeIp,
	}
	switch action {
	case nodeActionDelete:
		changedInfo.DeletedNodeInfo = append(changedInfo.DeletedNodeInfo, nodeInfo)
	case nodeActionAdd:
		changedInfo.AddedNodeInfo = append(changedInfo.AddedNodeInfo, nodeInfo)
	default:
		hwlog.RunLog.Errorf("invalid node action: %v", action)
		return
	}
	reportMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("generate new report message error: %v", err)
		return
	}
	reportMsg.SetRouter(common.NodeManagerName, common.CertUpdaterName, common.OptPost, common.ResNodeChanged)
	if err = reportMsg.FillContent(changedInfo); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return
	}
	if err = modulemgr.SendMessage(reportMsg); err != nil {
		hwlog.RunLog.Errorf("report node change info error: %v changed info:%+v", err, changedInfo)
	}
}
