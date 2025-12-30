/* Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MEF is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
*/

// Package common a series of common function
package common

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"

	"huawei.com/mindx/common/hwlog"
)

// GetPattern return pattern map
func GetPattern() map[string]string {
	return map[string]string{
		"vir310p": "Ascend310P-(1|2|4)c",
	}
}

var (
	allDeviceInfoLock sync.Mutex
)

// LockAllDeviceInfo lock for device info status
func LockAllDeviceInfo() {
	allDeviceInfoLock.Lock()
}

// UnlockAllDeviceInfo unlock for device info status
func UnlockAllDeviceInfo() {
	allDeviceInfoLock.Unlock()
}

// MapDeepCopy map deep copy
func MapDeepCopy(source map[string]string) map[string]string {
	dest := make(map[string]string, len(source))
	if source == nil {
		return dest
	}
	for key, value := range source {
		dest[key] = value
	}
	return dest
}

func setDeviceByPathWhen200RC(defaultDevices *[]string) {
	setDeviceByPath(defaultDevices, HiAi200RCEventSched)
	setDeviceByPath(defaultDevices, HiAi200RCHiDvpp)
	setDeviceByPath(defaultDevices, HiAi200RCLog)
	setDeviceByPath(defaultDevices, HiAi200RCMemoryBandwidth)
	setDeviceByPath(defaultDevices, HiAi200RCSVM0)
	setDeviceByPath(defaultDevices, HiAi200RCTsAisle)
	setDeviceByPath(defaultDevices, HiAi200RCUpgrade)
}

func setDeviceByPath(defaultDevices *[]string, device string) {
	if _, err := os.Stat(device); err == nil {
		*defaultDevices = append(*defaultDevices, device)
	}
}

// GetDefaultDevices get default device, for allocate mount
func GetDefaultDevices(getFdFlag bool) ([]string, error) {
	davinciManager, err := getDavinciManagerPath()
	if err != nil {
		return nil, err
	}
	var defaultDevices []string
	defaultDevices = append(defaultDevices, davinciManager)

	setDeviceByPath(&defaultDevices, HiAIHDCDevice)
	setDeviceByPath(&defaultDevices, HiAISVMDevice)
	if getFdFlag {
		setDeviceByPathWhen200RC(&defaultDevices)
	}

	var productType string
	if len(ParamOption.ProductTypes) == 1 {
		productType = ParamOption.ProductTypes[0]
	}
	if productType == Atlas200ISoc {
		socDefaultDevices, err := set200SocDefaultDevices()
		if err != nil {
			hwlog.RunLog.Errorf("get 200I soc default devices failed, err: %#v", err)
			return nil, err
		}
		defaultDevices = append(defaultDevices, socDefaultDevices...)
	}
	if ParamOption.RealCardType == Ascend310B {
		a310BDefaultDevices := set310BDefaultDevices()
		defaultDevices = append(defaultDevices, a310BDefaultDevices...)
	}
	return defaultDevices, nil
}

func getDavinciManagerPath() (string, error) {
	if ParamOption.RealCardType == Ascend310B {
		if _, err := os.Stat(HiAIManagerDeviceDocker); err == nil {
			return HiAIManagerDeviceDocker, nil
		}
		hwlog.RunLog.Warn("get davinci manager docker failed")
	}

	if _, err := os.Stat(HiAIManagerDevice); err != nil {
		return "", err
	}
	return HiAIManagerDevice, nil
}

// set200SocDefaultDevices set 200 soc defaults devices
func set200SocDefaultDevices() ([]string, error) {
	var socDefaultDevices = []string{
		Atlas200ISocVPC,
		Atlas200ISocVDEC,
		Atlas200ISocSYS,
		Atlas200ISocSpiSmbus,
		Atlas200ISocUserConfig,
		HiAi200RCTsAisle,
		HiAi200RCSVM0,
		HiAi200RCLog,
		HiAi200RCMemoryBandwidth,
		HiAi200RCUpgrade,
	}
	for _, devPath := range socDefaultDevices {
		if _, err := os.Stat(devPath); err != nil {
			return nil, err
		}
	}
	var socOptionsDevices = []string{
		HiAi200RCEventSched,
		Atlas200ISocXSMEM,
	}
	for _, devPath := range socOptionsDevices {
		if _, err := os.Stat(devPath); err != nil {
			hwlog.RunLog.Warnf("device %s not exist", devPath)
			continue
		}
		socDefaultDevices = append(socDefaultDevices, devPath)
	}
	return socDefaultDevices, nil
}

