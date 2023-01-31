// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to  provide containerized application template management.
package appmanager

import (
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/database"
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
func (r *repositoryImpl) createTemplate(template *AppTemplateDb) error {
	if err := r.db.Model(AppTemplateDb{}).Create(template).Error; err != nil {
		hwlog.RunLog.Errorf("create app template failed")
		return err
	}
	return nil
}

// DeleteTemplates batch delete app template
func (r *repositoryImpl) deleteTemplates(ids []uint64) error {
	if err := r.db.Where("Id in (?)", ids).Delete(AppTemplateDb{}).Error; err != nil {
		hwlog.RunLog.Error("delete app templates failed")
		return err
	}

	return nil
}

// updateTemplate modify app template
func (r *repositoryImpl) updateTemplate(template *AppTemplateDb) error {
	if err := r.db.Model(AppTemplateDb{}).Where("id = ?", template.ID).Updates(template).Error; err != nil {
		hwlog.RunLog.Errorf("update template failed: %s", err.Error())
		return err
	}
	return nil

}

func getTemplateByLikeName(page, pageSize uint64, appName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("INSTR(template_name, ?)", appName)
	}
}

// getTemplates get app template versions
func (r *repositoryImpl) getTemplates(name string, pageNum, pageSize uint64) ([]AppTemplateDb, error) {
	var templates []AppTemplateDb

	if err := r.db.Model(AppTemplateDb{}).Scopes(getTemplateByLikeName(pageNum,
		pageSize, name)).Find(&templates).Error; err != nil {
		hwlog.RunLog.Error("list appInfo db failed")
		return nil, err
	}

	return templates, nil
}

// GetTemplate get app template
func (r *repositoryImpl) getTemplate(id uint64) (*AppTemplateDb, error) {
	var template AppTemplateDb
	if err := r.db.Model(AppTemplateDb{}).Where("id = ?", id).First(&template).Error; err != nil {
		hwlog.RunLog.Error("get db template failed")
		return nil, err
	}
	return &template, nil
}

func (r *repositoryImpl) getTemplateCount(name string) (int64, error) {
	var totalTemplateCount int64
	if err := r.db.Model(AppTemplateDb{}).Where("INSTR(template_name, ?)",
		name).Count(&totalTemplateCount).Error; err != nil {
		hwlog.RunLog.Error("count list appInfo db failed")
		return 0, err
	}
	return totalTemplateCount, nil
}
