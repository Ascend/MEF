// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to define dto struct
package appmanager

import (
	"encoding/json"
	"errors"
	"huawei.com/mindx/common/hwlog"
)

// AppTemplateDto app template dto
type AppTemplateDto struct {
	Id          uint64      `json:"id"`
	Name        string      `json:"name"`
	CreatedAt   string      `json:"createdAt"`
	ModifiedAt  string      `json:"modifiedAt"`
	Description string      `json:"description"`
	Containers  []Container `json:"containers"`
}

// ListAppTemplateInfo encapsulate app list
type ListAppTemplateInfo struct {
	// AppTemplates app template info
	AppTemplates []AppTemplateDto `json:"appTemplates"`
	// Total is num of appInfos
	Total int64 `json:"total"`
}

// ReqDeleteTemplate request body to delete app template
type ReqDeleteTemplate struct {
	Ids []uint64 `json:"ids"`
}

// ToDb convert app template dto to db model
func (dto *AppTemplateDto) ToDb(template *AppTemplateDb) error {
	if template == nil {
		return errors.New("param is nil")
	}
	*template = AppTemplateDb{
		ID:          dto.Id,
		AppName:     dto.Name,
		Description: dto.Description,
	}

	containers, err := json.Marshal(dto.Containers)
	if err != nil {
		hwlog.RunLog.Error("marshal containers failed")
		return err
	}

	template.Containers = string(containers)

	return nil
}

// FromDb convert db model to app template dto
func (dto *AppTemplateDto) FromDb(template *AppTemplateDb) error {
	if template == nil {
		return errors.New("param is nil")
	}
	dto.Id = template.ID
	dto.Name = template.AppName
	dto.Description = template.Description
	dto.CreatedAt = template.CreatedAt
	dto.ModifiedAt = template.ModifiedAt

	if err := json.Unmarshal([]byte(template.Containers), &dto.Containers); err != nil {
		return err
	}

	return nil
}
