// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/kubeclient"
)

type appStatusServiceImpl struct {
	podInformer          cache.SharedIndexInformer
	daemonSetInformer    cache.SharedIndexInformer
	podStatusCache       map[string]string
	containerStatusCache map[string]containerStatus
	appStatusCacheLock   sync.RWMutex
}

type containerStatus struct {
	Status       string
	RestartCount int32
}

var appStatusService appStatusServiceImpl

func (a *appStatusServiceImpl) initAppStatusService() error {
	hwlog.RunLog.Info("start to init app status service for app manager")
	stopCh := make(chan struct{})
	clientSet := kubeclient.GetKubeClient().GetClientSet()
	if clientSet == nil {
		hwlog.RunLog.Error("init app status service failed, get k8s client failed")
		return errors.New("get k8s client set failed")
	}
	a.initInformer(clientSet, stopCh)
	if err := a.run(stopCh); err != nil {
		hwlog.RunLog.Error("sync app status service cache failed")
		return err
	}
	return nil
}

func (a *appStatusServiceImpl) initInformer(client *kubernetes.Clientset, stopCh <-chan struct{}) {
	a.podStatusCache = make(map[string]string)
	a.containerStatusCache = make(map[string]containerStatus)
	LabelSelector := labels.Set(map[string]string{common.AppManagerName: AppLabel}).
		AsSelector().String()
	informerFactory := informers.NewSharedInformerFactoryWithOptions(client, informerSyncInterval,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = LabelSelector
		}))
	appStatusService.podInformer = informerFactory.Core().V1().Pods().Informer()
	appStatusService.daemonSetInformer = informerFactory.Apps().V1().DaemonSets().Informer()
	a.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    a.addPod,
		UpdateFunc: a.updatePod,
		DeleteFunc: a.deletePod,
	})
	a.daemonSetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    a.addDaemonSet,
		UpdateFunc: a.updateDaemonSet,
		DeleteFunc: a.deleteDaemonSet,
	})
	informerFactory.Start(stopCh)
}

func (a *appStatusServiceImpl) run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	if err := AppRepositoryInstance().deleteAllRemainingInstance(); err != nil {
		hwlog.RunLog.Error("failed to delete remaining app instance before sync caches")
		return err
	}
	if err := AppRepositoryInstance().deleteAllRemainingDaemonSet(); err != nil {
		hwlog.RunLog.Error("failed to delete remaining daemon set instance before sync caches")
		return err
	}
	hwlog.RunLog.Info("Waiting for app status service caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, a.podInformer.HasSynced); !ok {
		hwlog.RunLog.Error("failed to wait for caches to sync ")
		return errors.New("sync app status service pod caches error")
	}
	if ok := cache.WaitForCacheSync(stopCh, a.daemonSetInformer.HasSynced); !ok {
		hwlog.RunLog.Error("failed to wait for caches to sync ")
		return errors.New("sync app status service daemon set caches error")
	}
	return nil
}

func (a *appStatusServiceImpl) updateStatusCache(pod *corev1.Pod) {
	a.appStatusCacheLock.Lock()
	defer a.appStatusCacheLock.Unlock()
	a.podStatusCache[pod.Name] = strings.ToLower(string(pod.Status.Phase))
	for _, cStatus := range pod.Status.ContainerStatuses {
		containerStatusKey := pod.Name + "-" + cStatus.Name
		a.containerStatusCache[containerStatusKey] = containerStatus{
			Status:       getContainerStatus(cStatus),
			RestartCount: cStatus.RestartCount,
		}
	}
}

func (a *appStatusServiceImpl) deleteStatusCache(pod *corev1.Pod) {
	a.appStatusCacheLock.Lock()
	defer a.appStatusCacheLock.Unlock()
	delete(a.podStatusCache, pod.Name)
	for _, cStatus := range pod.Status.ContainerStatuses {
		containerStatusKey := pod.Name + "-" + cStatus.Name
		delete(a.containerStatusCache, containerStatusKey)
	}
}

