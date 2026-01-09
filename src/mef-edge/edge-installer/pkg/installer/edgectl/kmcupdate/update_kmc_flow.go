// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package kmcupdate this file for update kmc flow
package kmcupdate

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

// UpdateKmcFlow is the struct for update kmc flow
type UpdateKmcFlow struct {
	configPathMgr *pathmgr.ConfigPathMgr
}

// NewUpdateKmcFlow create UpdateKmcTask instance
func NewUpdateKmcFlow(configPathMgr *pathmgr.ConfigPathMgr) *UpdateKmcFlow {
	return &UpdateKmcFlow{configPathMgr: configPathMgr}
}

func (ukm *UpdateKmcFlow) getModules() []string {
	return []string{constants.EdgeCore, constants.EdgeOm, constants.EdgeMain}
}

// RunFlow is the main func to start a task
func (ukm *UpdateKmcFlow) RunFlow() error {
	fmt.Println("start to update kmc keys")
	hwlog.RunLog.Info("start to update kmc keys")

	var failedModule []string
	for _, module := range ukm.getModules() {
		kmcDir := ukm.configPathMgr.GetCompKmcDir(module)
		ctx, err := ukm.initKmcCtx(kmcDir)
		if err != nil {
			return err
		}

		encryptedMap := ukm.getEncryptMap()
		encryptedList, exists := encryptedMap[module]
		if !exists {
			hwlog.RunLog.Errorf("unsupported module %s", module)
			return errors.New("unsupported module")
		}

		task := kmc.ManualUpdateKmcTask{
			UpdateKmcTask: kmc.UpdateKmcTask{
				ReEncryptParamList: encryptedList,
				Ctx:                ctx,
			},
		}

		hwlog.RunLog.Infof("start to update module %s's kmc keys", module)
		if err = task.RunTask(); err != nil {
			hwlog.RunLog.Errorf("update module %s's kmc keys failed: %s", module, err.Error())
			failedModule = append(failedModule, module)
			continue
		}
		hwlog.RunLog.Infof("update module %s's kmc keys success", module)
	}

	if len(failedModule) == 0 {
		fmt.Println("update kmc keys success")
		hwlog.RunLog.Info("update kmc keys success")
		return nil
	}
	fmt.Printf("update module %s's kmc key failed\n", failedModule)

	return errors.New("update kmc key failed")
}

func (ukm *UpdateKmcFlow) initKmcCtx(kmcDir string) (*kmc.Context, error) {
	kmcCfg, err := util.GetKmcConfig(kmcDir)
	if err != nil {
		hwlog.RunLog.Errorf("Get Kmc Config Failed %v", err.Error())
		fmt.Println("Get kmc Config failed")
		return nil, err
	}
	config := kmc.NewKmcInitConfig()
	config.PrimaryKeyStoreFile = kmcCfg.PrimaryKeyPath
	config.StandbyKeyStoreFile = kmcCfg.StandbyKeyPath
	config.SdpAlgId = kmcCfg.SdpAlgID
	c, err := kmc.KeInitializeEx(config)
	if err != nil {
		hwlog.RunLog.Errorf("Init kmc failed: %v", err.Error())
		fmt.Println("inin kmc failed")
		return nil, err
	}

	return &c, nil
}

func (ukm *UpdateKmcFlow) getEncryptMap() map[string][]kmc.ReEncryptParam {
	coreKeyParam := kmc.ReEncryptParam{
		Path:       ukm.configPathMgr.GetCompConfigDir(constants.EdgeCore),
		SuffixList: []string{kmc.KeySuffix},
	}
	omKeyParam := kmc.ReEncryptParam{
		Path:       ukm.configPathMgr.GetCompConfigDir(constants.EdgeOm),
		SuffixList: []string{kmc.KeySuffix},
	}
	mainKeyParam := kmc.ReEncryptParam{
		Path:       ukm.configPathMgr.GetCompConfigDir(constants.EdgeMain),
		SuffixList: []string{kmc.KeySuffix},
	}
	return map[string][]kmc.ReEncryptParam{
		constants.EdgeOm:   {omKeyParam},
		constants.EdgeMain: {mainKeyParam},
		constants.EdgeCore: {coreKeyParam},
	}
}
