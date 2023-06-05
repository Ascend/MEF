// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package kmcupdate this file for update kmc flow
package kmcupdate

import (
	"errors"
	"fmt"

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
	return []string{util.CertManagerName, util.NginxManagerName, util.EdgeManagerName, util.MefCenterRootName}
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
				ReEncryptedList: encryptedList,
				Ctx:             ctx,
			},
		}

		hwlog.RunLog.Infof("start to update module %s's kmc keys", module)
		if err = task.RunTask(); err != nil {
			hwlog.RunLog.Errorf("update module %s's kmc keys failed: %s", module, err.Error())
			failedModule = append(failedModule, module)
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

func (muk *UpdateKmcFlow) getEncryptMap() map[string][]string {
	return map[string][]string{
		util.CertManagerName: {
			muk.pathMgr.ConfigPathMgr.GetComponentKeyPath(util.CertManagerName),
			muk.pathMgr.ConfigPathMgr.GetHubSrvKeyPath(),
		},
		util.NginxManagerName: {
			muk.pathMgr.ConfigPathMgr.GetComponentKeyPath(util.NginxManagerName),
			muk.pathMgr.ConfigPathMgr.GetUserServerKeyPath(),
		},
		util.EdgeManagerName: {
			muk.pathMgr.ConfigPathMgr.GetComponentKeyPath(util.EdgeManagerName),
			muk.pathMgr.ConfigPathMgr.GetWebSocketKeyPath(),
		},
		util.MefCenterRootName: {muk.pathMgr.ConfigPathMgr.GetRootCaKeyPath()},
	}
}
