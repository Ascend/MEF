// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package envutils
package envutils

import (
	"context"
	"fmt"
	"testing"

	"huawei.com/mindx/common/hwlog"
)

func TestMain(m *testing.M) {
	if err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background()); err != nil {
		panic(err)
	}
	fmt.Printf("envutils dt exit %d\n", m.Run())
}
