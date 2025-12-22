// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package domainconfig for edge control command domain mapping config
package domainconfig

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"gorm.io/gorm"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

const (
	hostsFilePath       = "/etc/hosts"
	swapFilePath        = "/etc/hosts.mef.swap"
	hostsFileMode       = 0644
	hostsFileUmask      = 0133
	singleMappingLength = 2
	commentPrefix       = "#"
	defaultCfgNum       = 16
)

// DomainCfgFlow domain mapping config flow
type DomainCfgFlow struct {
	domain string
	ip     string
}

type importDomainCfgTask struct {
	domain string
	ip     string
}

// NewDomainCfgFlow create domain mapping config flow instance
func NewDomainCfgFlow(domain, ip string) *DomainCfgFlow {
	return &DomainCfgFlow{
		domain: domain,
		ip:     ip,
	}
}

// RunTasks run signature config task
func (dcf DomainCfgFlow) RunTasks() error {
	importTask := importDomainCfgTask{domain: dcf.domain, ip: dcf.ip}
	if err := importTask.runTask(); err != nil {
		return errors.New("import domain config failed")
	}

	return nil
}

func (cpt *importDomainCfgTask) runTask() error {
	var checkFunc = []func() error{
		cpt.checkParamDomain,
		cpt.checkParamIP,
		cpt.importDomainConfig,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err.Error())
			return err
		}
	}
	return nil
}

func (cpt *importDomainCfgTask) checkParamDomain() error {
	domainChecker := checker.GetDomainChecker("domain", true, true, true)
	if checkResult := domainChecker.Check(*cpt); !checkResult.Result {
		return errors.New(checkResult.Reason)
	}
	return nil
}

func (cpt *importDomainCfgTask) checkParamIP() error {
	ipChecker := checker.GetIpV4Checker("ip", true)
	if checkResult := ipChecker.Check(*cpt); !checkResult.Result {
		return errors.New(checkResult.Reason)
	}
	if utils.IsLocalIp(cpt.ip) {
		hwlog.RunLog.Error("IP can't be loopBack address")
		return errors.New("check localIP failed")
	}
	return nil
}

func (cpt *importDomainCfgTask) importDomainConfig() error {
	cfg, err := config.GetDomainCfg()
	if err == gorm.ErrRecordNotFound {
		return cpt.createDomainCfg()
	}
	if err != nil {
		return err
	}
	return cpt.updateDomainCfg(cfg)
}

func (cpt *importDomainCfgTask) updateDomainCfg(cfg *config.DomainConfigs) error {
	if len(cfg.Configs) >= defaultCfgNum {
		return errors.New("the number of domain config exceeds the maximum")
	}
	existFlag := false
	position := 0
	for i, cfg := range cfg.Configs {
		if cfg.Domain == cpt.domain && cfg.IP == cpt.ip {
			hwlog.RunLog.Warn("duplicate domain and ip, not need to import")
			return nil
		}
		if cfg.Domain == cpt.domain && cfg.IP != cpt.ip {
			existFlag = true
			position = i
			break
		}
	}

	if existFlag {
		if err := DeleteDomainCfgInFile(cpt.domain, cpt.ip); err != nil {
			return fmt.Errorf("delete duplicated domain failed, %v", err.Error())
		}
		if err := DeleteDomainCfgInFile(cfg.Configs[position].Domain, cfg.Configs[position].IP); err != nil {
			return fmt.Errorf("delete duplicated domain failed, %v", err.Error())
		}
		cfg.Configs = append(cfg.Configs[:position], cfg.Configs[position+1:]...)
	}
	domainConfig := config.DomainConfig{
		Domain: cpt.domain,
		IP:     cpt.ip,
	}
	cfg.Configs = append([]config.DomainConfig{domainConfig}, cfg.Configs...)
	if err := addDomainCfgToFile(domainConfig); err != nil {
		return errors.New("add domain to /etc/hosts failed")
	}
	return config.SetDomainCfg(cfg)
}

func (cpt *importDomainCfgTask) createDomainCfg() error {
	domainConfig := config.DomainConfig{
		Domain: cpt.domain,
		IP:     cpt.ip,
	}
	if err := config.SetDomainCfg(&config.DomainConfigs{Configs: []config.DomainConfig{domainConfig}}); err != nil {
		return fmt.Errorf("set domain config to db error, %s", err.Error())
	}

	// prepare host file and clear possibly duplicated domain and ip in /etc/hosts
	if fileutils.IsExist(hostsFilePath) {
		if err := DeleteDomainCfgInFile(cpt.domain, cpt.ip); err != nil {
			return fmt.Errorf("delete duplicated domain failed, %v", err.Error())
		}
	} else {
		if err := createHostsFile(); err != nil {
			return errors.New("/etc/hosts if not exist, try to create hosts failed")
		}
	}
	// append new domain/ip mapping config to /etc/hosts
	return addDomainCfgToFile(domainConfig)
}

