// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"edge-manager/pkg/kubeclient"
	"errors"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	halfMin = time.Second * 30
)

var (
	nodeStatusService nodeStatusServiceImpl
)

// NodeStatusService provide node status from k8s
type NodeStatusService interface {
	// ListNodeStatus list node status by nodeName-nodeStatus map
	ListNodeStatus() map[string]string
	// GetNodeStatus get specific node status by hostname
	GetNodeStatus(uniqueName string) string
}

type nodeStatusServiceImpl struct {
	informer        cache.SharedIndexInformer
	clientSet       *kubernetes.Clientset
	nodeStatusCache map[string]string
	cacheLock       sync.RWMutex
}

// NodeStatusServiceInstance get NodeStatusService singleton
func NodeStatusServiceInstance() NodeStatusService {
	return &nodeStatusService
}

// initNodeStatusService init k8s informer
func initNodeStatusService() error {
	client := kubeclient.GetKubeClient().GetClientSet()
	nodeStatusService = nodeStatusServiceImpl{
		clientSet:       client,
		nodeStatusCache: make(map[string]string),
	}
	stopCh := make(chan struct{})
	nodeStatusService.initNodeInformer(stopCh)
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

func (s *nodeStatusServiceImpl) initNodeInformer(stopCh <-chan struct{}) {
	nodeInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		s.clientSet,
		halfMin,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {}),
	)
	s.informer = nodeInformerFactory.Core().V1().Nodes().Informer()
	s.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.addNode,
		UpdateFunc: s.updateNode,
		DeleteFunc: s.deleteNode,
	})
	nodeInformerFactory.Start(stopCh)
}

func (s *nodeStatusServiceImpl) addNode(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		hwlog.RunLog.Warn("Recovered add object,But can't convert to node")
		return
	}
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.nodeStatusCache[node.Name] = nodeStatus(node)
}

func (s *nodeStatusServiceImpl) updateNode(oldObj, newObj interface{}) {
	oldNode, ok := oldObj.(*v1.Node)
	if !ok {
		hwlog.RunLog.Warn("Recovered update object,But can't convert to node")
		return
	}
	newNode, ok := newObj.(*v1.Node)
	if !ok {
		hwlog.RunLog.Warn("Recovered update object,But can't convert to node")
		return
	}
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	delete(s.nodeStatusCache, oldNode.Name)
	s.nodeStatusCache[newNode.Name] = nodeStatus(newNode)
}

func (s *nodeStatusServiceImpl) deleteNode(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		hwlog.RunLog.Warn("Recovered delete object,But can't convert to node")
		return
	}
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	delete(s.nodeStatusCache, node.Name)
}

func (s *nodeStatusServiceImpl) GetNodeStatus(nodeName string) string {
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()

	if s.nodeStatusCache == nil {
		hwlog.RunLog.Warn("Get node status failed, service has not initialized yet")
		return statusOffline
	}
	status, ok := s.nodeStatusCache[nodeName]
	if !ok {
		return statusOffline
	}
	return status
}

func (s *nodeStatusServiceImpl) ListNodeStatus() map[string]string {
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()

	if s.nodeStatusCache == nil {
		hwlog.RunLog.Warn("List node status failed, service has not initialized yet")
		return map[string]string{}
	}
	return deepCopy(s.nodeStatusCache)
}

func nodeStatus(node *v1.Node) string {
	curNodeStatus := statusOffline
	for _, cond := range node.Status.Conditions {
		if cond.Type == v1.NodeReady {
			if cond.Status == v1.ConditionTrue {
				curNodeStatus = statusReady
			} else if cond.Status == v1.ConditionFalse {
				curNodeStatus = statusNotReady
			} else if cond.Status == v1.ConditionUnknown {
				curNodeStatus = statusUnknown
			}
		}
	}
	return curNodeStatus
}

func deepCopy(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
