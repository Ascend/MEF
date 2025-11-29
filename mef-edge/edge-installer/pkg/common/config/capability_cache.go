// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package config

import (
	"context"
	"strings"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

var capabilityOnce sync.Once
var instance *CapabilityCache

// CapabilityCache a struct for capabilities cache
type CapabilityCache struct {
	capabilities sync.Map
	eventChan    chan interface{}
}

// GetCapabilityCache create or get a cache
func GetCapabilityCache() *CapabilityCache {
	capabilityOnce.Do(func() {
		const eventChanSize = 1024
		instance = &CapabilityCache{eventChan: make(chan interface{}, eventChanSize)}
	})
	return instance
}

// Set put a val
func (c *CapabilityCache) Set(key string, val bool) {
	c.capabilities.Store(key, val)
}

// SetEdgeOmCaps set capabilities by the messages from edge-om
func (c *CapabilityCache) SetEdgeOmCaps(info StaticInfo) {
	foundSharingConfig := false
	foundSharing := false
	for _, c := range info.ProductCapabilityEdge {
		if c == constants.CapabilityNpuSharingConfig {
			foundSharingConfig = true
		}
		if c == constants.CapabilityNpuSharing {
			foundSharing = true
		}
	}
	c.Set(constants.CapabilityNpuSharingConfig, foundSharingConfig)
	c.Set(constants.CapabilityNpuSharing, foundSharing)
}

// Notify notify a thread to report capability to fd
func (c *CapabilityCache) Notify() {
	c.eventChan <- struct{}{}
}

func (c *CapabilityCache) getCaps() []string {
	var effectiveCaps []string
	c.capabilities.Range(func(key, val interface{}) bool {
		capabilityName, ok := key.(string)
		if !ok {
			return true
		}
		valBool, ok := val.(bool)
		if !ok {
			return true
		}
		if valBool {
			effectiveCaps = append(effectiveCaps, capabilityName)
		}
		return true
	})
	return effectiveCaps
}

// HasCapability indicate if cache contains the capability
func (c *CapabilityCache) HasCapability(key string) bool {
	val, ok := c.capabilities.Load(key)
	if !ok {
		return false
	}
	valBool, ok := val.(bool)
	if !ok {
		return false
	}
	return valBool
}

// StartReportJob start job to report capabilities in cache to remote
func (c *CapabilityCache) StartReportJob(ctx context.Context) {
	hwlog.RunLog.Info("capability cache report job start")
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("capability cache report job stop")
			return
		case _, ok := <-c.eventChan:
			if ok {
				time.Sleep(time.Second)
				reportCapabilities()
			}
		}
	}
}

func reportCapabilities() {
	var toFdInfo StaticInfo
	toFdInfo.ProductCapabilityEdge = GetCapabilityCache().getCaps()
	if len(toFdInfo.ProductCapabilityEdge) == 0 {
		return
	}
	hwlog.RunLog.Infof("start to report capabilities: [%s]", strings.Join(toFdInfo.ProductCapabilityEdge, ","))
	newResponse, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("report capabilities failed, new message error: %v", err)
		return
	}
	newResponse.KubeEdgeRouter.Source = constants.SourceHardware
	newResponse.KubeEdgeRouter.Group = constants.GroupHub
	newResponse.KubeEdgeRouter.Operation = constants.OptUpdate
	newResponse.KubeEdgeRouter.Resource = constants.ResStatic
	newResponse.SetRouter("", constants.ModDeviceOm, "", "")
	if err = newResponse.FillContent(toFdInfo, true); err != nil {
		hwlog.RunLog.Errorf("fill fd intf content failed: %v", err)
		return
	}
	if err = modulemgr.SendMessage(newResponse); err != nil {
		hwlog.RunLog.Errorf("report capabilities failed, send message error: %v", err)
	}
}
