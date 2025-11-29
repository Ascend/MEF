// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package commands
package commands

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

const (
	serialNumberLen = 20
	sha256sumLen    = 32
)

type certRegisterInfo struct {
	name, path, description string
	needReducePriv          bool
}

type getCertInfoCmd struct {
	certName string
	certs    []certRegisterInfo
}

// NewGetCertInfoCmd edge control command get cert
func NewGetCertInfoCmd() common.Command {
	return &getCertInfoCmd{}
}

// Name command name
func (cmd *getCertInfoCmd) Name() string {
	return common.GetCertInfo
}

// Description command description
func (cmd *getCertInfoCmd) Description() string {
	return common.GetCertInfoDesc
}

// BindFlag command flag binding
func (cmd *getCertInfoCmd) BindFlag() bool {
	flag.StringVar(&cmd.certName, "certname", "",
		"the name of certificate to be obtained. Currently, only [center] is supported.")
	utils.MarkFlagRequired("certname")
	return true
}

// LockFlag command lock flag
func (cmd *getCertInfoCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *getCertInfoCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}
	var certRegistry *certRegisterInfo
	for _, c := range registerInfo(ctx.ConfigPathMgr) {
		if c.name == cmd.certName {
			certRegistry = &c
			break
		}
	}
	if certRegistry == nil {
		fmt.Println("the certificate name is not supported.")
		hwlog.RunLog.Error("the certificate name is not supported")
		return errors.New("the certificate name is not supported")
	}

	notExistedErr := errors.New("path does not exist")
	err := cmd.printCert(certRegistry)
	if err != nil && err.Error() == notExistedErr.Error() {
		fmt.Println("get cert info failed, cert does not exist. Execute command [netconfig] first.")
		hwlog.RunLog.Error("print cert failed, cert does not exist")
		return errors.New("print cert failed")
	}
	if err != nil {
		fmt.Println("get cert info failed")
		hwlog.RunLog.Errorf("print cert failed, error: %v", err)
		return errors.New("print cert failed")
	}
	return nil
}

func (cmd *getCertInfoCmd) printCert(certRegistry *certRegisterInfo) error {
	if certRegistry.needReducePriv {
		ugidMgr := util.NewEdgeUGidMgr()
		if err := ugidMgr.SetEUGidToEdge(); err != nil {
			hwlog.RunLog.Errorf("set euid/egid to mef-edge failed, error: %v", err)
			return errors.New("set euid/egid to mef-edge failed")
		}
		defer func() {
			if err := ugidMgr.ResetEUGid(); err != nil {
				hwlog.RunLog.Errorf("reset euid/egid failed, %v", err)
			}
		}()
	}

	mefCerts, err := x509.GetCerts(certRegistry.path)
	if err != nil {
		hwlog.RunLog.Errorf("get mef certs failed, error: %v", err)
		return err
	}
	if len(mefCerts) == 0 {
		hwlog.RunLog.Error("no cert found")
		return errors.New("no cert found")
	}

	fmt.Println(certRegistry.description + ":")
	for _, mefCert := range mefCerts {
		sha256sum := sha256.Sum256(mefCert.Raw)
		fmt.Println("    Issuer:", mefCert.Issuer)
		fmt.Println("    Subject:", mefCert.Subject)
		fmt.Println("    Serial Number:", utils.BinaryFormat(mefCert.SerialNumber.Bytes(), serialNumberLen))
		fmt.Println("    Validity")
		fmt.Println("        Not Before:", mefCert.NotBefore.In(time.Local).Format(constants.TimeFormat))
		fmt.Println("        Not After:", mefCert.NotAfter.In(time.Local).Format(constants.TimeFormat))
		fmt.Println("    FingerPrint Algorithm: sha256")
		fmt.Println("    FingerPrint:", utils.BinaryFormat(sha256sum[:], sha256sumLen))
	}
	return nil
}

func (cmd *getCertInfoCmd) PrintOpLogOk(user string, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

func (cmd *getCertInfoCmd) PrintOpLogFail(user string, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

func registerInfo(configPathMgr *pathmgr.ConfigPathMgr) []certRegisterInfo {
	return []certRegisterInfo{
		{
			name:           "center",
			path:           configPathMgr.GetHubSvrRootCertPath(),
			description:    "MEFCenter Southern Root Certificate",
			needReducePriv: true,
		},
	}
}
