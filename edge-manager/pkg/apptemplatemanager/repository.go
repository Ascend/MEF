// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package apptemplatemanager to  provide containerized application template management.
package apptemplatemanager

import (
	"edge-manager/pkg/common"
	"edge-manager/pkg/common/model"
	"edge-manager/pkg/database"
	"errors"
	"gorm.io/gorm"
	"strings"
	"sync"
)

var (
	onceInit   sync.Once
	repository Repository
)

// Repository app template db repository interface
type Repository interface {
	// Transaction db transaction
	Transaction(fc func(tx *gorm.DB) error) error
	// CreateGroup create app template group
	CreateGroup(group *model.TemplateGroupModel, tx *gorm.DB) error
	// DeleteGroup delete app template group
	DeleteGroup(id uint64, tx *gorm.DB) error
	// ModifyGroup modify app template group
	ModifyGroup(group *model.TemplateGroupModel, tx *gorm.DB) error
	// GetAllGroups get all app template groups
	GetAllGroups(tx *gorm.DB) ([]model.TemplateGroupModel, error)
	// GetGroup get app template group by id
	GetGroup(id uint64, tx *gorm.DB) (*model.TemplateGroupModel, error)
	// CreateVersion create app template version
	CreateVersion(version *model.TemplateVersionModel, tx *gorm.DB) error
	// DeleteVersion delete app template version
	DeleteVersion(groupId uint64, version string, tx *gorm.DB) error
	// ModifyVersion modify app template version
	ModifyVersion(version *model.TemplateVersionModel, tx *gorm.DB) error
	// GetVersions get app template versions in a group
	GetVersions(groupId uint64, tx *gorm.DB) ([]model.TemplateVersionModel, error)
	// GetVersion get app template version
	GetVersion(groupId uint64, version string, tx *gorm.DB) (*model.TemplateVersionModel, error)
	// GetVersionCount get the count of app template versions in a group
	GetVersionCount(groupId uint64, tx *gorm.DB) (int, error)
	// ExistsVersion whether the app template version name exists in a group
	ExistsVersion(groupId uint64, version string, tx *gorm.DB) (bool, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func (rep *repositoryImpl) getDb(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return rep.db
}

// Transaction db transaction
func (rep *repositoryImpl) Transaction(fc func(tx *gorm.DB) error) error {
	return rep.db.Transaction(fc)
}

// RepositoryInstance get app template repository service instance
func RepositoryInstance() Repository {
	onceInit.Do(func() {
		repository = &repositoryImpl{db: database.GetDb()}
	})
	return repository
}

// CreateGroup create app template group
func (rep *repositoryImpl) CreateGroup(group *model.TemplateGroupModel, tx *gorm.DB) error {
	if err := rep.getDb(tx).Create(group.ToDb()).Error; err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			return errors.New("the group name must be unique")
		}
		return errors.New("create group failed")
	}
	return nil
}

// DeleteGroup delete app template group
func (rep *repositoryImpl) DeleteGroup(id uint64, tx *gorm.DB) error {
	if err := rep.getDb(tx).Delete(&database.AppTemplateGroupDb{Id: id}).Error; err != nil {
		return errors.New("delete group failed")
	}
	return nil
}

// ModifyGroup modify app template group
func (rep *repositoryImpl) ModifyGroup(group *model.TemplateGroupModel, tx *gorm.DB) error {
	data := group.ToDb()
	if err := rep.getDb(tx).Model(data).UpdateColumns(database.AppTemplateGroupDb{
		Name:        data.Name,
		Description: data.Description,
		ModifiedAt:  data.ModifiedAt,
	}).Error; err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			return errors.New("the group name must be unique")
		}
		return errors.New("update group failed")
	}
	return nil
}

// GetAllGroups get all app template groups
func (rep *repositoryImpl) GetAllGroups(tx *gorm.DB) ([]model.TemplateGroupModel, error) {
	var data []database.AppTemplateGroupDb
	if err := rep.getDb(tx).Find(&data).Error; err != nil {
		return nil, errors.New("get all groups failed")
	}
	result := make([]model.TemplateGroupModel, len(data))
	for i, item := range data {
		(&result[i]).FromDb(&item)
	}
	return result, nil
}

// GetGroup get app template group by id
func (rep *repositoryImpl) GetGroup(id uint64, tx *gorm.DB) (*model.TemplateGroupModel, error) {
	var group database.AppTemplateGroupDb
	if err := rep.getDb(tx).Where(&database.AppTemplateGroupDb{Id: id}).First(&group).Error; err != nil {
		return nil, errors.New("get group failed")
	}
	result := &model.TemplateGroupModel{}
	result.FromDb(&group)
	return result, nil
}

// CreateVersion create app template version
func (rep *repositoryImpl) CreateVersion(version *model.TemplateVersionModel, tx *gorm.DB) error {
	var data []database.AppContainerTemplateDb
	var err error
	if data, err = version.ToDb(); err != nil {
		return err
	}
	if err = rep.getDb(tx).Create(&data).Error; err != nil {
		return errors.New("create version failed")
	}
	return nil
}

