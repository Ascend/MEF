// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
