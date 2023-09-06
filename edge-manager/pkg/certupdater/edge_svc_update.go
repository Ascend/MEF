// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certupdater cert update control module
package certupdater

import (
	"sync"

	"huawei.com/mindxedge/base/common"
)

type edgeSvcUpdater struct {
}

// edgeSvcNormalStopFlag: indicate update operation is stopped by normal way, not an error.
var edgeSvcNormalStopFlag bool
var edgeSvcCertUpdateFlag int64 = 0
var nodesChangeForSvcChan = make(chan changedNodeInfo, workingQueueSize)
var updateResultForSvcChan = make(chan NodeCertUpdateResult, common.MaxNode)
var forceUpdateSvcCertChan = make(chan CertUpdatePayload)
var edgeSvcWorkingLocker = sync.Mutex{}
var edgeSvcUpdaterInstance edgeSvcUpdater

// StartEdgeSvcCertUpdate  entry for edge service cert update operation
func StartEdgeSvcCertUpdate(payload *CertUpdatePayload) {
}
