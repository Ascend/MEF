// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package apptemplatemanager to define dto struct
package apptemplatemanager

import (
	"encoding/json"
	"errors"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"strconv"
	"time"
)

// AppTemplateDto app template dto
type AppTemplateDto struct {
	Id             uint64         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	CreateTime     string         `json:"createTime"`
	LastModifyTime string         `json:"lastModifyTime"`
	Containers     []ContainerDto `json:"containers"`
}

// TemplateSummaryDto app template summary dto
type TemplateSummaryDto struct {
	Id             uint64 `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CreateTime     string `json:"createTime"`
	LastModifyTime string `json:"lastModifyTime"`
}

// ContainerDto app template version container dto
type ContainerDto struct {
	Id             uint64    `json:"id"`
	Name           string    `json:"name"`
	ImageName      string    `json:"imageName"`
	ImageVersion   string    `json:"imageVersion"`
	CpuRequest     string    `json:"cpuRequest"`
	CpuLimit       string    `json:"cpuLimit"`
	MemRequest     string    `json:"memRequest"`
	MemLimit       string    `json:"memLimit"`
	Npu            string    `json:"npu"`
	Env            []Dic     `json:"env"`
	ContainerUser  string    `json:"containerUser"`
	ContainerGroup string    `json:"containerGroup"`
	PortMaps       []PortMap `json:"portMaps"`
	Command        []string  `json:"command"`
}

// ReqDeleteTemplate request body to delete app template
type ReqDeleteTemplate struct {
	Ids []uint64 `json:"ids"`
}

// ReqGetTemplates request body to get app template versions
type ReqGetTemplates struct {
	Name     string `json:"name"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
}

// ReqGetTemplateDetail request body to get app template detail
type ReqGetTemplateDetail struct {
	Id uint64 `json:"id"`
}

// Dic key-value dictionary
type Dic struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PortMap container port mapping config
type PortMap struct {
	Protocol      string `json:"protocol"`
	ContainerPort string `json:"containerPort"`
	HostIp        string `json:"hostIp"`
	HostPort      string `json:"hostPort"`
}

// ToDb convert app template dto to db model
func (dto *AppTemplateDto) ToDb(template *AppTemplateDb) error {
	if template == nil {
		return errors.New("param is nil")
	}
	now := time.Now().Format(common.TimeFormat)
	*template = AppTemplateDb{
		Id:          dto.Id,
		Name:        dto.Name,
		Description: dto.Description,
		CreatedAt:   now,
		ModifiedAt:  now,
	}
	template.Containers = make([]TemplateContainerDb, len(dto.Containers))
	for i, container := range dto.Containers {
		if err := (&container).ToDb(&(template.Containers[i])); err != nil {
			return err
		}
		template.Containers[i].TemplateId = template.Id
	}
	return nil
}

// FromDb convert db model to app template dto
func (dto *AppTemplateDto) FromDb(template *AppTemplateDb) error {
	if template == nil {
		return errors.New("param is nil")
	}
	dto.Id = template.Id
	dto.Name = template.Name
	dto.Description = template.Description
	dto.CreateTime = template.CreatedAt
	dto.LastModifyTime = template.ModifiedAt
	dto.Containers = make([]ContainerDto, len(template.Containers))
	for i, container := range template.Containers {
		if err := (&(dto.Containers[i])).FromDb(&container); err != nil {
			return err
		}
	}
	return nil
}

// FromDb convert model to app template version summary dto
func (dto *TemplateSummaryDto) FromDb(template *AppTemplateDb) {
	*dto = TemplateSummaryDto{
		Id:             template.Id,
		Name:           template.Name,
		Description:    template.Description,
		CreateTime:     template.CreatedAt,
		LastModifyTime: template.ModifiedAt,
	}
}

