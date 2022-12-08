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
	CreateTemplate(template *AppTemplate) error
	// DeleteTemplates batch delete app template
	DeleteTemplates(ids []uint64) error
	// UpdateTemplate modify app template
	UpdateTemplate(template *AppTemplate) error
	// GetTemplates get app template
	GetTemplates(name string, pageNum, pageSize int) ([]AppTemplate, error)
	// GetTemplate get app template
	GetTemplate(id uint64) (*AppTemplate, error)
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
func (rep *repositoryImpl) CreateTemplate(template *AppTemplate) error {
	if err := rep.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&template).Error; err != nil {
			hwlog.RunLog.Error("create db template failed")
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
		if err := tx.Where("Id in (?)", ids).Delete(AppTemplate{}).Error; err != nil {
			hwlog.RunLog.Error("delete db templates failed")
			return err
		}
		return nil
	}); err != nil {
		return errors.New("delete templates failed")
	}
	return nil
}

// UpdateTemplate modify app template
func (rep *repositoryImpl) UpdateTemplate(template *AppTemplate) error {
	if err := rep.db.Model(AppTemplate{}).Where("id = ?", template.ID).Update("containers", template.Containers).Error; err != nil {
		hwlog.RunLog.Errorf("update template failed: %s", err.Error())
		return err
	}
	return nil

}

// GetTemplates get app template versions
func (rep *repositoryImpl) GetTemplates(name string, pageNum, pageSize int) ([]AppTemplate, error) {
	var templates []AppTemplate
	if pageNum <= 0 {
		pageNum = common.DefaultPage
	}
	if pageSize <= 0 {
		pageSize = common.DefaultMaxPageSize
	}
	if err := rep.db.Model(AppTemplate{}).Where("Name like ?", "%"+name+"%").
		Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, errors.New("get templates failed")
	}
	return templates, nil
}

// GetTemplate get app template
func (rep *repositoryImpl) GetTemplate(id uint64) (*AppTemplate, error) {
	var template AppTemplate
	if err := rep.db.Where(&AppTemplate{ID: id}).First(&template).Error; err != nil {
		hwlog.RunLog.Error("get db template failed")
		return nil, errors.New("get template failed")
	}
	return &template, nil
}
