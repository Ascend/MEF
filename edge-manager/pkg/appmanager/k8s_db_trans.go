// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appmanager conversion between k8s and db
package appmanager

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertCmToK8S(configmapReq *ConfigmapReq) *v1.ConfigMap {
	configmapData := convertCmContentToMap(configmapReq)
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configmapReq.ConfigmapName,
		},
		Data: configmapData,
	}
}

func convertCmContentToMap(configmapReq *ConfigmapReq) map[string]string {
	configmapData := make(map[string]string)
	for idx := range configmapReq.ConfigmapContent {
		configmapContent := configmapReq.ConfigmapContent[idx]
		configmapData[configmapContent.Name] = configmapContent.Value
	}

	return configmapData
}
