// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager operation table configmap_infos according to the request
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/appmanager/appchecker"
	"edge-manager/pkg/types"
	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
)

func createConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("create the configmap start")

	var createCmReq ConfigmapReq
	if err := common.ParamConvert(input, &createCmReq); err != nil {
		hwlog.RunLog.Errorf("convert request param failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	if err := CheckItemCountInDB(); err != nil {
		return common.RespMsg{Status: common.ErrorCheckCmCount, Msg: err.Error(), Data: nil}
	}

	if checkResult := appchecker.NewCreateCmChecker().Check(createCmReq); !checkResult.Result {
		hwlog.RunLog.Errorf("create configmap param check failed, error: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	if err := NewCmSupplementalChecker(createCmReq).Check(); err != nil {
		hwlog.RunLog.Errorf("supplemental param check failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error(), Data: nil}
	}

	id, err := CmRepositoryInstance().createCm(&createCmReq)
	if err != nil {
		if err.Error() == common.ErrDbUniqueFailed {
			return common.RespMsg{Status: common.ErrorAppMrgDuplicate, Msg: "configmap name is duplicate", Data: nil}
		}
		return common.RespMsg{Status: common.ErrorCreateCm, Msg: err.Error(), Data: nil}
	}

	hwlog.RunLog.Infof("create configmap [%s] success", createCmReq.ConfigmapName)
	return common.RespMsg{Status: common.Success, Msg: "", Data: id}
}

func deleteConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("delete the configmap start")

	var deleteCmReq DeleteCmReq
	if err := common.ParamConvert(input, &deleteCmReq); err != nil {
		hwlog.RunLog.Errorf("convert request param failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	if checkResult := appchecker.NewDeleteCmChecker().Check(deleteCmReq); !checkResult.Result {
		hwlog.RunLog.Errorf("delete configmap param check failed, error: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	var res types.BatchResp
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap
	for _, configmapID := range deleteCmReq.ConfigmapIDs {

		if err := CmRepositoryInstance().deleteSingleCm(configmapID); err != nil {
			failedMap[strconv.Itoa(int(configmapID))] = err.Error()
			continue
		}

		res.SuccessIDs = append(res.SuccessIDs, configmapID)
	}

	if len(res.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteCm, Msg: "", Data: res}
	}

	hwlog.RunLog.Info("delete configmap success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func updateConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("update the configmap start")

	var updateCmReq ConfigmapReq
	var err error
	if err = common.ParamConvert(input, &updateCmReq); err != nil {
		hwlog.RunLog.Errorf("convert request param failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	if checkResult := appchecker.NewUpdateCmChecker().Check(updateCmReq); !checkResult.Result {
		hwlog.RunLog.Errorf("update configmap param check failed, error: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	if err = NewCmSupplementalChecker(updateCmReq).Check(); err != nil {
		hwlog.RunLog.Errorf("supplemental param check failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error(), Data: nil}
	}

	err = CmRepositoryInstance().updateCm(&updateCmReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.RespMsg{Status: common.ErrorAppMrgRecodeNoFound, Msg: "configmap does not exist", Data: nil}
		} else if err.Error() == common.ErrDbUniqueFailed {
			return common.RespMsg{Status: common.ErrorAppMrgDuplicate, Msg: "configmap name is duplicate", Data: nil}
		}
		return common.RespMsg{Status: common.ErrorUpdateCm, Msg: err.Error(), Data: nil}
	}

	hwlog.RunLog.Infof("update configmap [%s] success", updateCmReq.ConfigmapName)
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func queryConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("query the configmap start")

	configmapID, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("get configmap id failed: param type is not uint64")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "param type is not uint64", Data: nil}
	}

	if checkResult := appchecker.NewQueryCmChecker().Check(configmapID); !checkResult.Result {
		hwlog.RunLog.Errorf("query configmap param check failed, error: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	cmInfo, err := CmRepositoryInstance().queryCmByID(configmapID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			hwlog.RunLog.Errorf("configmap id [%d] does not exist", configmapID)
			return common.RespMsg{Status: common.ErrorAppMrgRecodeNoFound, Msg: "configmap does not exist", Data: nil}
		}

		hwlog.RunLog.Errorf("query configmap [%d] from db failed, error: %v", configmapID, err)
		return common.RespMsg{Status: common.ErrorQueryCm, Msg: "query configmap from db failed", Data: nil}
	}

	queryCmResp := ConfigmapInstance{
		ConfigmapID:   cmInfo.ID,
		ConfigmapName: cmInfo.ConfigmapName,
		Description:   cmInfo.Description,
		CreatedAt:     cmInfo.CreatedAt.Format(common.TimeFormat),
		UpdatedAt:     cmInfo.UpdatedAt.Format(common.TimeFormat),
	}
	if err = json.Unmarshal([]byte(cmInfo.ConfigmapContent), &queryCmResp.ConfigmapContent); err != nil {
		hwlog.RunLog.Errorf("unmarshal configmap content info failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorUnmarshalCm, Msg: "unmarshal configmap info failed", Data: nil}
	}
	if cmInfo.AssociatedAppList != "" { // 此处若直接对空切片进行反序列化：unexpected end of JSON input
		if err = json.Unmarshal([]byte(cmInfo.AssociatedAppList), &queryCmResp.AssociatedAppList); err != nil {
			hwlog.RunLog.Errorf("unmarshal configmap associated app info failed, error: %v", err)
			return common.RespMsg{Status: common.ErrorUnmarshalCm, Msg: "unmarshal configmap info failed", Data: nil}
		}
	}

	queryCmResp.AssociatedAppNum = uint64(len(queryCmResp.AssociatedAppList))

	hwlog.RunLog.Infof("query configmap [%d] from db success", configmapID)
	return common.RespMsg{Status: common.Success, Msg: "", Data: queryCmResp}
}

func listConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("list the configmap start")

	listReq, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("get list request failed: param type is not ListReq")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "convert list request error", Data: nil}
	}

	if checkResult := util.NewPaginationQueryChecker().Check(listReq); !checkResult.Result {
		hwlog.RunLog.Errorf("list configmap param check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	configmaps, err := getListConfigmapResp(listReq)
	if err != nil {
		hwlog.RunLog.Errorf("get configmap infos list from db failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorListCm, Msg: err.Error(), Data: nil}
	}

	configmaps.Total, err = CmRepositoryInstance().cmListCountByName(listReq.Name)
	if err != nil {
		hwlog.RunLog.Errorf("get configmap infos list count by name failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorListCm, Msg: "get cm list count error", Data: nil}
	}

	hwlog.RunLog.Info("list configmap items from db success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: configmaps}
}

func getListConfigmapResp(listReq types.ListReq) (*ListConfigmapResp, error) {
	cmInfoList, err := CmRepositoryInstance().listCmInfo(listReq.PageNum, listReq.PageSize, listReq.Name)
	if err != nil {
		return nil, err
	}

	var cmInstanceResp []ConfigmapInstance
	for _, cmInfo := range cmInfoList {
		instanceResp := ConfigmapInstance{
			ConfigmapID:   cmInfo.ID,
			ConfigmapName: cmInfo.ConfigmapName,
			Description:   cmInfo.Description,
			CreatedAt:     cmInfo.CreatedAt.Format(common.TimeFormat),
			UpdatedAt:     cmInfo.UpdatedAt.Format(common.TimeFormat),
		}
		if err = json.Unmarshal([]byte(cmInfo.ConfigmapContent), &instanceResp.ConfigmapContent); err != nil {
			hwlog.RunLog.Errorf("unmarshal configmap [%d] content failed, error: %v", cmInfo.ID, err)
			return nil, fmt.Errorf("unmarshal configmap [%d] content failed", cmInfo.ID)
		}
		if cmInfo.AssociatedAppList != "" { // 此处若直接对空切片进行反序列化：unexpected end of JSON input
			if err = json.Unmarshal([]byte(cmInfo.AssociatedAppList), &instanceResp.AssociatedAppList); err != nil {
				hwlog.RunLog.Errorf("unmarshal configmap associated app info failed, error: %v", err)
				return nil, fmt.Errorf("unmarshal configmap [%d] associated app failed", cmInfo.ID)
			}
		}
		instanceResp.AssociatedAppNum = uint64(len(instanceResp.AssociatedAppList))

		cmInstanceResp = append(cmInstanceResp, instanceResp)
	}

	return &ListConfigmapResp{Configmaps: cmInstanceResp}, nil
}

const maxCmItemNum = 1000

// CheckItemCountInDB check item count in db
func CheckItemCountInDB() error {
	total, err := common.GetItemCount(ConfigmapInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get table configmap_infos num failed, error: %v", err)
		return errors.New("get table configmap_infos num failed")
	}

	if total >= maxCmItemNum {
		hwlog.RunLog.Error("table configmap_infos item num is enough, can't be created")
		return errors.New("table configmap_infos item num is enough, can't be created")
	}

	hwlog.RunLog.Info("check item count in database success")
	return nil
}
