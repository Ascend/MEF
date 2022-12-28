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
	"edge-manager/pkg/util"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// SoftwareManagerInfo info required for updating software
type SoftwareManagerInfo struct {
	SoftwareIP   string
	SoftwarePort string
	SoftRoute    string
}

// HttpBody used to construct http body to software manager
type HttpBody struct {
	NodeId string `json:"nodeId"`
}

// RespMsg response message from software manager
type RespMsg struct {
	Status string              `json:"status"`
	Msg    string              `json:"msg"`
	Data   util.DealSfwContent `json:"data,omitempty"`
}

func sendToSoftwareManager(dealSfwReq *util.DownloadSfwReq) (*http.Response, error) {
	nodeInfo, err := constructHttpBody(dealSfwReq.NodeID)
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

func receiveRespFromHttp(req *http.Request) (*http.Response, error) {
	hwlog.RunLog.Info("edge-installer sends request to software manager")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			// 不校验服务端证书，直接信任
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

func constructHttpReq(dealSfwReq *util.DownloadSfwReq, nodeInfo []byte) (*http.Request, error) {
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
	// todo 核对新的url格式
	sfwUrl := fmt.Sprintf("http://%s:%s/%s/url?contentType=%s",
		sfwMgrInfo.SoftwareIP, sfwMgrInfo.SoftwarePort, sfwMgrInfo.SoftRoute, softwareName)
	req, err := http.NewRequest(HttpsMethod, sfwUrl, bytes.NewReader(nodeInfo))
	if err != nil {
		hwlog.RunLog.Errorf("new request for http failed, error: %v", err)
		return nil, err
	}

	return req, nil
}

func constructHttpBody(nodeId string) ([]byte, error) {
	httpBody := HttpBody{
		NodeId: nodeId,
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

func dealRespFromSfwManager(resp *http.Response, nodeId string) (*util.DealSfwContent, error) {
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

	dealSfwContent := respMsg.Data
	if err = CheckDataFromSfwMgr(&dealSfwContent, nodeId); err != nil {
		hwlog.RunLog.Errorf("check data from software manager failed, error: %v", err)
		return nil, err
	}

	dealSfwContent = mergeSfwInfo(dealSfwContent)
	if &dealSfwContent == nil {
		hwlog.RunLog.Error("merge software info failed")
		return nil, errors.New("merge software info failed")
	}
	return &dealSfwContent, nil
}

func mergeSfwInfo(dealSfwContent util.DealSfwContent) util.DealSfwContent {
	downloadUrl := dealSfwContent.Url
	dataBytes := strings.Split(downloadUrl, "=")
	if len(dataBytes) < LocationRespSfwName || len(dataBytes) < LocationRespSfwVersion {
		return util.DealSfwContent{}
	}
	softwareName := strings.Split(dataBytes[LocationRespSfwName], "&")[LocationSfw]
	softwareVersion := strings.Split(dataBytes[LocationRespSfwVersion], "&")[LocationSfw]
	dealSfwContent.SoftwareName = softwareName
	dealSfwContent.SoftwareVersion = softwareVersion
	return dealSfwContent
}

func downloadWithSfwMgr(dealSfwReq util.DownloadSfwReq) (*util.DealSfwContent, error) {
	var resp *http.Response
	var dealSfwContent *util.DealSfwContent
	var err error

	resp, err = sendToSoftwareManager(&dealSfwReq)
	if err != nil {
		hwlog.RunLog.Errorf("send to software manager failed, error: %v", err)
		return nil, err
	}

	if dealSfwContent, err = dealRespFromSfwManager(resp, dealSfwReq.NodeID); err != nil {
		hwlog.RunLog.Errorf("deal resp from software manager failed, error: %v", err)
		return nil, err
	}

	return dealSfwContent, nil
}

func upgradeWithSfwManager(upgradeSfwReq util.UpgradeSfwReq) (*util.DealSfwContent, error) {
	nodeIds, err := getNodeNum(upgradeSfwReq.NodeIDs)
	if err != nil {
		hwlog.RunLog.Errorf("get node unique name failed, error: %v", err)
		return nil, err
	}

	var dealSfwContent *util.DealSfwContent
	for _, nodeId := range nodeIds { // todo 只针对单个节点
		hwlog.RunLog.Infof("--------edge-installer %s upgrade software begin--------", nodeId)

		downloadSfwReq := util.DownloadSfwReq{
			NodeID:          nodeId,
			SoftwareName:    upgradeSfwReq.SoftwareName,
			SoftwareVersion: upgradeSfwReq.SoftwareVersion,
		}
		dealSfwContent, err = downloadWithSfwMgr(downloadSfwReq)
		if err != nil {
			hwlog.RunLog.Errorf("deal with software manager failed, error: %v", err)
			return nil, err
		}
		continue
	}

	return dealSfwContent, nil
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

func mergeContentAndSend(msg, resp *model.Message) {
	data, err := json.Marshal(msg.GetContent())
	if err != nil {
		hwlog.RunLog.Errorf("marshal message content failed, error: %v", err)
		return
	}
	respData, err := json.Marshal(resp.GetContent())
	if err != nil {
		hwlog.RunLog.Errorf("marshal resp message content failed, error: %v", err)
		return
	}
	content := make(map[string]interface{})
	if err = json.Unmarshal(data, &content); err != nil {
		hwlog.RunLog.Errorf("parse message content failed, error: %v", err)
		return
	}
	if err = json.Unmarshal(respData, &content); err != nil {
		hwlog.RunLog.Errorf("parse resp data failed, error: %v", err)
		return
	}

	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return
	}
	respMsg.SetRouter(common.EdgeInstallerName, common.EdgeConnectorName, common.Upgrade, common.Software)
	respMsg.FillContent(content)
	respMsg.SetIsSync(false)
	if err = modulemanager.SendMessage(respMsg); err != nil {
		hwlog.RunLog.Errorf("send message failed, error: %v", err)
		return
	}
}
