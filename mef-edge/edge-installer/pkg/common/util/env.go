// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util this file for get environment information
package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

const maxStatusFileSize = 4 * constants.KB

func getKmcDir() (string, error) {
	compKmcDir, err := path.GetCompSpecificDir(constants.KmcDir)
	if err != nil {
		return "", err
	}
	if err = fileutils.CreateDir(compKmcDir, fileutils.Mode700); err != nil {
		return "", err
	}
	return compKmcDir, nil
}

// GetKmcConfig init kmc config
func GetKmcConfig(kmcDir string) (*kmc.SubConfig, error) {
	if kmcDir == "" {
		getKmcDir, err := getKmcDir()
		if err != nil {
			return nil, err
		}
		kmcDir = getKmcDir
	}
	masterKmcPath := filepath.Join(kmcDir, constants.KmcMasterName)
	backupKmcPath := filepath.Join(kmcDir, constants.KmcBackupName)
	if err := fileutils.MakeSureDir(masterKmcPath); err != nil {
		return nil, err
	}
	return kmc.GetKmcCfg(masterKmcPath, backupKmcPath), nil
}

// IsValidVersion Check whether the new version is the previous version, the next version, or the same version as the
// old version.
// The rule is: 1.0 is the previous version of 2.0, and 3.0 is the next version of 2.0
func IsValidVersion(oldVersion, newVersion string) (bool, error) {
	oldNums := strings.Split(oldVersion, ".")
	newNums := strings.Split(newVersion, ".")
	for i := 0; i < len(oldNums) && i < len(newNums); i++ {
		oldNum, err := strconv.Atoi(oldNums[i])
		if err != nil {
			return false, err
		}
		newNum, err := strconv.Atoi(newNums[i])
		if err != nil {
			return false, err
		}
		if newNum == oldNum {
			continue
		}
		return newNum == oldNum+1 || newNum == oldNum-1, nil
	}
	return len(oldNums) == len(newNums), nil
}

// GetProcesses get running processes on device
func GetProcesses() ([]int, error) {
	reader, files, err := fileutils.ReadDir(constants.ProcPath)
	if err != nil {
		return nil, fmt.Errorf("read proc directory failed, error: %v", err)
	}
	defer fileutils.CloseFile(reader)

	var pids []int
	for _, fi := range files {
		if !fi.IsDir() {
			continue
		}

		name := fi.Name()
		if name[0] < '0' || name[0] > '9' {
			continue
		}

		// from this point forward, errors will be ignored,
		// because it might simply be that the process doesn't exist anymore.
		pid, err := strconv.ParseInt(name, constants.Base10, constants.BitSize0)
		if err != nil {
			continue
		}
		pids = append(pids, int(pid))
	}
	return pids, nil
}

// GetProcName get process name
func GetProcName(pid int) (string, error) {
	procNameFilePath := filepath.Join(constants.ProcPath, strconv.Itoa(pid), "comm")
	data, err := fileutils.LoadFile(procNameFilePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}

// CheckProcUser check process user
func CheckProcUser(pid int, user string) bool {
	uid, err := envutils.GetUid(user)
	if err != nil {
		return false
	}
	return checkProcessUserOrGroup(pid, uid, "Uid")
}

// CheckProcGroup check process group
func CheckProcGroup(pid int, group string) bool {
	gid, err := envutils.GetGid(group)
	if err != nil {
		return false
	}
	return checkProcessUserOrGroup(pid, gid, "Gid")
}

func checkProcessUserOrGroup(pid int, userOrGroupId uint32, typ string) bool {
	procStatusFilePath := filepath.Join(constants.ProcPath, strconv.Itoa(pid), "status")
	data, err := fileutils.ReadLimitBytes(procStatusFilePath, maxStatusFileSize)
	if err != nil {
		return false
	}

	const idCount = 2
	var ids = make(map[uint32]struct{})
	// Uid: [uid] [euid] [suid] [fsuid]
	// Gid: [gid] [egid] [sgid] [fsgid]
	for _, line := range bytes.Split(data, []byte("\n")) {
		if !bytes.HasPrefix(line, []byte(typ)) {
			continue
		}
		fields := strings.Fields(string(line))
		if len(fields) < idCount+1 {
			continue
		}
		// check whether expected user matches gid or egid
		for i := 1; i < idCount+1; i++ {
			id, err := strconv.ParseUint(fields[i], constants.Base10, constants.BitSize0)
			if err != nil {
				continue
			}
			ids[uint32(id)] = struct{}{}
		}
	}

	_, ok := ids[userOrGroupId]
	return ok
}

// IsFlagSet check whether the flag is set
func IsFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// IsProcessActive is process active
func IsProcessActive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		hwlog.RunLog.Warnf("find process failed, error: %v", err)
		return false
	}
	if err = proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}

