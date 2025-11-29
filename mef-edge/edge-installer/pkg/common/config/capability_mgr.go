// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package config dir for config
package config

import (
	"fmt"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-installer/pkg/common/constants"
)

const (
	npuSharingConfigKey = "npu_sharing_config"
	npuSharingKey       = "npu_sharing"
	yesCmd              = "Y\n"
	okSymbol            = "OK"
	trueSymbol          = "True"
	// SharedNpuResName the 310B chip report by device-plugin
	SharedNpuResName = "huawei.com/Ascend310"
	// NpuScale is the scale for downstream message when npu sharing is turned on
	NpuScale = 2
)

var closeNpuSharingArgs = []string{"set", "-t", "device-share", "-i", "0", "-d", "0"}
var openNpuSharingArgs = []string{"set", "-t", "device-share", "-i", "0", "-d", "1"}
var queryNpuStatusArgs = []string{"info", "-t", "device-share", "-i", "0", "-c", "0"}
var queryAllNpuInfoArgs = []string{"info"}
var allowedDeviceNames = []string{"Atlas 500 A2", "Atlas 200I A2"}
var allowedChips = []string{"310B"}
var queryProductNameArgs = []string{"info", "-t", "product", "-i", "0"}

// GetScaledNpu get the npu res after scaled
func GetScaledNpu(resName v1.ResourceName, originVal resource.Quantity) resource.Quantity {
	if resName.String() != SharedNpuResName {
		return originVal
	}
	if !GetCapabilityCache().HasCapability(constants.CapabilityNpuSharingConfig) {
		return originVal
	}
	return *resource.NewScaledQuantity(originVal.Value(), -NpuScale)
}

// ModifyNpuRes modify the npu resources, up-true: mef to fd, false:fd to mef
// if the data not qualified,do nothing. no need to return error.
func ModifyNpuRes(resourceList v1.ResourceList, up bool) {
	if !GetCapabilityCache().HasCapability(constants.CapabilityNpuSharingConfig) {
		return
	}
	npuCount, ok := resourceList[constants.CenterNpuName]
	if !ok {
		return
	}

	// increase value to prevent rounding up npu
	scale := resource.Milli
	if up {
		scale -= NpuScale
	} else {
		scale += NpuScale
	}

	resourceList[constants.CenterNpuName] = *resource.NewScaledQuantity(npuCount.MilliValue(), scale)
}

// CapabilityMgr manager to manage all capabilities
// method below this line should invoke from edge-om process
type CapabilityMgr struct {
	capabilities map[string]CapabilityIntf
}

// CapabilityIntf represent a capability interface
type CapabilityIntf interface {
	load()
	getName() string
	open() error
	close() error
	query() bool
}

// CapabilityItem represent a concrete capability
type CapabilityItem struct {
	name   string
	enable bool
	mgr    *CapabilityMgr
}

type npuInfo struct {
	name string
}

var capabilityInstance *CapabilityMgr

// GetCapabilityMgr get a singleton of capability manager
func GetCapabilityMgr() *CapabilityMgr {
	if capabilityInstance != nil {
		return capabilityInstance
	}
	capabilityInstance = &CapabilityMgr{}
	capabilities := map[string]CapabilityIntf{
		npuSharingConfigKey: &NpuSharingConfigCapability{
			CapabilityItem: CapabilityItem{
				name: npuSharingConfigKey,
				mgr:  capabilityInstance,
			}},
		npuSharingKey: &NpuSharingCapability{
			CapabilityItem: CapabilityItem{
				name: npuSharingKey,
				mgr:  capabilityInstance,
			}},
	}
	capabilityInstance.capabilities = capabilities
	for _, capability := range capabilities {

		if capability != nil {
			capability.load()
		}
	}
	return capabilityInstance
}

// GetCaps get all capabilities of this device
func (c *CapabilityMgr) GetCaps() []string {
	var ret []string
	for _, capItr := range c.capabilities {
		if capItr.query() {
			ret = append(ret, capItr.getName())
		}
	}
	return ret
}

// Query to query if this device has a capability
func (c *CapabilityMgr) Query(name string) bool {
	capability, ok := c.capabilities[name]
	if !ok {
		return false
	}
	return capability.query()
}

// Switch open or close a capability through onOrOff variable
func (c *CapabilityMgr) Switch(name string, on bool) error {
	capability, ok := c.capabilities[name]
	if !ok {
		return fmt.Errorf("no target capability: %s", name)
	}
	if on {
		return capability.open()
	}
	return capability.close()
}

func (c *CapabilityItem) load() {
	c.enable = true
}

func (c *CapabilityItem) getName() string {
	return c.name
}

func (c *CapabilityItem) open() error {
	return nil
}

func (c *CapabilityItem) close() error {
	return nil
}

func (c *CapabilityItem) query() bool {
	return c.enable
}

// NpuSharingConfigCapability ability to config npu sharing
type NpuSharingConfigCapability struct {
	CapabilityItem
}

