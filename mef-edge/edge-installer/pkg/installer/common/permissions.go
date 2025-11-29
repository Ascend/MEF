// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common this file for get permission setting when install components
package common

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

const (
	all             = "all"
	dir             = "dir"
	file            = "file"
	service         = "*.service"
	target          = "*.target"
	db              = "*.db"
	dbJournal       = "*.db-journal"
	dbBackupJournal = "*.db.backup-journal"
	dbJsonBackup    = "*.backup"
	innerCertBackup = "**/*.backup"
	peerCertBackup  = "**/*/*.backup"
	log             = "*.log"
	json            = "*.json"
	sh              = "*.sh"
	txt             = "**/*.txt"
	tarGz           = "*.tar.gz"
	gz              = "*.gz"
	crt             = "**/*.crt"
	key             = "**/*.key"
	ks              = "**/*.ks"
	edgeCtl         = "edgectl"
	upgrade         = "upgrade"
	innerCtl        = "innerctl"
)

// UserCfg the struct of user configuration
type UserCfg struct {
	userUid uint32
	userGid uint32
	dirList []string
}

// ModeCfg the struct of mode configuration
type ModeCfg struct {
	mode    os.FileMode
	types   string
	dirList []string
}

func (pm *PermissionMgr) edgeInstallerUserCfg() UserCfg {
	dirList := []string{
		pm.WorkAbsPathMgr.GetCompWorkDir(pm.CompName),
		pm.LogPathMgr.GetComponentLogDir(pm.CompName),
		pm.LogPathMgr.GetComponentLogBackupDir(pm.CompName),
	}
	dirList = append(dirList, pm.getCfgDirs(pm.ConfigPathMgr)...)
	return UserCfg{
		userUid: constants.RootUserUid,
		userGid: constants.RootUserGid,
		dirList: dirList,
	}
}

func (pm *PermissionMgr) edgeOmUserCfg() UserCfg {
	dirList := []string{
		pm.WorkAbsPathMgr.GetCompWorkDir(pm.CompName),
		pm.LogPathMgr.GetComponentLogDir(pm.CompName),
		pm.LogPathMgr.GetComponentLogBackupDir(pm.CompName),
	}
	dirList = append(dirList, pm.getCfgDirs(pm.ConfigPathMgr)...)
	return UserCfg{
		userUid: constants.RootUserUid,
		userGid: constants.RootUserGid,
		dirList: dirList,
	}
}

func (pm *PermissionMgr) edgeCoreUserCfg() UserCfg {
	dirList := []string{
		pm.WorkAbsPathMgr.GetCompWorkDir(pm.CompName),
		pm.LogPathMgr.GetComponentLogDir(pm.CompName),
		pm.LogPathMgr.GetComponentLogBackupDir(pm.CompName),
	}
	dirList = append(dirList, pm.getCfgDirs(pm.ConfigPathMgr)...)
	return UserCfg{
		userUid: constants.RootUserUid,
		userGid: constants.RootUserGid,
		dirList: dirList,
	}
}

func (pm *PermissionMgr) devicePluginUserCfg() UserCfg {
	return UserCfg{
		userUid: constants.RootUserUid,
		userGid: constants.RootUserGid,
		dirList: []string{
			pm.WorkAbsPathMgr.GetCompWorkDir(pm.CompName),
			pm.LogPathMgr.GetComponentLogDir(pm.CompName),
			pm.LogPathMgr.GetComponentLogBackupDir(pm.CompName),
		},
	}
}

func (pm *PermissionMgr) edgeMainUserCfg() (UserCfg, error) {
	edgeUserId, edgeGroupId, err := util.GetMefId()
	if err != nil {
		return UserCfg{}, fmt.Errorf("get edge user id or group id failed, error: %v", err)
	}
	if uint64(edgeUserId) > math.MaxInt || uint64(edgeGroupId) > math.MaxInt {
		return UserCfg{}, errors.New("edge user id or edge group id is out of range")
	}
	dirList := []string{
		pm.WorkAbsPathMgr.GetCompWorkDir(pm.CompName),
		pm.LogPathMgr.GetComponentLogDir(pm.CompName),
		pm.LogPathMgr.GetComponentLogBackupDir(pm.CompName),
	}
	dirList = append(dirList, pm.getCfgDirs(pm.ConfigPathMgr)...)
	return UserCfg{
		userUid: edgeUserId,
		userGid: edgeGroupId,
		dirList: dirList,
	}, nil
}

// GetUserCfgMap get user config map for setting owner and group
func (pm *PermissionMgr) GetUserCfgMap() (map[string]UserCfg, error) {
	edgeMainUserCfg, err := pm.edgeMainUserCfg()
	if err != nil {
		return nil, err
	}
	userCfgMap := map[string]UserCfg{
		constants.EdgeInstaller: pm.edgeInstallerUserCfg(),
		constants.EdgeOm:        pm.edgeOmUserCfg(),
		constants.EdgeMain:      edgeMainUserCfg,
		constants.EdgeCore:      pm.edgeCoreUserCfg(),
		constants.DevicePlugin:  pm.devicePluginUserCfg(),
	}
	return userCfgMap, nil
}

