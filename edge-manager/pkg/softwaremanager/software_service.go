// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager to deal software info
package softwaremanager

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

const (
	maxSftUrlCount = 16
)

func updateAuthInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to update software auth info")
	var req SftAuthInfo

	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("parse parameter failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	defer func() {
		if req.Password != nil {
			common.ClearSliceByteMemory(*req.Password)
		}
	}()

	if checkResult := newSftAuthInfoChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("software auth info check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	encryptedData, err := common.EncryptContent(*req.Password, nil)
	if err != nil {
		hwlog.RunLog.Errorf("encrypt software auth info failed: %v", err)
		return common.RespMsg{Status: common.ErrorEncryptAuthInfo, Msg: "encrypt software auth info failed", Data: nil}
	}

	var sftAuthInfo = SftAuthInfo{UserName: req.UserName, Password: &encryptedData}

	data, err := json.Marshal(sftAuthInfo)
	if err != nil {
		hwlog.RunLog.Errorf("marshal software auth info failed: %v", err)
		return common.RespMsg{Status: common.ErrorMarshalFailed, Msg: "marshal software auth info failed", Data: nil}
	}

	if err = sftRepositoryInstance().insertOrUpdate("auth_info", string(data)); err != nil {
		hwlog.RunLog.Errorf("update software url info failed: %v", err)
		return common.RespMsg{Status: common.ErrorUpdateAuthInfo, Msg: "update software auth info failed", Data: nil}
	}

	hwlog.RunLog.Info("update software auth info success")
	return common.RespMsg{Status: common.Success, Msg: "update software auth info success", Data: nil}
}

func splitUrlInfo(req UrlUpdateInfo) map[string][]UrlInfo {
	var res = make(map[string][]UrlInfo)
	for _, urlInfo := range req.UrlInfos {
		res[urlInfo.Type] = append(res[urlInfo.Type], urlInfo)
	}

	return res
}

func getUrlInfoFromDb(urlType string) []UrlInfo {
	var urlInfos []UrlInfo
	data, err := sftRepositoryInstance().query(urlType)
	if err != nil {
		hwlog.RunLog.Errorf("query software url info failed: %v", err)
		return urlInfos
	}

	if data == "" {
		hwlog.RunLog.Infof("[%s] software url info is none", urlType)
		return urlInfos
	}

	if err = json.Unmarshal([]byte(data), &urlInfos); err != nil {
		hwlog.RunLog.Errorf("unmarshal software url info failed: %v", err)
		return urlInfos
	}

	return urlInfos
}

func getAuthInfoFromDb() (SftAuthInfo, error) {
	var sftAuthInfo SftAuthInfo
	data, err := sftRepositoryInstance().query("auth_info")
	if err != nil {
		hwlog.RunLog.Errorf("get software auth info failed: %v", err)
		return sftAuthInfo, errors.New("get software auth info failed")
	}

	if err = json.Unmarshal([]byte(data), &sftAuthInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal software url info failed: %v", err)
		return sftAuthInfo, errors.New("get software auth info failed")
	}

	return sftAuthInfo, nil
}

func setUrlInfoToDb(urlType string, urlInfos []UrlInfo) error {
	data, err := json.Marshal(urlInfos)
	if err != nil {
		hwlog.RunLog.Errorf("marshal url info failed: %v", err)
		return errors.New("marshal url info failed")
	}

	if err = sftRepositoryInstance().insertOrUpdate(urlType, string(data)); err != nil {
		hwlog.RunLog.Errorf("update url info failed: %v", err)
		return errors.New("update info failed")
	}

	return nil
}

func updateSftUrlInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to update software url info")
	var req UrlUpdateInfo
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("parse parameter failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	if checkResult := newSfwUrlInfoChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("software url info check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	reqUrlInfos := splitUrlInfo(req)
	for urlType, urlInfo := range reqUrlInfos {
		oldUrlInfo := getUrlInfoFromDb(urlType)

		urlOpr := newUrlOperator(oldUrlInfo, req.Option)
		err := urlOpr.operate(urlInfo)
		if err != nil {
			hwlog.RunLog.Errorf("update software url info failed :%v", err)
			return common.RespMsg{Status: common.ErrorUpdateUrlInfo, Msg: "update software url info failed", Data: nil}
		}

		if err = setUrlInfoToDb(urlType, urlOpr.urlInfos); err != nil {
			hwlog.RunLog.Errorf("update software url info failed :%v", err)
			return common.RespMsg{Status: common.ErrorUpdateUrlInfo, Msg: "update software url info failed", Data: nil}
		}
	}

	hwlog.RunLog.Info("update software url info success")
	return common.RespMsg{Status: common.Success, Msg: "update software url info success", Data: nil}
}

func innerGetSftDownloadInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to get software url info")

	urlType, ok := input.(string)
	if !ok {
		hwlog.RunLog.Error("get message content failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message content failed", Data: nil}
	}

	urlInfos := getUrlInfoFromDb(urlType)
	if len(urlInfos) == 0 {
		hwlog.RunLog.Errorf("no [%s] software url info in db", urlType)
		return common.RespMsg{Status: common.ErrorInnerGetData, Msg: "no software url info in db", Data: nil}
	}

	sftAuthInfo, err := getAuthInfoFromDb()
	if err != nil {
		hwlog.RunLog.Errorf("get software auth info failed")
		return common.RespMsg{Status: common.ErrorInnerGetData, Msg: "get software  auth info failed", Data: nil}
	}

	decryptedData, err := common.DecryptContent(*sftAuthInfo.Password, nil)
	if err != nil {
		hwlog.RunLog.Errorf("decrypt software auth info failed: %v", err)
		return common.RespMsg{Status: common.ErrorDecryptAuthInfo, Msg: "decrypt software auth info failed", Data: nil}
	}

	var downloadInfo = types.DownloadInfo{Package: urlInfos[0].Url,
		UserName: sftAuthInfo.UserName,
		Password: &decryptedData}

	hwlog.RunLog.Infof("inner get software [%s] download info success", urlType)
	return common.RespMsg{Status: common.Success, Msg: "inner get software download info success", Data: downloadInfo}
}
