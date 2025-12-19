// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

// GetOwnerMgr [method] for set owner for mef-center config dir or config-backup dir
func GetOwnerMgr(configPathMgr *ConfigPathMgr) *ownerMgr {
	return &ownerMgr{
		configPathMgr: configPathMgr,
	}
}

type ownerMgr struct {
	configPathMgr *ConfigPathMgr
	mefUid        uint32
	mefGid        uint32
}

func (om *ownerMgr) SetConfigOwner() error {
	if om.configPathMgr == nil {
		hwlog.RunLog.Error("configPathMgr is nil")
		return errors.New("configPathMgr is nil")
	}

	uid, gid, err := GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get mef uid or gid failed: %s", err.Error())
		return errors.New("get mef uid or gid failed")
	}
	om.mefUid, om.mefGid = uid, gid

	var prepareCertsTasks = []func() error{
		om.setComponentConfigOwnerGroup,
		om.setRootCaDirOwnerGroup,
	}

	fmt.Println("start to set owner of mef config dirs")
	for _, function := range prepareCertsTasks {
		if err := function(); err != nil {
			return err
		}
	}
	fmt.Println("set owner of mef config dirs success")
	return nil
}

func (om *ownerMgr) setComponentConfigOwnerGroup() error {
	param := fileutils.SetOwnerParam{
		Path:       om.configPathMgr.GetConfigPath(),
		Uid:        om.mefUid,
		Gid:        om.mefGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetConfigPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	param = fileutils.SetOwnerParam{
		Path:       om.configPathMgr.GetConfigPath(),
		Uid:        RootUid,
		Gid:        RootGid,
		Recursive:  false,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetConfigPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	return nil
}

func (om *ownerMgr) setRootCaDirOwnerGroup() error {
	param := fileutils.SetOwnerParam{
		Path:       om.configPathMgr.GetRootCaKeyDirPath(),
		Uid:        RootUid,
		Gid:        RootGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetRootCaKeyDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	param = fileutils.SetOwnerParam{
		Path:       om.configPathMgr.GetRootCaDirPath(),
		Uid:        RootUid,
		Gid:        RootGid,
		Recursive:  false,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetRootCaDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	if !fileutils.IsExist(om.configPathMgr.GetRootKmcDirPath()) {
		return nil
	}
	param = fileutils.SetOwnerParam{
		Path:       om.configPathMgr.GetRootKmcDirPath(),
		Uid:        RootUid,
		Gid:        RootGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetRootKmcDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	return nil
}
