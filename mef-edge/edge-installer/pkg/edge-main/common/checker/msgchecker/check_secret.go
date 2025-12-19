// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
