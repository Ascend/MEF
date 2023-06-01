// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package kmcupdate this file for update kmc flow
package kmcupdate

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const maxKeyCount = 100

// UpdateKmcTask is the struct for update kmc task which is to update single component's kmc key
type UpdateKmcTask struct {
	configPathMgr *util.ConfigPathMgr
	module        string
	ctx           kmc.Context
}

func (ukt *UpdateKmcTask) initKmcCtx() error {
	var kmcKeyPath, kmcBackKeyPath string
	if ukt.module == util.MefCenterRootName {
		kmcKeyPath = ukt.configPathMgr.GetRootMasterKmcPath()
		kmcBackKeyPath = ukt.configPathMgr.GetRootBackKmcPath()
	} else {
		kmcKeyPath = ukt.configPathMgr.GetComponentMasterKmcPath(ukt.module)
		kmcBackKeyPath = ukt.configPathMgr.GetComponentBackKmcPath(ukt.module)
	}
	kmcConfig := kmc.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)

	config := kmc.NewKmcInitConfig()
	config.PrimaryKeyStoreFile = kmcConfig.PrimaryKeyPath
	config.StandbyKeyStoreFile = kmcConfig.StandbyKeyPath
	config.SdpAlgId = kmcConfig.SdpAlgID
	c, err := kmc.KeInitializeEx(config)
	if err != nil {
		hwlog.RunLog.Errorf("Init kmc failed: %v", err.Error())
		fmt.Println("inin kmc failed")
		return err
	}

	ukt.ctx = c
	return nil
}

func (ukt *UpdateKmcTask) updateRk() error {
	hwlog.RunLog.Infof("start to update module %s's kmc root key", ukt.module)
	if err := ukt.ctx.UpdateRootKey(); err != nil {
		hwlog.RunLog.Errorf("update module %s's kmc root key failed: %s", ukt.module, err.Error())
		return fmt.Errorf("update module %s's kmc root key failed", ukt.module)
	}

	hwlog.RunLog.Infof("update module %s's kmc root key success", ukt.module)
	return nil
}

func (ukt *UpdateKmcTask) updateMk() error {
	hwlog.RunLog.Infof("start to update module %s's kmc master key", ukt.module)
	if err := ukt.ctx.KeActiveNewKeyEx(kmc.DefaultDoMainId); err != nil {
		hwlog.RunLog.Errorf("update module %s's kmc master key failed: %s", ukt.module, err.Error())
		return fmt.Errorf("update module %s's kmc master key failed", ukt.module)
	}

	hwlog.RunLog.Infof("update module %s's kmc master key success", ukt.module)
	return nil
}

func (ukt *UpdateKmcTask) smoothMk() error {
	maxKeyId, err := ukt.ctx.KeGetMaxMkIDEx(kmc.DefaultDoMainId)
	if err != nil {
		hwlog.RunLog.Errorf("get module %s's kmc max master key id failed: %s", ukt.module, err.Error())
		return fmt.Errorf("get module %s's kmc max master key id failed", ukt.module)
	}

	if maxKeyId <= maxKeyCount {
		return nil
	}

	deleteKeyId := maxKeyId - maxKeyCount
	if err = ukt.ctx.KeRemoveKeyByIDEx(kmc.DefaultDoMainId, deleteKeyId); err != nil {
		hwlog.RunLog.Errorf("remove module %s's oldest master key failed: %s", ukt.module, err.Error())
		return fmt.Errorf("remove module %s's oldest master key failed", ukt.module)
	}

	return nil
}

func (ukt *UpdateKmcTask) reEncrypt() error {
	hwlog.RunLog.Infof("start to re-encrypted module %s", ukt.module)
	encryptMap := ukt.getEncryptMap()
	fileList, exists := encryptMap[ukt.module]
	if !exists {
		hwlog.RunLog.Error("unsupported module name")
		return errors.New("unsupported module name")
	}

	for _, singleFile := range fileList {
		reEncryptFunc, err := ukt.getReEncryptFunc(singleFile)
		if err != nil {
			return err
		}

		if err = reEncryptFunc(singleFile); err != nil {
			return err
		}
	}

	hwlog.RunLog.Infof("re-encrypted module %s success", ukt.module)
	return nil
}

