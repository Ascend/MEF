// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package model provide common struct
package model

import (
	"edge-manager/pkg/common"
	"edge-manager/pkg/database"
	"encoding/json"
	"time"
)

// TemplateGroupModel app template group model
type TemplateGroupModel struct {
	Id          uint64
	Name        string
	Description string
	CreatedAt   string
	ModifiedAt  string
}

// TemplateVersionModel app template version model
type TemplateVersionModel struct {
	GroupId     uint64
	VersionName string
	CreatedAt   string
	ModifiedAt  string
	Containers  []ContainerModel
}

// ContainerModel container config model
type ContainerModel struct {
	Id             uint64
	ContainerName  string
	ImageName      string
	ImageVersion   string
	CpuRequest     string
	CpuLimit       string
	MemRequest     string
	MemLimit       string
	Npu            string
	Env            []Dic
	ContainerUser  string
	ContainerGroup string
	PortMaps       []PortMap
	Command        []string
}

// AppTemplateGroupDto app template group
type AppTemplateGroupDto struct {
	Id          uint64 `json:"Id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TemplateGroupSummaryDto app template group summary
type TemplateGroupSummaryDto struct {
	GroupId          uint64 `json:"group_id"`
	GroupName        string `json:"group_name"`
	Description      string `json:"description"`
	CreateTime       string `json:"create_time"`
	LastModifyTime   string `json:"last_modify_time"`
	SubVersionNumber int    `json:"sub_version_number"`
}

// AppTemplateVersionDto app template version
type AppTemplateVersionDto struct {
	GroupId    uint64         `json:"group_id"`
	Version    string         `json:"version"`
	Containers []ContainerDto `json:"containers"`
}

// VersionSummaryDto app template version summary
type VersionSummaryDto struct {
	GroupId    uint64 `json:"group_id"`
	Version    string `json:"version"`
	CreateTime string `json:"create_time"`
}

// VersionDetailDto app template version detail
type VersionDetailDto struct {
	GroupName      string `json:"group_name"`
	CreateTime     string `json:"create_time"`
	LastModifyTime string `json:"last_modify_time"`
	AppTemplateVersionDto
}

// ContainerDto item of app template version containers
type ContainerDto struct {
	Id             uint64    `json:"id"`
	ContainerName  string    `json:"container_name"`
	ImageName      string    `json:"image_name"`
	ImageVersion   string    `json:"image_version"`
	CpuRequest     string    `json:"cpu_request"`
	CpuLimit       string    `json:"cpu_limit"`
	MemRequest     string    `json:"mem_request"`
	MemLimit       string    `json:"mem_limit"`
	Npu            string    `json:"npu"`
	Env            []Dic     `json:"env"`
	ContainerUser  string    `json:"container_user"`
	ContainerGroup string    `json:"container_group"`
	PortMaps       []PortMap `json:"port_maps"`
	Command        []string  `json:"command"`
}

// Dic key-value dictionary
type Dic struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PortMap container port mapping config
type PortMap struct {
	Protocol      string `json:"protocol"`
	ContainerPort string `json:"container_port"`
	HostIp        string `json:"host_ip"`
	HostPort      string `json:"host_port"`
}

// Check whether the app template group is valid
func (dto *AppTemplateGroupDto) Check() error {
	return nil
}

// Check whether the app template version is valid
func (dto *AppTemplateVersionDto) Check() error {
	return nil
}

// ToDb convert app template group model to db model
func (model *TemplateGroupModel) ToDb() *database.AppTemplateGroupDb {
	return &database.AppTemplateGroupDb{
		Id:          model.Id,
		Name:        model.Name,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		ModifiedAt:  model.ModifiedAt,
	}
}

// FromDb convert db model to app template group model
func (model *TemplateGroupModel) FromDb(data *database.AppTemplateGroupDb) {
	if data == nil {
		return
	}
	model.Id = data.Id
	model.Name = data.Name
	model.Description = data.Description
	model.CreatedAt = data.CreatedAt
	model.ModifiedAt = data.ModifiedAt
}

// ToDb convert app template version model to db model
func (model *TemplateVersionModel) ToDb() ([]database.AppContainerTemplateDb, error) {
	result := make([]database.AppContainerTemplateDb, len(model.Containers))
	for i, container := range model.Containers {
		env, err := json.Marshal(container.Env)
		if err != nil {
			return nil, err
		}
		var command []byte
		command, err = json.Marshal(container.Command)
		if err != nil {
			return nil, err
		}
		var portMaps []byte
		portMaps, err = json.Marshal(container.PortMaps)
		if err != nil {
			return nil, err
		}
		result[i] = database.AppContainerTemplateDb{
			Id:             container.Id,
			GroupId:        model.GroupId,
			VersionName:    model.VersionName,
			ContainerName:  container.ContainerName,
			ImageNme:       container.ImageName,
			ImageVersion:   container.ImageVersion,
			CpuRequest:     container.CpuRequest,
			CpuLimit:       container.CpuLimit,
			MemoryRequest:  container.MemRequest,
			MemoryLimit:    container.MemLimit,
			Npu:            container.Npu,
			Env:            string(env),
			ContainerUser:  container.ContainerUser,
			ContainerGroup: container.ContainerGroup,
			PortMaps:       string(portMaps),
			Command:        string(command),
			CreatedAt:      model.CreatedAt,
			ModifiedAt:     model.ModifiedAt,
		}
	}
	return result, nil
}

// FromDb convert db model to app template version model
func (model *TemplateVersionModel) FromDb(data []database.AppContainerTemplateDb) error {
	if len(data) == 0 {
		return nil
	}
	model.GroupId = data[0].GroupId
	model.VersionName = data[0].VersionName
	model.CreatedAt = data[0].CreatedAt
	model.ModifiedAt = data[0].ModifiedAt
	model.Containers = make([]ContainerModel, len(data))
	for i, container := range data {
		var env []Dic
		if container.Env != "" {
			if err := json.Unmarshal([]byte(container.Env), &env); err != nil {
				return err
			}
		}
		var command []string
		if container.Command != "" {
			if err := json.Unmarshal([]byte(container.Command), &command); err != nil {
				return err
			}
		}
		var portMaps []PortMap
		if container.PortMaps != "" {
			if err := json.Unmarshal([]byte(container.PortMaps), &portMaps); err != nil {
				return err
			}
		}
		model.Containers[i] = ContainerModel{
			Id:             container.Id,
			ContainerName:  container.ContainerName,
			ImageName:      container.ImageNme,
			ImageVersion:   container.ImageVersion,
			CpuRequest:     container.CpuRequest,
			CpuLimit:       container.CpuLimit,
			MemRequest:     container.MemoryRequest,
			MemLimit:       container.MemoryLimit,
			Npu:            container.Npu,
			Env:            env,
			ContainerUser:  container.ContainerUser,
			ContainerGroup: container.ContainerGroup,
			PortMaps:       portMaps,
			Command:        command,
		}
	}
	return nil
}

// ToModel convert dto to app template group model
func (dto *AppTemplateGroupDto) ToModel() *TemplateGroupModel {
	now := time.Now().Format(common.TimeFormat)
	return &TemplateGroupModel{
		Id:          dto.Id,
		Name:        dto.Name,
		Description: dto.Description,
		CreatedAt:   now,
		ModifiedAt:  now,
	}
}

// ToModel convert dto to app template version model
func (dto *AppTemplateVersionDto) ToModel() *TemplateVersionModel {
	now := time.Now().Format(common.TimeFormat)
	result := &TemplateVersionModel{
		GroupId:     dto.GroupId,
		VersionName: dto.Version,
		CreatedAt:   now,
		ModifiedAt:  now,
	}
	result.Containers = make([]ContainerModel, len(dto.Containers))
	for i, item := range dto.Containers {
		result.Containers[i] = *item.ToModel()
	}
	return result
}

// ToModel convert dto to app template container model
func (container *ContainerDto) ToModel() *ContainerModel {
	return &ContainerModel{
		Id:             container.Id,
		ContainerName:  container.ContainerName,
		ImageName:      container.ImageName,
		ImageVersion:   container.ImageVersion,
		CpuRequest:     container.CpuRequest,
		CpuLimit:       container.CpuLimit,
		MemRequest:     container.MemRequest,
		MemLimit:       container.MemLimit,
		Npu:            container.Npu,
		Env:            container.Env,
		ContainerUser:  container.ContainerUser,
		ContainerGroup: container.ContainerGroup,
		PortMaps:       container.PortMaps,
		Command:        container.Command,
	}
}

// ToDto convert model to app template group dto
func (model *TemplateGroupModel) ToDto(versionCount int) *TemplateGroupSummaryDto {
	return &TemplateGroupSummaryDto{
		GroupId:          model.Id,
		GroupName:        model.Name,
		Description:      model.Description,
		CreateTime:       model.CreatedAt,
		LastModifyTime:   model.ModifiedAt,
		SubVersionNumber: versionCount,
	}
}

// ToSummaryDto convert model to app template version summary dto
func (model *TemplateVersionModel) ToSummaryDto() *VersionSummaryDto {
	return &VersionSummaryDto{
		GroupId:    model.GroupId,
		Version:    model.VersionName,
		CreateTime: model.CreatedAt,
	}
}

// ToDetailDto convert model to app template version detail dto
func (model *TemplateVersionModel) ToDetailDto(groupName string) *VersionDetailDto {
	result := &VersionDetailDto{
		GroupName:      groupName,
		CreateTime:     model.CreatedAt,
		LastModifyTime: model.ModifiedAt,
	}
	result.GroupId = model.GroupId
	result.Version = model.VersionName
	result.Containers = make([]ContainerDto, len(model.Containers))
	for i, item := range model.Containers {
		result.Containers[i] = *item.ToDto()
	}
	return result
}

// ToDto convert model to app template container dto
func (model *ContainerModel) ToDto() *ContainerDto {
	return &ContainerDto{
		Id:             model.Id,
		ContainerName:  model.ContainerName,
		ImageName:      model.ImageName,
		ImageVersion:   model.ImageVersion,
		CpuRequest:     model.CpuRequest,
		CpuLimit:       model.CpuLimit,
		MemRequest:     model.MemRequest,
		MemLimit:       model.MemLimit,
		Npu:            model.Npu,
		Env:            model.Env,
		ContainerUser:  model.ContainerUser,
		ContainerGroup: model.ContainerGroup,
		PortMaps:       model.PortMaps,
		Command:        model.Command,
	}
}
