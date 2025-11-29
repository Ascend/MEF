// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlermgr
package handlermgr

import "huawei.com/mindx/common/hwlog"

type initFunc func() error

var initFuncList = []initFunc{initConfig}

// Enable module enable
func (hm *handlerManger) Enable() bool {
	for _, fn := range initFuncList {
		if err := fn(); err != nil {
			hwlog.RunLog.Errorf("%s, config manager cannot enable", err.Error())
			return false
		}
	}

	return hm.enable
}