func (ukt *UpdateKmcTask) reEncryptKey(fileName string) error {
	if !utils.IsExist(fileName) {
		return nil
	}

	encryptedData, err := utils.LoadFile(fileName)
	if err != nil {
		hwlog.RunLog.Errorf("load file %s failed: %s", fileName, err.Error())
		return fmt.Errorf("load file %s failed", fileName)
	}

	plainData, err := ukt.ctx.KeDecryptByDomainEx(kmc.DefaultDoMainId, encryptedData)
	if err != nil {
		hwlog.RunLog.Errorf("decrypt %s file failed: %s", fileName, err.Error())
		return fmt.Errorf("decrypt %s file failed", fileName)
	}
	defer utils.ClearSliceByteMemory(plainData)

	reEncryptedData, err := ukt.ctx.KeEncryptByDomainEx(kmc.DefaultDoMainId, plainData)
	if err != nil {
		hwlog.RunLog.Errorf("re-encrypt %s file failed: %s", fileName, err.Error())
		return fmt.Errorf("re-encrypt %s file failed", fileName)
	}

	if err = utils.WriteData(fileName, reEncryptedData); err != nil {
		hwlog.RunLog.Errorf("write into re-encrypted data into %s failed: %s", fileName, err.Error())
		return fmt.Errorf("write into re-encrypted data into %s failed", fileName)
	}

	return nil
}

func (ukt *UpdateKmcTask) getReEncryptFunc(fileName string) (func(string) error, error) {
	funcMap := map[string]func(string) error{
		util.KeySuffix: ukt.reEncryptKey,
	}

	fileSuffix := filepath.Ext(fileName)
	function, exists := funcMap[fileSuffix]
	if !exists {
		return nil, errors.New("unsupported suffix")
	}

	return function, nil
}

func (ukt *UpdateKmcTask) getEncryptMap() map[string][]string {
	return map[string][]string{
		util.CertManagerName: {
			ukt.configPathMgr.GetComponentKeyPath(util.CertManagerName),
			ukt.configPathMgr.GetHubSrvKeyPath(),
		},
		util.NginxManagerName: {
			ukt.configPathMgr.GetComponentKeyPath(util.NginxManagerName),
			ukt.configPathMgr.GetUserServerKeyPath(),
		},
		util.EdgeManagerName: {
			ukt.configPathMgr.GetComponentKeyPath(util.EdgeManagerName),
			ukt.configPathMgr.GetWebSocketKeyPath(),
		},
		util.MefCenterRootName: {ukt.configPathMgr.GetRootCaKeyPath()},
	}
}

// ManualUpdateKmcTask is the struct to manually update kmc keys
type ManualUpdateKmcTask struct {
	UpdateKmcTask
}

// RunTask is the main func to start manually update kmc task
func (muk *ManualUpdateKmcTask) RunTask() error {
	if err := muk.initKmcCtx(); err != nil {
		return fmt.Errorf("init module %s's kmc context failed", muk.module)
	}

	defer func() {
		if err := muk.ctx.KeFinalizeEx(); err != nil {
			hwlog.RunLog.Errorf("finalize module %s's kmc context failed", muk.module)
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

// NewManualUpdateKmcTask is the task to update init a ManualUpdateKmcTask instance
func NewManualUpdateKmcTask(configPathMgr *util.ConfigPathMgr, module string) *ManualUpdateKmcTask {
	return &ManualUpdateKmcTask{
		UpdateKmcTask: UpdateKmcTask{
			configPathMgr: configPathMgr,
			module:        module,
		},
	}
}
