// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller for the table edge_account_infos operation
package edgeinstaller

import (
	"errors"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/database"
	"huawei.com/mindxedge/base/common"
)

var (
	edgeAccountRepositoryInitOnce sync.Once
	edgeAccountRepository         EdgeAccountRepository
)

// EdgeAccountRepositoryImpl edge account service struct
type EdgeAccountRepositoryImpl struct {
	db *gorm.DB
}

// EdgeAccountRepository for edge account method to operate db
type EdgeAccountRepository interface {
	createEdgeAccountInfo(edgeAccountInfo *EdgeAccountInfo) error
	updateEdgeAccountInfo(edgeAccountInfo *EdgeAccountInfo) error
	setEdgeAccountInfo(edgeAccountInfo *EdgeAccountInfo) error
}

// EdgeAccountRepositoryInstance returns the singleton instance of configmap service
func EdgeAccountRepositoryInstance() EdgeAccountRepository {
	edgeAccountRepositoryInitOnce.Do(func() {
		edgeAccountRepository = &EdgeAccountRepositoryImpl{db: database.GetDb()}
	})
	return edgeAccountRepository
}

func (ai *EdgeAccountRepositoryImpl) createEdgeAccountInfo(edgeAccountInfo *EdgeAccountInfo) error {
	if err := ai.db.Model(EdgeAccountInfo{}).Create(edgeAccountInfo).Error; err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			return errors.New("account name is duplicate")
		}
		return err
	}

	hwlog.RunLog.Info("create edge account in db success")
	return nil
}

func (ai *EdgeAccountRepositoryImpl) updateEdgeAccountInfo(edgeAccountInfo *EdgeAccountInfo) error {
	if err := ai.db.Model(EdgeAccountInfo{}).Where("account = ?", edgeAccountInfo.Account).Updates(edgeAccountInfo).
		Error; err != nil {
		return err
	}

	hwlog.RunLog.Info("update edge account in db success")
	return nil
}

func (ai *EdgeAccountRepositoryImpl) setEdgeAccountInfo(edgeAccountInfo *EdgeAccountInfo) error {
	count, err := database.GetItemCount(EdgeAccountInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get item count in table edge_account_infos failed, error: %v", err)
		return err
	}

	if count == 0 {
		edgeAccountInfo.CreatedAt = time.Now().Format(common.TimeFormat)
		if err = EdgeAccountRepositoryInstance().createEdgeAccountInfo(edgeAccountInfo); err != nil {
			hwlog.RunLog.Errorf("create edge account in db failed, error: %v", err)
			return err
		}
	}

	if err = EdgeAccountRepositoryInstance().updateEdgeAccountInfo(edgeAccountInfo); err != nil {
		hwlog.RunLog.Errorf("update edge account in db failed, error: %v", err)
		return err
	}

	return nil
}
