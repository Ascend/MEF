// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
