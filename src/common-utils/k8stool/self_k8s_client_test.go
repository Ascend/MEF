// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package k8s offer the k8s client with support encoded kubeConfig file
package k8stool

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

func getAbsPath(relPath string, t *testing.T) string {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		t.Fatalf("get client key abs path failed")
	}
	return absPath
}

// TestBuildConfigFromFlags test function for BuildConfigFromFlags
func TestBuildConfigFromFlags(t *testing.T) {
	kubeconfigBytes, err := fileutils.ReadLimitBytes(getAbsPath("./testdata/test.conf", t), fileutils.Size10M)
	if err != nil {
		return
	}
	initStub := gomonkey.ApplyFunc(bytes.Contains, func(s, prefix []byte) bool {
		return false
	})
	defer initStub.Reset()
	kmc2 := gomonkey.ApplyFunc(kmc.Initialize, func(sdpAlgID int, primaryKey, standbyKey string) error {
		return nil
	})
	defer kmc2.Reset()
	decrypt := gomonkey.ApplyFunc(kmc.Decrypt, func(domainID uint, data []byte) ([]byte, error) {
		return kubeconfigBytes, nil
	})
	defer decrypt.Reset()
	convey.Convey("relative path", t, func() {
		config, err := BuildConfigFromFlags("", getAbsPath("./testdata/test.conf", t))
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(config.Host, convey.ShouldEqual, "https://127.0.0.1:6443")
	})
	convey.Convey("no path provide", t, func() {
		config, err := BuildConfigFromFlags("", "")
		convey.So(err.Error(), convey.ShouldEqual, "no ExplicitPath set")
		convey.So(config, convey.ShouldEqual, nil)
	})
	kubeconfigBytes, err = fileutils.ReadLimitBytes(getAbsPath("./testdata/test.conf", t), fileutils.Size10M)
	if err != nil {
		return
	}
	convey.Convey("init client", t, func() {
		rawEnv := os.Getenv("KUBECONFIG")
		if err := os.Setenv("KUBECONFIG", getAbsPath("./testdata/test.conf", t)); err != nil {
			fmt.Println("set env failed")
			t.FailNow()
		}
		defer func() {
			if err := os.Setenv("KUBECONFIG", rawEnv); err != nil {
				fmt.Println("set env failed")
			}
		}()

		cli, err := K8sClient("")
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(cli, convey.ShouldNotEqual, nil)
	})
}

func TestK8sClientFor(t *testing.T) {
	convey.Convey("get from init client", t, func() {
		cli, err := K8sClientFor("", "")
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(cli, convey.ShouldNotEqual, nil)
	})
}
