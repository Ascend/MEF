// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package x509 provides the capability of x509.
package x509

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

// ConstError error constant type
type ConstError string

// Error implement the error interface
func (e ConstError) Error() string { return string(e) }

const (
	// ErrCheckSUMEmpty error if checkSum is empty
	ErrCheckSUMEmpty = ConstError("the checkSum can't be empty")
	// ErrVerified verify check sum error
	ErrVerified = ConstError("the data verify failed")
	// ErrInstanceEmpty instance have no required data
	ErrInstanceEmpty = ConstError("the instance can't be empty")
)

// BackUpInstance important file backup and check tools
type BackUpInstance struct {
	data       []byte
	split      []byte
	checkSum   []byte
	mainPath   string
	backUpPath string
}

// NewBKPInstance return a new instance of BackUpInstance
func NewBKPInstance(data []byte, mainPath, backPath string) (*BackUpInstance, error) {
	if mainPath == "" || backPath == "" {
		return nil, errors.New("the instance path can't be empty")
	}
	var err error
	mainPath, err = fileutils.CheckOriginPath(mainPath)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	backPath, err = fileutils.CheckOriginPath(backPath)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}

	var checkSum []byte
	if len(data) != 0 {
		checkSum = utils.GetSha256Code(data)
	}
	return &BackUpInstance{
		data:       data,
		split:      []byte(utils.SplitFlag),
		checkSum:   checkSum,
		mainPath:   mainPath,
		backUpPath: backPath,
	}, nil
}

// WriteToDisk write data to disk, param needPadding control whether you need safety write
func (b *BackUpInstance) WriteToDisk(mode os.FileMode, needPadding bool) error {
	if err := commonValid(b); err != nil {
		return err
	}
	finalData := append(append(b.data, b.split...), b.checkSum...)
	if needPadding {
		if err := OverridePassWdFile(b.mainPath, finalData, mode); err != nil {
			hwlog.RunLog.Error("write file with padding to main path failed")
			return err
		}
		if err := OverridePassWdFile(b.backUpPath, finalData, mode); err != nil {
			hwlog.RunLog.Error("write file with padding to backup path failed")
			return err
		}
		hwlog.RunLog.Debug("write file with  padding successfully")
		return nil
	}
	if err := ioutil.WriteFile(b.mainPath, finalData, mode); err != nil {
		hwlog.RunLog.Error("write file to main path failed")
		return err
	}
	if err := ioutil.WriteFile(b.backUpPath, finalData, mode); err != nil {
		hwlog.RunLog.Error("write file to backup path failed")
		return err
	}
	hwlog.RunLog.Debug("write file with  successfully")
	return nil
}

func commonValid(b *BackUpInstance) error {
	if b == nil || len(b.data) == 0 {
		return errors.New("the instance can't be empty")
	}
	if b.mainPath == "" || b.backUpPath == "" {
		return errors.New("the instance path can't be empty")
	}
	if b.mainPath == b.backUpPath {
		return errors.New("the path can't be same")
	}
	return nil
}

// ReadFromDisk load file content from main file path or back up file path
func (b *BackUpInstance) ReadFromDisk(mode os.FileMode, needPadding bool) ([]byte, error) {
	if b == nil || b.mainPath == "" || b.backUpPath == "" {
		return nil, ErrInstanceEmpty
	}
	if fileutils.IsExist(b.mainPath) {
		return readFromFile(b, true, needPadding, mode)
	} else {
		if !fileutils.IsExist(b.backUpPath) {
			return nil, errors.New("both two file is not exist")
		}
		return readFromFile(b, false, needPadding, mode)
	}
}

func readFromFile(b *BackUpInstance, isMain, needPadding bool, mode os.FileMode) ([]byte, error) {
	path := b.backUpPath
	if isMain {
		path = b.mainPath
	}
	fullData, err := fileutils.LoadFile(path)
	if err != nil {
		if isMain {
			return readFromFile(b, false, needPadding, mode)
		}
		return nil, err
	}
	idx := bytes.LastIndex(fullData, b.split)
	if idx == -1 {
		// old version, need skip verify
		hwlog.RunLog.Warn("no checksum found, skip verify and back up")
		b.data = fullData
		b.split = nil
		return fullData, nil
	}
	b.data = fullData[0:idx]
	b.checkSum = fullData[idx+len(b.split):]
	if err = b.Verify(); err != nil {
		hwlog.RunLog.Warnf("checksum verify failed,is main file: %+v", isMain)
		if isMain {
			return readFromFile(b, false, needPadding, mode)
		}
		return nil, err
	}
	return b.data, b.WriteToDisk(mode, needPadding)
}

// Verify whether  the data match with checksum
func (b *BackUpInstance) Verify() error {
	if b == nil || len(b.data) == 0 {
		return ErrInstanceEmpty
	}
	if len(b.checkSum) == 0 {
		return ErrCheckSUMEmpty
	}
	if bytes.Equal(utils.GetSha256Code(b.data), b.checkSum) {
		return nil
	}
	return ErrVerified
}
