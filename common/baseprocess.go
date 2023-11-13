// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base process used
package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/rand"
)

// Router struct
type Router struct {
	Source      string
	Destination string
	Option      string
	Resource    string
}

const (
	maxRandomLen = 10240
	minSaltLen   = 16
)

// ClearSliceByteMemory clears slice in memory
func ClearSliceByteMemory(sliceByte []byte) {
	for i := 0; i < len(sliceByte); i++ {
		sliceByte[i] = 0
	}
}

// SendSyncMessageByRestful send sync message by restful
func SendSyncMessageByRestful(input interface{}, router *Router, timeout time.Duration) RespMsg {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("new message error")
		return RespMsg{Status: ErrorsSendSyncMessageByRestful, Msg: "", Data: nil}
	}

	msg.SetRouter(router.Source, router.Destination, router.Option, router.Resource)
	msg.FillContent(input)

	respMsg, err := modulemgr.SendSyncMessage(msg, timeout)
	if err != nil {
		hwlog.RunLog.Error("get response error")
		return RespMsg{Status: ErrorsSendSyncMessageByRestful, Msg: "", Data: nil}
	}
	return marshalResponse(respMsg)
}

func marshalResponse(respMsg *model.Message) RespMsg {
	content := respMsg.GetContent()
	respStr, err := json.Marshal(content)
	if err != nil {
		return RespMsg{Status: ErrorGetResponse, Msg: "", Data: nil}
	}
	var resp RespMsg
	if err := json.Unmarshal(respStr, &resp); err != nil {
		return RespMsg{Status: ErrorGetResponse, Msg: "", Data: nil}
	}
	return resp
}

// ParamConvert convert request parameter from restful module
func ParamConvert(input interface{}, reqType interface{}) error {
	inputStr, ok := input.(string)
	if !ok {
		hwlog.RunLog.Error("param type is not string")
		return errors.New("param type error")
	}
	dec := json.NewDecoder(strings.NewReader(inputStr))
	if err := dec.Decode(reqType); err != nil {
		hwlog.RunLog.Error("param decode failed")
		return errors.New("param decode error")
	}
	return nil
}

// Combine to combine option and resource to find url method
func Combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}

// GetSafeRandomBytes get safe random bytes
func GetSafeRandomBytes(saltLength int) ([]byte, error) {
	if saltLength > maxRandomLen || saltLength < minSaltLen {
		return nil, errors.New("salt length invalid")
	}
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// GracefulShutDown is the func to close a process gracefully
func GracefulShutDown(cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	select {
	case _, ok := <-signalChan:
		if !ok {
			hwlog.RunLog.Info("catch stop signal channel is closed")
		}
	}
	cancelFunc()
}

// LimitChecker check limit parameter
func LimitChecker(param httpsmgr.ServerParam, maxConcurrency, maxIPConnLimit int64) error {
	if res := checker.GetRegChecker("", limiter.IPReqLimitReg, true).Check(param.LimitIPReq); !res.Result {
		return fmt.Errorf("limitIPReq is invalid")
	}
	if res := checker.GetIntChecker("", 1, maxConcurrency, true).Check(param.LimitTotalConn); !res.Result {
		return fmt.Errorf("limitTotalConn %d is not in [%d, %d]", param.LimitTotalConn, 1, maxConcurrency)
	}
	if res := checker.GetIntChecker("", 1, limiter.DefaultDataLimit, true).Check(param.CacheSize); !res.Result {
		return fmt.Errorf("cacheSize %d is not in [%d, %d]", param.CacheSize, 1, limiter.DefaultDataLimit)
	}
	if res := checker.GetIntChecker("", 1, maxConcurrency, true).Check(param.Concurrency); !res.Result {
		return fmt.Errorf("concurrency %d is not in [%d, %d]", param.Concurrency, 1, maxConcurrency)
	}
	if res := checker.GetIntChecker("", 1, maxIPConnLimit, true).Check(param.LimitIPConn); !res.Result {
		return fmt.Errorf("limitIPConn %d is not in [%d, %d]", param.LimitIPConn, 1, maxIPConnLimit)
	}
	if res := checker.GetIntChecker("", 1, limiter.DefaultDataLimit, true).Check(param.BodySizeLimit); !res.Result {
		return fmt.Errorf("dataLimit %d is not in [%d, %d]", param.BodySizeLimit, 1, limiter.DefaultDataLimit)
	}
	return nil
}
