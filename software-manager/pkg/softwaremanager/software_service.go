// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager this file is for business logic
package softwaremanager

import (
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"

	"software-manager/pkg/restfulservice"
)

var (
	maximumSize float64
	// RepositoryFilesPath is to define the root path of imported software
	RepositoryFilesPath string
	// IP is the address of service
	IP string
	// Port is to define the service entrance
	Port int
)

func init() {
	flag.StringVar(&RepositoryFilesPath, "repositoryFilesPath", defaultFilesPath,
		"The path of repository")
	flag.Float64Var(&maximumSize, "maximumSize", 0,
		"The path of repository")
	flag.StringVar(&IP, "ip", "",
		"The listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.IntVar(&Port, "port", 0, "The server port of the http service,range[1025-40000]")
}

func batchDeleteSoftware(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start batch delete software")
	var req []int
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	var notDeleteID []int
	for i := 0; i < len(req); i++ {
		deleteSoftware(req[i], &notDeleteID)
	}
	result := batchDeleteResult{
		NotDeleteID: notDeleteID,
	}
	if len(notDeleteID) == 0 {
		return common.RespMsg{Status: common.Success}
	}
	return common.RespMsg{Status: common.ErrorGetResponse, Msg: "delete error", Data: result}
}

func downloadSoftware(input interface{}) common.RespMsg {
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "class convert error"}
	}
	result := checkDownloadRight(info.UserName, info.Password, info.NodeID)
	if !result {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "wrong user, password or nodeId"}
	}
	err := checkFields(info.ContentType, info.Version)
	if err != nil {
		hwlog.RunLog.Error(fmt.Sprintf("%s%s", err.Error(), "in downloadSoftware func"))
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	exist, err := checkSoftwareExist(info.ContentType, info.Version)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	if !exist {
		hwlog.RunLog.Errorf("%s%s dose not exist", info.ContentType, info.Version)
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "software dose not exist"}
	}
	path := softwarePathJoin(info.ContentType, info.Version)
	return common.RespMsg{Status: common.Success, Data: path}
}

func uploadSoftware(input interface{}) common.RespMsg {
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "class convert error"}
	}
	err := checkFields(info.ContentType, info.Version)
	if err != nil {
		hwlog.RunLog.Error(fmt.Sprintf("%s%s", err.Error(), "in uploadSoftware func"))
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	exist, err := checkSoftwareExist(info.ContentType, info.Version)
	if err != nil {
		hwlog.RunLog.Error("check software exist error in uploadSoftware func")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "check software exist error"}
	}
	if exist {
		hwlog.RunLog.Errorf("%s%s dose not exist", info.ContentType, info.Version)
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "software already exists"}
	}
	file := info.File
	ok, err = checkFile(file)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	if !ok {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "wrong file format"}
	}
	dst, err := creatDir(info.ContentType, info.Version)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "create directory error"}
	}
	if err = saveUploadedFile(file, dst+"/"+info.ContentType+".zip"); err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "save file error"}
	}
	err = SoftwareDbCtlInstance().addSoftware(&info)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	return common.RespMsg{Status: common.Success}
}

func listRepository(input interface{}) common.RespMsg {
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "class convert error"}
	}
	softwareRecords, total, err := SoftwareDbCtlInstance().listSoftware(&info)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	return common.RespMsg{Status: common.Success, Data: queryResult{*softwareRecords, total}}
}

func getURL(input interface{}) common.RespMsg {
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "class convert error"}
	}
	var err error
	if info.Version == "" {
		err = checkContentType(info.ContentType)
	} else {
		err = checkFields(info.ContentType, info.Version)
	}
	if err != nil {
		hwlog.RunLog.Error(fmt.Sprintf("%s%s", err.Error(), "in getURL func"))
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	if !checkNodeID(info.NodeID) {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "incorrect node_id"}
	}
	if info.Version == "" {
		info.Version, err = returnLatestVer(info.ContentType)
		if err != nil {
			hwlog.RunLog.Error(err.Error())
			return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
		}
	}
	exist, err := checkSoftwareExist(info.ContentType, info.Version)
	if err != nil {
		hwlog.RunLog.Error("check software exist error in getURL func")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	if !exist {
		hwlog.RunLog.Errorf("%s%s dose not exist", info.ContentType, info.Version)
		return common.RespMsg{Status: common.ErrorGetResponse,
			Msg: "software dose not exist. Need to import software first"}
	}
	downloadInfo := downloadData{}
	if err = fillDownloadData(&downloadInfo, &info); err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	return common.RespMsg{Status: common.Success, Data: downloadInfo}
}
