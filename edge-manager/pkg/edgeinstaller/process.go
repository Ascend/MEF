// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller processing used in edge-installer module
package edgeinstaller

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"edge-manager/pkg/database"
	"edge-manager/pkg/nodemanager"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func sendToSoftwareManager(nodeID string, dealSfwReq *DownloadSfwReqToSfwMgr) (*http.Response, error) {
	nodeInfo, err := constructHttpBody(nodeID)
	if err != nil {
		hwlog.RunLog.Errorf("construct http body failed, error: %v", err)
		return nil, err
	}

	req, err := constructHttpReq(dealSfwReq, nodeInfo)
	if err != nil {
		hwlog.RunLog.Errorf("construct http request failed, error: %v", err)
		return nil, err
	}

	var resp *http.Response
	if resp, err = receiveRespFromHttp(req); err != nil {
		hwlog.RunLog.Errorf("receive response from http failed, error: %v", err)
		return nil, err
	}

	return resp, nil
}

func constructHttpBody(nodeID string) ([]byte, error) {
	httpBody := HttpBody{
		NodeID: nodeID,
	}

	var nodeInfo []byte
	var err error
	nodeInfo, err = json.Marshal(httpBody)
	if err != nil {
		hwlog.RunLog.Errorf("marshal http body failed, error: %v", err)
		return nil, err
	}

	return nodeInfo, nil
}

func constructHttpReq(dealSfwReq *DownloadSfwReqToSfwMgr, nodeInfo []byte) (*http.Request, error) {
	var sfwUrl string
	sfwMgrInfoFromTable := &SoftwareMgrInfo{}
	if err := readInSfwMgrInfo(sfwMgrInfoFromTable); err != nil {
		hwlog.RunLog.Errorf("read in table software manager info failed, error: %v", err)
		return nil, err
	}

	sfwMgrInfo := &SoftwareManagerInfo{
		SoftwareIP:   sfwMgrInfoFromTable.Address,
		SoftwarePort: sfwMgrInfoFromTable.Port,
		SoftRoute:    sfwMgrInfoFromTable.Route,
	}

	softwareName := dealSfwReq.SoftwareName
	if dealSfwReq.SoftwareVersion == "" { // todo 后续需与软件仓统一修改为https
		sfwUrl = fmt.Sprintf("http://%s:%s/%s/url?contentType=%s",
			sfwMgrInfo.SoftwareIP, sfwMgrInfo.SoftwarePort, sfwMgrInfo.SoftRoute, softwareName)
	} else {
		softwareVersion := dealSfwReq.SoftwareVersion
		sfwUrl = fmt.Sprintf("http://%s:%s/%s/url?contentType=%s&&version=%s",
			sfwMgrInfo.SoftwareIP, sfwMgrInfo.SoftwarePort, sfwMgrInfo.SoftRoute, softwareName, softwareVersion)
	}

	req, err := http.NewRequest(HttpsMethod, sfwUrl, bytes.NewReader(nodeInfo))
	if err != nil {
		hwlog.RunLog.Errorf("new request for http failed, error: %v", err)
		return nil, err
	}

	return req, nil
}

func receiveRespFromHttp(req *http.Request) (*http.Response, error) {
	hwlog.RunLog.Info("edge-installer sends request to software manager")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			// todo 证书校验
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   HttpTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		hwlog.RunLog.Errorf("get response from software manager failed, error: %v", err)
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			hwlog.RunLog.Errorf("close response body from software manager error: %v", err)
			return
		}
	}(req.Body)

	hwlog.RunLog.Info("edge-manager receives download url from software manager success")
	return resp, nil
}

func dealRespFromSfwManager(resp *http.Response, nodeId string) (*ContentToConnector, error) {
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		hwlog.RunLog.Errorf("edge-manager could not read response body from software manager: %v", err)
		return nil, err
	}

	respMsg := &RespMsg{}
	if err = json.Unmarshal(resBody, respMsg); err != nil {
		hwlog.RunLog.Error("parse response body from http to respMsg failed")
		return nil, err
	}

	if respMsg.Status != common.Success {
		hwlog.RunLog.Error("get response from software manager failed")
		return nil, errors.New("get response from software manager failed")
	}

	respDataFromSfwMgr := respMsg.Data
	if err = CheckRespDataFromSfwMgr(&respDataFromSfwMgr, nodeId); err != nil {
		hwlog.RunLog.Errorf("check response data from software manager failed, error: %v", err)
		return nil, err
	}

	contentToConnector := constructContentToConnector(respDataFromSfwMgr)
	if &contentToConnector == nil {
		hwlog.RunLog.Error("construct content to edge-connector failed")
		return nil, errors.New("construct content to edge-connector failed")
	}
	return &contentToConnector, nil
}

