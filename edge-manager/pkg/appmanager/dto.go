// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to define dto struct
package appmanager

import (
	"encoding/json"
	"errors"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"time"
)

// AppTemplate app template dto
type AppTemplate struct {
	Id          uint64      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CreatedAt   string      `json:"createdAt"`
	ModifiedAt  string      `json:"modifiedAt"`
	Containers  []Container `json:"containers"`
}

// ListAppTemplateInfo encapsulate app list
type ListAppTemplateInfo struct {
	// AppTemplates app template info
	AppTemplates []AppTemplate `json:"appTemplates"`
	// Total is num of appInfos
	Total int64 `json:"total"`
}

// ReqDeleteTemplate request body to delete app template
type ReqDeleteTemplate struct {
	Ids []uint64 `json:"ids"`
}

// ToDb convert app template dto to db model
func (dto *AppTemplate) ToDb(template *AppTemplateDb) error {
	if template == nil {
		return errors.New("param is nil")
	}
	now := time.Now().Format(common.TimeFormat)
	*template = AppTemplateDb{
		ID:           dto.Id,
		TemplateName: dto.Name,
		Description:  dto.Description,
		CreatedAt:    now,
		ModifiedAt:   now,
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
func (dto *AppTemplate) FromDb(template *AppTemplateDb) error {
	if template == nil {
		return errors.New("param is nil")
	}
	dto.Id = template.ID
	dto.Name = template.TemplateName
	dto.Description = template.Description
	dto.CreatedAt = template.CreatedAt
	dto.ModifiedAt = template.ModifiedAt

	if err := json.Unmarshal([]byte(template.Containers), &dto.Containers); err != nil {
		return err
	}

	return nil
}
