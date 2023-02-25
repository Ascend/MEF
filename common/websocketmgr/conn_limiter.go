// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package websocketmgr for websocket manager
package websocketmgr

import (
	"fmt"

	"huawei.com/mindxedge/base/common"
)

var clientLimitNum = common.MaxNode

// InitConnLimiter initialize max limiter number
func InitConnLimiter(limitNum int) error {
	if limitNum > 0 && limitNum <= common.MaxNode {
		clientLimitNum = limitNum
		return nil
	}
	return fmt.Errorf("invalid edge client limit num: %v", limitNum)
}

// RemoveClientNum sub one from the clients counter
func RemoveClientNum(sp *WsServerProxy) {
	sp.CounterLock.Lock()
	defer sp.CounterLock.Unlock()
	sp.ClientNum -= 1
}

// CheckAndAddClientNum if check passed, atomically adds one to clients counter
// if check failed, do nothing and return false
func CheckAndAddClientNum(sp *WsServerProxy) bool {
	sp.CounterLock.Lock()
	defer sp.CounterLock.Unlock()
	if sp.ClientNum < clientLimitNum {
		sp.ClientNum += 1
		return true
	}
	return false
}
