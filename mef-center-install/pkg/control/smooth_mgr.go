// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package control

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
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
		usm.setPriv,
		usm.smoothAlarmManager,
		usm.smoothBackupConfig,
	}

	for _, task := range tasks {
		if err := task(); err != nil {
			return err
		}
	}

	return nil
}

func (usm *upgradeSmoothMgr) setPriv() error {
	configPath := usm.installPathMgr.GetConfigPath()
	ownerChecker := fileutils.NewFileOwnerChecker(true, false, fileutils.RootUid, fileutils.RootGid)
	linkChecker := fileutils.NewFileLinkChecker(false)
	ownerChecker.SetNext(linkChecker)
	if err := fileutils.SetParentPathPermission(configPath, common.Mode755, ownerChecker); err != nil {
		hwlog.RunLog.Errorf("set install parent path permission failed: %s", err.Error())
		return errors.New("set install parent path permission failed")
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
	configPath := usm.installPathMgr.ConfigPathMgr.GetComponentConfigPath(component)
	if !fileutils.IsExist(configPath) {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.PrepareComponentCertDir(usm.installPathMgr.GetConfigPath()); err != nil {
			return fmt.Errorf("prepare %s's cert dir failed: %s", component, err.Error())
		}

		rootCaFilePath := usm.installPathMgr.ConfigPathMgr.GetRootCaCertPath()
		rootPrivFilePath := usm.installPathMgr.ConfigPathMgr.GetRootCaKeyPath()
		kmcKeyPath := usm.installPathMgr.ConfigPathMgr.GetRootMasterKmcPath()
		kmcBackKeyPath := usm.installPathMgr.ConfigPathMgr.GetRootBackKmcPath()

		rootKmcCfg := kmc.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)
		rootCaMgr := certutils.InitRootCertMgr(rootCaFilePath, rootPrivFilePath,
			common.MefCertCommonNamePrefix, rootKmcCfg)
		if err := componentMgr.PrepareComponentCert(rootCaMgr, usm.installPathMgr.ConfigPathMgr); err != nil {
			return fmt.Errorf("prepare %s's cert failed: %s", component, err.Error())
		}

		if err := componentMgr.PrepareComponentConfig(usm.installPathMgr.ConfigPathMgr); err != nil {
			return fmt.Errorf("prepare %s's config failed: %s", component, err.Error())
		}
	}

	if err := setSingleOwner(configPath); err != nil {
		return err
	}

	dealFunc, ok := postSmoothFuncMap[component]
	if ok {
		return dealFunc(usm.installPathMgr)
	}
	return nil
}

type postSmoothFunc func(installPathMgr *util.InstallDirPathMgr) error

var postSmoothFuncMap = map[string]postSmoothFunc{
	util.AlarmManagerName: postSmoothAlarmManager,
}

func postSmoothAlarmManager(installPathMgr *util.InstallDirPathMgr) (err error) {
	defer func() {
		if resetErr := util.ResetPriv(); resetErr != nil {
			err = resetErr
			hwlog.RunLog.Errorf("reset euid/gid back to root failed: %v", err)
		}
	}()
	if err := util.ReducePriv(); err != nil {
		return err
	}

	configDir := installPathMgr.GetConfigPath()
	alarmConfigDir := filepath.Join(configDir, util.AlarmManagerName)
	alarmDbMgr := common.NewDbMgr(alarmConfigDir, common.AlarmConfigDBName)
	alarmDbPath := filepath.Join(alarmConfigDir, common.AlarmConfigDBName)
	if fileutils.IsExist(alarmDbPath) {
		if err := alarmDbMgr.InitDB(); err != nil {
			return errors.New("init alarm manager database failed")
		}

		if err := database.CreateTableIfNotExist(common.AlarmConfig{}); err != nil {
			hwlog.RunLog.Errorf("create alarm config table failed, error: %v", err)
			return errors.New("create alarm config table failed")
		}
		return nil
	}

	if err := alarmDbMgr.InitDB(); err != nil {
		return errors.New("init alarm manager database failed")
	}
	if err := util.InitAndSetAlarmCfgTable(alarmConfigDir); err != nil {
		hwlog.RunLog.Errorf("init and set alarm config to table failed, error: %v", err)
		return errors.New("init and set alarm config to table failed")
	}
	if err := setSingleOwner(alarmDbPath); err != nil {
		hwlog.RunLog.Errorf("set alarm db owner failed, error: %v", err)
		return errors.New("set alarm db owner failed")
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

	if err := setSingleOwner(alarmLogPath); err != nil {
		return err
	}

	if err := setSingleOwner(alarmLogBackPath); err != nil {
		return err
	}

	return nil
}

func setSingleOwner(path string) error {
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

func (usm *upgradeSmoothMgr) smoothBackupConfig() error {
	path := usm.installPathMgr.ConfigPathMgr.GetConfigPath()
	if err := backuputils.NewBackupDirMgr(path, backuputils.JsonFileType, backuputils.CrtFileType,
		backuputils.CrlFileType, backuputils.KeyFileType).BackUp(); err != nil {
		hwlog.RunLog.Errorf("create mef backup dir failed: %s", err.Error())
		return fmt.Errorf("create mef backup dir failed: %s", err.Error())
	}

	if err := util.GetOwnerMgr(usm.installPathMgr.ConfigPathMgr).SetConfigOwner(); err != nil {
		hwlog.RunLog.Errorf("set owner for mef backup files failed: %s", err.Error())
		return fmt.Errorf("set owner for mef backup files failed: %s", err.Error())
	}
	return nil
}
