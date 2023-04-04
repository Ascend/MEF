// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager operation table configmap_infos according to the request body
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
)

const maxConfigmapItemNum = 64

func createConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("create the configmap item start")

	var createCmReq ConfigmapReq
	if err := common.ParamConvert(input, &createCmReq); err != nil {
		hwlog.RunLog.Errorf("convert request parameter from restful service module failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	if err := checkItemCountInDB(); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	checker := configmapParaChecker{req: &createCmReq}
	if err := checker.Check(); err != nil {
		hwlog.RunLog.Errorf("configmap para check failed before creating, error: %v", err)
		return common.RespMsg{Status: "",
			Msg: fmt.Sprintf("configmap para check failed before creating, error: %s", err.Error()), Data: nil}
	}

	if err := createCmByK8S(&createCmReq); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	configmap, err := createCmReq.toDb()
	if err != nil {
		hwlog.RunLog.Errorf("convert request to configmapInfo failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("convert request to configmapInfo failed, error: %s",
			err.Error()), Data: nil}
	}

	if err = ConfigmapRepositoryInstance().createConfigmap(configmap); err != nil {
		hwlog.RunLog.Errorf("create configmap [%s] in db failed, error: %v", createCmReq.ConfigmapName, err)
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("create configmap [%s] in db failed, error: %s",
			createCmReq.ConfigmapName, err.Error()), Data: nil}
	}

	hwlog.RunLog.Infof("create configmap [%s] success", createCmReq.ConfigmapName)
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func createCmByK8S(createCmReq *ConfigmapReq) error {
	configmapToK8S := convertCmToK8S(createCmReq)

	_, err := kubeclient.GetKubeClient().CreateConfigMap(configmapToK8S)
	if err != nil {
		hwlog.RunLog.Errorf("create configmap [%s] by k8s failed, error: %v", createCmReq.ConfigmapName, err)
		return fmt.Errorf("create configmap [%s] by k8s failed, error: %s", createCmReq.ConfigmapName, err.Error())
	}

	hwlog.RunLog.Infof("create configmap [%s] by k8s success", createCmReq.ConfigmapName)
	return nil
}

func deleteConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("delete the configmap item start")

	var deleteCmReq DeleteConfigmapReq
	if err := common.ParamConvert(input, &deleteCmReq); err != nil {
		hwlog.RunLog.Errorf("convert request parameter from restful service module failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	var failedDeleteConfigmapIDs = make([]int64, 0, len(deleteCmReq.ConfigmapIDs))
	for _, configmapID := range deleteCmReq.ConfigmapIDs {
		configmapInfoFromDB, err := ConfigmapRepositoryInstance().queryConfigmapByID(configmapID)
		if err != nil {
			hwlog.RunLog.Errorf("query configmap [%d] from db failed, error: %v", configmapID, err)
			failedDeleteConfigmapIDs = append(failedDeleteConfigmapIDs, configmapID)
			continue
		}

		if ok := deleteCmByK8S(configmapInfoFromDB.ConfigmapName, configmapID); !ok {
			failedDeleteConfigmapIDs = append(failedDeleteConfigmapIDs, configmapID)
			continue
		}

		if err = ConfigmapRepositoryInstance().deleteConfigmapByID(configmapID); err != nil {
			hwlog.RunLog.Errorf("delete configmap [%d] from db failed, error: %v", configmapID, err)
			failedDeleteConfigmapIDs = append(failedDeleteConfigmapIDs, configmapID)
			continue
		}
	}

	if len(failedDeleteConfigmapIDs) > 0 {
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("delete configmap %d failed",
			failedDeleteConfigmapIDs), Data: nil}
	}

	hwlog.RunLog.Info("delete configmap success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func deleteCmByK8S(configmapName string, configmapID int64) bool {
	if err := kubeclient.GetKubeClient().DeleteConfigMap(configmapName); err != nil {
		hwlog.RunLog.Errorf("delete configmap [%d] by k8s failed, error: %v", configmapID, err)
		return false
	}

	hwlog.RunLog.Infof("delete configmap [%d] by k8s success", configmapID)
	return true
}

func updateConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("update the configmap item start")

	var updateCmReq ConfigmapReq
	var err error
	if err = common.ParamConvert(input, &updateCmReq); err != nil {
		hwlog.RunLog.Errorf("convert request parameter from restful service module failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	checker := configmapParaChecker{req: &updateCmReq}
	if err = checker.Check(); err != nil {
		hwlog.RunLog.Errorf("configmap para check failed before updating, error: %v", err)
		return common.RespMsg{Status: "",
			Msg: fmt.Sprintf("configmap para check failed before updating, error: %s", err.Error()), Data: nil}
	}

	if err = updateCmByK8S(&updateCmReq); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	configmapInfo, err := ConfigmapRepositoryInstance().queryConfigmapByName(updateCmReq.ConfigmapName)
	if err != nil {
		hwlog.RunLog.Errorf("query configmap [%s] from db failed, error: %v", updateCmReq.ConfigmapName, err)
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("query configmap [%s] from db failed, error: %s",
			updateCmReq.ConfigmapName, err.Error()), Data: nil}
	}

	configmapInfo.Description = updateCmReq.Description
	if err = convertCmContentInReqToDB(&updateCmReq, configmapInfo); err != nil {
		hwlog.RunLog.Errorf("convert configmap content in request to db failed: %v", err)
		return common.RespMsg{Status: "",
			Msg: fmt.Sprintf("convert configmap content in request to db failed: %s", err.Error()), Data: nil}
	}

	if err = ConfigmapRepositoryInstance().updateConfigmapByName(configmapInfo.ConfigmapName, configmapInfo); err != nil {
		hwlog.RunLog.Errorf("update configmap [%s] to db failed, error: %v", configmapInfo.ConfigmapName, err)
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("update configmap [%s] to db failed, error: %s",
			configmapInfo.ConfigmapName, err.Error()), Data: nil}
	}

	hwlog.RunLog.Infof("update configmap [%s] success", updateCmReq.ConfigmapName)
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func convertCmContentInReqToDB(updateCmReq *ConfigmapReq, configmapInfo *ConfigmapInfo) error {
	content, err := json.Marshal(updateCmReq.ConfigmapContent)
	if err != nil {
		return errors.New("marshal configmapContent info failed")
	}

	configmapInfo.ConfigmapContent = string(content)
	return nil
}

func updateCmByK8S(updateCmReq *ConfigmapReq) error {
	configmapK8S := convertCmToK8S(updateCmReq)

	_, err := kubeclient.GetKubeClient().UpdateConfigMap(configmapK8S)
	if err != nil {
		hwlog.RunLog.Errorf("update configmap [%s] by k8s failed, error: %v", updateCmReq.ConfigmapName, err)
		return fmt.Errorf("update configmap [%s] by k8s failed, error: %s", updateCmReq.ConfigmapName, err.Error())
	}

	hwlog.RunLog.Infof("update configmap [%s] by k8s success", updateCmReq.ConfigmapName)
	return nil
}

func queryConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("query the configmap item start")

	configmapID, ok := input.(int64)
	if !ok {
		hwlog.RunLog.Error("get configmap id failed: param type is not int64")
		return common.RespMsg{Status: "", Msg: "param type is not int64", Data: nil}
	}

	configmapInfo, err := ConfigmapRepositoryInstance().queryConfigmapByID(configmapID)
	if err != nil {
		hwlog.RunLog.Errorf("query configmap [%d] from db failed, error: %v", configmapID, err)
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("query configmap failed, error: %s",
			err.Error()), Data: nil}
	}

	createdAt := configmapInfo.CreatedAt.Format(common.TimeFormat)
	updatedAt := configmapInfo.UpdatedAt.Format(common.TimeFormat)
	queryCmReturnInfo := QueryConfigmapReturnInfo{
		ConfigmapID:   configmapInfo.ConfigmapID,
		ConfigmapName: configmapInfo.ConfigmapName,
		Description:   configmapInfo.Description,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
	if err = json.Unmarshal([]byte(configmapInfo.ConfigmapContent), &queryCmReturnInfo.ConfigmapContent); err != nil {
		hwlog.RunLog.Errorf("unmarshal configmap content info failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("unmarshal configmap content info failed, error: %s",
			err.Error()), Data: nil}
	}

	hwlog.RunLog.Infof("query configmap [%d] from db success", configmapID)
	return common.RespMsg{Status: common.Success, Msg: "", Data: queryCmReturnInfo}
}

