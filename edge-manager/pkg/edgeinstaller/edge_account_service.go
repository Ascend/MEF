// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller for operation table edge_account_infos according to the request body
package edgeinstaller

import (
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/passutils"
)

func setEdgeAccount(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("set edge account start")

	var setEdgeAccountReq SetEdgeAccountReq
	if err := common.ParamConvert(input, &setEdgeAccountReq); err != nil {
		hwlog.RunLog.Errorf("convert request parameter from restful service module failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	defer func() {
		common.ClearStringMemory(*setEdgeAccountReq.Password)
		common.ClearStringMemory(*setEdgeAccountReq.ConfirmPassword)
	}()

	if setEdgeAccountReq.Account != DefaultAccountName {
		hwlog.RunLog.Error("account is not default account")
		return common.RespMsg{Status: common.ErrorAccountOrPassword, Msg: "incorrect account or password", Data: nil}
	}

	if checkResult := NewSetEdgeAccountChecker().Check(setEdgeAccountReq); !checkResult.Result {
		hwlog.RunLog.Errorf("check setEdgeAccountReq param failed before creating, error: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "parameter invalid", Data: nil}
	}
	hwlog.RunLog.Info("check setEdgeAccountReq param success")

	encryptPassWord, salt, err := passutils.GetEncryptPassword(setEdgeAccountReq.Password)
	if err != nil {
		hwlog.RunLog.Errorf("get encrypt password failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSetEdgeAccountPassword, Msg: "set edge account password failed", Data: nil}
	}

	edgeAccountInfo := &EdgeAccountInfo{
		Account:   setEdgeAccountReq.Account,
		Password:  encryptPassWord,
		Salt:      salt,
		UpdatedAt: time.Now().Format(common.TimeFormat),
	}

	if err = EdgeAccountRepositoryInstance().setEdgeAccountInfo(edgeAccountInfo); err != nil {
		hwlog.RunLog.Errorf("set edge account in db failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSetEdgeAccount, Msg: "set edge account failed", Data: nil}
	}

	hwlog.RunLog.Info("set edge account success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