func (a *appStatusServiceImpl) addPod(obj interface{}) {
	pod, err := parsePod(obj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered add object, parse pod error: %v", err)
		return
	}
	appStatusService.updateStatusCache(pod)
	appInstance, err := parsePodToInstance(pod)
	if err != nil {
		hwlog.RunLog.Errorf("recovered add pod, parse pod to app instance error: %v", err)
		return
	}
	if err = AppRepositoryInstance().addPod(appInstance); err != nil {
		hwlog.RunLog.Errorf("recovered add object, add instance to db error: %v", err)
		return
	}
}

func (a *appStatusServiceImpl) updatePod(oldObj, newObj interface{}) {
	oldPod, err := parsePod(oldObj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered update object, parse old pod error: %v", err)
		return
	}
	appStatusService.deleteStatusCache(oldPod)

	newPod, err := parsePod(newObj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered update object, parse new pod error: %v", err)
		return
	}
	appStatusService.updateStatusCache(newPod)
	appInstance, err := parsePodToInstance(newPod)
	if err != nil {
		hwlog.RunLog.Errorf("recovered update object, parse pod to app instance error: %v", err)
		return
	}
	if err = AppRepositoryInstance().updatePod(appInstance); err != nil {
		hwlog.RunLog.Errorf("recovered update object, update instance to db error: %v", err)
		return
	}
}

func (a *appStatusServiceImpl) deletePod(obj interface{}) {
	pod, err := parsePod(obj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered delete object, parse pod error: %v", err)
		return
	}
	appStatusService.deleteStatusCache(pod)
	if err = AppRepositoryInstance().deletePod(&AppInstance{PodName: pod.Name}); err != nil {
		hwlog.RunLog.Errorf("recovered delete object, delete instance from db error: %v", err)
		return
	}
}

func (a *appStatusServiceImpl) deleteTerminatingPod() {
	list := a.podInformer.GetStore().List()
	for _, podContent := range list {
		pod, ok := podContent.(*corev1.Pod)
		if !ok {
			hwlog.RunLog.Error("convert pod error")
			continue
		}
		if pod.ObjectMeta.DeletionTimestamp == nil {
			continue
		}
		hwlog.RunLog.Infof("find timeout terminating pod, start to delete pod: %v", pod.Name)
		if err := kubeclient.GetKubeClient().DeletePodByForce(pod); err != nil {
			hwlog.RunLog.Errorf("remove timeout terminating pod error: %v", err)
			continue
		}
	}
}

func (a *appStatusServiceImpl) getContainerInfos(instance AppInstance, nodeStatus string) ([]ContainerInfo, error) {
	var containerInfos []ContainerInfo
	if err := json.Unmarshal([]byte(instance.ContainerInfo), &containerInfos); err != nil {
		return nil, errors.New("unmarshal app container info failed")
	}
	a.appStatusCacheLock.Lock()
	defer a.appStatusCacheLock.Unlock()
	for i := range containerInfos {
		containerStatusKey := instance.PodName + "-" + containerInfos[i].Name
		cStatus, ok := appStatusService.containerStatusCache[containerStatusKey]
		containerInfos[i].Status = cStatus.Status
		containerInfos[i].RestartCount = cStatus.RestartCount
		if !ok || nodeStatus != nodeStatusReady {
			containerInfos[i].Status = containerStateUnknown
		}
	}
	return containerInfos, nil
}

func (a *appStatusServiceImpl) getPodStatusFromCache(appName, nodeStatus string) string {
	a.appStatusCacheLock.Lock()
	defer a.appStatusCacheLock.Unlock()
	podStatus, ok := appStatusService.podStatusCache[appName]
	if !ok || nodeStatus != nodeStatusReady {
		podStatus = podStatusUnknown
	}
	return podStatus
}

func parsePod(obj interface{}) (*corev1.Pod, error) {
	eventPod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil, errors.New("convert object to pod error")
	}
	return eventPod, nil
}

