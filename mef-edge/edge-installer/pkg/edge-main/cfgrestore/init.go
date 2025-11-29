// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package cfgrestore
package cfgrestore

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

const (
	maxGoRoutine = 1024
)

var (
	goRoutineCount int32
)

// cfgRestore config module
type cfgRestore struct {
	ctx    context.Context
	enable bool
	worker worker
}

// NewCfgRestore new status msg manager module
func NewCfgRestore(ctx context.Context, enable bool) model.Module {
	return &cfgRestore{
		enable: enable,
		ctx:    ctx,
		worker: worker{ctx: ctx},
	}
}

// Name module name
func (cr *cfgRestore) Name() string {
	return constants.CfgRestore
}

// Enable module enable
func (cr *cfgRestore) Enable() bool {
	return cr.enable
}

// Start module start running
func (cr *cfgRestore) Start() {
	for {
		select {
		case <-cr.ctx.Done():
			hwlog.RunLog.Info("----------------cfg restore module proxy exit-------------------")
			return
		default:
		}
		message, err := modulemgr.ReceiveMessage(cr.Name())
		if err != nil {
			hwlog.RunLog.Errorf("failed to receive message, %v", err)
			continue
		}
		if err := cr.dispatchMessage(message); err != nil {
			hwlog.RunLog.Errorf("failed to dispatch message, %v", err)
		}
	}
}

func (cr *cfgRestore) dispatchMessage(message *model.Message) error {
	if message.KubeEdgeRouter.Resource == constants.ActionPodsData &&
		message.KubeEdgeRouter.Operation == constants.OptDelete {
		if goRoutineCount > maxGoRoutine {
			hwlog.RunLog.Errorf("number of go routine over limit, discard message, router: %v, route: %v",
				message.Router, message.KubeEdgeRouter)
			return errors.New("number of go routine over limit")
		}
		atomic.AddInt32(&goRoutineCount, 1)
		go cr.worker.DeletePodsData()
		return nil
	}
	return fmt.Errorf("unable to dispatch message, route: %v", message.KubeEdgeRouter)
}
