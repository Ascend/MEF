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

	"huawei.com/mindx/common/hwlog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"edge-manager/pkg/kubeclient"

	"huawei.com/mindxedge/base/common"
)

const (
	maxPodNum      = 21000
	namespaceField = "metadata.namespace"
)

type appStatusServiceImpl struct {
	podInformer          cache.SharedIndexInformer
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
	if err := initNamespace(); err != nil {
		hwlog.RunLog.Errorf("init user namespace failed: %s", err.Error())
		return err
	}
	a.initInformer(clientSet, stopCh)
	if err := a.run(stopCh); err != nil {
		hwlog.RunLog.Error("sync app status service cache failed")
		return err
	}
	if err := initDefaultImagePullSecret(); err != nil {
		hwlog.RunLog.Errorf("create default image pull secret failed: %s", err.Error())
		return err
	}
	return nil
}

func initNamespace() error {
	userNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.MefUserNs,
			Namespace: common.MefUserNs,
		},
	}
	_, err := kubeclient.GetKubeClient().CreateNamespace(userNs)
	if err != nil {
		return fmt.Errorf("create default user namespace failed: %s", err.Error())
	}
	return nil
}

func initDefaultImagePullSecret() error {
	secret, err := kubeclient.GetKubeClient().GetSecret(kubeclient.DefaultImagePullSecretKey)
	if err == nil {
		secretData, ok := secret.Data[corev1.DockerConfigJsonKey]
		if !ok {
			return nil
		}
		common.ClearSliceByteMemory(secretData)
		return nil
	}
	if !strings.Contains(err.Error(), kubeclient.K8sNotFoundErrorFragment) {
		return fmt.Errorf("check image pull secret fialed")
	}
	defaultImagePullSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeclient.DefaultImagePullSecretKey,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{corev1.DockerConfigJsonKey: []byte(kubeclient.DefaultImagePullSecretValue)},
	}
	if _, err := kubeclient.GetKubeClient().CreateOrUpdateSecret(defaultImagePullSecret); err != nil {
		return fmt.Errorf("create default image pull secret fialed")
	}
	return nil
}

func (a *appStatusServiceImpl) initInformer(client *kubernetes.Clientset, stopCh <-chan struct{}) {
	a.podStatusCache = make(map[string]string)
	a.containerStatusCache = make(map[string]containerStatus)
	labelSelector := labels.Set{common.AppManagerName: AppLabel}.AsSelector().String()
	fieldSelector := fields.Set{namespaceField: common.MefUserNs}.AsSelector().String()

	informerFactory := informers.NewSharedInformerFactoryWithOptions(client, informerSyncInterval,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector
			options.FieldSelector = fieldSelector
		}))
	appStatusService.podInformer = informerFactory.Core().V1().Pods().Informer()
	a.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    a.addPod,
		UpdateFunc: a.updatePod,
		DeleteFunc: a.deletePod,
	})
	informerFactory.Start(stopCh)
}

func (a *appStatusServiceImpl) run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	if err := AppRepositoryInstance().deleteAllRemainingInstance(); err != nil {
		hwlog.RunLog.Error("failed to delete remaining app instance before sync caches")
		return err
	}
	hwlog.RunLog.Info("Waiting for app status service caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, a.podInformer.HasSynced); !ok {
		hwlog.RunLog.Error("failed to wait for caches to sync ")
		return errors.New("sync app status service pod caches error")
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
	count, err := GetTableCount(AppInstance{})
	if err != nil {
		hwlog.RunLog.Errorf("get pod table count error:%v", err)
		return
	}
	if count >= maxPodNum {
		hwlog.RunLog.Errorf("pod count cannot exceed %d, please delete no need pod", maxPodNum)
		return
	}

	appInstance, err := parsePodToInstance(pod)
	if err != nil {
		hwlog.RunLog.Errorf("recovered add pod, parse pod to app instance error: %v", err)
		return
	}

	appStatusService.updateStatusCache(pod)

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
	appInstance, err := parsePodToInstance(newPod)
	if err != nil {
		hwlog.RunLog.Errorf("recovered update object, parse pod to app instance error: %v", err)
		return
	}

	appStatusService.updateStatusCache(newPod)
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