// GetModeCfgMap get mode config map for setting mode
func (pm *PermissionMgr) GetModeCfgMap() map[string][]ModeCfg {
	modeCfgMap := map[string][]ModeCfg{
		constants.EdgeInstaller: pm.getEdgeInstallerModeList(),
		constants.EdgeOm:        pm.getEdgeOmModeList(),
		constants.EdgeMain:      pm.getEdgeMainModeList(),
		constants.EdgeCore:      pm.getEdgeCoreModeList(),
		constants.DevicePlugin:  pm.getDevicePluginModeList(),
	}
	return modeCfgMap
}

func (pm *PermissionMgr) getMode400() ModeCfg {
	cfg := pm.getMode400WithoutCfg()
	cfg.dirList = append(cfg.dirList, pm.getCfgDirs(pm.ConfigPathMgr)...)
	return cfg
}

func (pm *PermissionMgr) getMode400WithoutCfg() ModeCfg {
	return ModeCfg{
		mode:  constants.Mode400,
		types: all,
		dirList: []string{
			pm.WorkAbsPathMgr.GetCompWorkDir(pm.CompName),
			pm.LogPathMgr.GetComponentLogDir(pm.CompName),
			pm.LogPathMgr.GetComponentLogBackupDir(pm.CompName),
		},
	}
}

func (pm *PermissionMgr) getMode700() ModeCfg {
	cfg := pm.getMode700WithoutCfg()
	cfg.dirList = append(cfg.dirList, pm.getCfgDirs(pm.ConfigPathMgr)...)
	return cfg
}

func (pm *PermissionMgr) getMode700WithoutCfg() ModeCfg {
	return ModeCfg{
		mode:  constants.Mode700,
		types: dir,
		dirList: []string{
			pm.WorkAbsPathMgr.GetCompWorkDir(pm.CompName),
		},
	}
}

func (pm *PermissionMgr) getMode750() ModeCfg {
	return ModeCfg{
		mode:  constants.Mode750,
		types: dir,
		dirList: []string{
			pm.LogPathMgr.GetComponentLogDir(pm.CompName),
			pm.LogPathMgr.GetComponentLogBackupDir(pm.CompName),
		},
	}
}

func (pm *PermissionMgr) get600ConfigFileList() ModeCfg {
	fileTypes := []string{json, ks, db, dbJournal, dbBackupJournal, dbJsonBackup, innerCertBackup, peerCertBackup}
	var dirList []string
	return ModeCfg{
		mode:    constants.Mode600,
		types:   file,
		dirList: pm.getDirList(dirList, fileTypes),
	}
}

func (pm *PermissionMgr) getEdgeInstallerModeList() []ModeCfg {
	return []ModeCfg{
		pm.getMode400(),
		pm.getMode700(),
		pm.getMode750(),
		{
			mode:  constants.Mode600,
			types: file,
			dirList: []string{
				pm.WorkAbsPathMgr.GetServicePath(service),
				pm.WorkAbsPathMgr.GetServicePath(target),
			},
		},
		{
			mode:  constants.Mode640,
			types: file,
			dirList: []string{
				pm.LogPathMgr.GetComponentLogPath(pm.CompName, log),
			},
		},
		{
			mode:  constants.Mode400,
			types: file,
			dirList: []string{
				pm.LogPathMgr.GetComponentLogBackupPath(pm.CompName, gz),
				pm.WorkAbsPathMgr.GetCompScriptFilePath(pm.CompName, txt),
			},
		},
		{
			mode:  constants.Mode500,
			types: file,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompBinFilePath(pm.CompName, edgeCtl),
				pm.WorkAbsPathMgr.GetCompBinFilePath(pm.CompName, upgrade),
				pm.WorkAbsPathMgr.GetCompBinFilePath(pm.CompName, innerCtl),
				pm.WorkAbsPathMgr.GetCompScriptFilePath(pm.CompName, sh),
				pm.WorkAbsPathMgr.GetCompScriptFilePath(pm.CompName, filepath.Join(constants.DockerIsolate, sh)),
			},
		},
		{
			mode:  constants.Mode500,
			types: dir,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompBinDir(pm.CompName),
				pm.WorkAbsPathMgr.GetCompScriptDir(pm.CompName),
				pm.WorkAbsPathMgr.GetCompScriptFilePath(pm.CompName, constants.DockerIsolate),
			},
		},
	}
}

