// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package envutils

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"huawei.com/mindx/common/fileutils"
)

const (
	socketFlag      = "socket"
	socketStartLoc  = 8
	inodeColumn     = 9
	portStateColumn = 3
	addrColumn      = 1
	addrPortColumn  = 1
	hexadecimal     = 16
	bitSize64       = 64
	allToken        = "ALL"
)

// const for process_port_mgr
const (
	ListenState      = "LISTEN"
	EstablishedState = "ESTABLISHED"
	AllState         = "ALL"

	TcpProtocol = "tcp"
)

// ProcessPortMgr is the struct to manage the port occupied by process with its pid
type ProcessPortMgr struct {
	Pid    int
	inodes map[string]string
}

func (pm *ProcessPortMgr) getFdDirPath() string {
	return filepath.Join("/proc", strconv.Itoa(pm.Pid), "fd")
}

func (pm *ProcessPortMgr) initSocketInodes() error {
	fdPath := pm.getFdDirPath()

	handle, files, err := fileutils.ReadDir(fdPath)
	if err != nil {
		return fmt.Errorf("read fd dir failed: %s", err.Error())
	}
	defer handle.Close()

	inodeMap := map[string]string{}
	for _, file := range files {
		filePath := filepath.Join(fdPath, file.Name())
		realPath, err := fileutils.ReadLink(filePath)
		if err != nil {
			// file maybe normally closed before read operation
			continue
		}

		if !strings.Contains(realPath, socketFlag) {
			continue
		}

		fileRealName := filepath.Base(realPath)
		if len(fileRealName) <= socketStartLoc {
			continue
		}

		inode := fileRealName[socketStartLoc : len(fileRealName)-1]
		inodeMap[inode] = ""
	}

	pm.inodes = inodeMap

	return nil
}

func (pm *ProcessPortMgr) getPortStateMap() map[string]string {
	return map[string]string{
		ListenState:      "0A",
		EstablishedState: "01",
		AllState:         allToken,
	}
}

func (pm *ProcessPortMgr) getPortFileMap() map[string]string {
	return map[string]string{
		TcpProtocol: filepath.Join("/proc", strconv.Itoa(pm.Pid), "net", "tcp"),
	}
}

func (pm *ProcessPortMgr) findPort(protocol, portStat string) ([]int, error) {
	netFilePath, exists := pm.getPortFileMap()[protocol]
	if !exists {
		return nil, errors.New("unsupported protocol")
	}

	requiredStatToken, exists := pm.getPortStateMap()[portStat]
	if !exists {
		return nil, errors.New("unsupported portStat")
	}

	content, err := fileutils.LoadFile(netFilePath)
	if err != nil {
		return nil, fmt.Errorf("read net file path failed: %s", err.Error())
	}

	var ports []int
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		port, found := pm.findSinglePort(line, requiredStatToken)
		if !found {
			continue
		}

		decimalPort, err := strconv.ParseInt(port, hexadecimal, bitSize64)
		if err != nil {
			return nil, fmt.Errorf("transform port into decimal failed: %s", err.Error())
		}

		ports = append(ports, int(decimalPort))
	}
	return ports, nil
}

func (pm *ProcessPortMgr) findSinglePort(line, requiredStatSig string) (string, bool) {
	columns := strings.Fields(line)
	if len(columns) <= inodeColumn {
		return "", false
	}

	inode := columns[inodeColumn]
	_, exists := pm.inodes[inode]
	if !exists {
		return "", false
	}

	portStatToken := columns[portStateColumn]
	if requiredStatSig != allToken && portStatToken != requiredStatSig {
		return "", false
	}

	addr := columns[addrColumn]
	splitAddr := strings.Split(addr, ":")
	if len(splitAddr) <= addrPortColumn {
		return "", false
	}

	port := splitAddr[addrPortColumn]
	return port, true
}

// GetPortByPid is the func to get the port occupied by a pid by its protocol and portStat
func (pm *ProcessPortMgr) GetPortByPid(protocol, portStat string) ([]int, error) {
	if err := pm.initSocketInodes(); err != nil {
		return nil, err
	}

	ports, err := pm.findPort(protocol, portStat)
	if err != nil {
		return nil, err
	}

	return ports, nil
}