func parsePodToInstance(eventPod *corev1.Pod) (*AppInstance, error) {
	appName, appId, err := getAppNameAndId(eventPod)
	if err != nil {
		return nil, fmt.Errorf("get app name or id error, %v", err)
	}
	nodeId, nodeName, err := getNodeInfoByUniqueName(eventPod)
	if err != nil {
		return nil, fmt.Errorf("get node info error, %v", err)
	}
	nodeGroupId, err := getNodeGroupId(eventPod)
	if err != nil {
		return nil, fmt.Errorf("get group info error, %v", err)
	}
	containerInfos, err := getContainerInfoString(eventPod)
	if err != nil {
		return nil, fmt.Errorf("get container info error, %v", err)
	}

	newAppInstance := AppInstance{
		PodName:        eventPod.Name,
		NodeName:       nodeName,
		NodeUniqueName: eventPod.Spec.NodeName,
		NodeID:         nodeId,
		NodeGroupID:    nodeGroupId,
		AppName:        appName,
		AppID:          appId,
		ContainerInfo:  containerInfos,
	}
	return &newAppInstance, nil
}

func getAppNameAndId(eventPod *corev1.Pod) (string, uint64, error) {
	podLabels := eventPod.Labels
	if podLabels == nil {
		return "", 0, errors.New("node selector is nil")
	}
	appName, ok := podLabels[AppName]
	if !ok {
		return "", 0, errors.New("app name label do not exist")
	}
	value, ok := podLabels[AppId]
	if !ok {
		return "", 0, errors.New("app id label do not exist")
	}
	appId, err := strconv.ParseUint(value, common.BaseHex, common.BitSize64)
	if err != nil {
		return "", 0, err
	}
	return appName, appId, nil
}

func getNodeGroupId(eventPod *corev1.Pod) (uint64, error) {
	nodeSelector := eventPod.Spec.NodeSelector
	if nodeSelector == nil {
		return 0, errors.New("node selector is nil")
	}
	var nodeGroupId uint64
	var err error
	for labelKey := range nodeSelector {
		if !strings.HasPrefix(labelKey, common.NodeGroupLabelPrefix) {
			continue
		}
		nodeGroupId, err = strconv.ParseUint(strings.TrimPrefix(labelKey, common.NodeGroupLabelPrefix),
			common.BaseHex, common.BitSize64)
		if err != nil {
			return 0, err
		}
	}
	return nodeGroupId, nil
}

func getContainerInfoString(eventPod *corev1.Pod) (string, error) {
	var containerInfos []ContainerInfo
	for _, container := range eventPod.Spec.Containers {
		containerInfo := ContainerInfo{
			Name:   container.Name,
			Image:  container.Image,
			Status: "",
		}
		containerInfos = append(containerInfos, containerInfo)
	}
	containerInfosDate, err := json.Marshal(containerInfos)
	if err != nil {
		hwlog.RunLog.Error("marshal container info error")
		return "", err
	}
	return string(containerInfosDate), nil
}

func getContainerStatus(containerStatus corev1.ContainerStatus) string {
	if containerStatus.State.Waiting != nil {
		return containerStateWaiting
	}
	if containerStatus.State.Running != nil {
		return containerStateRunning
	}
	if containerStatus.State.Terminated != nil {
		return containerStateTerminated
	}
	return containerStateUnknown
}

func (a *appStatusServiceImpl) addDaemonSet(obj interface{}) {
	daemonSet, err := parseDaemonSet(obj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered add object, parse daemon set error: %v", err)
		return
	}
	appDaemonSet, err := parseDaemonSetToDB(daemonSet)
	if err != nil {
		hwlog.RunLog.Errorf("recovered add daemon set, generate app daemon set error: %v", err)
		return
	}
	if err = updateAllocatedNodeRes(daemonSet, appDaemonSet.NodeGroupID, false); err != nil {
		hwlog.RunLog.Errorf("recovered add daemon set, update allocated node resource error: %v", err)
		return
	}
	if err = AppRepositoryInstance().addDaemonSet(appDaemonSet); err != nil {
		hwlog.RunLog.Error("recovered add object, add daemon set to db error")
		return
	}
}