func (c *NpuSharingConfigCapability) load() {
	var npuRet string
	var err error
	if npuRet, err = envutils.RunCommand(constants.NpuSmiCmd, envutils.DefCmdTimeoutSec,
		queryProductNameArgs...); err != nil {
		c.enable = false
		hwlog.RunLog.Errorf("query product name failed, error: %s", err.Error())
		return
	}
	if !stringContainsAny(npuRet, allowedDeviceNames) {
		c.enable = false
		return
	}
	if npuRet, err = envutils.RunCommand(constants.NpuSmiCmd, envutils.DefCmdTimeoutSec,
		queryAllNpuInfoArgs...); err != nil {
		c.enable = false
		hwlog.RunLog.Errorf("query npu info failed, error: %s", err.Error())
		return
	}
	chipName, ok := findChipName(npuRet)
	if !ok {
		c.enable = false
		return
	}
	if !stringContainsAny(chipName, allowedChips) {
		c.enable = false
		return
	}
	c.enable = true
}

func (c *NpuSharingConfigCapability) getName() string {
	return c.name
}

func (c *NpuSharingConfigCapability) open() error {
	return c.CapabilityItem.open()
}

func (c *NpuSharingConfigCapability) close() error {
	return c.CapabilityItem.close()
}

func (c *NpuSharingConfigCapability) query() bool {
	return c.enable
}

// NpuSharingCapability to show if the npu sharing is open or close
type NpuSharingCapability struct {
	CapabilityItem
}

func (c *NpuSharingCapability) load() {
	var npuRet string
	var err error
	if npuRet, err = envutils.RunCommand(constants.NpuSmiCmd, envutils.DefCmdTimeoutSec,
		queryNpuStatusArgs...); err != nil {
		hwlog.RunLog.Errorf("load capability failed, error: %s", err.Error())
		return
	}
	npuRet = strings.Replace(npuRet, "\t", "", -1)
	npuRet = strings.Replace(npuRet, "\n", "", -1)
	statuses := strings.Split(npuRet, ":")
	minStatusLen := 2
	if len(statuses) < minStatusLen {
		c.enable = false
		return
	}
	status := strings.Trim(statuses[1], " ")
	if status == trueSymbol {
		c.enable = true
		return
	}
	c.enable = false
}

func (c *NpuSharingCapability) open() error {
	if !c.mgr.Query(npuSharingConfigKey) {
		return fmt.Errorf("cannot open at this device")
	}
	if c.enable {
		return fmt.Errorf("already open, do nothing")
	}
	ret, err := envutils.RunInteractCommand(constants.NpuSmiCmd, yesCmd, envutils.DefCmdTimeoutSec,
		openNpuSharingArgs...)
	if err != nil {
		hwlog.RunLog.Errorf("open npu sharing failed: %s", err.Error())
		return fmt.Errorf("open npu sharing failed: %s", err.Error())
	}
	if strings.Contains(ret, okSymbol) {
		c.enable = true
	}
	return nil
}

func (c *NpuSharingCapability) close() error {
	if !c.mgr.Query(npuSharingConfigKey) {
		return fmt.Errorf("cannot turn on at this type of machine")
	}
	if !c.enable {
		return fmt.Errorf("already close, no need to do so")
	}
	ret, err := envutils.RunInteractCommand(constants.NpuSmiCmd, yesCmd, envutils.DefCmdTimeoutSec,
		closeNpuSharingArgs...)
	if err != nil {
		hwlog.RunLog.Errorf("close npu sharing failed: %s", err.Error())
		return fmt.Errorf("close npu sharing failed: %s", err.Error())
	}
	if strings.Contains(ret, okSymbol) {
		c.enable = false
	}
	return nil
}

func (c *NpuSharingCapability) query() bool {
	if !c.mgr.Query(npuSharingConfigKey) {
		return false
	}
	c.load()
	return c.enable
}

func findChipName(npuRet string) (string, bool) {
	excludeWords := []string{"+", "NPU", "Chip", "npu-smi"}
	lines := strings.Split(npuRet, "\n")
	const maxLineCount = 100
	const shouldReadLine = 2
	const shouldItemsCount = 15
	var iterationCount int
	var npuReadLineCount int
	var npuContent []string
	var myNpuInfos []npuInfo
	for _, line := range lines {
		if iterationCount > maxLineCount {
			break
		}
		if stringContainsAny(line, excludeWords) {
			continue
		}
		if strings.Contains(line, "Version") {
			continue
		}
		replacedLine := strings.ReplaceAll(line, "|", "")
		splitLines := strings.Split(replacedLine, " ")
		npuContent = append(npuContent, splitLines...)
		var npuContentModified []string
		for _, content := range npuContent {
			if len(content) == 0 {
				continue
			}
			npuContentModified = append(npuContentModified, content)
		}
		npuReadLineCount++
		if npuReadLineCount != shouldReadLine {
			continue
		}
		npuReadLineCount = 0
		if len(npuContentModified) != shouldItemsCount {
			continue
		}
		nInfo := npuInfo{name: npuContentModified[1]}
		myNpuInfos = append(myNpuInfos, nInfo)
		iterationCount++
	}
	if len(myNpuInfos) != 1 {
		return "", false
	}
	return myNpuInfos[0].name, true
}

func stringContainsAny(originStr string, targets []string) bool {
	for _, target := range targets {
		if strings.Contains(originStr, target) {
			return true
		}
	}
	return false
}
