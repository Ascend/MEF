// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to define dto struct
package appmanager

import (
	"encoding/json"
	"errors"
	"time"

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
		CreatedAt:   time.Now().Format(common.TimeFormat),
		ModifiedAt:  time.Now().Format(common.TimeFormat),
	}, nil

}

func (req *CreateAppReq) fromDb() (*AppInfo, error) {
	containers, err := json.Marshal(req.Containers)
	if err != nil {
		hwlog.RunLog.Error("marshal containers failed")
		return nil, err
	}

	return &AppInfo{
		AppName:     req.AppName,
		Description: req.Description,
		Containers:  string(containers),
		CreatedAt:   time.Now().Format(common.TimeFormat),
		ModifiedAt:  time.Now().Format(common.TimeFormat),
	}, nil
}

// ToDb convert app template dto to db model
func (dto *AppTemplateReq) ToDb(template *AppTemplateDb) error {
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
func (dto *AppTemplateReq) FromDb(template *AppTemplateDb) error {
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
