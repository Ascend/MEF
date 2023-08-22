// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

import (
	"context"
	"time"
)

var (
	alwaysOpenEmptyChannel = make(chan struct{})
	alwaysOpenTimeChannel  = make(chan time.Time)
	closedEmptyChannel     = make(chan struct{})
	closedContext          context.Context
)

func init() {
	close(closedEmptyChannel)

	var cancel context.CancelFunc
	closedContext, cancel = context.WithCancel(context.Background())
	cancel()
}
