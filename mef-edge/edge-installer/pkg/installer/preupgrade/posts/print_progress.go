// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