// GetUuid get uuid
func GetUuid() (string, error) {
	uuid, err := envutils.RunCommand(constants.Cat, envutils.DefCmdTimeoutSec, constants.UuidPath)
	if err != nil {
		return "", fmt.Errorf("get uuid failed, error: %v", err)
	}
	return uuid, nil
}

// MemoryInfo system memory info
type MemoryInfo struct {
	Total uint64 // memory total space, unit is Byte
	Avail uint64 // memory available space, unit is Byte
}

// getMemoryInfo return free command output, unit is KB
func getMemoryInfo() (MemoryInfo, error) {
	const (
		memTotalPattern = `MemTotal:\s*([0-9]+) kB`
		memAvailPattern = `MemAvailable:\s*([0-9]+) kB`
		memInfoPath     = "/proc/meminfo"
	)
	var memoryInfo MemoryInfo
	out, err := fileutils.LoadFile(memInfoPath)
	if err != nil {
		return memoryInfo, err
	}
	total, err := parseMemInfo(memTotalPattern, out)
	if err != nil {
		return memoryInfo, err
	}
	available, err := parseMemInfo(memAvailPattern, out)
	if err != nil {
		return memoryInfo, err
	}

	memoryInfo = MemoryInfo{
		Total: total * constants.KB,
		Avail: available * constants.KB,
	}
	return memoryInfo, nil
}

func parseMemInfo(pattern string, memInfo []byte) (uint64, error) {
	const matchResultLength = 2
	matches := regexp.MustCompile(pattern).FindSubmatch(memInfo)
	if len(matches) != matchResultLength {
		return 0, fmt.Errorf("meminfo format error")
	}
	res, err := strconv.ParseUint(string(matches[1]), constants.Base10, constants.BitSize64)
	if err != nil {
		return 0, errors.New("memory info parse failed")
	}
	return res, nil
}

const (
	cpuTimeUser = iota
	cpuTimeNice
	cpuTimeSystem
	cpuTimeIdle
	cpuTimeIOWait
	cpuTimeIrq
	cpuTimeSoftIrq
	cpuTimeSteal

	statInfoPath            = "/proc/stat"
	statInfoFilePartsNumber = 2
	cpuSumStatName          = "cpu"
	cpuLineIndex            = 0
	cpuItemsSize            = 11
	cpuNameIndex            = 0

	cpuUsageQueueMaxSize = 10
)

var cpuTransientUsageQueue []float64

func updateCPUUsageQueue(info float64) {
	if len(cpuTransientUsageQueue) >= cpuUsageQueueMaxSize {
		// only save latest 10 times cpu use rate, remove oldest usage
		var tmpQueue []float64
		tmpQueue = append(tmpQueue, cpuTransientUsageQueue[1:]...)
		cpuTransientUsageQueue = tmpQueue
	}
	cpuTransientUsageQueue = append(cpuTransientUsageQueue, info)
}

