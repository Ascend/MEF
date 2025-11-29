// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// package alarm this file for register mef alarm manager

package alarm

import (
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/configpara"
)

func (am *alarmManager) registerManager() error {
	for i := 0; i < constants.TryConnectNet; i++ {
		netTypeStr, err := configpara.GetNetType()
		if err != nil {
			time.Sleep(constants.StartWsWaitTime)
			continue
		}
		hwlog.RunLog.Infof("current netType: %s", netTypeStr)
		switch netTypeStr {
		case constants.MEF:
			am.proxyManger = NewAlarmMEFManager(am.ctx)
			return nil
		default:
			time.Sleep(constants.StartWsWaitTime)
			continue
		}
	}
	return errors.New("get netType failed, reached the maximum number of the connection attempts")
}