func addDomainCfgToFile(domainCfg config.DomainConfig) error {
	if err := checkHostsFile(); err != nil {
		return fmt.Errorf("check /etc/hosts owner and permission failed, %s", err.Error())
	}
	fileData, err := fileutils.LoadFile(hostsFilePath)
	if err != nil {
		return fmt.Errorf("load /etc/hosts failed, %s", err.Error())
	}
	cfg := fmt.Sprintf(domainCfg.IP + " " + domainCfg.Domain + "\n")
	newData := append([]byte(cfg), fileData...)
	if err = overwriteHostsFile(newData); err != nil {
		return fmt.Errorf("create image registry domain/ip config failed, %s", err.Error())
	}
	return nil
}

func createHostsFile() error {
	file, err := os.OpenFile(hostsFilePath, os.O_RDWR|os.O_CREATE, hostsFileMode)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			hwlog.RunLog.Errorf("close file failed when creating file, %s", err.Error())
		}
	}()
	return nil
}

func checkHostsFile() error {
	absPath, err := fileutils.CheckOwnerAndPermission(hostsFilePath, hostsFileUmask, constants.RootUserGid)
	if err != nil {
		return fmt.Errorf("check /etc/hosts owner and permission failed, %s", err.Error())
	}
	if absPath != hostsFilePath {
		return errors.New("check /etc/hosts path failed, path contains symbolic links")
	}
	return nil
}

// DeleteDomainCfgInFile for delete domain/ip mapping config in file /etc/hosts
func DeleteDomainCfgInFile(domain, ip string) error {
	if ok := fileutils.IsExist(hostsFilePath); !ok {
		hwlog.RunLog.Info("file /etc/hosts is not exist, and no need to clear")
		return nil
	}

	if err := checkHostsFile(); err != nil {
		return err
	}
	fileData, err := fileutils.LoadFile(hostsFilePath)
	if err != nil {
		return fmt.Errorf("load /etc/hosts failed, %s", err.Error())
	}
	newData, err := getModifiedHostsFile(fileData, domain, ip)
	if err != nil {
		return fmt.Errorf("clear image registry domain/ip config failed: %s", err.Error())
	}
	if err = overwriteHostsFile(newData); err != nil {
		return fmt.Errorf("clear image registry domain/ip config failed: %s", err.Error())
	}
	return nil
}

func getModifiedHostsFile(data []byte, domain, ip string) ([]byte, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	var newData []byte
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read hosts file in line failed, %s", err.Error())
		}

		// exclude comments and non-target line in /etc/hosts
		if strings.HasPrefix(line, commentPrefix) || !(strings.Contains(line, domain) && strings.Contains(line, ip)) {
			newData = append(newData, []byte(line)...)
			continue
		}
		newLine := getNewLineBesidesTargetDomain(domain, line)
		newData = append(newData, newLine...)
	}
	return newData, nil
}

func getNewLineBesidesTargetDomain(domain, line string) []byte {
	var newLine []byte
	subStrings := strings.Fields(line)
	if len(subStrings) == singleMappingLength {
		return newLine
	}

	var lineBuilder strings.Builder
	for i, str := range subStrings {
		if str == domain {
			lineBuilder.WriteString("")
		} else {
			lineBuilder.WriteString(subStrings[i] + " ")
		}
	}
	lineBuilder.WriteByte('\n')
	newLine = []byte(lineBuilder.String())
	return newLine
}

func overwriteHostsFile(data []byte) error {
	var copyFileError error
	file, err := os.OpenFile(swapFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, hostsFileMode)
	if err != nil {
		return fmt.Errorf("open swap file failed, %s", err.Error())
	}
	defer func() {
		if err = file.Close(); err != nil {
			hwlog.RunLog.Errorf("close swap file error: %s", err.Error())
		}
		if copyFileError != nil {
			hwlog.RunLog.Errorf("swap file will not be clear up, " +
				"please restore hosts file from /etc/hosts.mef.swap manually")
			return
		}
		if err = fileutils.DeleteFile(swapFilePath); err != nil {
			hwlog.RunLog.Errorf("clear swap file error: %s", err.Error())
		}
	}()

	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("lock swap file failed, %v", err.Error())
	}
	defer func() {
		if err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
			hwlog.RunLog.Errorf("unlock file[%s] failed:%v", swapFilePath, err)
		}
	}()
	if _, err = file.Write(data); err != nil {
		return err
	}
	if copyFileError = fileutils.CopyFile(swapFilePath, hostsFilePath); copyFileError != nil {
		return fmt.Errorf("update /etc/hosts by swap file failed, %s", err.Error())
	}
	return nil
}
