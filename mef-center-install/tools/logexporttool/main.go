// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/common/logmgmt/hwlogconfig"
	"huawei.com/mindxedge/base/common/logmgmt/logcollect"
)

const (
	defaultRunLogFile     = "/var/log/mindx-edge/edge-manager/run.log"
	defaultOperateLogFile = "/var/log/mindx-edge/edge-manager/operate.log"
	defaultBackupDirName  = "/var/log_backup/mindx-edge/edge-manager"
	defaultOpLogMaxSize   = 100
	defaultRunLogMaxSize  = 100
	rootCaPath            = "/home/data/inner-root-ca/RootCA.crt"
	clientCertKeyFile     = "/home/data/config/mef-certs/edge-manager.key"
	clientCertFile        = "/home/data/config/mef-certs/edge-manager.crt"
	failExitCode          = 1
	maxRetry              = 30
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: defaultRunLogFile, FileMaxSize: defaultRunLogMaxSize,
		BackupDirName: defaultBackupDirName, OnlyToFile: true}
	serverOpConf = &hwlog.LogConfig{LogFileName: defaultOperateLogFile, FileMaxSize: defaultOpLogMaxSize,
		BackupDirName: defaultBackupDirName, OnlyToFile: true}
	nodes string
)

type batchResp struct {
	SuccessIDs []interface{} `json:"successIDs"`
	FailIDs    []interface{} `json:"failIDs"`
}

func main() {
	fmt.Println("start to collect logs")
	flag.Parse()
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s.\n", err.Error())
		os.Exit(failExitCode)
		return
	}
	if err := exportEdgeLogs(strings.Split(nodes, ",")); err != nil {
		fmt.Printf("failed to collect logs, %v\n", err)
		os.Exit(failExitCode)
		return
	}
	fmt.Println("collect logs successful")
}

func init() {
	flag.StringVar(&nodes, "nodes", "", "the serial-numbers of node to collect log")
	hwlogconfig.BindFlags(serverOpConf, serverRunConf)
}

func exportEdgeLogs(edgeNodes []string) error {
	baseUrl := fmt.Sprintf("https://%s:%d", common.EdgeMgrDns, common.EdgeMgrPort)
	fmt.Println("create log collection tasks")
	runningNodes, err := createTasks(baseUrl, edgeNodes)
	if err != nil {
		return err
	}
	if len(runningNodes) == 0 {
		return errors.New("none of task will be execute")
	}

	var (
		successNodes []string
		failNodes    []string
		retry        int
	)
	for {
		s, f, err := queryTaskProgress(baseUrl, runningNodes)
		if err != nil {
			return err
		}
		successNodes = append(successNodes, s...)
		if len(s) == 0 {
			time.Sleep(time.Second)
			continue
		}
		paths, err := queryTaskPath(baseUrl, s)
		if err != nil {
			return err
		}
		for i := range s {
			if len(paths) <= i {
				continue
			}
			fmt.Printf("collect edge log success, node name is %s, file name is %s", s[i], paths[i])
		}
		failNodes = append(failNodes, f...)
		if len(successNodes)+len(failNodes) == len(runningNodes) {
			break
		}
		retry += 1
		if retry >= maxRetry {
			return errors.New("too many retries")
		}
		time.Sleep(time.Second)
	}

	return nil
}

func getTlsCertInfo() certutils.TlsCertInfo {
	return certutils.TlsCertInfo{
		RootCaPath: rootCaPath,
		CertPath:   clientCertFile,
		KeyPath:    clientCertKeyFile,
		SvrFlag:    false,
	}
}

func createTasks(baseUrl string, edgeNodes []string) ([]string, error) {
	req := httpsmgr.GetHttpsReq(
		baseUrl+common.LogCollectPathPrefix+common.ResRelLogTask, getTlsCertInfo())

	reqData := logcollect.BatchQueryTaskReq{
		Module:    logcollect.ModuleEdge,
		EdgeNodes: edgeNodes,
	}
	postData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request, %v", err)
	}
	respBytes, err := req.PostJson(postData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request, %v", err)
	}
	br, err := parseBatchResp(respBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response, %v", err)
	}
	if br == nil {
		return edgeNodes, nil
	}
	var result []string
	for _, s := range br.SuccessIDs {
		node, ok := s.(string)
		if !ok {
			return nil, errors.New("failed to parse response, bad node name type")
		}
		result = append(result, node)
	}
	for _, f := range br.FailIDs {
		fmt.Printf("failed to colect log for node %s\n", f)
	}
	return result, nil
}