// FromDb convert model to app template container dto
func (dto *ContainerDto) FromDb(container *TemplateContainerDb) error {
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
	*dto = ContainerDto{
		Id:             container.Id,
		Name:           container.Name,
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
	return nil
}

// ToDb convert dto to app template container db model
func (dto *ContainerDto) ToDb(container *TemplateContainerDb) error {
	if container == nil {
		return errors.New("param is nil")
	}
	env, err := json.Marshal(dto.Env)
	if err != nil {
		return err
	}
	var command []byte
	command, err = json.Marshal(dto.Command)
	if err != nil {
		return err
	}
	var portMaps []byte
	portMaps, err = json.Marshal(dto.PortMaps)
	if err != nil {
		return err
	}
	*container = TemplateContainerDb{
		Id:             dto.Id,
		Name:           dto.Name,
		ImageNme:       dto.ImageName,
		ImageVersion:   dto.ImageVersion,
		CpuRequest:     dto.CpuRequest,
		CpuLimit:       dto.CpuLimit,
		MemoryRequest:  dto.MemRequest,
		MemoryLimit:    dto.MemLimit,
		Npu:            dto.Npu,
		Env:            string(env),
		ContainerUser:  dto.ContainerUser,
		ContainerGroup: dto.ContainerGroup,
		PortMaps:       string(portMaps),
		Command:        string(command),
	}
	return nil
}

// Check whether app template dto is valid
func (dto *AppTemplateDto) Check() error {
	validator := common.NewValidator().ValidateAppName("name", dto.Name).
		ValidateAppDesc("description", dto.Description)
	validateTemplateContainers(validator, "containers", dto.Containers)
	return validator.Error()
}

func validateTemplateContainers(v *common.Validator, paramName string, containers []ContainerDto) {
	v.ValidateCount(paramName, len(containers), common.AppTemplateContainersMin, common.AppTemplateContainersMax)
	for i, container := range containers {
		prefix := paramName + "[" + strconv.Itoa(i) + "]."
		v.ValidateUnique(prefix+"name", "name", container.Name).
			ValidateContainerName(prefix+"name", container.Name).
			ValidateImageName(prefix+"imageName", container.ImageName).
			ValidateImageVersion(prefix+"imageVersion", container.ImageVersion).
			ValidateCpu(prefix+"cpuRequest", container.CpuRequest).
			ValidateMemory(prefix+"memRequest", container.MemRequest)
		if container.CpuLimit != "" {
			v.ValidateCpu(prefix+"cpuLimit", container.CpuRequest).
				ValidateGtEq(prefix+"cpuLimit", prefix+"cpuRequest", container.CpuLimit, container.CpuRequest)
		}
		if container.MemLimit != "" {
			v.ValidateMemory(prefix+"memLimit", container.MemRequest).
				ValidateGtEq(prefix+"memLimit", prefix+"memRequest", container.MemLimit, container.MemRequest)
		}
		if container.Npu != "" {
			v.ValidateNpu(prefix+"npu", container.Npu)
		}
		if container.ContainerUser != "" {
			v.ValidateContainerUid(prefix+"containerUser", container.ContainerUser)
		}
		if container.ContainerGroup != "" {
			v.ValidateContainerGid(prefix+"containerGroup", container.ContainerGroup)
		}
		if container.Env != nil {
			ValidateEnv(v, prefix+"env", container.Env)
		}
		if container.PortMaps != nil {
			ValidatePortMaps(v, prefix+"portMaps", container.PortMaps)
		}
	}
}

// ValidatePortMaps validate port maps
func ValidatePortMaps(v *common.Validator, paramName string, portMaps []PortMap) {
	v.ValidateCount(paramName, len(portMaps), 0, common.PortMapsMax)
	for j, pm := range portMaps {
		prefixPm := paramName + "[" + strconv.Itoa(j) + "]."
		v.ValidateContainerPort(prefixPm+"containerPort", pm.ContainerPort).
			ValidateHostPort(prefixPm+"hostPort", pm.HostPort).
			ValidateHostIp(prefixPm+"hostIp", pm.HostIp).
			ValidatePortProtocol(prefixPm+"protocol", pm.Protocol)
	}
}

// ValidateEnv validate environment variables
func ValidateEnv(v *common.Validator, paramName string, env []Dic) {
	v.ValidateCount(paramName, len(env), 0, common.EnvCountMax)
	for j, kv := range env {
		prefixEnv := paramName + "[" + strconv.Itoa(j) + "]."
		v.ValidateEnvKey(prefixEnv+"key", kv.Key).ValidateEnvValue(prefixEnv+"value", kv.Value)
	}
}

// UnmarshalJSON custom JSON unmarshal
func (req *ReqGetTemplates) UnmarshalJSON(input []byte) error {
	objMap := make(map[string][]string)
	if err := json.Unmarshal(input, &objMap); err != nil {
		return err
	}
	if names, ok := objMap["name"]; ok && len(names) > 0 {
		req.Name = names[0]
	} else {
		hwlog.RunLog.Warn("get param name failed")
	}
	if err := common.GetIntParam(objMap, "pageNum", &(req.PageNum)); err != nil {
		hwlog.RunLog.Warn("get param pageNum failed")
	}
	if err := common.GetIntParam(objMap, "pageSize", &(req.PageSize)); err != nil {
		hwlog.RunLog.Warn("get param pageSize failed")
	}
	return nil
}

// UnmarshalJSON custom JSON unmarshal
func (req *ReqGetTemplateDetail) UnmarshalJSON(input []byte) error {
	objMap := make(map[string][]string)
	if err := json.Unmarshal(input, &objMap); err != nil {
		return err
	}
	if err := common.GetUintParam(objMap, "id", &(req.Id)); err != nil {
		return errors.New("get param id failed")
	}
	return nil
}
