// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package control

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const (
	upgradeFlow = "upgrade"
)

// GetSmoother func returns a smoother by a flow
func GetSmoother(flow string, pathMgr *util.InstallDirPathMgr) (Smoother, error) {
	if flow == upgradeFlow {
		return &upgradeSmoothMgr{
			smoothMgr: smoothMgr{
				InstallPathMgr: pathMgr,
			},
		}, nil
	}
	return nil, errors.New("unsupported flow type")
}

// Smoother is an interface to do smooth func in upgrade or rollback flow
type Smoother interface {
	smooth() error
}

type smoothMgr struct {
	InstallPathMgr *util.InstallDirPathMgr
}

type upgradeSmoothMgr struct {
	smoothMgr
}

func (usm *upgradeSmoothMgr) smooth() error {
	tasks := []func() error{
		usm.smoothAlarmManager,
	}

	for _, task := range tasks {
		if err := task(); err != nil {
			return err
		}
	}

	return nil
}

func (usm *upgradeSmoothMgr) smoothAlarmManager() error {
	if err := usm.smoothSingleComponent(util.AlarmManagerName); err != nil {
		return err
	}
	return nil
}

func (usm *upgradeSmoothMgr) smoothSingleComponent(component string) error {
	alarmConfigPath := usm.InstallPathMgr.ConfigPathMgr.GetComponentConfigPath(component)
	if utils.IsExist(alarmConfigPath) {
		if err := usm.setSingleOwner(component); err != nil {
			return err
		}

		return nil
	}

	componentMgr := util.GetComponentMgr(component)
	if err := componentMgr.PrepareComponentCertDir(usm.InstallPathMgr.GetConfigPath()); err != nil {
		return fmt.Errorf("prepare %s's cert dir failed: %s", component, err.Error())
	}

	rootCaFilePath := usm.InstallPathMgr.ConfigPathMgr.GetRootCaCertPath()
	rootPrivFilePath := usm.InstallPathMgr.ConfigPathMgr.GetRootCaKeyPath()
	kmcKeyPath := usm.InstallPathMgr.ConfigPathMgr.GetRootMasterKmcPath()
	kmcBackKeyPath := usm.InstallPathMgr.ConfigPathMgr.GetRootBackKmcPath()

	rootKmcCfg := kmc.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)
	rootCaMgr := certutils.InitRootCertMgr(rootCaFilePath, rootPrivFilePath, component, rootKmcCfg)
	if err := componentMgr.PrepareComponentCert(rootCaMgr, usm.InstallPathMgr.ConfigPathMgr); err != nil {
		return fmt.Errorf("prepare %s's cert failed: %s", component, err.Error())
	}

	if err := componentMgr.PrepareComponentConfig(usm.InstallPathMgr.ConfigPathMgr); err != nil {
		return fmt.Errorf("prepare %s's config failed: %s", component, err.Error())
	}

	if err := usm.setSingleOwner(component); err != nil {
		return err
	}

	return nil
}

func (usm *upgradeSmoothMgr) setSingleOwner(component string) error {
	mefUid, mefGid, err := util.GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get mef uid or gid failed: %s", err.Error())
		return errors.New("get mef uid or gid failed")
	}

	param := fileutils.SetOwnerParam{
		Path:       usm.InstallPathMgr.ConfigPathMgr.GetComponentConfigPath(component),
		Uid:        mefUid,
		Gid:        mefGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	if err = fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set alarm config right failed: %s", err.Error())
		return errors.New("set alarm config right failed")
	}

	return nil
}
