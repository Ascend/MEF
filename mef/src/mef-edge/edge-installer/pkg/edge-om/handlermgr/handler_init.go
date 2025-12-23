// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
