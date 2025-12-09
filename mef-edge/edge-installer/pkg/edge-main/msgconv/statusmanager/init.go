// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package statusmanager
package statusmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/database"
)

var (
	mgrsLock sync.Mutex
	mgrs     []statusMgrImpl
	// ErrNotFound record not found
	ErrNotFound = gorm.ErrRecordNotFound
)

// StatusMgr to handle the status cache
type StatusMgr interface {
	Set(key string, data interface{}) error
	Patch(key string, data []byte) error
	Get(key string) (string, error)
	GetAll() (map[string]string, error)
	Delete(key string) error
}

// GetPodStatusMgr getStatusMgr pod status manager
func GetPodStatusMgr() StatusMgr {
	return getStatusMgr(constants.ResourceTypePod)
}

// GetNodeStatusMgr getStatusMgr node status manager
func GetNodeStatusMgr() StatusMgr {
	return getStatusMgr(constants.ResourceTypeNode)
}

// GetConfigMapStatusMgr getStatusMgr configmap status manager
func GetConfigMapStatusMgr() StatusMgr {
	return getStatusMgr(constants.ResourceTypeConfigMap)
}

// GetAlarmStatusMgr getStatusMgr alarm status manager
func GetAlarmStatusMgr() StatusMgr {
	return getStatusMgr(constants.ResourceTypeAlarm)
}

// Get the getStatusMgr method
func getStatusMgr(category string) StatusMgr {
	mgrsLock.Lock()
	defer mgrsLock.Unlock()

	for index := range mgrs {
		mgr := &mgrs[index]
		if mgr.typ == category {
			return mgr
		}
	}

	mgrs = append(mgrs, statusMgrImpl{typ: category})
	return &mgrs[len(mgrs)-1]
}

type statusMgrImpl struct {
	typ string
}

// Set the set method
func (sm *statusMgrImpl) Set(key string, value interface{}) error {
	var (
		jsonValue string
	)

	switch value.(type) {
	case string:
		jsonValue = value.(string)
	default:
		valueBytes, err := json.Marshal(value)
		if err != nil {
			return errors.New("failed to marshal valueData")
		}
		jsonValue = string(valueBytes)
	}

	err := database.GetMetaRepository().CreateOrUpdate(database.Meta{
		Key:   key,
		Type:  sm.typ,
		Value: jsonValue,
	})
	if err != nil {
		return fmt.Errorf("create or update data failed, [type: %s, key: %s] ", sm.typ, key)
	}
	return nil
}

// Patch the patch method
func (sm *statusMgrImpl) Patch(key string, data []byte) error {
	var (
		patchBytes []byte
		err        error
	)
	patchBytes = model.UnformatMsg(data)

	oldMeta, err := database.GetMetaRepository().GetByKey(key)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to load meta(key=%s), db query error", key)
		}
		err := database.GetMetaRepository().CreateOrUpdate(database.Meta{
			Key:   key,
			Type:  sm.typ,
			Value: string(patchBytes),
		})
		if err != nil {
			return fmt.Errorf("create or update data failed, [type: %s, key: %s] ", sm.typ, key)
		}
		return nil
	}

	jsonValueBytes, err := util.MergePatch([]byte(oldMeta.Value), patchBytes)
	if err != nil {
		return fmt.Errorf("failed to create merge patch, %v", err)
	}
	jsonValue := string(jsonValueBytes)
	if jsonValue == "{}" {
		if err := database.GetMetaRepository().DeleteByKey(key); err != nil {
			return fmt.Errorf("delete data failed, [type: %s, key: %s] ", sm.typ, key)
		}
		return nil
	}

	err = database.GetMetaRepository().CreateOrUpdate(database.Meta{
		Key:   key,
		Type:  sm.typ,
		Value: jsonValue,
	})
	if err != nil {
		return fmt.Errorf("create or update data failed, [type: %s, key: %s] ", sm.typ, key)
	}
	return nil
}

func (sm *statusMgrImpl) Get(key string) (string, error) {
	meta, err := database.GetMetaRepository().GetByKey(key)
	if err != nil {
		return "", fmt.Errorf("get data failed, [type: %s, key: %s] ", sm.typ, key)
	}
	return meta.Value, nil
}

func (sm *statusMgrImpl) GetAll() (map[string]string, error) {
	metas, err := database.GetMetaRepository().GetByType(sm.typ)
	if err != nil {
		return nil, fmt.Errorf("get data by type %s failed", sm.typ)
	}

	result := make(map[string]string)
	for _, meta := range metas {
		result[meta.Key] = meta.Value
	}
	return result, nil
}

func (sm *statusMgrImpl) Delete(key string) error {
	if err := database.GetMetaRepository().DeleteByKey(key); err != nil {
		return fmt.Errorf("delete data failed, [type: %s, key: %s] ", sm.typ, key)
	}
	return nil
}
