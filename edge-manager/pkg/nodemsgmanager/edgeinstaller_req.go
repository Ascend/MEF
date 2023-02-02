// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package nodemsgmanager request used when downloading and upgrading software
package nodemsgmanager

// DownloadSfwReqToSfwMgr download software request to software manager
type DownloadSfwReqToSfwMgr struct {
	NodeID          string `json:"nodeID,omitempty"`
	SoftwareName    string `json:"softwareName"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
}

// RespDataFromSfwMgr response data from software manager
type RespDataFromSfwMgr struct { // todo json tag与软件仓统一，后续可与软件仓统一修改
	NodeId      string `json:"nodeID"`
	DownloadUrl string `json:"url"`
	Username    string `json:"userName"`
	Password    string `json:"password"`
}

// ContentToConnector content to edge-installer
type ContentToConnector struct {
	DownloadUrl     string `json:"downloadUrl"`
	SoftwareName    string `json:"softwareName"`
	SoftwareVersion string `json:"softwareVersion"`
	Username        string `json:"username"`
	Password        string `json:"password"`
}

// HttpBody used to construct http body to software manager
type HttpBody struct {
	NodeID string `json:"nodeID"`
}

// RespMsg response message from software manager
type RespMsg struct {
	Status string             `json:"status"`
	Msg    string             `json:"msg"`
	Data   RespDataFromSfwMgr `json:"data,omitempty"`
}

// SoftwareManagerInfo info required for updating software
type SoftwareManagerInfo struct {
	SoftwareIP   string
	SoftwarePort string
	SoftRoute    string
}
