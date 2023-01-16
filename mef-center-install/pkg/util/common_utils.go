// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import "huawei.com/mindxedge/base/common"

// GetArch is used to get the arch info
func GetArch() (string, error) {
	arch, err := common.RunCommand(ArchCommand, true, "-i")
	if err != nil {
		return "", err
	}
	return arch, nil
}
