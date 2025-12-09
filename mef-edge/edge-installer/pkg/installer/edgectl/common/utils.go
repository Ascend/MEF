// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package common

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

// InitEdgeOmResource get init database of edge-om
func InitEdgeOmResource() error {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed, error: %v", err)
		return errors.New("get config path manager failed")
	}
	dbPath := configPathMgr.GetEdgeOmDbPath()
	if err = database.InitDB(dbPath); err != nil {
		hwlog.RunLog.Errorf("init database failed, error: %v", err)
		return errors.New("init database failed")
	}
	if err = database.CreateTableIfNotExist(config.Configuration{}); err != nil {
		hwlog.RunLog.Errorf("table configurations create failed, error: %v", err)
		return errors.New("table configurations create failed")
	}
	return nil
}

// LockProcessFlag lock flag for process
func LockProcessFlag(flagPath string, operation string) error {
	processFlag := util.FlagLockInstance(flagPath, constants.ProcessFlag, operation)
	if err := processFlag.Lock(); err != nil {
		return fmt.Errorf("lock control process failed: %v", err)
	}
	return nil
}

// UnlockProcessFlag unlock flag for process
func UnlockProcessFlag(flagPath string, operation string) error {
	processFlag := util.FlagLockInstance(flagPath, constants.ProcessFlag, operation)
	if err := processFlag.Unlock(); err != nil {
		return fmt.Errorf("unlock control process failed: %v", err)
	}
	return nil
}
