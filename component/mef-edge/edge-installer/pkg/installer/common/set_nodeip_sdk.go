// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package common for setting nodeIP to edge core configuration file
package common

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

// SetNodeIPToEdgeCore set nodeIP to edge core configuration file
func SetNodeIPToEdgeCore() error {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		return fmt.Errorf("get config path manager failed, error: %v", err)
	}

	nodeIP, err := getNodeIP(configPathMgr.GetCompConfigDir(constants.EdgeOm))
	if err != nil {
		return fmt.Errorf("get node ip failed: %v", err)
	}

	if nodeIP == "" {
		return nil
	}

	edgeCoreConfigPath := configPathMgr.GetEdgeCoreConfigPath()
	if err = config.SetNodeIP(edgeCoreConfigPath, nodeIP); err != nil {
		return fmt.Errorf("set node ip failed: %v", err)
	}

	return nil
}

func getNodeIP(edgeOmConfigDir string) (string, error) {
	dbMgr := config.NewDbMgr(edgeOmConfigDir, constants.DbEdgeOmPath)
	netConfig, err := config.GetNetManager(dbMgr)
	if err != nil {
		return "", fmt.Errorf("get net config failed, error: %v", err)
	}

	defer utils.ClearSliceByteMemory(netConfig.Token)
	if netConfig.NetType != constants.MEF {
		hwlog.RunLog.Info("netType is not MEF, do not need set nodeIP")
		return "", nil
	}

	address := net.JoinHostPort(netConfig.IP, strconv.Itoa(netConfig.Port))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return "", errors.New("test connect to server failed")
	}
	defer func() {
		if err = conn.Close(); err != nil {
			hwlog.RunLog.Warn("close the test connection failed when getting the node ip")
		}
	}()

	localAddr, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		return "", errors.New("get local address failed")
	}

	ip := strings.Split(localAddr.String(), ":")[0]
	return ip, nil
}
