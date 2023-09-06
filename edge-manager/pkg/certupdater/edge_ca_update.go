// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater cert update control module
package certupdater

import (
	"sync"

	"huawei.com/mindxedge/base/common"
)

type edgeCaUpdater struct {
}

// indicate update operation is stopped by normal way, not an error.
var edgeCaNormalStopFlag bool
var edgeCaCertUpdateFlag int64 = 0
var nodesChangeForCaChan = make(chan changedNodeInfo, workingQueueSize)
var updateResultForCaChan = make(chan NodeCertUpdateResult, common.MaxNode)
var forceUpdateCaCertChan = make(chan CertUpdatePayload)
var edgeCaWorkingLocker = sync.Mutex{}
var edgeCaUpdaterInstance edgeCaUpdater

// StartEdgeCaCertUpdate  entry for edge root ca cert update operation
func StartEdgeCaCertUpdate(payload *CertUpdatePayload) {
}
