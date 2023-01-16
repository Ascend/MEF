// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"fmt"

	"huawei.com/mindxedge/base/common"
)

const (
	startCommand = "./nginx"
)

// Start do the start nginx job
func cmdStart() error {
	_, err := common.RunCommand(startCommand, true)
	if err != nil {
		return fmt.Errorf("start error is %v", err)
	}
	return nil
}
