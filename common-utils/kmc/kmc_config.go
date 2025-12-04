// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package kmc for
package kmc

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

// kmc config used
const (
	Aes128gcmId                = 8
	Aes256gcmId                = 9
	DefaultDoMainId       uint = 0
	DefaultPrimaryKeyPath      = "/home/data/config/kmc/master.ks"
	DefaultStandKeyPath        = "/home/data/config/kmc/backup.ks"
)

var (
	algMap = map[string]int{
		"Aes128gcm": Aes128gcmId,
		"Aes256gcm": Aes256gcmId,
	}
	sdpAlgID = Aes256gcmId
	doMainId = DefaultDoMainId
)

type kmcCfgData struct {
	Algorithms string `json:"algorithms"`
}

// SubConfig kmc config struct
type SubConfig struct {
	SdpAlgID       int
	PrimaryKeyPath string
	StandbyKeyPath string
	DoMainId       uint
}

// GetKmcCfg init a kmc config structure
func GetKmcCfg(keyPath string, backKeyPath string) *SubConfig {
	return &SubConfig{
		SdpAlgID:       sdpAlgID,
		PrimaryKeyPath: keyPath,
		StandbyKeyPath: backKeyPath,
		DoMainId:       doMainId,
	}
}

// GetDefKmcCfg get default kmc config
func GetDefKmcCfg() *SubConfig {
	return &SubConfig{
		SdpAlgID:       Aes256gcmId,
		PrimaryKeyPath: DefaultPrimaryKeyPath,
		StandbyKeyPath: DefaultStandKeyPath,
		DoMainId:       DefaultDoMainId,
	}
}

// InitKmcCfg init kmc config path json
func InitKmcCfg(cfgPath string) error {
	data, err := fileutils.LoadFile(cfgPath)
	if data == nil {
		return fmt.Errorf("load kmc config file failed, error: %v", err)
	}
	var kmcCfg kmcCfgData
	if err = json.Unmarshal(data, &kmcCfg); err != nil {
		return fmt.Errorf("json umarshal kmc config data failed, error: %v", err)
	}
	algId, ok := algMap[kmcCfg.Algorithms]
	if !ok {
		return errors.New("the kmcCfg Algorithms not supported")
	}
	sdpAlgID = algId
	return nil
}

// EncryptContent encrypt content with kmc
func EncryptContent(content []byte, kmcCfg *SubConfig) ([]byte, error) {
	if kmcCfg == nil {
		kmcCfg = GetDefKmcCfg()
	}
	config := NewKmcInitConfig()
	config.LogLevel = Error
	if !fileutils.IsExist(kmcCfg.PrimaryKeyPath) {
		hwlog.RunLog.Infof("Primary Key %s will be created.", kmcCfg.PrimaryKeyPath)
	}
	if !fileutils.IsExist(kmcCfg.StandbyKeyPath) {
		hwlog.RunLog.Infof("Standby Key %s will be created.", kmcCfg.StandbyKeyPath)
	}
	config.PrimaryKeyStoreFile = kmcCfg.PrimaryKeyPath
	config.StandbyKeyStoreFile = kmcCfg.StandbyKeyPath
	config.SdpAlgId = kmcCfg.SdpAlgID
	c, err := KeInitializeEx(config)
	if err != nil {
		utils.ClearSliceByteMemory(content)
		return nil, errors.New("initialize kmc failed")
	}
	defer func() {
		if err = c.KeFinalizeEx(); err != nil {
			hwlog.RunLog.Errorf("%s", err.Error())
		}
	}()
	defer utils.ClearSliceByteMemory(content)
	if err = c.KeRefreshMkMaskEx(); err != nil {
		hwlog.RunLog.Errorf("%s", err.Error())
		return nil, err
	}
	encryptByte, err := c.KeEncryptByDomainEx(kmcCfg.DoMainId, content)
	if err != nil {
		return nil, err
	}
	if err = c.KeRefreshMkMaskEx(); err != nil {
		hwlog.RunLog.Errorf("%s", err.Error())
		return nil, err
	}
	hwlog.RunLog.Infof("KMC Key %s will used to encrypt content.", config.PrimaryKeyStoreFile)
	return encryptByte, nil
}

// DecryptContent decrypt content with kmc
func DecryptContent(encryptByte []byte, kmcCfg *SubConfig) ([]byte, error) {
	if kmcCfg == nil {
		kmcCfg = GetDefKmcCfg()
	}
	config := NewKmcInitConfig()
	config.LogLevel = Error
	if !fileutils.IsExist(kmcCfg.PrimaryKeyPath) {
		hwlog.RunLog.Infof("Primary Key %s will be created.", kmcCfg.PrimaryKeyPath)
	}
	if !fileutils.IsExist(kmcCfg.StandbyKeyPath) {
		hwlog.RunLog.Infof("Standby Key %s will be created.", kmcCfg.StandbyKeyPath)
	}
	config.PrimaryKeyStoreFile = kmcCfg.PrimaryKeyPath
	config.StandbyKeyStoreFile = kmcCfg.StandbyKeyPath
	config.SdpAlgId = kmcCfg.SdpAlgID
	c, err := KeInitializeEx(config)
	if err != nil {
		return nil, errors.New("initialize kmc failed")
	}
	defer func() {
		if err = c.KeFinalizeEx(); err != nil {
			hwlog.RunLog.Errorf("%s", err.Error())
		}
	}()
	if err = c.KeRefreshMkMaskEx(); err != nil {
		hwlog.RunLog.Errorf("%s", err.Error())
		return nil, err
	}
	decryptByte, err := c.KeDecryptByDomainEx(kmcCfg.DoMainId, encryptByte)
	if err != nil {
		return nil, err
	}
	if err = c.KeRefreshMkMaskEx(); err != nil {
		hwlog.RunLog.Errorf("%s", err.Error())
		return nil, err
	}
	hwlog.RunLog.Infof("KMC Key %s will used to decrypt content.", config.PrimaryKeyStoreFile)
	return decryptByte, nil
}
