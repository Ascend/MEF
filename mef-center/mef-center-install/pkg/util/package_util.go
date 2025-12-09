// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

// ClearPakEnv if install failed ,need to clear package file
func ClearPakEnv(path string) {
	fmt.Println("install failed, start to clear environment")
	hwlog.RunLog.Info("-----Start to clear environment-----")
	if err := fileutils.DeleteAllFileWithConfusion(path); err != nil {
		fmt.Println("clear environment failed, please clear manually")
		hwlog.RunLog.Warnf("clear environment meets err:%s, need to do it manually", err.Error())
		hwlog.RunLog.Info("-----End to clear environment-----")
		return
	}
	fmt.Println("clear environment success")
	hwlog.RunLog.Info("clear environment success")
	hwlog.RunLog.Info("-----End to clear environment-----")
}