func (a *appStatusServiceImpl) updateDaemonSet(_, newObj interface{}) {
	daemonSet, err := parseDaemonSet(newObj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered update object, parse daemon set error: %v", err)
		return
	}
	appDaemonSet, err := parseDaemonSetToDB(daemonSet)
	if err != nil {
		hwlog.RunLog.Errorf("recovered update daemon set, generate app daemon set error: %v", err)
		return
	}
	if err = AppRepositoryInstance().updateDaemonSet(appDaemonSet); err != nil {
		hwlog.RunLog.Error("recovered update object, update daemon set to db error")
		return
	}
}

func (a *appStatusServiceImpl) deleteDaemonSet(obj interface{}) {
	daemonSet, err := parseDaemonSet(obj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered delete object, parse daemon set error: %v", err)
		return
	}
	nodeGroupID, err := getNodeGroupIdFromDaemonSet(daemonSet)
	if err != nil {
		hwlog.RunLog.Errorf("recovered delete daemon set, get node group id error, %v", err)
		return
	}
	if err = updateAllocatedNodeRes(daemonSet, nodeGroupID, true); err != nil {
		hwlog.RunLog.Errorf("recovered delete daemon set, update allocated node resource error: %v", err)
		return
	}
	if err = AppRepositoryInstance().deleteDaemonSet(daemonSet.Name); err != nil {
		hwlog.RunLog.Error("recovered add object, add daemon set to db error")
		return
	}
}

func parseDaemonSet(obj interface{}) (*appv1.DaemonSet, error) {
	set, ok := obj.(*appv1.DaemonSet)
	if !ok {
		return nil, errors.New("convert object to daemon set error")
	}
	return set, nil
}

func parseDaemonSetToDB(eventSet *appv1.DaemonSet) (*AppDaemonSet, error) {
	appId, err := getAppIdFromDaemonSet(eventSet)
	if err != nil {
		return nil, fmt.Errorf("get app id error, %v", err)
	}
	nodeGroupId, err := getNodeGroupIdFromDaemonSet(eventSet)
	if err != nil {
		return nil, fmt.Errorf("get node group id error, %v", err)
	}
	nodeGroupInfos, err := getNodeGroupInfos([]uint64{nodeGroupId})
	if err != nil {
		return nil, fmt.Errorf("get group name or id error, %v", err)
	}
	if len(nodeGroupInfos) != 1 {
		return nil, errors.New("get group name or id nums error")
	}
	nodeGroupInfo := nodeGroupInfos[0]
	set := AppDaemonSet{
		DaemonSetName: eventSet.Name,
		AppID:         appId,
		NodeGroupID:   nodeGroupInfo.NodeGroupID,
		NodeGroupName: nodeGroupInfo.NodeGroupName,
	}
	return &set, nil
}

func getNodeGroupIdFromDaemonSet(eventSet *appv1.DaemonSet) (uint64, error) {
	if eventSet == nil {
		return 0, errors.New("event daemon set is nil")
	}
	nodeSelector := eventSet.Spec.Template.Spec.NodeSelector
	if nodeSelector == nil {
		return 0, errors.New("node selector is nil")
	}
	var nodeGroupId uint64
	var err error
	for labelKey := range nodeSelector {
		if !strings.HasPrefix(labelKey, common.NodeGroupLabelPrefix) {
			continue
		}
		nodeGroupId, err = strconv.ParseUint(strings.TrimPrefix(labelKey, common.NodeGroupLabelPrefix),
			common.BaseHex, common.BitSize64)
		if err != nil {
			return 0, err
		}
	}
	return nodeGroupId, nil
}

func getAppIdFromDaemonSet(eventSet *appv1.DaemonSet) (uint64, error) {
	podLabels := eventSet.Spec.Template.Labels
	value, ok := podLabels[AppId]
	if !ok {
		return 0, errors.New("app id label do not exist")
	}
	appId, err := strconv.ParseUint(value, common.BaseHex, common.BaseHex)
	if err != nil {
		return 0, err
	}
	return appId, nil
}
