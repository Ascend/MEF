// Copyright (c) 2023. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
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
	if err := utils.SetPathOwnerGroup(om.configPathMgr.GetConfigPath(), om.mefUid,
		om.mefGid, true, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetConfigPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	if err := utils.SetPathOwnerGroup(om.configPathMgr.GetConfigPath(), RootUid,
		RootGid, false, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetConfigPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	return nil
}

func (om *ownerMgr) setRootCaDirOwnerGroup() error {
	if err := utils.SetPathOwnerGroup(om.configPathMgr.GetRootCaKeyDirPath(), RootUid,
		RootGid, true, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetRootCaKeyDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	if err := utils.SetPathOwnerGroup(om.configPathMgr.GetRootCaDirPath(), RootUid,
		RootGid, false, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetRootCaDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	if !utils.IsExist(om.configPathMgr.GetRootKmcDirPath()) {
		return nil
	}
	if err := utils.SetPathOwnerGroup(om.configPathMgr.GetRootKmcDirPath(), RootUid,
		RootGid, true, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			om.configPathMgr.GetRootKmcDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	return nil
}
