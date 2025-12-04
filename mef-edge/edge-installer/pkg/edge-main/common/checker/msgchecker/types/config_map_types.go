// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package types for config map
package types

// ConfigMap [struct] define config map struct
type ConfigMap struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Immutable  *bool             `json:"immutable,omitempty" binding:"isdefault,omitempty"`
	Data       map[string]string `json:"data,omitempty" binding:"max=256,dive,keys,max=64,endkeys,max=2048"`
	BinaryData map[string][]byte `json:"binaryData,omitempty" binding:"isdefault,omitempty"`
}
