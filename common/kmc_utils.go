// Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.

// Package common kmc utils
package common

import (
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/x509"
)

const (
	// Aes256gcm Aes256gcm id
	Aes256gcm = 9
	// DoMainId default do main id for kmc
	DoMainId = 0
	// DefaultPrimaryKeyPath kmc primary key
	DefaultPrimaryKeyPath = "/home/data/mef/kmc/master.ks"
	// DefaultStandKeyPath kmc stand key
	DefaultStandKeyPath = "/home/data/mef/kmc/backup.ks"
)

// KmcCfg kmc config struct
type KmcCfg struct {
	SdpAlgID       int
	PrimaryKeyPath string
	StandbyKeyPath string
	DoMainId       uint
}

// GetDefKmcCfg get default kmc config
func GetDefKmcCfg() *KmcCfg {
	return &KmcCfg{
		SdpAlgID:       Aes256gcm,
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
	if kmcCfg == nil {
		kmcCfg = GetDefKmcCfg()
	}
	if err := kmc.Initialize(kmcCfg.SdpAlgID, kmcCfg.PrimaryKeyPath, kmcCfg.StandbyKeyPath); err != nil {
		return nil, err
	}
	defer func() {
		if err := kmc.Finalize(); err != nil {
			hwlog.RunLog.Errorf("%s", err.Error())
		}
	}()
	decryptByte, err := kmc.Decrypt(kmcCfg.DoMainId, encryptByte)
	if err != nil {
		return nil, err
	}
	return decryptByte, nil
}