// WatchAndUpdateCPUTransientUsage watch and update the CPU usage within one second
func WatchAndUpdateCPUTransientUsage() {
	workPre, totalPre, err := getCPUTransientStatus()
	if err != nil {
		hwlog.RunLog.Errorf("get CPU transient status failed, reason: %s", err.Error())
		return
	}

	time.Sleep(time.Second)

	workAfter, totalAfter, err := getCPUTransientStatus()
	if err != nil {
		hwlog.RunLog.Errorf("get CPU transient status failed, reason: %s", err.Error())
		return
	}

	work := workAfter - workPre
	total := totalAfter - totalPre
	if total == 0 {
		hwlog.RunLog.Error("total cpu is zero")
		return
	}
	updateCPUUsageQueue(float64(work) / float64(total))
}

func getCPUTransientStatus() (workTime, totalTime int, err error) {
	contents, err := fileutils.LoadFile(statInfoPath)
	if err != nil {
		return 0, 0, err
	}

	lines := strings.SplitN(string(contents), "\n", statInfoFilePartsNumber)
	if len(lines) == 0 {
		return 0, 0, errors.New("invalid statistic file, invalid cpu status")
	}

	fields := strings.Fields(lines[cpuLineIndex])
	if len(fields) != cpuItemsSize || fields[cpuNameIndex] != cpuSumStatName {
		return 0, 0, errors.New("invalid statistic file, invalid cpu status phase info")
	}

	var cpuStat []int
	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseInt(fields[i], constants.Base10, constants.BitSize64)
		if err != nil {
			return 0, 0, err
		}
		cpuStat = append(cpuStat, int(val))
	}

	if len(cpuStat) <= cpuTimeSteal {
		return 0, 0, errors.New("invalid statistic file, invalid cpu status phase number")
	}
	workTime = cpuStat[cpuTimeUser] + cpuStat[cpuTimeNice] +
		cpuStat[cpuTimeSystem] + cpuStat[cpuTimeIrq] + cpuStat[cpuTimeSoftIrq] + cpuStat[cpuTimeSteal]
	idleTime := cpuStat[cpuTimeIdle] + cpuStat[cpuTimeIOWait]

	return workTime, workTime + idleTime, nil
}

func getCPUAverageUsage() float64 {
	if len(cpuTransientUsageQueue) == 0 {
		return 0
	}
	var total float64
	for _, val := range cpuTransientUsageQueue {
		total += val
	}
	return total / float64(len(cpuTransientUsageQueue))
}

// IsSystemMemoryEnough check memory available space is enough, threshold unit is byte
func IsSystemMemoryEnough(threshold uint64) (bool, error) {
	memoryInfo, err := getMemoryInfo()
	if err != nil {
		return false, err
	}
	return memoryInfo.Avail > threshold, nil
}

// IsSystemCPUAvailable check cpu is available, threshold is cpu usage percentage
func IsSystemCPUAvailable(threshold float64) bool {
	cpuPercentage := getCPUAverageUsage()
	return cpuPercentage < threshold
}

// IsSystemStorageEnough check storage available space is enough
func IsSystemStorageEnough(path string, threshold uint64) (bool, error) {
	availSpace, err := envutils.GetDiskFree(path)
	if err != nil {
		return false, err
	}
	return availSpace >= threshold, nil
}

// RemoveContainer remove containers
func RemoveContainer() error {
	containerID, err := envutils.RunCommand(constants.DockerCmd, envutils.DefCmdTimeoutSec, "ps", "-aq")
	if err != nil {
		fmt.Println("warning: remove containers failed, containers may not be removed, " +
			"please check and remove them manually.")
		hwlog.RunLog.Warnf("get container command error: %v", err)
		return nil
	}
	if containerID == "" {
		hwlog.RunLog.Info("no container exists, do not need to remove")
		return nil
	}
	var errContainer []string
	containers := strings.Split(containerID, "\n")
	iterationCount := 1
	for _, id := range containers {
		if iterationCount > constants.MaxIterationCount {
			break
		}
		if _, err := envutils.RunCommand(constants.DockerCmd, envutils.DefCmdTimeoutSec, "stop", id); err != nil {
			errContainer = append(errContainer, id)
		}
		if _, err = envutils.RunCommand(constants.DockerCmd, envutils.DefCmdTimeoutSec, "rm", id); err != nil {
			errContainer = append(errContainer, id)
		}
		iterationCount++
	}
	if len(errContainer) != 0 {
		fmt.Println("warning: remove containers failed, some containers are not removed, " +
			"please remove them manually.")
		hwlog.RunLog.Warnf("remove container error: %v", errContainer)
		return nil
	}

	hwlog.RunLog.Info("remove containers success")
	return nil
}

