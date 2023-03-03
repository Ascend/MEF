// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common kmc utils
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

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

var lock sync.Mutex

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
	lock.Lock()
	defer lock.Unlock()
	if kmcCfg == nil {
		kmcCfg = GetDefKmcCfg()
	}
	defer x509.PaddingAndCleanSlice(content)

	if err := kmc.Initialize(kmcCfg.SdpAlgID, kmcCfg.PrimaryKeyPath, kmcCfg.StandbyKeyPath); err != nil {
		return nil, err
	}
	defer func() {
		if err := kmc.Finalize(); err != nil {
			hwlog.RunLog.Errorf("%s", err.Error())
		}
	}()
	encryptByte, err := kmc.Encrypt(kmcCfg.DoMainId, content)
	if err != nil {
		return nil, err
	}
	return encryptByte, nil
}

// DecryptContent decrypt content with kmc
func DecryptContent(encryptByte []byte, kmcCfg *KmcCfg) ([]byte, error) {
	lock.Lock()
	defer lock.Unlock()
	if kmcCfg == nil {
		kmcCfg = GetDefKmcCfg()
	}
	if err := kmc.Initialize(kmcCfg.SdpAlgID, kmcCfg.PrimaryKeyPath, kmcCfg.StandbyKeyPath); err != nil {
		return nil, err
	}
	defer func() {
		if err := kmc.Finalize(); err != nil {
			hwlog.RunLog.Errorf("kmc finalize failed: %s", err.Error())
		}
	}()
	decryptByte, err := kmc.Decrypt(kmcCfg.DoMainId, encryptByte)
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
