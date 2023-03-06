// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common kmc utils
package common

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
)

const (
	// Aes256gcmId Aes256gcmId id
	Aes256gcmId = 9
	// Aes128gcmId Aes128gcmId id
	Aes128gcmId = 8
	// DoMainId default do main id for kmc
	DoMainId uint = 0
	// DefaultPrimaryKeyPath kmc primary key
	DefaultPrimaryKeyPath = "/home/data/config/kmc/master.ks"
	// DefaultStandKeyPath kmc stand key
	DefaultStandKeyPath = "/home/data/config/kmc/backup.ks"
)

var algMap = map[string]int{
	"Aes256gcm": Aes256gcmId,
	"Aes128gcm": Aes128gcmId,
}

var sdpAlgID = Aes256gcmId
var doMainId = DoMainId

// KmcCfg kmc config struct
type KmcCfg struct {
	SdpAlgID       int
	PrimaryKeyPath string
	StandbyKeyPath string
	DoMainId       uint
}

// GetKmcCfg is the func to init a kmc config structure
func GetKmcCfg(keyPath string, backKeyPath string) *KmcCfg {
	return &KmcCfg{
		SdpAlgID:       sdpAlgID,
		PrimaryKeyPath: keyPath,
		StandbyKeyPath: backKeyPath,
		DoMainId:       doMainId,
	}
}

// GetDefKmcCfg get default kmc config
func GetDefKmcCfg() *KmcCfg {
	return &KmcCfg{
		SdpAlgID:       Aes256gcmId,
		PrimaryKeyPath: DefaultPrimaryKeyPath,
		StandbyKeyPath: DefaultStandKeyPath,
		DoMainId:       DoMainId,
	}
}

// EncryptContent encrypt content with kmc
func EncryptContent(content []byte, kmcCfg *KmcCfg) ([]byte, error) {
	if kmcCfg == nil {
		kmcCfg = GetDefKmcCfg()
	}
	config := kmc.NewKmcInitConfig()
	config.LogLevel = kmc.Error
	config.PrimaryKeyStoreFile = kmcCfg.PrimaryKeyPath
	config.StandbyKeyStoreFile = kmcCfg.StandbyKeyPath
	config.SdpAlgId = kmcCfg.SdpAlgID
	c, err := kmc.KeInitializeEx(config)
	if err != nil {
		return nil, errors.New("initialize kmc failed")
	}
	defer func() {
		if err := c.KeFinalizeEx(); err != nil {
			hwlog.RunLog.Errorf("%s", err.Error())
		}
	}()
	defer x509.PaddingAndCleanSlice(content)
	encryptByte, err := c.KeEncryptByDomainEx(kmcCfg.DoMainId, content)
	if err != nil {
		return nil, err
	}
	return encryptByte, nil
}

// DecryptContent decrypt content with kmc
func DecryptContent(encryptByte []byte, kmcCfg *KmcCfg) ([]byte, error) {
	if kmcCfg == nil {
		kmcCfg = GetDefKmcCfg()
	}
	config := kmc.NewKmcInitConfig()
	config.LogLevel = kmc.Error
	config.PrimaryKeyStoreFile = kmcCfg.PrimaryKeyPath
	config.StandbyKeyStoreFile = kmcCfg.StandbyKeyPath
	config.SdpAlgId = kmcCfg.SdpAlgID
	c, err := kmc.KeInitializeEx(config)
	if err != nil {
		return nil, errors.New("initialize kmc failed")
	}
	defer func() {
		if err := c.KeFinalizeEx(); err != nil {
			hwlog.RunLog.Errorf("%s", err.Error())
		}
	}()
	decryptByte, err := c.KeDecryptByDomainEx(kmcCfg.DoMainId, encryptByte)
	if err != nil {
		return nil, err
	}
	return decryptByte, nil
}

type kmcCfgData struct {
	Algorithms string `json:"algorithms"`
}

// InitKmcCfg [method] for Init kmc config path json
func InitKmcCfg(cfgPath string) error {
	data, err := utils.LoadFile(cfgPath)
	if data == nil {
		return fmt.Errorf("load kmc config file failed: %v", err)
	}
	var kmcCfg kmcCfgData
	err = json.Unmarshal(data, &kmcCfg)
	if err != nil {
		return fmt.Errorf("json umarshal kmc config data failed: %v", err)
	}
	algId, ok := algMap[kmcCfg.Algorithms]
	if !ok {
		return errors.New("the kmcCfg Algorithms not supported")
	}
	sdpAlgID = algId
	return nil
}
