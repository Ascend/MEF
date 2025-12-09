// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common for log path permissions setting
package common

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/envutils"

	"edge-installer/pkg/common/constants"
)

// LogPerm the permissions of log path
type LogPerm struct {
	userUid   uint32
	userGid   uint32
	modeUmask os.FileMode
	dir       string
}

// GetLogPermList get modeUmask config map for setting modeUmask
func (lpm LogPermissionMgr) GetLogPermList() []LogPerm {
	logPermList := []LogPerm{
		lpm.getMefEdgePerm(),
		lpm.getEdgeInstallerPerm(),
		lpm.getEdgeOmPerm(),
		lpm.getEdgeCorePerm(),
		lpm.getDevicePluginPerm(),
	}
	return logPermList
}

func (lpm LogPermissionMgr) getMefEdgePerm() LogPerm {
	return LogPerm{
		userUid:   constants.RootUserUid,
		userGid:   constants.RootUserGid,
		modeUmask: constants.ModeUmask022,
		dir:       lpm.LogPath,
	}
}

func (lpm LogPermissionMgr) getEdgeInstallerPerm() LogPerm {
	return LogPerm{
		userUid:   constants.RootUserUid,
		userGid:   constants.RootUserGid,
		modeUmask: constants.ModeUmask027,
		dir:       filepath.Join(lpm.LogPath, constants.EdgeInstaller),
	}
}

func (lpm LogPermissionMgr) getEdgeOmPerm() LogPerm {
	return LogPerm{
		userUid:   constants.RootUserUid,
		userGid:   constants.RootUserGid,
		modeUmask: constants.ModeUmask027,
		dir:       filepath.Join(lpm.LogPath, constants.EdgeOm),
	}
}

func (lpm LogPermissionMgr) getEdgeMainPerm() (LogPerm, error) {
	edgeUserId, err := envutils.GetUid(constants.EdgeUserName)
	if err != nil {
		return LogPerm{}, fmt.Errorf("get edge user id failed, error: %v", err)
	}
	edgeGroupId, err := envutils.GetGid(constants.EdgeUserGroup)
	if err != nil {
		return LogPerm{}, fmt.Errorf("get edge group id failed, error: %v", err)
	}
	if uint64(edgeUserId) > math.MaxInt || uint64(edgeGroupId) > math.MaxInt {
		return LogPerm{}, errors.New("edge user id or edge group id is out of range")
	}
	return LogPerm{
		userUid:   edgeUserId,
		userGid:   edgeGroupId,
		modeUmask: constants.ModeUmask027,
		dir:       filepath.Join(lpm.LogPath, constants.EdgeMain),
	}, nil
}

func (lpm LogPermissionMgr) getEdgeCorePerm() LogPerm {
	return LogPerm{
		userUid:   constants.RootUserUid,
		userGid:   constants.RootUserGid,
		modeUmask: constants.ModeUmask027,
		dir:       filepath.Join(lpm.LogPath, constants.EdgeCore),
	}
}

func (lpm LogPermissionMgr) getDevicePluginPerm() LogPerm {
	return LogPerm{
		userUid:   constants.RootUserUid,
		userGid:   constants.RootUserGid,
		modeUmask: constants.ModeUmask027,
		dir:       filepath.Join(lpm.LogPath, constants.DevicePlugin),
	}
}