func constructContentToConnector(respDataFromSfwMgr RespDataFromSfwMgr) ContentToConnector {
	dataBytes := strings.Split(respDataFromSfwMgr.DownloadUrl, "=")
	if len(dataBytes) < LocationRespSfwName || len(dataBytes) < LocationRespSfwVersion {
		return ContentToConnector{}
	}
	softwareName := strings.Split(dataBytes[LocationRespSfwName], "&")[LocationSfw]
	softwareVersion := strings.Split(dataBytes[LocationRespSfwVersion], "&")[LocationSfw]
	contentToConnector := ContentToConnector{
		DownloadUrl:     respDataFromSfwMgr.DownloadUrl,
		SoftwareName:    softwareName,
		SoftwareVersion: softwareVersion,
		Username:        respDataFromSfwMgr.Username,
		Password:        respDataFromSfwMgr.Password,
	}

	return contentToConnector
}

func downloadWithSfwMgr(nodeID string, dealSfwReq DownloadSfwReqToSfwMgr) (*ContentToConnector, error) {
	var resp *http.Response
	var contentToConnector *ContentToConnector
	var err error

	resp, err = sendToSoftwareManager(nodeID, &dealSfwReq)
	if err != nil {
		hwlog.RunLog.Errorf("send to software manager failed, error: %v", err)
		return nil, err
	}

	if contentToConnector, err = dealRespFromSfwManager(resp, nodeID); err != nil {
		hwlog.RunLog.Errorf("deal resp from software manager failed, error: %v", err)
		return nil, err
	}

	return contentToConnector, nil
}

func getNodeNum(nodeNums []int64) ([]string, error) {
	var nodeIds []string
	for _, nodeNum := range nodeNums {
		var node nodemanager.NodeInfo
		if err := database.GetDb().Model(nodemanager.NodeInfo{}).Where("id = ?", nodeNum).First(&node).Error; err != nil {
			hwlog.RunLog.Errorf("get nodeInfo failed, error: %v", err)
			return []string{}, err
		}
		nodeIds = append(nodeIds, node.UniqueName)
	}

	return nodeIds, nil
}

func getContentToConnector(upgradeSfwReqWithUrl *UpgradeSfwReq) *ContentToConnector {
	downloadUrl := fmt.Sprintf("%s %s", common.OptPost, upgradeSfwReqWithUrl.DownloadUrlFromUser)
	contentToConnector := &ContentToConnector{
		DownloadUrl:     downloadUrl,
		SoftwareName:    upgradeSfwReqWithUrl.SoftwareName,
		SoftwareVersion: upgradeSfwReqWithUrl.SoftwareVersion,
		Username:        upgradeSfwReqWithUrl.Username,
		Password:        upgradeSfwReqWithUrl.Password,
	}
	return contentToConnector
}

func respRestful(message *model.Message) (*UpgradeSfwReq, error) {
	var respContent = common.RespMsg{Status: common.Success, Msg: "", Data: nil}
	upgradeSfwReq, err := constructContentToRestful(message)
	if err != nil {
		respContent = common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
		return nil, errors.New("edge-installer construct content to restful module failed")
	}
	respToRestful, err := message.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("edge-installer new response failed, error: %v", err)
		return nil, errors.New("edge-installer new response failed")
	}
	respToRestful.FillContent(respContent)
	if err = modulemanager.SendMessage(respToRestful); err != nil {
		hwlog.RunLog.Errorf("%s send response to restful failed", common.EdgeInstallerName)
		return nil, err
	}

	return upgradeSfwReq, nil
}

func constructContentToRestful(message *model.Message) (*UpgradeSfwReq, error) {
	var upgradeSfwReq UpgradeSfwReq
	if err := common.ParamConvert(message.GetContent(), &upgradeSfwReq); err != nil {
		return nil, err
	}

	if err := upgradeSfwReq.checkUpgradeSfwReq(); err != nil {
		return nil, err
	}

	return &upgradeSfwReq, nil
}
