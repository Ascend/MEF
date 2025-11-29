// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

package innercommands

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/common/tasks"
)

const (
	curConfigTmp = "cur_config_temp"
	newConfigTmp = "new_config_temp"
)

// RestoreCfg is used to restore default config
type RestoreCfg struct {
	configPathMgr     *pathmgr.ConfigPathMgr
	curCfgTmpDir      string
	newCfgTmpDir      string
	clearCurCfgTmpDir bool
}

// NewRestoreCfg an RestoreCfg struct
func NewRestoreCfg(configPathMgr *pathmgr.ConfigPathMgr) *RestoreCfg {
	return &RestoreCfg{
		configPathMgr:     configPathMgr,
		curCfgTmpDir:      filepath.Join(configPathMgr.GetMefEdgeDir(), curConfigTmp),
		newCfgTmpDir:      filepath.Join(configPathMgr.GetMefEdgeDir(), newConfigTmp),
		clearCurCfgTmpDir: true,
	}
}

// Run restore default config task
func (rc *RestoreCfg) Run() error {
	var postFunc = []func() error{
		rc.stopService,
		rc.restoreCfg,
		rc.removeContainer,
	}
	for _, function := range postFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (rc *RestoreCfg) stopService() error {
	mgr := common.NewComponentMgr(rc.configPathMgr.GetInstallRootDir())
	if err := mgr.StopAll(); err != nil {
		return fmt.Errorf("stop service failed, error: %v", err)
	}
	return nil
}

func (rc *RestoreCfg) restoreCfg() error {
	defer rc.cleanEnv()
	if err := copyCfgDir(rc.configPathMgr.GetConfigBackupDir(), rc.newCfgTmpDir); err != nil {
		return fmt.Errorf("restore default config to config temp dir failed, error: %v", err)
	}

	if err := rc.writeSystemInfo(); err != nil {
		return err
	}

	if err := fileutils.RenameFile(rc.configPathMgr.GetConfigDir(), rc.curCfgTmpDir); err != nil {
		return fmt.Errorf("backup cur config dir failed, error: %v", err)
	}
	if err := fileutils.RenameFile(rc.newCfgTmpDir, rc.configPathMgr.GetConfigDir()); err != nil {
		rc.recoveryEnv()
		return fmt.Errorf("restore default config failed, error: %v", err)
	}

	hwlog.RunLog.Info("restore default config success")
	return nil
}

func (rc *RestoreCfg) writeSystemInfo() error {
	logRootDir, err := path.GetLogRootDir(rc.configPathMgr.GetInstallRootDir())
	if err != nil {
		return fmt.Errorf("get log root dir failed, error: %v", err)
	}
	logBackupRootDir, err := path.GetLogBackupRootDir(rc.configPathMgr.GetInstallRootDir())
	if err != nil {
		return fmt.Errorf("get log backup root dir failed, error: %v", err)
	}
	setSystemInfo := tasks.SetSystemInfoTask{
		ConfigDir:     rc.newCfgTmpDir,
		ConfigPathMgr: rc.configPathMgr,
		LogPathMgr:    pathmgr.NewLogPathMgr(logRootDir, logBackupRootDir),
	}
	if err = setSystemInfo.Run(); err != nil {
		return errors.New("set system info into config files failed")
	}
	return nil
}

func (rc *RestoreCfg) recoveryEnv() {
	fmt.Println("restore default config failed, recover current config now")
	hwlog.RunLog.Warn("restore default config failed, recover current config now")
	if err := fileutils.RenameFile(rc.curCfgTmpDir, rc.configPathMgr.GetConfigDir()); err != nil {
		rc.clearCurCfgTmpDir = false
		hwlog.RunLog.Errorf("recover current config [%s] failed, error: %v", rc.curCfgTmpDir, err)
		errMsg := fmt.Errorf("recover current config [%s] failed, "+
			"please manually recover it by renaming it to [%s]", rc.curCfgTmpDir, rc.configPathMgr.GetConfigDir())
		fmt.Println(errMsg)
		hwlog.RunLog.Error(errMsg)
	}
}

func (rc *RestoreCfg) cleanEnv() {
	tmpDirs := []string{rc.curCfgTmpDir, rc.newCfgTmpDir}
	for _, tmpDir := range tmpDirs {
		if tmpDir == rc.curCfgTmpDir && !rc.clearCurCfgTmpDir {
			continue
		}
		if err := fileutils.DeleteAllFileWithConfusion(tmpDir); err != nil {
			warnMsg := fmt.Errorf("warning: remove config temp dir [%s] failed: %v, "+
				"please manually remove it", tmpDir, err)
			fmt.Println(warnMsg)
			hwlog.RunLog.Warn(warnMsg)
		} else {
			hwlog.RunLog.Infof("clean temp config dir [%s] success", tmpDir)
		}
	}
}

func (rc *RestoreCfg) removeContainer() error {
	return util.RemoveContainer()
}

func copyCfgDir(dirSrc, dirDst string) error {
	if err := fileutils.CreateDir(dirDst, constants.Mode755); err != nil {
		return fmt.Errorf("create dir [%s] failed, error: %v", dirDst, err)
	}

	compDirs := []string{
		constants.EdgeInstaller,
		constants.EdgeOm,
		constants.EdgeCore,
	}
	for _, dirName := range compDirs {
		subDirSrc := filepath.Join(dirSrc, dirName)
		subDirDst := filepath.Join(dirDst, dirName)
		if err := fileutils.CopyDir(subDirSrc, subDirDst); err != nil {
			return fmt.Errorf("copy dir [%s] failed: %v", subDirSrc, err)
		}
	}

	edgeMainDirSrc := filepath.Join(dirSrc, constants.EdgeMain)
	edgeMainDirDst := filepath.Join(dirDst, constants.EdgeMain)
	uid, gid, err := util.GetMefId()
	if err != nil {
		return fmt.Errorf("get uid or gid failed: %v", err)
	}
	if err = fileutils.CreateDir(edgeMainDirDst, constants.Mode700); err != nil {
		return fmt.Errorf("create dir [%s] failed, error: %v", edgeMainDirDst, err)
	}

	param := fileutils.SetOwnerParam{
		Path:       edgeMainDirDst,
		Uid:        uid,
		Gid:        gid,
		Recursive:  false,
		IgnoreFile: true,
	}
	if err = fileutils.SetPathOwnerGroup(param); err != nil {
		return fmt.Errorf("set dir [%s] owner and group failed, error: %v", edgeMainDirDst, err)
	}
	if _, err = envutils.RunCommandWithUser(constants.CpCmd, envutils.DefCmdTimeoutSec, uid, gid, "-r",
		edgeMainDirSrc, dirDst); err != nil {
		return fmt.Errorf("copy dir [%s] failed: %v", edgeMainDirSrc, err)
	}
	return nil
}
