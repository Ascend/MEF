// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package types some common struct which all modules can use
package types

// ModelFileInfo msg struct from FD
type ModelFileInfo struct {
	Operation  string      `json:"operation,omitempty"`
	Target     string      `json:"target,omitempty"`
	Uuid       string      `json:"uuid,omitempty"`
	ModelFiles []ModelFile `json:"modelfiles,omitempty"`
}

// ModelFile msg struct from FD
type ModelFile struct {
	Name       string         `json:"name,omitempty"`
	Version    string         `json:"version,omitempty"`
	CheckType  string         `json:"check_type,omitempty"`
	CheckCode  string         `json:"check_code,omitempty"`
	Size       string         `json:"size,omitempty"`
	FileServer FileServerInfo `json:"file_server,omitempty"`
}

// ModelBrief struct for model file brief info
type ModelBrief struct {
	Uuid   string `json:"uuid,omitempty"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}

// FileServerInfo msg struct from FD
type FileServerInfo struct {
	Protocol string `json:"protocol,omitempty"`
	Path     string `json:"path,omitempty"`
	UserName string `json:"user_name,omitempty"`
	PassWord string `json:"password,omitempty"`
}

// ModelFileEffectInfo model file effect info struct
type ModelFileEffectInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	ActiveType string `json:"active_type"`
}

// ContainerInfo container info struct
type ContainerInfo struct {
	ModelFile []ModelFileEffectInfo `json:"modelfile"`
}

// UpdateContainerInfo update container info struct
type UpdateContainerInfo struct {
	Operation string          `json:"operation"`
	Source    string          `json:"source"`
	PodName   string          `json:"pod_name"`
	PodUid    string          `json:"pod_uid"`
	Uuid      string          `json:"uuid"`
	Container []ContainerInfo `json:"container"`
}

// OperateModelFileContent operate model file content struct
type OperateModelFileContent struct {
	Operate     string
	OperateInfo map[string]string
	UsedFiles   []string
	CurrentUuid string
}

// Compare compare two model file is same
func (m ModelFile) Compare(target ModelFile) bool {
	return m.Version == target.Version &&
		m.CheckCode == target.CheckCode && m.CheckType == target.CheckType
}

// ModelStatusType the type of model status
type ModelStatusType string

// the actual type for status
const (
	StatusDownloading ModelStatusType = "downloading"
	StatusInactive    ModelStatusType = "inactive"
	StatusActive      ModelStatusType = "active"
	StatusFail        ModelStatusType = "fail"
)

func (m ModelStatusType) String() string {
	return string(m)
}
