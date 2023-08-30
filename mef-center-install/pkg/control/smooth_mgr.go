// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package control

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const (
	upgradeFlow = "upgrade"
)

// GetSmoother func returns a smoother by a flow
func GetSmoother(flow string, pathMgr *util.InstallDirPathMgr, logPathMgr *util.LogDirPathMgr) (Smoother, error) {
	if flow == upgradeFlow {
		return &upgradeSmoothMgr{
			smoothMgr: smoothMgr{
				installPathMgr: pathMgr,
				logPathMgr:     logPathMgr,
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
	installPathMgr *util.InstallDirPathMgr
	logPathMgr     *util.LogDirPathMgr
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
	if err := usm.smoothSingleComponentConfig(util.AlarmManagerName); err != nil {
		return err
	}

	if err := usm.smoothSingleComponentLog(util.AlarmManagerName); err != nil {
		return err
	}
	return nil
}

func (usm *upgradeSmoothMgr) smoothSingleComponentConfig(component string) error {
	alarmConfigPath := usm.installPathMgr.ConfigPathMgr.GetComponentConfigPath(component)
	if !fileutils.IsExist(alarmConfigPath) {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.PrepareComponentCertDir(usm.installPathMgr.GetConfigPath()); err != nil {
			return fmt.Errorf("prepare %s's cert dir failed: %s", component, err.Error())
		}

		rootCaFilePath := usm.installPathMgr.ConfigPathMgr.GetRootCaCertPath()
		rootPrivFilePath := usm.installPathMgr.ConfigPathMgr.GetRootCaKeyPath()
		kmcKeyPath := usm.installPathMgr.ConfigPathMgr.GetRootMasterKmcPath()
		kmcBackKeyPath := usm.installPathMgr.ConfigPathMgr.GetRootBackKmcPath()

		rootKmcCfg := kmc.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)
		rootCaMgr := certutils.InitRootCertMgr(rootCaFilePath, rootPrivFilePath, component, rootKmcCfg)
		if err := componentMgr.PrepareComponentCert(rootCaMgr, usm.installPathMgr.ConfigPathMgr); err != nil {
			return fmt.Errorf("prepare %s's cert failed: %s", component, err.Error())
		}

		if err := componentMgr.PrepareComponentConfig(usm.installPathMgr.ConfigPathMgr); err != nil {
			return fmt.Errorf("prepare %s's config failed: %s", component, err.Error())
		}
	}

	if err := usm.setSingleOwner(alarmConfigPath); err != nil {
		return err
	}

	return nil
}

func (usm *upgradeSmoothMgr) smoothSingleComponentLog(component string) error {
	alarmLogPath := usm.logPathMgr.GetComponentLogPath(component)
	alarmLogBackPath := usm.logPathMgr.GetComponentBackupLogPath(component)
	if !fileutils.IsExist(alarmLogPath) {
		if err := fileutils.CreateDir(alarmLogPath, fileutils.Mode600); err != nil {
			hwlog.RunLog.Errorf("create component %s's log dir failed: %s", component, err.Error())
			return fmt.Errorf("create component %s's log dir failed", component)
		}
		hwlog.RunLog.Infof("create component %s's log dir success", component)
	}

	if !fileutils.IsExist(alarmLogBackPath) {
		if err := fileutils.CreateDir(alarmLogBackPath, fileutils.Mode600); err != nil {
			hwlog.RunLog.Errorf("create component %s's log back up dir failed: %s", component, err.Error())
			return fmt.Errorf("create component %s's log back up dir failed", component)
		}
		hwlog.RunLog.Infof("create component %s's log back up dir success", component)
	}

	if err := usm.setSingleOwner(alarmLogPath); err != nil {
		return err
	}

	if err := usm.setSingleOwner(alarmLogBackPath); err != nil {
		return err
	}

	return nil
}

func (usm *upgradeSmoothMgr) setSingleOwner(path string) error {
	mefUid, mefGid, err := util.GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get mef uid or gid failed: %s", err.Error())
		return errors.New("get mef uid or gid failed")
	}

	param := fileutils.SetOwnerParam{
		Path:       path,
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
