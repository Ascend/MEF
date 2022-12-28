// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to define dto struct
package appmanager

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

func (req *CreateAppReq) toDb() (*AppInfo, error) {
	containers, err := json.Marshal(req.Containers)
	if err != nil {
		hwlog.RunLog.Error("marshal containers failed")
		return nil, err
	}

	return &AppInfo{
		AppName:     req.AppName,
		Description: req.Description,
		Containers:  string(containers),
	}, nil
}

func (req *UpdateAppReq) toDb() (*AppInfo, error) {
	containers, err := json.Marshal(req.Containers)
	if err != nil {
		hwlog.RunLog.Error("marshal containers failed")
		return nil, err
	}

	return &AppInfo{
		ID:          req.AppID,
		AppName:     req.AppName,
		Description: req.Description,
		Containers:  string(containers),
	}, nil
}

// ToDb convert app template dto to db model
func (dto *CreateTemplateReq) ToDb(template *AppTemplateDb) error {
	if template == nil {
		return errors.New("param is nil")
	}
	*template = AppTemplateDb{
		TemplateName: dto.Name,
		Description:  dto.Description,
	}

	containers, err := json.Marshal(dto.Containers)
	if err != nil {
		hwlog.RunLog.Error("marshal containers failed")
		return err
	}

	template.Containers = string(containers)

	return nil
}

// ToDb convert app template dto to db model
func (dto *UpdateTemplateReq) ToDb(template *AppTemplateDb) error {
	if template == nil {
		return errors.New("param is nil")
	}
	*template = AppTemplateDb{
		ID:           dto.Id,
		TemplateName: dto.Name,
		Description:  dto.Description,
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
	dto.CreatedAt = template.CreatedAt.Format(common.TimeFormat)
	dto.ModifiedAt = template.UpdatedAt.Format(common.TimeFormat)

	if err := json.Unmarshal([]byte(template.Containers), &dto.Containers); err != nil {
		return err
	}

	return nil
}