// GetContentMap get map from k8s message content which is an interface
func GetContentMap(content interface{}) (map[string]interface{}, error) {
	contentBytes, err := GetContentData(content)
	if err != nil {
		return nil, err
	}
	var ret map[string]interface{}
	if err = json.Unmarshal(contentBytes, &ret); err != nil {
		return nil, errors.New("convert content unmarshal err")
	}
	return ret, nil
}

// GetContentData get []byte from k8s message content which is an interface
func GetContentData(content interface{}) ([]byte, error) {
	if data, ok := content.([]byte); ok {
		return data, nil
	}

	if data, ok := content.(string); ok {
		return []byte(data), nil
	}

	data, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("marshal interface to []byte failed: %v", err)
	}
	return data, nil
}

// GetBoolPointer get pointer based on bool value
// If the query or update value is 0 in db, the query or update fails. Use the pointer can solve the problem.
func GetBoolPointer(value bool) *bool {
	pointer := new(bool)
	*pointer = value
	return pointer
}

// GetMefId get MEFEdge uid & gid
func GetMefId() (uint32, uint32, error) {
	uid, err := envutils.GetUid(constants.EdgeUserName)
	if err != nil {
		return 0, 0, err
	}
	gid, err := envutils.GetGid(constants.EdgeUserGroup)
	if err != nil {
		return 0, 0, err
	}
	return uid, gid, nil
}

// NewEdgeUGidMgr the constructor func for EdgeGUidMgr
func NewEdgeUGidMgr() EdgeGUidMgr {
	mgr := EdgeGUidMgr{
		uid: syscall.Getuid(),
		gid: syscall.Getgid(),
	}
	return mgr
}

// EdgeGUidMgr a helper to set/reset euid/egid from the current user to mef-edge
type EdgeGUidMgr struct {
	uid, gid int
}

// SetEUGidToEdge set euid/egid to mef-edge
func (u *EdgeGUidMgr) SetEUGidToEdge() error {
	uid, gid, err := GetMefId()
	if err != nil {
		return fmt.Errorf("get mef-edge uid/gid failed, %v", err)
	}
	return SetEuidAndEgid(int(uid), int(gid))
}

// ResetEUGid reset euid/guid to the original
func (u *EdgeGUidMgr) ResetEUGid() error {
	return SetEuidAndEgid(u.uid, u.gid)
}

// SetEuidAndEgid set euid and egid
func SetEuidAndEgid(uid, gid int) error {
	if err := syscall.Setegid(gid); err != nil {
		return fmt.Errorf("set gid to %d failed, %v", gid, err)
	}
	if err := syscall.Seteuid(uid); err != nil {
		return fmt.Errorf("set uid to %d failed, %v", uid, err)
	}
	return nil
}

// CheckNecessaryCommands check commands used in shell
func CheckNecessaryCommands() error {
	var necessaryCommands = []string{
		"arch", "awk", "basename", "blockdev", "cat", "chattr", "chmod", "chown", "cp", "date", "dirname", "docker",
		"file", "find", "grep", "iptables", "ln", "mkdir", "mknod", "mount", "mountpoint", "sleep", "stat", "readlink",
		"realpath", "rm", "sed", "systemctl", "tar", "touch", "umount", "unlink",
	}

	for _, command := range necessaryCommands {
		if err := envutils.CheckCommandAllowedSugid(command); err != nil {
			hwlog.RunLog.Errorf("check necessary commands failed, [%s] is abnormal, error: %s", command, err.Error())
			return fmt.Errorf("check necessary commands failed, [%s] is abnormal", command)
		}
	}
	return nil
}
