// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package configmanager to init config manager service
package configmanager

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindx/common/xcrypto"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"edge-manager/pkg/config"
	"edge-manager/pkg/configmanager/configchecker"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
)

const (
	saltLen   = 16
	retryTime = 5
)

// downloadConfig download image config
func downloadConfig(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to generate configuration of image registry")
	var req ImageConfig
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("parse parameter failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	defer func() {
		common.ClearSliceByteMemory(req.Password)
	}()
	if checkResult := configchecker.NewConfigChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("image config para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	imageAddress, err := createSecret(req)
	if err != nil {
		hwlog.RunLog.Error("create k8s secret failed")
		return common.RespMsg{Status: common.ErrorCreateSecret, Msg: "create k8s secret failed", Data: nil}
	}
	if err := fetchCertToClient(imageAddress); err != nil {
		hwlog.RunLog.Errorf("distribute cert file to client failed, error:%v", err)
	}
	hwlog.RunLog.Info("create image config success")
	return common.RespMsg{Status: common.Success, Msg: "create image config success", Data: nil}
}

func createSecret(config ImageConfig) (string, error) {
	auth := []byte(config.Account + ":" + string(config.Password))
	base64Auth := make([]byte, base64.StdEncoding.EncodedLen(len(auth)))
	base64.StdEncoding.Encode(base64Auth, auth)
	registryPath := config.IP + ":" + strconv.Itoa(int(config.Port))
	if config.Domain != "" {
		registryPath = config.Domain + ":" + strconv.Itoa(int(config.Port))
	}
	data := assembleSecretData(registryPath, base64Auth, &config)
	defer func() {
		common.ClearSliceByteMemory(auth)
		common.ClearSliceByteMemory(data)
	}()
	userSecret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeclient.DefaultImagePullSecretKey,
		},
		Type: apiv1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{apiv1.DockerConfigJsonKey: data},
	}
	if _, err := kubeclient.GetKubeClient().CreateOrUpdateSecret(userSecret); err != nil {
		hwlog.RunLog.Error("create or update secret failed")
		return "", err
	}
	return registryPath, nil
}

func fetchCertToClient(registryPath string) error {
	certRes, err := util.GetCertContent(common.ImageCertName)
	if err != nil {
		hwlog.RunLog.Errorf("get cert content failed, error: %v", err)
		return errors.New("get cert content failed")
	}
	if certRes.CertContent == "" {
		hwlog.RunLog.Warnf(" %s cert content should be imported", certRes.CertName)
		return nil
	}
	certRes.ImageAddress = registryPath
	// send message connector
	if err := reportCertToClient(certRes); err != nil {
		return errors.New("update cert content failed")
	}
	return nil
}

func assembleSecretData(registryPath string, base64Auth []byte, config *ImageConfig) []byte {
	var data bytes.Buffer
	data.WriteString(`{"auths":{"` + registryPath + `":{"auth":"`)
	data.Write(base64Auth)
	data.WriteString(`","docker-password":"`)
	data.Write(config.Password)
	data.WriteString(`","docker-username":"` + config.Account + `"}}}`)
	defer func() {
		common.ClearSliceByteMemory(config.Password)
		common.ClearSliceByteMemory(base64Auth)
	}()
	defer data.Reset()
	return data.Bytes()
}