func queryTaskProgress(baseUrl string, edgeNodes []string) ([]string, []string, error) {
	url := baseUrl + common.LogCollectPathPrefix + common.ResRelLogTaskProgress + "?module=edge"
	for _, node := range edgeNodes {
		url += "&node=" + node
	}
	req := httpsmgr.GetHttpsReq(url, getTlsCertInfo())
	respBytes, err := req.Get()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request, %v", err)
	}
	br, err := parseBatchResp(respBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse response, %v", err)
	}
	if br == nil {
		return nil, nil, errors.New("failed to parse response, bad batch response type")
	}
	if len(br.FailIDs) > 0 {
		return nil, nil, errors.New("get progress failed")
	}
	successIDs, _, err := parseTaskResp(br)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse response, %v", err)
	}
	var result []string
	for _, s := range successIDs {
		fmt.Printf("%v\n", s)
		dataBytes, err := json.Marshal(s.Data)
		if err != nil {
			return nil, nil, errors.New("failed to parse response")
		}
		var p logcollect.TaskProgress
		if err := json.Unmarshal(dataBytes, &p); err != nil {
			return nil, nil, errors.New("failed to parse response")
		}
		if p.Status == common.Success && p.Progress == logcollect.ProgressMax {
			result = append(result, s.EdgeNode)
		}
		if p.Status != common.Success {
			fmt.Printf("collect edge log failed, node name is %s\n", s.EdgeNode)
		}
	}
	return result, nil, nil
}

func queryTaskPath(baseUrl string, edgeNodes []string) ([]string, error) {
	url := baseUrl + common.LogCollectPathPrefix + common.ResRelLogTaskPath + "?module=edge"
	for _, node := range edgeNodes {
		url += "&node=" + node
	}
	req := httpsmgr.GetHttpsReq(url, getTlsCertInfo())
	respBytes, err := req.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to send request, %v", err)
	}

	br, err := parseBatchResp(respBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response, %v", err)
	}
	if br == nil {
		return nil, errors.New("failed to parse response, bad data type")
	}
	if len(br.FailIDs) > 0 {
		return nil, errors.New("failed to parse response, can't get path")
	}

	successIDs, _, err := parseTaskResp(br)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response, %v", err)
	}

	taskPaths := make(map[string]string)
	for _, s := range successIDs {
		path, ok := s.Data.(string)
		if !ok {
			return nil, errors.New("failed to parse response, bad data type")
		}
		taskPaths[s.EdgeNode] = path
	}

	var result []string
	for _, n := range edgeNodes {
		result = append(result, taskPaths[n])
	}
	return result, nil
}

func parseBatchResp(respBytes []byte) (*batchResp, error) {
	var resp common.RespMsg
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, errors.New("failed to unmarshal response")
	}

	if resp.Data == nil {
		if resp.Status == common.Success {
			return nil, nil
		}
		return nil, errors.New(resp.Msg)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, errors.New("failed to parse response")
	}
	var br batchResp
	if err := json.Unmarshal(dataBytes, &br); err != nil {
		return nil, errors.New("failed to parse response")
	}

	return &br, nil
}

func parseTaskResp(resp *batchResp) ([]logcollect.QueryTaskResp, []logcollect.QueryTaskResp, error) {
	var successIDs, failIDs []logcollect.QueryTaskResp
	for _, itemObj := range resp.SuccessIDs {
		dataBytes, err := json.Marshal(itemObj)
		if err != nil {
			return nil, nil, errors.New("failed to parse progress")
		}
		var qr logcollect.QueryTaskResp
		if err := json.Unmarshal(dataBytes, &qr); err != nil {
			return nil, nil, errors.New("failed to parse progress")
		}
		successIDs = append(successIDs, qr)
	}
	for _, itemObj := range resp.FailIDs {
		dataBytes, err := json.Marshal(itemObj)
		if err != nil {
			return nil, nil, errors.New("failed to parse progress")
		}
		var qr logcollect.QueryTaskResp
		if err := json.Unmarshal(dataBytes, &qr); err != nil {
			return nil, nil, errors.New("failed to parse progress")
		}
		failIDs = append(failIDs, qr)
	}
	return successIDs, failIDs, nil
}