func listConfigmap(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("list the configmap items start")

	listReq, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("get list request failed: para type is not ListReq")
		return common.RespMsg{Status: "", Msg: "para type is not ListReq", Data: nil}
	}

	configmaps, err := getListConfigmapReturnInfo(listReq)
	if err != nil {
		hwlog.RunLog.Errorf("get configmap infos list from db failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("get configmap infos list from db failed, error: %s",
			err.Error()), Data: nil}
	}

	configmaps.Total, err = ConfigmapRepositoryInstance().configmapInfosListCountByName(listReq.Name)
	if err != nil {
		hwlog.RunLog.Errorf("get configmap infos list count by name failed, error: %v", err)
		return common.RespMsg{Status: "",
			Msg: fmt.Sprintf("get configmap infos list count by name failed, error: %s", err.Error()), Data: nil}
	}

	hwlog.RunLog.Info("list configmap items from db success")
	return common.RespMsg{Status: common.Success, Msg: "list configmap items success", Data: configmaps}
}

func getListConfigmapReturnInfo(listReq types.ListReq) (*ListConfigmapReturnInfo, error) {
	cmInfoList, err := ConfigmapRepositoryInstance().listConfigmapInfo(listReq.PageNum,
		listReq.PageSize, listReq.Name)
	if err != nil {
		return nil, err
	}

	var cmInstanceResp []ConfigmapInstanceResp
	var cmInfoFromDB *ConfigmapInfo
	for _, cmInfo := range cmInfoList {
		cmInfoFromDB, err = ConfigmapRepositoryInstance().queryConfigmapByID(cmInfo.ConfigmapID)
		if err != nil {
			hwlog.RunLog.Errorf("query configmap [%d] from db failed, error: %v", cmInfo.ConfigmapID, err)
			return nil, err
		}

		createdAt := cmInfo.CreatedAt.Format(common.TimeFormat)
		updatedAt := cmInfo.UpdatedAt.Format(common.TimeFormat)
		instanceResp := ConfigmapInstanceResp{
			ConfigmapID:   cmInfoFromDB.ConfigmapID,
			ConfigmapName: cmInfoFromDB.ConfigmapName,
			Description:   cmInfoFromDB.Description,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		}

		if err = json.Unmarshal([]byte(cmInfo.ConfigmapContent), &instanceResp.ConfigmapContent); err != nil {
			hwlog.RunLog.Errorf("unmarshal configmap [%d] content failed, error: %v", cmInfo.ConfigmapID, err)
			return nil, fmt.Errorf("unmarshal configmap [%d] content failed", cmInfo.ConfigmapID)
		}

		cmInstanceResp = append(cmInstanceResp, instanceResp)
	}

	return &ListConfigmapReturnInfo{
		ConfigmapInstance: cmInstanceResp,
	}, nil
}
