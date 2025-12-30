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

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// UpdateKmcFlow is the struct for update kmc flow
type UpdateKmcFlow struct {
	pathMgr *util.InstallDirPathMgr
}

// NewUpdateKmcFlow create UpdateKmcTask instance
func NewUpdateKmcFlow(pathMgr *util.InstallDirPathMgr) *UpdateKmcFlow {
	return &UpdateKmcFlow{pathMgr: pathMgr}
}

func (muk *UpdateKmcFlow) getModules() []string {
	components := util.GetCompulsorySlice()
	return append(components, util.MefCenterRootName)
}

// RunFlow is the main func to start a task
func (muk *UpdateKmcFlow) RunFlow() error {
	var failedModule []string
	for _, module := range muk.getModules() {
		ctx, err := muk.initKmcCtx(module)
		if err != nil {
			return err
		}

		encryptedMap := muk.getEncryptMap()
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
		return nil
	}
	fmt.Printf("update module %s's kmc key failed\n", failedModule)

	return errors.New("update kmc key failed")
}

func (muk *UpdateKmcFlow) initKmcCtx(module string) (*kmc.Context, error) {
	var kmcKeyPath, kmcBackKeyPath string
	if module == util.MefCenterRootName {
		kmcKeyPath = muk.pathMgr.ConfigPathMgr.GetRootMasterKmcPath()
		kmcBackKeyPath = muk.pathMgr.ConfigPathMgr.GetRootBackKmcPath()
	} else {
		kmcKeyPath = muk.pathMgr.ConfigPathMgr.GetComponentMasterKmcPath(module)
		kmcBackKeyPath = muk.pathMgr.ConfigPathMgr.GetComponentBackKmcPath(module)
	}
	if err := fileutils.IsSoftLink(kmcKeyPath); err != nil {
		hwlog.RunLog.Errorf("check path [%s] failed: %s, cannot initKmc", kmcKeyPath, err.Error())
		return nil, err
	}
	if err := fileutils.IsSoftLink(kmcBackKeyPath); err != nil {
		hwlog.RunLog.Errorf("check path [%s] failed: %s, cannot initKmc", kmcBackKeyPath, err.Error())
		return nil, err
	}
	kmcConfig := kmc.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)

	config := kmc.NewKmcInitConfig()
	config.PrimaryKeyStoreFile = kmcConfig.PrimaryKeyPath
	config.StandbyKeyStoreFile = kmcConfig.StandbyKeyPath
	config.SdpAlgId = kmcConfig.SdpAlgID
	c, err := kmc.KeInitializeEx(config)
	if err != nil {
		hwlog.RunLog.Errorf("Init kmc failed: %v", err.Error())
		fmt.Println("init kmc failed")
		return nil, err
	}

	return &c, nil
}

func (muk *UpdateKmcFlow) getEncryptMap() map[string][]kmc.ReEncryptParam {
	return map[string][]kmc.ReEncryptParam{
		util.CertManagerName: {
			kmc.ReEncryptParam{
				Path:       muk.pathMgr.ConfigPathMgr.GetComponentConfigPath(util.CertManagerName),
				SuffixList: []string{kmc.KeySuffix},
			},
		},
		util.NginxManagerName: {
			kmc.ReEncryptParam{
				Path:       muk.pathMgr.ConfigPathMgr.GetComponentConfigPath(util.NginxManagerName),
				SuffixList: []string{kmc.KeySuffix},
			},
		},
		util.EdgeManagerName: {
			kmc.ReEncryptParam{
				Path:       muk.pathMgr.ConfigPathMgr.GetComponentConfigPath(util.EdgeManagerName),
				SuffixList: []string{kmc.KeySuffix},
			},
		},
		util.AlarmManagerName: {
			kmc.ReEncryptParam{
				Path:       muk.pathMgr.ConfigPathMgr.GetComponentConfigPath(util.AlarmManagerName),
				SuffixList: []string{kmc.KeySuffix},
			},
		},
		util.MefCenterRootName: {
			kmc.ReEncryptParam{
				Path: muk.pathMgr.ConfigPathMgr.GetRootCaKeyPath(),
			},
		},
	}
}
