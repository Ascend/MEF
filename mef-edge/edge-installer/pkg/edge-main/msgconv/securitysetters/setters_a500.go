// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_A500

// Package securitysetters
package securitysetters

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

// SetPodUpdate sets pod update
func SetPodUpdate(_ *model.Message, content interface{}) error {
	pod, ok := content.(*v1.Pod)
	if !ok {
		return errors.New("invalid content type")
	}
	if pod == nil {
		return errors.New("nil pointer content is not allowed")
	}

	modifyPodStatusAndTypeMeta(pod)

	if isPodGraceDelete(pod.DeletionTimestamp) {
		modifyPodMetaDataGraceDeletionPara(pod)
		modifyPodSpecGraceDeletionPara(pod)
		return nil
	}

	modifyPodMetaDataCreatePara(pod)
	modifyPodSpecCreatePara(pod)
	return nil
}

// SetConfigmapUpdate sets configmap update
func SetConfigmapUpdate(_ *model.Message, content interface{}) error {
	configmap, ok := content.(*v1.ConfigMap)
	if !ok {
		return errors.New("invalid content type")
	}
	if configmap == nil {
		return errors.New("nil pointer content is not allowed")
	}

	*configmap = *createNewUpdateConfigmapFromFdConfigmap(configmap)
	return nil
}

// SetModelFileUpdate set model file delete
func SetModelFileUpdate(_ *model.Message, content interface{}) error {
	modelFileInfo, ok := content.(*types.ModelFileInfo)
	if !ok {
		return errors.New("invalid content type")
	}
	if modelFileInfo == nil {
		return errors.New("nil pointer content is not allowed")
	}

	if modelFileInfo.Operation != constants.OptDelete {
		hwlog.RunLog.Infof("no need to modify model msg for operation: %s", modelFileInfo.Operation)
		return nil
	}

	modifyDelete(modelFileInfo)
	return nil
}

// SetSecretUpdate set secret update
func SetSecretUpdate(_ *model.Message, content interface{}) error {
	secret, ok := content.(*v1.Secret)
	if !ok {
		return errors.New("invalid content type")
	}
	if secret == nil {
		return errors.New("nil pointer content is not allowed")
	}

	*secret = *createNewUpdateSecretFromFdSecret(secret)
	return nil
}

// SetPodDelete set secret delete
func SetPodDelete(_ *model.Message, content interface{}) error {
	pod, ok := content.(*v1.Pod)
	if !ok {
		return errors.New("invalid content type")
	}
	if pod == nil {
		return errors.New("nil pointer content is not allowed")
	}

	modifyPodStatusAndTypeMeta(pod)

	modifyPodMetaDataDeletePara(pod)
	modifyPodSpecDeletePara(pod)
	return nil
}

// SetConfigmapDelete set configmap delete
func SetConfigmapDelete(_ *model.Message, content interface{}) error {
	configmap, ok := content.(*v1.ConfigMap)
	if !ok {
		return errors.New("invalid content type")
	}
	if configmap == nil {
		return errors.New("nil pointer content is not allowed")
	}

	*configmap = *createNewDeleteConfigmapFromFdConfigmap(configmap)
	return nil
}
