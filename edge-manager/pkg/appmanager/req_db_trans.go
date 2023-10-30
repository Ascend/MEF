// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to define dto struct
package appmanager

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/hwlog"
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

func (cr *ConfigmapReq) toDb() (*ConfigmapInfo, error) {
	configmapContentBytes, err := json.Marshal(cr.ConfigmapContent)
	if err != nil {
		hwlog.RunLog.Errorf("marshal configmap content failed, error: %v", err)
		return nil, errors.New("marshal configmap content failed")
	}

	return &ConfigmapInfo{
		ConfigmapName:    cr.ConfigmapName,
		ConfigmapContent: string(configmapContentBytes),
		Description:      cr.Description,
	}, nil
}
