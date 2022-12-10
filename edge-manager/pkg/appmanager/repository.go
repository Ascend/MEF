// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package apptemplatemanager to  provide containerized application template management.
package appmanager

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
	// createTemplate create app template
	createTemplate(template *AppTemplateDb) error
	// deleteTemplates batch delete app template
	deleteTemplates(ids []uint64) error
	// updateTemplate modify app template
	updateTemplate(template *AppTemplateDb) error
	// getTemplates get app template
	getTemplates(name string, pageNum, pageSize uint64) ([]AppTemplateDb, error)
	// getTemplate get app template
	getTemplate(id uint64) (*AppTemplateDb, error)

	// countListAppsInfo get app template
	getTemplateCount(name string) (int64, error)
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

// createTemplate create app template
func (req *repositoryImpl) createTemplate(template *AppTemplateDb) error {
	if err := req.db.Transaction(func(tx *gorm.DB) error {
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
func (req *repositoryImpl) deleteTemplates(ids []uint64) error {
	if err := req.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("Id in (?)", ids).Delete(AppTemplateDb{}).Error; err != nil {
			hwlog.RunLog.Error("delete db templates failed")
			return err
		}
		return nil
	}); err != nil {
		return errors.New("delete templates failed")
	}
	return nil
}

// updateTemplate modify app template
func (req *repositoryImpl) updateTemplate(template *AppTemplateDb) error {
	if err := req.db.Model(AppTemplateDb{}).Where("id = ?", template.ID).Updates(template).Error; err != nil {
		hwlog.RunLog.Errorf("update template failed: %s", err.Error())
		return err
	}
	return nil

}

// getTemplates get app template versions
func (req *repositoryImpl) getTemplates(name string, pageNum, pageSize uint64) ([]AppTemplateDb, error) {
	var templates []AppTemplateDb

	if err := req.db.Model(AppTemplateDb{}).Scopes(getAppInfoByLikeName(pageNum,
		pageSize, name)).Find(&templates).Error; err != nil {
		hwlog.RunLog.Error("list appInfo db failed")
		return nil, err
	}

	return templates, nil
}

// GetTemplate get app template
func (req *repositoryImpl) getTemplate(id uint64) (*AppTemplateDb, error) {
	var template AppTemplateDb
	if err := req.db.Where(&AppTemplateDb{ID: id}).First(&template).Error; err != nil {
		hwlog.RunLog.Error("get db template failed")
		return nil, errors.New("get template failed")
	}
	return &template, nil
}

func (req *repositoryImpl) getTemplateCount(name string) (int64, error) {
	var totalTemplateCount int64
	if err := req.db.Model(AppTemplateDb{}).Where("template_name like ?",
		"%"+name+"%").Count(&totalTemplateCount).Error; err != nil {
		hwlog.RunLog.Error("count list appInfo db failed")
		return 0, err
	}
	return totalTemplateCount, nil
}