// updateConfig update image config
func updateConfig(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("update cert content start")
	var updateCert certutils.UpdateClientCert
	if err := common.ParamConvert(input, &updateCert); err != nil {
		hwlog.RunLog.Error("update cert info failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "update cert info request convert error", Data: nil}
	}
	certRes := certutils.ClientCertResp{
		CertName:    updateCert.CertName,
		CertContent: string(updateCert.CertContent),
		CertOpt:     updateCert.CertOpt,
	}
	if updateCert.CertName == common.ImageCertName {
		address, err := util.GetImageAddress()
		if err != nil {
			hwlog.RunLog.Errorf("get image registry address failed, error:%v", err)
			return common.RespMsg{Status: common.ErrorGetSecret, Msg: err.Error(), Data: nil}
		}
		if address == "" {
			hwlog.RunLog.Warn("image registry address should be configured")
			return common.RespMsg{Status: common.Success, Msg: "update cert content success", Data: certRes.CertName}
		}
		certRes.ImageAddress = address
	}
	// send message to connector
	if err := reportCertToClient(certRes); err != nil {
		hwlog.RunLog.Errorf("update cert content failed, %v", err)
		return common.RespMsg{Status: common.ErrorDistributeRootCa, Msg: "update cert content failed", Data: nil}
	}
	hwlog.RunLog.Info("update cert content success")
	return common.RespMsg{Status: common.Success, Msg: "update cert content success", Data: certRes.CertName}
}

func reportCertToClient(req certutils.ClientCertResp) error {
	content, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal info failed, error: %v", err)
	}
	nodes, err := getAllNodeInfo()
	if err != nil {
		return fmt.Errorf("get all node info failed, error: %v", err)
	}
	router := common.Router{
		Source:      common.ConfigManagerName,
		Destination: common.CloudHubName,
		Option:      common.OptPost,
		Resource:    common.ResDownLoadCert,
	}
	for _, node := range nodes {
		if err := sendMessageToNode(node.SerialNumber, string(content), router); err != nil {
			hwlog.RunLog.Warnf("send message to node [%s], error: %v", node.SerialNumber, err)
			continue
		}
	}
	return nil
}

func sendMessageToNode(serialNumber string, content string, router common.Router) error {
	sendMsg, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("create new message failed, error: %v", err)
	}
	sendMsg.SetNodeId(serialNumber)
	sendMsg.SetRouter(router.Source, router.Destination, router.Option, router.Resource)
	sendMsg.FillContent(content)
	if err = modulemgr.SendMessage(sendMsg); err != nil {
		return fmt.Errorf("%s sends message to %s failed, error: %v",
			common.ConfigManagerName, common.CloudHubName, err)
	}
	return nil
}

func exportToken(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("export token start")
	plainText, err := generateToken()
	if err != nil {
		return common.RespMsg{Status: common.ErrorExportToken}
	}
	defer utils.ClearSliceByteMemory(plainText)

	hwlog.RunLog.Info("export token success")
	return common.RespMsg{Status: common.Success, Msg: "export token success", Data: string(plainText)}
}

func generateToken() ([]byte, error) {
	var plainText []byte
	var err error
	var index int
	for index = 0; index < retryTime; index++ {
		plainText, err = x509.GetRandomPass()
		if err != nil && strings.Contains(err.Error(), "the password is too simple") {
			hwlog.RunLog.Warn("generate token is simple, try again")
			continue
		} else if err != nil {
			hwlog.RunLog.Errorf("generate token failed: %v", err)
			return nil, errors.New("generate raw token failed")
		}
		break
	}
	if index >= retryTime {
		return nil, errors.New("generate token is simple")
	}
	salt, err := common.GetSafeRandomBytes(saltLen)
	if err != nil {
		hwlog.RunLog.Errorf("generate salt failed: %v", err)
		return nil, errors.New("generate salt failed")
	}

	token, err := xcrypto.Pbkdf2WithSha256(plainText, salt, common.Pbkdf2IterationCount, common.BytesOfEncryptedString)
	if err != nil {
		hwlog.RunLog.Errorf("encrypt token failed: %v", err)
		return nil, errors.New("encrypt token failed")
	}

	durationTime := time.Duration(config.GetAuthConfig().TokenExpireTime) * common.OneDay
	tokenInfo := TokenInfo{
		Token:      token,
		Salt:       salt,
		ExpireTime: time.Now().Add(durationTime).Unix(),
	}
	if err := ConfigRepositoryInstance().saveToken(tokenInfo); err != nil {
		hwlog.RunLog.Error(err)
		return nil, errors.New("save token failed")
	}
	return plainText, nil
}

func checkAndUpdateToken() {
	expire, err := ConfigRepositoryInstance().ifTokenExpire()
	if err != nil {
		hwlog.RunLog.Errorf("check if token expire error:%v", err)
		return
	}
	if !expire {
		return
	}
	if err = ConfigRepositoryInstance().revokeToken(); err != nil {
		hwlog.RunLog.Errorf("token is expire, error :%v", err)
		return
	}
	hwlog.RunLog.Info("token is expire, system auto revoke token")
}