func (pm *PermissionMgr) getEdgeOmModeList() []ModeCfg {
	fileTypes := []string{constants.PodCfgFile, constants.ContainerCfgFile, crt, key}
	dirList := []string{pm.LogPathMgr.GetComponentLogBackupPath(pm.CompName, gz)}
	return []ModeCfg{
		pm.getMode400(),
		pm.getMode700(),
		pm.getMode750(),
		pm.get600ConfigFileList(),
		{
			mode:  constants.Mode640,
			types: file,
			dirList: []string{
				pm.LogPathMgr.GetComponentLogPath(pm.CompName, log),
			},
		},
		{
			mode:    constants.Mode400,
			types:   file,
			dirList: pm.getDirList(dirList, fileTypes),
		},
		{
			mode:  constants.Mode500,
			types: file,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompBinFilePath(pm.CompName, constants.EdgeOmFileName),
			},
		},
		{
			mode:  constants.Mode500,
			types: dir,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompBinDir(pm.CompName),
			},
		},
	}
}

func (pm *PermissionMgr) getEdgeMainModeList() []ModeCfg {
	fileTypes := []string{crt, key}
	dirList := []string{pm.LogPathMgr.GetComponentLogBackupPath(pm.CompName, gz)}
	return []ModeCfg{
		pm.getMode400(),
		pm.getMode700(),
		pm.getMode750(),
		pm.get600ConfigFileList(),
		{
			mode:  constants.Mode640,
			types: file,
			dirList: []string{
				pm.LogPathMgr.GetComponentLogPath(pm.CompName, log),
			},
		},
		{
			mode:    constants.Mode400,
			types:   file,
			dirList: pm.getDirList(dirList, fileTypes),
		},
		{
			mode:  constants.Mode500,
			types: file,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompBinFilePath(pm.CompName, constants.EdgeMainFileName),
			},
		},
		{
			mode:  constants.Mode500,
			types: dir,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompBinDir(pm.CompName),
			},
		},
	}
}

func (pm *PermissionMgr) getEdgeCoreModeList() []ModeCfg {
	fileTypes := []string{crt, key}
	dirList := []string{pm.LogPathMgr.GetComponentLogBackupPath(pm.CompName, gz)}
	return []ModeCfg{
		pm.getMode400(),
		pm.getMode700(),
		pm.getMode750(),
		pm.get600ConfigFileList(),
		{
			mode:  constants.Mode600,
			types: file,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompBinFilePath(pm.CompName, tarGz),
			},
		},
		{
			mode:  constants.Mode640,
			types: file,
			dirList: []string{
				pm.LogPathMgr.GetComponentLogPath(pm.CompName, log),
			},
		},
		{
			mode:    constants.Mode400,
			types:   file,
			dirList: pm.getDirList(dirList, fileTypes),
		},
		{
			mode:  constants.Mode500,
			types: file,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompScriptFilePath(pm.CompName, sh),
				pm.WorkAbsPathMgr.GetCompBinFilePath(pm.CompName, constants.EdgeCoreFileName),
			},
		},
		{
			mode:  constants.Mode500,
			types: dir,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompScriptDir(pm.CompName),
			},
		},
	}
}

func (pm *PermissionMgr) getDevicePluginModeList() []ModeCfg {
	return []ModeCfg{
		pm.getMode400WithoutCfg(),
		pm.getMode700WithoutCfg(),
		pm.getMode750(),
		{
			mode:  constants.Mode640,
			types: file,
			dirList: []string{
				pm.LogPathMgr.GetComponentLogPath(pm.CompName, log),
			},
		},
		{
			mode:  constants.Mode400,
			types: file,
			dirList: []string{
				pm.LogPathMgr.GetComponentLogBackupPath(pm.CompName, gz),
			},
		},
		{
			mode:  constants.Mode500,
			types: file,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompScriptFilePath(pm.CompName, sh),
				pm.WorkAbsPathMgr.GetCompBinFilePath(pm.CompName, constants.DevicePluginFileName),
			},
		},
		{
			mode:  constants.Mode500,
			types: dir,
			dirList: []string{
				pm.WorkAbsPathMgr.GetCompBinDir(pm.CompName),
				pm.WorkAbsPathMgr.GetCompScriptDir(pm.CompName),
			},
		},
	}
}

func (pm *PermissionMgr) getCfgDirs(configPathMgr *pathmgr.ConfigPathMgr) []string {
	cfgBackupTempDir := configPathMgr.GetConfigBackupTempDir()
	if fileutils.IsExist(cfgBackupTempDir) {
		return []string{filepath.Join(cfgBackupTempDir, pm.CompName)}
	}

	cfgDirs := []string{configPathMgr.GetCompConfigDir(pm.CompName)}
	cfgBackupDir := configPathMgr.GetConfigBackupDir()
	if fileutils.IsExist(cfgBackupDir) {
		cfgDirs = append(cfgDirs, filepath.Join(cfgBackupDir, pm.CompName))
	}
	return cfgDirs
}

func (pm *PermissionMgr) getDirList(dirList, fileTypes []string) []string {
	cfgDirs := pm.getCfgDirs(pm.ConfigPathMgr)
	for _, cfgDir := range cfgDirs {
		for _, fileType := range fileTypes {
			dirList = append(dirList, filepath.Join(cfgDir, fileType))
		}
	}
	return dirList
}
