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

// AppTemplateDto app template dto
type AppTemplateDto struct {
	Id          uint64      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Containers  []Container `json:"containers"`
}

// ReqDeleteTemplate request body to delete app template
type ReqDeleteTemplate struct {
	Ids []uint64 `json:"ids"`
}

// ToDb convert app template dto to db model
func (dto *AppTemplateDto) ToDb(template *AppTemplate) error {
	if template == nil {
		return errors.New("param is nil")
	}
	now := time.Now().Format(common.TimeFormat)
	*template = AppTemplate{
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
func (dto *AppTemplateDto) FromDb(template *AppTemplate) error {
	if template == nil {
		return errors.New("param is nil")
	}
	dto.Id = template.ID
	dto.Name = template.TemplateName
	dto.Description = template.Description

	if err := json.Unmarshal([]byte(template.Containers), &dto.Containers); err != nil {
		return err
	}

	return nil
}
