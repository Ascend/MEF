// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package apptemplatemanager to  provide containerized application template management.
package apptemplatemanager

import (
	"edge-manager/pkg/common"
	"edge-manager/pkg/common/model"
	"errors"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
)

// CreateGroup create app template group
func CreateGroup(group *model.AppTemplateGroupDto) error {
	hwlog.RunLog.Info("create app template group,start")
	if group == nil {
		err := errors.New("request body is nil")
		hwlog.RunLog.Errorf("create app template group,failed,error:%v", err)
		return err
	}
	if err := group.Check(); err != nil {
		hwlog.RunLog.Errorf("create app template group,failed,error:%v", err)
		return err
	}
	if err := RepositoryInstance().CreateGroup(group.ToModel(), nil); err != nil {
		hwlog.RunLog.Errorf("create app template group,failed,error:%v", err)
		return err
	}
	hwlog.RunLog.Info("create app template group,success")
	return nil
}

// DeleteGroup delete app template group
func DeleteGroup(id uint64) error {
	hwlog.RunLog.Info("delete app template group,start")
	if id == 0 {
		hwlog.RunLog.Error("delete app template group,failed,error:group id invalid")
		return errors.New("group id invalid")
	}
	if err := RepositoryInstance().Transaction(func(tx *gorm.DB) error {
		if err := RepositoryInstance().DeleteGroup(id, nil); err != nil {
			return err
		}
		return RepositoryInstance().DeleteVersion(id, "", nil)
	}); err != nil {
		hwlog.RunLog.Errorf("delete app template group,failed,error:%v", err)
		return err
	}
	hwlog.RunLog.Info("delete app template group,success")
	return nil
}

// ModifyGroup modify app template group
func ModifyGroup(group *model.AppTemplateGroupDto) error {
	hwlog.RunLog.Info("modify app template group,start")
	if group == nil {
		err := errors.New("request body is nil")
		hwlog.RunLog.Errorf("modify app template group,failed,error:%v", err)
		return err
	}
	if err := group.Check(); err != nil {
		hwlog.RunLog.Errorf("modify app template group,failed,error:%v", err)
		return err
	}
	if err := RepositoryInstance().ModifyGroup(group.ToModel(), nil); err != nil {
		hwlog.RunLog.Errorf("modify app template group,failed,error:%v", err)
		return err
	}
	hwlog.RunLog.Info("modify app template group,success")
	return nil
}

// GetAllGroups get all app template groups
func GetAllGroups() ([]model.TemplateGroupSummaryDto, error) {
	hwlog.RunLog.Info("get all app template group,start")
	groups, err := RepositoryInstance().GetAllGroups(nil)
	if err != nil {
		hwlog.RunLog.Errorf("get all app template group,failed,error:%v", err)
		return nil, err
	}
	result := make([]model.TemplateGroupSummaryDto, len(groups))
	for i, group := range groups {
		count := 0
		if count, err = RepositoryInstance().GetVersionCount(group.Id, nil); err != nil {
			hwlog.RunLog.Errorf("get all app template group,failed,error:%v", err)
			return nil, err
		}
		result[i] = *group.ToDto(count)
	}
	hwlog.RunLog.Info("get all app template group,success")
	return result, nil
}

// CreateVersion create app template version
func CreateVersion(version *model.AppTemplateVersionDto) error {
	hwlog.RunLog.Info("create app template version,start")
	var err error
	if version == nil {
		err = errors.New("request body is nil")
		hwlog.RunLog.Errorf("create app template version,failed,error:%v", err)
		return err
	}
	if err = version.Check(); err != nil {
		hwlog.RunLog.Errorf("create app template version,failed,error:%v", err)
		return err
	}
	var exist bool
	if exist, err = RepositoryInstance().ExistsVersion(version.GroupId, version.Version, nil); err != nil {
		hwlog.RunLog.Errorf("create app template version,failed,error:%v", err)
		return err
	}
	if exist {
		err = errors.New("version name must be unique in a group")
		hwlog.RunLog.Errorf("create app template version,failed,error:%v", err)
		return err
	}
	var count int
	if count, err = RepositoryInstance().GetVersionCount(version.GroupId, nil); err != nil {
		hwlog.RunLog.Errorf("create app template version,failed,error:%v", err)
		return err
	}
	if count >= common.AppTemplateGroupVersionsLimit {
		err = errors.New("the number of versions in the template group has reached the upper limit")
		hwlog.RunLog.Errorf("create app template version,failed,error:%v", err)
		return err
	}
	if err = RepositoryInstance().CreateVersion(version.ToModel(), nil); err != nil {
		hwlog.RunLog.Errorf("create app template version,failed,error:%v", err)
		return err
	}
	hwlog.RunLog.Info("create app template version,success")
	return nil
}

// DeleteVersion delete app template version
func DeleteVersion(groupId uint64, version string) error {
	hwlog.RunLog.Info("delete app template version,start")
	if groupId == 0 {
		hwlog.RunLog.Error("delete app template version,failed,error:group id invalid")
		return errors.New("group id invalid")
	}
	if version == "" {
		hwlog.RunLog.Error("delete app template version,failed,error:version name invalid")
		return errors.New("version name invalid")
	}
	if err := RepositoryInstance().DeleteVersion(groupId, version, nil); err != nil {
		hwlog.RunLog.Errorf("delete app template version,failed,error:%v", err)
		return err
	}
	hwlog.RunLog.Info("delete app template version,success")
	return nil
}

// ModifyVersion modify app template version
func ModifyVersion(version *model.AppTemplateVersionDto) error {
	hwlog.RunLog.Info("modify app template version,start")
	if version == nil {
		err := errors.New("request body is nil")
		hwlog.RunLog.Errorf("modify app template version,failed,error:%v", err)
		return err
	}
	if err := version.Check(); err != nil {
		hwlog.RunLog.Errorf("modify app template version,failed,error:%v", err)
		return err
	}
	if err := RepositoryInstance().ModifyVersion(version.ToModel(), nil); err != nil {
		hwlog.RunLog.Errorf("modify app template version,failed,error:%v", err)
		return err
	}
	hwlog.RunLog.Info("modify app template version,success")
	return nil
}

// GetVersions get app template versions
func GetVersions(groupId uint64) ([]model.VersionSummaryDto, error) {
	hwlog.RunLog.Info("get app template versions in a group,start")
	models, err := RepositoryInstance().GetVersions(groupId, nil)
	if err != nil {
		hwlog.RunLog.Errorf("get app template versions in a group,failed,error:%v", err)
		return nil, err
	}
	result := make([]model.VersionSummaryDto, len(models))
	for i, item := range models {
		result[i] = *item.ToSummaryDto()
	}
	hwlog.RunLog.Info("get app template versions in a group,success")
	return result, nil
}

// GetVersionDetail get app template version detail
func GetVersionDetail(groupId uint64, version string) (*model.VersionDetailDto, error) {
	hwlog.RunLog.Info("get app template version detail,start")
	versionModel, err := RepositoryInstance().GetVersion(groupId, version, nil)
	if err != nil {
		hwlog.RunLog.Errorf("get app template version detail,failed,error:%v", err)
		return nil, err
	}
	group, err := RepositoryInstance().GetGroup(groupId, nil)
	if err != nil {
		hwlog.RunLog.Errorf("get app template version detail,failed,error:%v", err)
		return nil, err
	}
	dto := versionModel.ToDetailDto(group.Name)
	hwlog.RunLog.Info("get app template version detail,success")
	return dto, nil
}
