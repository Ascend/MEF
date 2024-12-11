// Copyright (c) 2024. Huawei Technologies Co., Ltd.  All rights reserved.

package certmanager

import (
	"context"

	"huawei.com/mindxedge/base/common"
)

func newMockEdgeSvcCertUpdater() *EdgeSvcCertUpdater {
	var instance EdgeSvcCertUpdater
	instance.CaCertName = common.WsCltName
	instance.ctx, instance.cancel = context.WithCancel(context.Background())
	return &instance
}

func newMockEdgeCaCertUpdater() *EdgeCaCertUpdater {
	var instance EdgeCaCertUpdater
	instance.CaCertName = common.WsCltName
	instance.ctx, instance.cancel = context.WithCancel(context.Background())
	return &instance
}
