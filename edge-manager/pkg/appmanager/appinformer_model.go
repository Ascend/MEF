// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager
package appmanager

import (
	"sync"

	"gorm.io/gorm"

	"edge-manager/pkg/database"
)

var (
	serviceSingletonInstance appInformerService
	informerSingleton        sync.Once
)

type appInformerServiceImp struct {
	db *gorm.DB
}

type appInformerService interface {
	addPod(obj *AppInstance) error
	updatePod(obj *AppInstance) error
	deletePod(obj *AppInstance) error
}

func kubeServiceInstance() appInformerService {
	informerSingleton.Do(func() {
		serviceSingletonInstance = &appInformerServiceImp{db: database.GetDb()}
	})
	return serviceSingletonInstance
}

func (k *appInformerServiceImp) addPod(obj *AppInstance) error {
	return k.db.Model(AppInstance{}).Create(obj).Error
}

func (k *appInformerServiceImp) updatePod(obj *AppInstance) error {
	return k.db.Model(AppInstance{}).UpdateColumns(obj).Error
}

func (k *appInformerServiceImp) deletePod(obj *AppInstance) error {
	return k.db.Model(AppInstance{}).Where("pod_name = ?", obj.PodName).Delete(obj).Error
}
