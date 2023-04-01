// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"sync"
)

var (
	once         sync.Once
	whiteListMap = map[string]struct{}{}
)

// InWhiteList [method] for check out if target path is in whitelist
func InWhiteList(hostPath string, whiteList []string) bool {
	once.Do(func() {
		for _, white := range whiteList {
			whiteListMap[white] = struct{}{}
		}
	})
	_, ok := whiteListMap[hostPath]
	return ok
}
