// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager
package appmanager

import (
	"errors"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// Informer is used to sync pod info
type Informer struct {
	PodInformer cache.SharedIndexInformer
}

var appInformer Informer

// InitInformer init informer of certain k8s
func (ai *Informer) InitInformer(client *kubernetes.Clientset, stopCh <-chan struct{}) error {
	hwlog.RunLog.Info("start to init informer for app manager")
	ai.initPodInformer(client, stopCh)
	if err := ai.run(stopCh); err != nil {
		hwlog.RunLog.Error("sync informer cache failed")
		return err
	}
	return nil
}

func (ai *Informer) initPodInformer(client *kubernetes.Clientset, stopCh <-chan struct{}) {
	podLabelSelector := labels.Set(map[string]string{common.AppManagerName: AppLabel}).
		AsSelector().String()
	podInformerFactory := informers.NewSharedInformerFactoryWithOptions(client, informerSyncInterval,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = podLabelSelector
		}))
	appInformer.PodInformer = podInformerFactory.Core().V1().Pods().Informer()
	ai.PodInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addPod,
		UpdateFunc: updatePod,
		DeleteFunc: deletePod,
	})
	podInformerFactory.Start(stopCh)
}

func (ai *Informer) run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	hwlog.RunLog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, ai.PodInformer.HasSynced); !ok {
		hwlog.RunLog.Error("failed to wait for caches to sync ")
		return errors.New("sync informer caches error")
	}
	return nil
}

func addPod(obj interface{}) {
	appInstance, err := parsePod(obj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered add object, parse pod error: %v", err)
		return
	}
	if err = kubeServiceInstance().addPod(appInstance); err != nil {
		hwlog.RunLog.Errorf("recovered add object, add instance to db error: %v", err)
	}
	return
}

// updatePod do nothing to the app instance
func updatePod(_, _ interface{}) {
	return
}

func deletePod(obj interface{}) {
	appInstance, err := parsePod(obj)
	if err != nil {
		hwlog.RunLog.Errorf("recovered delete object, parse pod error: %v", err)
		return
	}
	if err = kubeServiceInstance().deletePod(&AppInstance{PodName: appInstance.PodName}); err != nil {
		hwlog.RunLog.Errorf("recovered delete object, delete instance from db error: %v", err)
	}
	return
}

func parsePod(obj interface{}) (*AppInstance, error) {
	eventPod, ok := obj.(*v1.Pod)
	if !ok {
		hwlog.RunLog.Error("recovered add object, but can't convert to pod")
		return nil, errors.New("convert to pod error")
	}

	var nodeGroupName string
	var nodeGroupId int
	podLabels := eventPod.Labels
	nodeSelector := eventPod.Spec.NodeSelector
	appName := podLabels[AppName]
	value, ok := podLabels[AppId]
	if !ok {
		hwlog.RunLog.Error("assert pod label error")
		return nil, errors.New("app id do not exist")
	}
	appId, err := strconv.Atoi(value)
	if err != nil {
		hwlog.RunLog.Error("assert pod app id label error")
		return nil, err

	}
	for labelKey, value := range nodeSelector {
		if !strings.HasPrefix(labelKey, common.NodeGroupLabelPrefix) {
			continue
		}
		nodeGroupName = value
		nodeGroupId, err = strconv.Atoi(strings.TrimPrefix(labelKey, common.NodeGroupLabelPrefix))
		if err != nil {
			hwlog.RunLog.Error("assert pod node group id label error")
			return nil, err
		}
	}

	newAppInstance := AppInstance{
		PodName:       eventPod.Name,
		NodeName:      eventPod.Spec.NodeName,
		NodeGroupName: nodeGroupName,
		NodeGroupID:   int64(nodeGroupId),
		Status:        string(eventPod.Status.Phase),
		AppName:       appName,
		AppID:         int64(appId),
		CreatedAt:     time.Now().Format(common.TimeFormat),
		ChangedAt:     time.Now().Format(common.TimeFormat),
	}
	return &newAppInstance, nil
}
