// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configmanager for
package configmanager

import (
	"errors"
	"sync"

	"gorm.io/gorm"

	"edge-manager/pkg/database"
)

var (
	repositoryInitOnce sync.Once
	configRepository   ConfigRepository
)

type configRepositoryImpl struct {
	db *gorm.DB
}

// ConfigRepository for config method to operate db
type ConfigRepository interface {
	generateAndSaveToken(*TokenInfo) error
	GetToken() ([]byte, []byte, error)
}

// ConfigRepositoryInstance returns the singleton instance of config service
func ConfigRepositoryInstance() ConfigRepository {
	repositoryInitOnce.Do(func() {
		configRepository = &configRepositoryImpl{db: database.GetDb()}
	})
	return configRepository
}

func (c *configRepositoryImpl) generateAndSaveToken(tokenInfo *TokenInfo) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&TokenInfo{}).Error; err != nil {
			return errors.New("delete old token error")
		}
		if err := tx.Model(TokenInfo{}).Create(tokenInfo).Error; err != nil {
			return errors.New("create token db error")
		}
		return nil
	})
}

func (c *configRepositoryImpl) GetToken() ([]byte, []byte, error) {
	var tokenInfo []TokenInfo
	if err := c.db.Model(TokenInfo{}).Find(&tokenInfo).Error; err != nil {
		return nil, nil, errors.New("get token from db error")
	}
	if len(tokenInfo) != 1 {
		return nil, nil, errors.New("token number exceed 1")
	}
	return tokenInfo[0].Token, tokenInfo[0].Salt, nil
}
