// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init hw logger
package common

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"

	"huawei.com/mindx/common/hwlog"
)

// InitHwlogger initialize run and operate logger
func InitHwlogger(ServerRunConf, ServerOpConf *hwlog.LogConfig) error {
	err := hwlog.InitRunLogger(ServerRunConf, context.Background())
	if err != nil {
		return err
	}
	err = hwlog.InitOperateLogger(ServerOpConf, context.Background())
	if err != nil {
		return err
	}
	return nil
}

// RespMsg response msg
type RespMsg struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
}

// ConstructResp construct response
func ConstructResp(c *gin.Context, errorCode string, msg string, data interface{}) {
	if msg == "" {
		msg = ErrorMap[errorCode]
	}
	result := RespMsg{
		Status: errorCode,
		Msg:    msg,
		Data:   data,
	}
	c.JSON(http.StatusOK, result)
}
