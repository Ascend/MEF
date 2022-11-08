// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package kubeclient to init kubeclient
package kubeclient

import (
	"fmt"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/k8stool"

	"k8s.io/client-go/kubernetes"
)

// ClientK8s ClientK8s struct
type ClientK8s struct {
	Clientset kubernetes.Interface
}

// NewClientK8s create ClientK8s
func NewClientK8s() (*ClientK8s, error) {
	client, err := k8stool.K8sClientFor("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}
	hwlog.RunLog.Info("init k8s success")
	return &ClientK8s{Clientset: client}, nil
}
