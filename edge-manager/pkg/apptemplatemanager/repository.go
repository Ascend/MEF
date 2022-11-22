// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package apptemplatemanager to  provide containerized application template management.
package apptemplatemanager

import (
	"edge-manager/pkg/database"
	"errors"
	"strings"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

var (
	onceInit   sync.Once
	repository Repository
)

// Repository app template db repository interface
type Repository interface {
	// CreateTemplate create app template
	CreateTemplate(template *AppTemplateDb) error
	// DeleteTemplates batch delete app template
	DeleteTemplates(ids []uint64) error
	// ModifyTemplate modify app template
	ModifyTemplate(template *AppTemplateDb) error
	// GetTemplates get app template
	GetTemplates(name string, pageNum, pageSize int) ([]AppTemplateDb, error)
	// GetTemplate get app template
	GetTemplate(id uint64) (*AppTemplateDb, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

// RepositoryInstance get app template repository service instance
func RepositoryInstance() Repository {
	onceInit.Do(func() {
		repository = &repositoryImpl{db: database.GetDb()}
	})
	return repository
}

// CreateTemplate create app template
func (rep *repositoryImpl) CreateTemplate(template *AppTemplateDb) error {
	if err := rep.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&template).Error; err != nil {
			hwlog.RunLog.Error("create db template failed")
			return err
		}
		for i := range template.Containers {
			template.Containers[i].TemplateId = template.Id
		}
		if err := tx.Create(template.Containers).Error; err != nil {
			hwlog.RunLog.Error("create db containers failed")
			return err
		}
		return nil
	}); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			return errors.New("the template name and container name must be unique")
		}
		return errors.New("create app template failed")
	}
	return nil
}

// DeleteTemplates batch delete app template
func (rep *repositoryImpl) DeleteTemplates(ids []uint64) error {
	if err := rep.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("Id in (?)", ids).Delete(AppTemplateDb{}).Error; err != nil {
			hwlog.RunLog.Error("delete db templates failed")
			return err
		}
		if err := tx.Where("TemplateId in (?)", ids).Delete(TemplateContainerDb{}).Error; err != nil {
			hwlog.RunLog.Error("delete db containers failed")
			return err
		}
		return nil
	}); err != nil {
		return errors.New("delete templates failed")
	}
	return nil
}

// ModifyTemplate modify app template
func (rep *repositoryImpl) ModifyTemplate(template *AppTemplateDb) error {
	var exists []TemplateContainerDb
	if err := rep.db.Where(TemplateContainerDb{TemplateId: template.Id}).Find(&exists).Error; err != nil {
		hwlog.RunLog.Error("get db containers failed")
		return errors.New("modify template failed")
	}
	toCreate, toModify, toDeleteIds := getContainerChanges(exists, template.Containers)
	if err := rep.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("Id in (?)", toDeleteIds).Delete(TemplateContainerDb{}).Error; err != nil {
			hwlog.RunLog.Error("batch delete db containers failed")
			return err
		}
		if len(toCreate) > 0 {
			if err := tx.Create(&toCreate).Error; err != nil {
				hwlog.RunLog.Error("batch create db containers failed")
				return err
			}
		}
		for _, container := range toModify {
			if err := tx.Updates(&container).Error; err != nil {
				hwlog.RunLog.Error("update db container failed")
				return err
			}
		}
		template.CreatedAt = ""
		if err := tx.Updates(&template).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			return errors.New("the template name and container name must be unique")
		}
		return errors.New("modify app template failed")
	}
	return nil
}

// GetTemplates get app template versions
func (rep *repositoryImpl) GetTemplates(name string, pageNum, pageSize int) ([]AppTemplateDb, error) {
	var templates []AppTemplateDb
	if pageNum <= 0 {
		pageNum = common.DefaultPage
	}
	if pageSize <= 0 {
		pageSize = common.DefaultMaxPageSize
	}
	if err := rep.db.Model(AppTemplateDb{}).Where("Name like ?", "%"+name+"%").
		Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, errors.New("get templates failed")
	}
	return templates, nil
}

// GetTemplate get app template
func (rep *repositoryImpl) GetTemplate(id uint64) (*AppTemplateDb, error) {
	var template AppTemplateDb
	if err := rep.db.Where(&AppTemplateDb{Id: id}).First(&template).Error; err != nil {
		hwlog.RunLog.Error("get db template failed")
		return nil, errors.New("get template failed")
	}
	if err := rep.db.Where(TemplateContainerDb{TemplateId: id}).Find(&(template.Containers)).Error; err != nil {
		hwlog.RunLog.Error("get db containers failed")
		return nil, errors.New("get template failed")
	}
	return &template, nil
}

func getContainerChanges(exists []TemplateContainerDb, containers []TemplateContainerDb) (
	toCreate, toModify []TemplateContainerDb, toDeleteIds []uint64) {
	deleteDic := make(map[uint64]struct{})
	dic := make(map[uint64]struct{})
	for _, exist := range exists {
		dic[exist.Id] = struct{}{}
		deleteDic[exist.Id] = struct{}{}
	}
	toModify = make([]TemplateContainerDb, 0)
	toCreate = make([]TemplateContainerDb, 0)
	for _, container := range containers {
		if _, ok := dic[container.Id]; ok {
			toModify = append(toModify, container)
			delete(deleteDic, container.Id)
		} else {
			toCreate = append(toCreate, container)
		}
	}
	toDeleteIds = make([]uint64, 0)
	for id := range deleteDic {
		toDeleteIds = append(toDeleteIds, id)
	}
	return toCreate, toModify, toDeleteIds
}
