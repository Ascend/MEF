// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package msgchecker for Secret
package msgchecker

import (
	"fmt"
	"strings"

	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
)

const (
	imageSecretDataKey     = ".dockerconfigjson"
	fdDockerImageSecretUid = "fusion-director-docker-registry-secret"
)

func (mv *MsgValidator) auxCheckSecret(secret *types.Secret) error {
	defer func() {
		for key := range secret.Data {
			utils.ClearSliceByteMemory(secret.Data[key])
		}
	}()

	if err := validateStruct(secret); err != nil {
		if !strings.Contains(err.Error(), "Secret.ObjectMeta.UID") || (secret.UID != fdDockerImageSecretUid) {
			return err
		}
	}

	if err := checkSecretDataKey(secret); err != nil {
		return err
	}

	return nil
}

func checkSecretDataKey(secret *types.Secret) error {
	if _, found := secret.Data[imageSecretDataKey]; !found {
		return fmt.Errorf("fd secret data key invalid")
	}

	return nil
}
