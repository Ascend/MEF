// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package configmanager for
package configmanager

import (
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"

	"huawei.com/mindx/common/database"
)

var (
	repositoryInitOnce sync.Once
	configRepository   ConfigRepository
)

type configRepositoryImpl struct {
}

// ConfigRepository for config method to operate db
type ConfigRepository interface {
	saveToken(TokenInfo) error
	GetToken() ([]byte, []byte, error)
	ifTokenExpire() (bool, error)
	revokeToken() error
}

// ConfigRepositoryInstance returns the singleton instance of config service
func ConfigRepositoryInstance() ConfigRepository {
	repositoryInitOnce.Do(func() {
		configRepository = &configRepositoryImpl{}
	})
	return configRepository
}

func (c *configRepositoryImpl) db() *gorm.DB {
	return database.GetDb()
}

func (c *configRepositoryImpl) saveToken(info TokenInfo) error {
	return database.Transaction(c.db(), func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&TokenInfo{}).Error; err != nil {
			return errors.New("delete old token error")
		}
		if err := tx.Model(TokenInfo{}).Create(info).Error; err != nil {
			return errors.New("create token db error")
		}
		return nil
	})
}

func (c *configRepositoryImpl) GetToken() ([]byte, []byte, error) {
	tokenInfo, err := c.getTokenInfo()
	if err != nil {
		return nil, nil, err
	}
	return tokenInfo.Token, tokenInfo.Salt, nil
}

func (c *configRepositoryImpl) getTokenInfo() (TokenInfo, error) {
	var tokenInfo []TokenInfo
	if err := c.db().Model(TokenInfo{}).Find(&tokenInfo).Error; err != nil {
		return TokenInfo{}, errors.New("get token from db error")
	}
	if len(tokenInfo) == 0 {
		return TokenInfo{}, gorm.ErrRecordNotFound
	}
	if len(tokenInfo) != 1 {
		return TokenInfo{}, errors.New("token number exceed 1")
	}
	return tokenInfo[0], nil
}

func (c *configRepositoryImpl) ifTokenExpire() (bool, error) {
	tokenInfo, err := c.getTokenInfo()
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	if tokenInfo.ExpireTime > time.Now().Unix() {
		return false, nil
	}
	return true, nil
}

func (c *configRepositoryImpl) revokeToken() error {
	if err := c.db().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&TokenInfo{}).Error; err != nil {
		return errors.New("revoke token error")
	}
	return nil
}
