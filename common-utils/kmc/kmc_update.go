// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package kmc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

// define the constant that used in kmc update and key re-encrypt
const (
	maxKeyCount = 100
	KeySuffix   = ".key"
)

// ReEncryptParam is the struct that defines how a path should be re-encrypted after kmc-updating
type ReEncryptParam struct {
	Path       string
	SuffixList []string
}

// UpdateKmcTask is the struct for update kmc task which is to update single component's kmc key
type UpdateKmcTask struct {
	ReEncryptParamList []ReEncryptParam
	Ctx                *Context
}

func (ukt *UpdateKmcTask) updateRk() error {
	if err := ukt.Ctx.UpdateRootKey(); err != nil {
		return fmt.Errorf("update kmc root key failed: %s", err.Error())
	}

	return nil
}

func (ukt *UpdateKmcTask) updateMk() error {
	if err := ukt.Ctx.KeActiveNewKeyEx(DefaultDoMainId); err != nil {
		return fmt.Errorf("update kmc master key failed: %s", err.Error())
	}

	return nil
}

func (ukt *UpdateKmcTask) smoothMk() error {
	maxKeyId, err := ukt.Ctx.KeGetMaxMkIDEx(DefaultDoMainId)
	if err != nil {
		return fmt.Errorf("get kmc max master key id failed: %s", err.Error())
	}

	if maxKeyId <= maxKeyCount {
		return nil
	}

	deleteKeyId := maxKeyId - maxKeyCount
	if err = ukt.Ctx.KeRemoveKeyByIDEx(DefaultDoMainId, deleteKeyId); err != nil {
		return fmt.Errorf("remove oldest master key failed: %s", err.Error())
	}

	return nil
}

func (ukt *UpdateKmcTask) reEncrypt() error {
	for _, reEncryptParam := range ukt.ReEncryptParamList {
		if fileutils.IsDir(reEncryptParam.Path) {
			if err := ukt.reEncryptFromDir(reEncryptParam); err != nil {
				return err
			}
		}

		if fileutils.IsFile(reEncryptParam.Path) {
			if err := ukt.reEncryptFromFile(reEncryptParam.Path); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ukt *UpdateKmcTask) reEncryptFromDir(param ReEncryptParam) error {
	return filepath.Walk(param.Path, func(file string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path [%s] failed, error: %s", param.Path, err.Error())
		}

		if ukt.ifNeedReEncrypt(file, param.SuffixList) {
			if err = ukt.reEncryptFromFile(file); err != nil {
				return err
			}
		}

		return nil
	})
}

func (ukt *UpdateKmcTask) reEncryptFromFile(file string) error {
	reEncryptFunc, err := ukt.getReEncryptFunc(filepath.Ext(file))
	if err != nil {
		return err
	}

	if err = reEncryptFunc(file); err != nil {
		return err
	}

	return nil
}

func (ukt *UpdateKmcTask) ifNeedReEncrypt(fileName string, suffixList []string) bool {
	for _, suffix := range suffixList {
		if filepath.Ext(fileName) == suffix {
			return true
		}
	}

	return false
}

func (ukt *UpdateKmcTask) reEncryptKey(fileName string) error {
	if !fileutils.IsExist(fileName) {
		return nil
	}

	encryptedData, err := fileutils.LoadFile(fileName)
	if err != nil {
		return fmt.Errorf("load file %s failed: %s", fileName, err.Error())
	}

	plainData, err := ukt.Ctx.KeDecryptByDomainEx(DefaultDoMainId, encryptedData)
	if err != nil {
		return fmt.Errorf("decrypt %s file failed: %s", fileName, err.Error())
	}
	defer utils.ClearSliceByteMemory(plainData)

	reEncryptedData, err := ukt.Ctx.KeEncryptByDomainEx(DefaultDoMainId, plainData)
	if err != nil {
		return fmt.Errorf("re-encrypt %s file failed: %s", fileName, err.Error())
	}

	if err = fileutils.WriteData(fileName, reEncryptedData); err != nil {
		return fmt.Errorf("write into re-encrypted data into %s failed: %s", fileName, err.Error())
	}

	if err = backuputils.BackUpFiles(fileName); err != nil {
		return fmt.Errorf("back up file %s failed: %v", fileName, err)
	}

	return nil
}

func (ukt *UpdateKmcTask) getReEncryptFunc(fileName string) (func(string) error, error) {
	funcMap := map[string]func(string) error{
		KeySuffix: ukt.reEncryptKey,
	}

	fileSuffix := filepath.Ext(fileName)
	function, exists := funcMap[fileSuffix]
	if !exists {
		return nil, errors.New("unsupported suffix")
	}

	return function, nil
}

// ManualUpdateKmcTask is the struct to manually update kmc keys
type ManualUpdateKmcTask struct {
	UpdateKmcTask
}

// RunTask is the main func to start manually update kmc task
func (muk *ManualUpdateKmcTask) RunTask() error {
	defer func() {
		if err := muk.Ctx.KeFinalizeEx(); err != nil {
			hwlog.RunLog.Errorf("finalize kmc context failed: %s", err.Error())
		}
	}()

	tasks := []func() error{
		muk.updateRk,
		muk.updateMk,
		muk.smoothMk,
		muk.reEncrypt,
	}

	for _, task := range tasks {
		if err := task(); err != nil {
			return err
		}
	}

	return nil
}
