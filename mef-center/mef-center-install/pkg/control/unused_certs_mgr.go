// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package control for
package control

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// UnusedCertsMgr manages unused certs
type UnusedCertsMgr struct {
	pathMgr *util.InstallDirPathMgr
	operate string
	caName  string
}

// InitUnusedCertsMgr creates a new UnusedCertsMgr
func InitUnusedCertsMgr(pathMgr *util.InstallDirPathMgr, operate, caName string) *UnusedCertsMgr {
	return &UnusedCertsMgr{
		pathMgr: pathMgr,
		operate: operate,
		caName:  caName,
	}
}

// DoOperate operates the unused certs
func (ucm UnusedCertsMgr) DoOperate() error {
	if err := ucm.checkParams(); err != nil {
		hwlog.RunLog.Errorf("failed to check parameters, %v", err)
		return err
	}

	if err := util.ReducePriv(); err != nil {
		hwlog.RunLog.Errorf("unable to reduce privilege, error: %v", err)
		return errors.New("unable to reduce privilege")
	}
	defer func() {
		if err := util.ResetPriv(); err != nil {
			hwlog.RunLog.Errorf("unable to reset privilege, error: %v", err)
		}
	}()

	switch ucm.operate {
	case util.GetUnusedCertOperateFlag:
		return ucm.getUnusedCerts()
	case util.RestoreCertOperateFlag:
		return ucm.restoreCert()
	case util.DeleteCertOperateFlag:
		return ucm.deleteCert()
	default:
		return errors.New("unknown operation")
	}
}

func (ucm UnusedCertsMgr) getUnusedCerts() error {
	unusedCertPath := ucm.getUnusedCertPath()
	if fileutils.IsLexist(unusedCertPath) {
		fmt.Println(unusedCertPath)
	} else {
		fmt.Println("no unused certificate found")
	}
	return nil
}

func (ucm UnusedCertsMgr) restoreCert() error {
	unusedCertPath := ucm.getUnusedCertPath()
	currentCertPath := ucm.getCurrentCertPath()
	if !fileutils.IsLexist(unusedCertPath) {
		fmt.Printf("the unused cert [%s] doesn't exist\n", ucm.caName)
		return fmt.Errorf("the unused cert [%s] doesn't exist", ucm.caName)
	}
	caBytes, err := fileutils.LoadFile(unusedCertPath)
	if err != nil {
		hwlog.RunLog.Errorf("load unused cert [%s] failed, %v", unusedCertPath, err)
		return errors.New("load cert failed")
	}
	caChainMgr, err := x509.NewCaChainMgr(caBytes)
	if err != nil {
		hwlog.RunLog.Errorf("parse unused cert [%s] failed, %v", unusedCertPath, err)
		return errors.New("parse cert failed")
	}
	if err := caChainMgr.CheckCertsOverdue(0); err != nil {
		hwlog.RunLog.Errorf("unable to restore %s's ca cert because it is overdue", ucm.caName)
		return fmt.Errorf("unable to restore %s's ca cert because it is overdue", ucm.caName)
	}
	// north ca is readonly, set it writable
	if err := fileutils.SetPathPermission(currentCertPath, common.Mode600, false, true); err != nil {
		hwlog.RunLog.Errorf("unable to change permission of %s's ca cert, error: %v", ucm.caName, err)
		return fmt.Errorf("unable to change permission of %s's ca cert because it is overdue", ucm.caName)
	}
	if err := fileutils.CopyFile(unusedCertPath, currentCertPath); err != nil {
		hwlog.RunLog.Errorf("restore unused cert [%s] failed, %v", currentCertPath, err)
		return errors.New("restore cert failed")
	}
	if err := backuputils.BackUpFiles(currentCertPath); err != nil {
		hwlog.RunLog.Errorf("backup cert [%s] failed, %v", currentCertPath, err)
		return errors.New("backup cert failed")
	}
	if err := fileutils.DeleteFile(unusedCertPath); err != nil {
		hwlog.RunLog.Errorf("delete unused cert [%s] failed, %v", unusedCertPath, err)
		return errors.New("delete cert failed")
	}
	return nil
}

func (ucm UnusedCertsMgr) deleteCert() error {
	unusedCertPath := ucm.getUnusedCertPath()
	if !fileutils.IsLexist(unusedCertPath) {
		fmt.Printf("the unused cert [%s] doesn't exist\n", ucm.caName)
		return fmt.Errorf("the unused cert [%s] doesn't exist", ucm.caName)
	}
	question := fmt.Sprintf("Do you want to permanently delete file [%s]?", unusedCertPath)
	if err := ucm.interactiveConfirmation(question); err != nil {
		fmt.Printf("deletion was interrupted by user: %v\n", err)
		return fmt.Errorf("deletion was interrupted by user, %v", err)
	}
	if err := fileutils.DeleteFile(unusedCertPath); err != nil {
		hwlog.RunLog.Errorf("delete unused cert [%s] failed, %v", unusedCertPath, err)
		return err
	}
	return nil
}

func (ucm UnusedCertsMgr) getCurrentCertPath() string {
	return map[string]string{
		common.SoftwareCertName: ucm.pathMgr.ConfigPathMgr.GetSoftwareCertPath(),
		common.ImageCertName:    ucm.pathMgr.ConfigPathMgr.GetImageCertPath(),
		common.NorthernCertName: ucm.pathMgr.ConfigPathMgr.GetNorthernCertPath(),
	}[ucm.caName]
}

func (ucm UnusedCertsMgr) getUnusedCertPath() string {
	return ucm.getCurrentCertPath() + common.PreviousCertSuffix
}

func (ucm UnusedCertsMgr) checkParams() error {
	if ucm.getCurrentCertPath() == "" {
		return fmt.Errorf("unsupported ca [%s]", ucm.caName)
	}
	return nil
}

func (ucm UnusedCertsMgr) interactiveConfirmation(question string) error {
	fmt.Printf("%s\n(yes/no):", question)
	const bufferSize = 256
	buffer := make([]byte, bufferSize)
	nRead, err := os.Stdin.Read(buffer)
	if err != nil {
		hwlog.RunLog.Errorf("failed to read user input, error: %v", err)
		return errors.New("unexpected error while read input")
	}
	input := string(bytes.TrimSpace(buffer[:nRead]))

	switch input {
	case "yes":
		return nil
	case "no":
		return errors.New("cancelled")
	default:
		return errors.New("invalid option")
	}
}
