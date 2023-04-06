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

// URLReq [struct] to get software url info
type URLReq struct {
	ContentType string `json:"contentType"`
	Version     string `json:"version"`
}

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
	path := softwarePathJoin(info.ContentType, info.Version, info.FileName)
	hwlog.RunLog.Infof("download software pkg :%s", path)
	return common.RespMsg{Status: common.Success, Data: path}
}

func uploadSoftware(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start upload software")
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
	dst, err := creatDir(info.ContentType, info.Version)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "create directory error"}
	}
	if err = saveUploadedFile(file, dst+"/"+info.ContentType+".zip"); err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "save file error"}
	}

	if err = extraZipFile(dst+"/"+info.ContentType+".zip", dst); err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "unzip file error"}
	}

	err = SoftwareDbCtlInstance().addSoftware(&info)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	return common.RespMsg{Status: common.Success}
}

func listRepository(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list software")
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

// getURL [method] get software url info
func getURL(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get software url")
	var urlReq URLReq
	if err := common.ParamConvert(input, &urlReq); err != nil {
		hwlog.RunLog.Error("class convert error")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	exist, err := checkSoftwareExist(urlReq.ContentType, urlReq.Version)
	if err != nil {
		hwlog.RunLog.Error("check software exist error in getURL func")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	if !exist {
		hwlog.RunLog.Errorf("%s%s dose not exist", urlReq.ContentType, urlReq.Version)
		return common.RespMsg{Status: common.ErrorGetResponse,
			Msg: "software dose not exist. Need to import software first"}
	}
	downloadInfo := DownloadInfo{}
	if err = fillDownloadData(&downloadInfo, &urlReq); err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: err.Error()}
	}
	return common.RespMsg{Status: common.Success, Data: downloadInfo}
}
