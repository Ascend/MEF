// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package msgconv
package msgconv

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// PodResp represent pod struct
type PodResp struct {
	// the content
	Object *v1.Pod `json:"Object"`
	// error info
	Err errors.StatusError `json:"Err"`
}

// NodeResp represent node struct
type NodeResp struct {
	// the content
	Object *v1.Node `json:"Object"`
	// error info
	Err errors.StatusError `json:"Err"`
}