func set310BDefaultDevices() []string {
	var a310BDefaultDevices = []string{
		Atlas310BDvppCmdlist,
		Atlas310BPngd,
		Atlas310BVenc,
		HiAi200RCUpgrade,
		Atlas200ISocSYS,
		HiAi200RCSVM0,
		Atlas200ISocVDEC,
		Atlas200ISocVPC,
		HiAi200RCTsAisle,
		HiAi200RCLog,
		Atlas310BAcodec,
		Atlas310BAi,
		Atlas310BAo,
		Atlas310BVo,
		Atlas310BHdmi,
	}
	var available310BDevices []string
	for _, devPath := range a310BDefaultDevices {
		if _, err := os.Stat(devPath); err != nil {
			hwlog.RunLog.Warnf("device %s not exist", devPath)
			continue
		}
		available310BDevices = append(available310BDevices, devPath)
	}
	return available310BDevices
}

// VerifyPathAndPermission used to verify the validity of the path and permission and return resolved absolute path
func VerifyPathAndPermission(verifyPath string) (string, bool) {
	hwlog.RunLog.Debug("starting check device socket file path.")
	absVerifyPath, err := filepath.Abs(verifyPath)
	if err != nil {
		hwlog.RunLog.Errorf("abs current path failed: %#v", err)
		return "", false
	}
	pathInfo, err := os.Stat(absVerifyPath)
	if err != nil {
		hwlog.RunLog.Errorf("abs current path failed: %#v", err)
		return "", false
	}
	realPath, err := filepath.EvalSymlinks(absVerifyPath)
	if err != nil {
		hwlog.RunLog.Errorf("evaluation of any symbolic failed: %#v", err)
		return "", false
	}
	if absVerifyPath != realPath {
		hwlog.RunLog.Error("Symlinks is not allowed")
		return "", false
	}

	stat, ok := pathInfo.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != RootUID || stat.Gid != RootGID {
		hwlog.RunLog.Error("Non-root owner group of the path")
		return "", false
	}
	return realPath, true
}

// NewFileWatch is used to watch socket file
func NewFileWatch() (*FileWatch, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatch{FileWatcher: watcher}, nil
}

// WatchFile add file to watch
func (fw *FileWatch) WatchFile(fileName string) error {
	if _, err := os.Stat(fileName); err != nil {
		return err
	}
	return fw.FileWatcher.Add(fileName)
}

// NewSignWatcher new sign watcher
func NewSignWatcher(osSigns ...os.Signal) chan os.Signal {
	// create signs chan
	signChan := make(chan os.Signal, 1)
	for _, sign := range osSigns {
		signal.Notify(signChan, sign)
	}
	return signChan
}

// CheckFileUserSameWithProcess to check whether the owner of the log file is the same as the uid
func CheckFileUserSameWithProcess(loggerPath string) bool {
	curUid := os.Getuid()
	if curUid == RootUID {
		return true
	}
	pathInfo, err := os.Lstat(loggerPath)
	if err != nil {
		path := filepath.Dir(loggerPath)
		pathInfo, err = os.Lstat(path)
		if err != nil {
			fmt.Printf("get logger path stat failed, error is %#v\n", err)
			return false
		}
	}
	stat, ok := pathInfo.Sys().(*syscall.Stat_t)
	if !ok {
		fmt.Printf("get logger file stat failed\n")
		return false
	}
	if int(stat.Uid) != curUid || int(stat.Gid) != curUid {
		fmt.Printf("check log file failed, owner not right\n")
		return false
	}
	return true
}

// IsContainAtlas300IDuo in ProductTypes list, is contain Atlas 300I Duo card
func IsContainAtlas300IDuo() bool {
	for _, product := range ParamOption.ProductTypes {
		if product == Atlas300IDuo {
			return true
		}
	}
	return false
}
