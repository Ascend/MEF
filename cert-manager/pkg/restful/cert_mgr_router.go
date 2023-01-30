// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful init cert restful service
package restful

import (
	"fmt"

	"cert-manager/pkg/certconstant"
	"cert-manager/pkg/certmgr"

	"github.com/gin-gonic/gin"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

type csrJson struct {
	CertName string `json:"certName"`
	Csr      string `json:"csr"`
}

func queryRootCA(c *gin.Context) {
	certName := c.Query("certName")
	// todo 增加checker 验证
	ca, err := certmgr.QueryRootCa(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query cert [%s] root ca failed: %v", certName, err)
		common.ConstructResp(c, certconstant.ErrorGetRootCa, "Query root ca failed", "")
		return
	}
	hwlog.RunLog.Infof("query [%s] root ca success", certName)
	common.ConstructResp(c, common.Success, "query ca success", string(ca))
}

func issueServiceCa(c *gin.Context) {
	var csrJsonData csrJson
	err := c.BindJSON(&csrJsonData)
	if err != nil {
		hwlog.RunLog.Errorf("issue service cert failed, bind json data failed: %v", err)
		common.ConstructResp(c, certconstant.ErrorIssueSrvCert, "bind json data failed", nil)
		return
	}
	// todo 增加checker 验证
	cert, err := certmgr.IssueServiceCert(csrJsonData.CertName, csrJsonData.Csr)
	if err != nil {
		hwlog.RunLog.Errorf("issue service certificate failed: %v", err)
		common.ConstructResp(c, certconstant.ErrorIssueSrvCert, "issue service certificate failed", nil)
		return
	}
	hwlog.RunLog.Infof("issue [%s] service certificate success", csrJsonData.CertName)
	common.ConstructResp(c, common.Success, "issue success", string(cert))
}

func importRootCa(c *gin.Context) {
	// todo 待实现
	common.ConstructResp(c, common.Success, "import success", nil)
}

func queryAlert(c *gin.Context) {
	// todo 待实现
	var alertList = [...]string{"alert 1"}
	common.ConstructResp(c, common.Success, "query alert success", alertList)
}

func versionQuery(c *gin.Context) {
	msg := fmt.Sprintf("%s version: %s", BuildNameStr, BuildVersionStr)
	common.ConstructResp(c, common.Success, "", msg)
}
