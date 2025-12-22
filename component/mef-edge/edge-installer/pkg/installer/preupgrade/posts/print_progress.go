// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package posts for flow task post-processing print progress
package posts

import (
	"fmt"

	"edge-installer/pkg/installer/common"
)

// PrintProgress print progress
func PrintProgress(item *common.FlowItem) error {
	if item.Error != nil {
		fmt.Printf("progress:%d%%,%s,failed:%v\n", item.Progress, item.Description, item.Error)
		return nil
	}
	fmt.Printf("progress:%d%%,%s,finished\n", item.Progress, item.Description)
	return nil
}
