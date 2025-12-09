// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package securitysetters
package securitysetters

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"edge-installer/pkg/common/types"
)

func createNewUpdateConfigmapFromFdConfigmap(origin *corev1.ConfigMap) *corev1.ConfigMap {
	var newConfigmap = &corev1.ConfigMap{}
	newConfigmap.ObjectMeta = *newObjectMeta(origin.ObjectMeta)
	newConfigmap.Data = origin.Data
	return newConfigmap
}

func createNewDeleteConfigmapFromFdConfigmap(origin *corev1.ConfigMap) *corev1.ConfigMap {
	var newConfigmap = &corev1.ConfigMap{}
	newConfigmap.ObjectMeta = *newObjectMeta(origin.ObjectMeta)
	return newConfigmap
}

func createNewUpdateSecretFromFdSecret(originalSecret *corev1.Secret) *corev1.Secret {
	var newSecret = &corev1.Secret{}
	newSecret.ObjectMeta = *newObjectMeta(originalSecret.ObjectMeta)
	newSecret.Data = originalSecret.Data
	newSecret.Type = originalSecret.Type

	return newSecret

}

// newObjectMeta [method] generate objectMeta for other resources
func newObjectMeta(originObjectMeta metav1.ObjectMeta) *metav1.ObjectMeta {
	newMeta := &metav1.ObjectMeta{
		Name:              originObjectMeta.Name,
		Namespace:         originObjectMeta.Namespace,
		UID:               originObjectMeta.UID,
		ResourceVersion:   originObjectMeta.ResourceVersion,
		CreationTimestamp: originObjectMeta.CreationTimestamp,
	}
	return newMeta
}

func modifyDelete(updateInfo *types.ModelFileInfo) {
	for index := range updateInfo.ModelFiles {
		updateInfo.ModelFiles[index].CheckType = ""
		updateInfo.ModelFiles[index].CheckCode = ""
		updateInfo.ModelFiles[index].Size = ""
		updateInfo.ModelFiles[index].FileServer = types.FileServerInfo{}
	}
}
