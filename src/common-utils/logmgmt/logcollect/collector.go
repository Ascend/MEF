// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package logcollect provides utils for log collection
package logcollect

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

const (
	maxDepth = 2
	maxFiles = 1024
	logExt   = ".log"
	logGzExt = ".log.gz"
)

var validSuffixes = []string{logGzExt, logExt}

// CheckFunc custom file checker
type CheckFunc func(string) error

// LogGroup defines a group of log files
type LogGroup struct {
	RootDir   string
	BaseDir   string
	CheckFunc CheckFunc
}

// Collector provides interface for log collection
type Collector interface {
	// Collect collects logs, return packed file path if success
	Collect() (string, error)
}

type collector struct {
	destination          string
	groups               []LogGroup
	packMaxSize          int64
	collectPathWhiteList []string
}

// NewCollector creates a new Collector instance
func NewCollector(destination string, groups []LogGroup, packMaxSize int64, collectPathWhiteList []string) Collector {
	return &collector{
		destination:          destination,
		groups:               groups,
		packMaxSize:          packMaxSize,
		collectPathWhiteList: collectPathWhiteList,
	}
}

// Collect implements log collection
func (l *collector) Collect() (string, error) {
	if !l.inCollectPathWhiteList() {
		return "", fmt.Errorf("pack file path [%s] is not unsupported", l.destination)
	}
	if _, err := fileutils.CheckOriginPath(l.destination); err != nil {
		return "", fmt.Errorf("check file [%s] failed, %v", l.destination, err)
	}
	if err := fileutils.DeleteFile(l.destination); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("delete file [%s] failed, %v", l.destination, err)
	}
	diskFree, err := envutils.GetDiskFree(filepath.Dir(l.destination))
	if err != nil {
		return "", fmt.Errorf("get free disk space on [%s] failed, %v", filepath.Dir(l.destination), err)
	}
	if diskFree < uint64(l.packMaxSize) {
		return "", errors.New("no enough disk space")
	}
	dstFile, err := os.OpenFile(l.destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileutils.Mode600)
	if err != nil {
		return "", fmt.Errorf("open file [%s] failed, %v", l.destination, err)
	}
	var success bool
	defer func() {
		if err := dstFile.Close(); err != nil {
			hwlog.RunLog.Error("failed to close pack file")
		}
		if success {
			return
		}
		if err := fileutils.DeleteFile(l.destination); err != nil {
			hwlog.RunLog.Error("failed to delete pack file")
		}
	}()

	if err := l.doCollect(dstFile); err != nil {
		return "", err
	}
	success = true
	return l.destination, nil
}

func (l *collector) doCollect(dstFile *os.File) (firstErr error) {
	gzipWriter, err := gzip.NewWriterLevel(dstFile, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer func() {
		if err := gzipWriter.Close(); err != nil {
			if firstErr != nil {
				hwlog.RunLog.Error("failed to close gzip")
			} else {
				firstErr = errors.New("failed to close gzip")
			}
		}
	}()
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if err := tarWriter.Close(); err != nil {
			if firstErr != nil {
				hwlog.RunLog.Error("failed to close tar")
			} else {
				firstErr = errors.New("failed to close tar")
			}
		}
	}()

	for _, group := range l.groups {
		relPaths, err := group.listFiles()
		if err != nil {
			hwlog.RunLog.Errorf("list log files for dir [%s] failed, %v", group.RootDir, err)
			return errors.New("list log files for dir failed")
		}
		for _, relPath := range relPaths {
			tarEntryPath := filepath.Join(group.BaseDir, relPath)
			sourceFile := filepath.Join(group.RootDir, relPath)
			if err := collectFile(tarWriter, sourceFile, tarEntryPath, group.CheckFunc); err != nil {
				hwlog.RunLog.Errorf("collect log [%s] failed, %v", sourceFile, err)
				return errors.New("collect log failed")
			}
			fi, err := dstFile.Stat()
			if err != nil {
				return fmt.Errorf("stat file failed, %v", err)
			}
			if fi.Size() > l.packMaxSize {
				return errors.New("pack file is too large")
			}
		}
	}
	return nil
}

func (l *collector) inCollectPathWhiteList() bool {
	for _, allowPath := range l.collectPathWhiteList {
		if l.destination == allowPath {
			return true
		}
	}
	return false
}

func collectFile(tarWriter *tar.Writer, fileName, tarEntryPath string, checkFunc CheckFunc) error {
	fileStat, err := os.Stat(fileName)
	if err != nil {
		return fmt.Errorf("stat file failed, %v", err)
	}
	fileSize := fileStat.Size()
	if checkFunc != nil {
		if err := checkFunc(fileName); err != nil {
			return fmt.Errorf("check file failed, %v", err)
		}
	}

	tarHeader := &tar.Header{
		Name: tarEntryPath,
		Mode: fileutils.Mode400,
		Size: fileSize,
	}
	if err := tarWriter.WriteHeader(tarHeader); err != nil {
		return fmt.Errorf("write header failed, %v", err)
	}

	inputFile, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("open file [%s] failed, %v", fileName, err)
	}
	defer func() {
		if err := inputFile.Close(); err != nil {
			hwlog.RunLog.Error("failed to close source file")
		}
	}()

	limitedReader := io.LimitReader(inputFile, fileSize)
	if _, err := io.Copy(tarWriter, limitedReader); err != nil {
		return fmt.Errorf(" copy log content failed, %v", err)
	}
	return nil
}

func (g LogGroup) listFiles() ([]string, error) {
	absRoot, err := fileutils.RealDirCheck(g.RootDir, false, false)
	if err != nil {
		return nil, fmt.Errorf("check dir [%s] failed, %v", g.RootDir, err)
	}
	fileNames, err := walkLogDir(absRoot, 0)
	if err != nil {
		return nil, err
	}
	relativeFileNames := make([]string, 0, len(fileNames))
	for _, fileName := range fileNames {
		relPath, err := filepath.Rel(absRoot, fileName)
		if err != nil {
			return nil, fmt.Errorf("get relative path [%s,%s] failed, %v", absRoot, fileName, err)
		}
		relativeFileNames = append(relativeFileNames, relPath)
	}
	return relativeFileNames, nil
}

func walkLogDir(root string, depth int) ([]string, error) {
	if depth >= maxDepth {
		return nil, nil
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("get absolute path [%s] failed, %v", root, err)
	}
	entries, err := os.ReadDir(absRoot)
	if err != nil {
		return nil, fmt.Errorf("read dir [%s] failed, %v", absRoot, err)
	}
	if len(entries) > maxFiles {
		return nil, fmt.Errorf("too many files under dir [%s]", absRoot)
	}
	resolved, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, fmt.Errorf("eval symlink for file [%s] failed, %v", absRoot, err)
	}
	if resolved != absRoot {
		return nil, fmt.Errorf("symlink [%s] is not allowed", absRoot)
	}
	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		entryPath := filepath.Join(absRoot, entry.Name())
		if entry.IsDir() {
			childFileNames, err := walkLogDir(entryPath, depth+1)
			if err != nil {
				return nil, err
			}
			fileNames = append(fileNames, childFileNames...)
			continue
		}
		if !isSuffixValid(entryPath) {
			continue
		}
		fileNames = append(fileNames, entryPath)
	}
	return fileNames, nil
}

func isSuffixValid(fileName string) bool {
	var suffixIsValid bool
	for _, suffix := range validSuffixes {
		if strings.HasSuffix(fileName, suffix) {
			suffixIsValid = true
			break
		}
	}
	return suffixIsValid
}