// DeleteVersion delete app template version
func (rep *repositoryImpl) DeleteVersion(groupId uint64, version string, tx *gorm.DB) error {
	if err := rep.getDb(tx).Where(&database.AppContainerTemplateDb{GroupId: groupId, VersionName: version}).
		Delete(&database.AppContainerTemplateDb{}).Error; err != nil {
		return errors.New("delete version failed")
	}
	return nil
}

// ModifyVersion modify app template version
func (rep *repositoryImpl) ModifyVersion(version *model.TemplateVersionModel, tx *gorm.DB) error {
	containers, err := version.ToDb()
	if err != nil {
		return err
	}
	var exists []database.AppContainerTemplateDb
	if err = rep.getDb(tx).Where(&database.AppContainerTemplateDb{GroupId: version.GroupId,
		VersionName: version.VersionName}).Find(&exists).Error; err != nil {
		return errors.New("modify version failed")
	}
	toCreate, toModify, toDeleteIds := getContainerChanges(exists, containers)
	if err = rep.getDb(tx).Transaction(func(tx *gorm.DB) error {
		for _, toDeleteId := range toDeleteIds {
			if err := rep.getDb(tx).Where(&database.AppContainerTemplateDb{Id: toDeleteId}).
				Delete(&database.AppContainerTemplateDb{}).Error; err != nil {
				return err
			}
		}
		if len(toCreate) > 0 {
			if err := rep.db.Create(&toCreate).Error; err != nil {
				return err
			}
		}
		for _, item := range toModify {
			item.CreatedAt = ""
			if err := rep.db.Updates(&item).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return errors.New("modify version failed")
	}
	return nil
}

// GetVersions get app template versions in a group
func (rep *repositoryImpl) GetVersions(groupId uint64, tx *gorm.DB) ([]model.TemplateVersionModel, error) {
	var containers []database.AppContainerTemplateDb
	if err := rep.getDb(tx).Where(&database.AppContainerTemplateDb{GroupId: groupId}).
		Find(&containers).Error; err != nil {
		return nil, errors.New("get versions failed")
	}
	dic := make(map[string][]database.AppContainerTemplateDb)
	for _, item := range containers {
		if _, ok := dic[item.VersionName]; !ok {
			dic[item.VersionName] = make([]database.AppContainerTemplateDb, 1)
		}
		dic[item.VersionName] = append(dic[item.VersionName], item)
	}
	result := make([]model.TemplateVersionModel, len(dic))
	for _, v := range dic {
		item := &model.TemplateVersionModel{}
		if err := item.FromDb(v); err != nil {
			return nil, err
		}
		result = append(result, *item)
	}
	return result, nil
}

// GetVersion get app template version
func (rep *repositoryImpl) GetVersion(groupId uint64, version string, tx *gorm.DB) (
	*model.TemplateVersionModel, error) {
	var containers []database.AppContainerTemplateDb
	if err := rep.getDb(tx).Where(&database.AppContainerTemplateDb{GroupId: groupId, VersionName: version}).
		Find(&containers).Error; err != nil {
		return nil, errors.New("get version failed")
	}
	result := &model.TemplateVersionModel{}
	if err := result.FromDb(containers); err != nil {
		return nil, err
	}
	return result, nil
}

// GetVersionCount get the count of app template versions in a group
func (rep *repositoryImpl) GetVersionCount(groupId uint64, tx *gorm.DB) (int, error) {
	var containers []database.AppContainerTemplateDb
	if err := rep.getDb(tx).Where(&database.AppContainerTemplateDb{GroupId: groupId}).
		Find(&containers).Error; err != nil {
		return 0, errors.New("get version count failed")
	}
	dic := make(map[string]struct{})
	for _, item := range containers {
		if _, ok := dic[item.VersionName]; !ok {
			dic[item.VersionName] = struct{}{}
		}
	}
	return len(dic), nil
}

// ExistsVersion whether the app template version name exists in a group
func (rep *repositoryImpl) ExistsVersion(groupId uint64, version string, tx *gorm.DB) (bool, error) {
	var count int64
	if err := rep.getDb(tx).Model(&database.AppContainerTemplateDb{}).
		Where(&database.AppContainerTemplateDb{GroupId: groupId, VersionName: version}).
		Count(&count).Error; err != nil {
		return false, errors.New("query version exist failed")
	}
	return count > 0, nil
}

func getContainerChanges(exists []database.AppContainerTemplateDb, containers []database.AppContainerTemplateDb) (
	toCreate, toModify []database.AppContainerTemplateDb, toDeleteIds []uint64) {
	deleteDic := make(map[uint64]struct{})
	dic := make(map[uint64]struct{})
	for _, exist := range exists {
		dic[exist.Id] = struct{}{}
		deleteDic[exist.Id] = struct{}{}
	}
	toModify = make([]database.AppContainerTemplateDb, 0)
	toCreate = make([]database.AppContainerTemplateDb, 0)
	for _, container := range containers {
		if _, ok := dic[container.Id]; ok {
			toModify = append(toModify, container)
			delete(deleteDic, container.Id)
		} else {
			toCreate = append(toCreate, container)
		}
	}
	toDeleteIds = make([]uint64, len(deleteDic))
	for id := range deleteDic {
		toDeleteIds = append(toDeleteIds, id)
	}
	return toCreate, toModify, toDeleteIds
}
