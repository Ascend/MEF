// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common for parameter validate
package common

import (
	"huawei.com/mindxedge/base/common/checker"
)

// ValidateAppName validate app name
func (v *Validator) ValidateAppName(paramName, templateName string) *Validator {
	return v.ValidateStringRegex(paramName, templateName, RegAppTemplate)
}

// ValidateAppDesc validate app description
func (v *Validator) ValidateAppDesc(paramName, description string) *Validator {
	return v.ValidateStringLength(paramName, description, AppTemplateDesMin, AppTemplateDesMax)
}

// ValidateContainerName validate app container name
func (v *Validator) ValidateContainerName(paramName, containerName string) *Validator {
	return v.ValidateStringRegex(paramName, containerName, RegContainerName)
}

// ValidateImageName validate container image name
func (v *Validator) ValidateImageName(paramName, imageName string) *Validator {
	return v.ValidateStringRegex(paramName, imageName, RegImageName)
}

// ValidateImageVersion validate container image version
func (v *Validator) ValidateImageVersion(paramName, imageVersion string) *Validator {
	return v.ValidateStringRegex(paramName, imageVersion, RegImageVersion)
}

// ValidateCpu validate container cpu count setting
func (v *Validator) ValidateCpu(paramName, cpu string) *Validator {
	return v.ValidateFloat(paramName, cpu, CpuMin, CpuMax, CpuDecimalsNum)
}

// ValidateMemory validate container memory count setting
func (v *Validator) ValidateMemory(paramName, memory string) *Validator {
	return v.ValidateInt(paramName, memory, MemoryMin, MemoryMax)
}

// ValidateNpu validate container npu count setting
func (v *Validator) ValidateNpu(paramName, npu string) *Validator {
	return v.ValidateFloat(paramName, npu, NpuMin, NpuMax, NpuDecimalsNum)
}

// ValidateContainerUid validate container user id
func (v *Validator) ValidateContainerUid(paramName, userId string) *Validator {
	return v.ValidateInt(paramName, userId, ContainerUserIdMin, ContainerUserIdMax)
}

// ValidateContainerGid validate container group id
func (v *Validator) ValidateContainerGid(paramName, groupId string) *Validator {
	return v.ValidateInt(paramName, groupId, ContainerGroupIdMin, ContainerGroupIdMax)
}

// ValidateEnvKey validate environment variable key
func (v *Validator) ValidateEnvKey(paramName, keyName string) *Validator {
	return v.ValidateStringRegex(paramName, keyName, RegEnvKey)
}

// ValidateEnvValue validate environment variable value
func (v *Validator) ValidateEnvValue(paramName, value string) *Validator {
	return v.ValidateStringLength(paramName, value, TemplateEnvValueMin, TemplateEnvValueMax)
}

// ValidateContainerPort validate container port
func (v *Validator) ValidateContainerPort(paramName, port string) *Validator {
	return v.ValidateInt(paramName, port, ContainerPortMin, ContainerPortMax)
}

// ValidateHostPort validate host port
func (v *Validator) ValidateHostPort(paramName, port string) *Validator {
	return v.ValidateInt(paramName, port, HostPortMin, HostPortMax)
}

// ValidateHostIp validate host ip
func (v *Validator) ValidateHostIp(paramName, ip string) *Validator {
	if ok, err := checker.IsIpValid(ip); err != nil || !ok {
		return v.AppendError(paramName, "ip invalid")
	}
	return v
}

// ValidatePortProtocol validate port mapping protocol
func (v *Validator) ValidatePortProtocol(paramName, protocol string) *Validator {
	return v.ValidateIn(paramName, protocol, []string{Tcp, Udp})
}
