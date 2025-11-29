// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types for Secret
package types

// Secret [struct] to describe secret info
type Secret struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Immutable  *bool             `json:"immutable,omitempty" binding:"isdefault,omitempty"`
	Data       map[string][]byte `json:"data,omitempty" binding:"max=1,dive,keys,max=64,endkeys,max=2048"`
	StringData map[string]string `json:"stringData,omitempty" binding:"isdefault,omitempty"`
	Type       string            `json:"type,omitempty" binding:"omitempty,eq=kubernetes.io/dockerconfigjson"`
}
