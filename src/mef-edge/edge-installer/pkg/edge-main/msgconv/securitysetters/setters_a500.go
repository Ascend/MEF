// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package securitysetters
package securitysetters

import (
	"errors"

	"k8s.io/api/core/v1"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

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
