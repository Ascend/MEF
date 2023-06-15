// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logcollect provides utils for log collection
package logcollect

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

const (
	maxDepth = 2
)

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
	destination string
	groups      []LogGroup
	packMaxSize int64
}

// NewCollector creates a new Collector instance
func NewCollector(destination string, groups []LogGroup, packMaxSize int64) Collector {
	return &collector{
		destination: destination,
		groups:      groups,
		packMaxSize: packMaxSize,
	}
}

// Collect implements log collection
func (l *collector) Collect() (string, error) {
	if _, err := utils.CheckPath(l.destination); err != nil {
		return "", err
	}
	diskFree, err := envutils.GetDiskFree(filepath.Dir(l.destination))
	if err != nil {
		return "", err
	}
	if diskFree < uint64(l.packMaxSize) {
		return "", errors.New("no enough disk space")
	}
	if utils.IsExist(l.destination) {
		if err := common.DeleteFile(l.destination); err != nil {
			return "", err
		}
	}
	dstFile, err := os.OpenFile(l.destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, common.Mode600)
	if err != nil {
		return "", err
	}
	var success bool
	defer func() {
		if err := dstFile.Close(); err != nil {
			hwlog.RunLog.Error("failed to close pack file")
		}
		if success {
			return
		}
		if err := common.DeleteFile(l.destination); err != nil {
			hwlog.RunLog.Error("failed to delete pack file")
		}
	}()

	if err := l.doCollect(dstFile); err != nil {
		return "", err
	}
	success = true
	return l.destination, nil
}

func (l *collector) doCollect(dstFile *os.File) error {
	gzipWriter := gzip.NewWriter(dstFile)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, group := range l.groups {
		relPaths, err := group.listFiles()
		if err != nil {
			return err
		}
		for _, relPath := range relPaths {
			tarEntryPath := filepath.Join(group.BaseDir, relPath)
			sourceFile := filepath.Join(group.RootDir, relPath)
			if err := collectFile(tarWriter, sourceFile, tarEntryPath, group.CheckFunc); err != nil {
				return err
			}
			fi, err := dstFile.Stat()
			if err != nil {
				return err
			}
			if fi.Size() > l.packMaxSize {
				return errors.New("pack file is too large")
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		return errors.New("failed to close tar")
	}
	if err := gzipWriter.Close(); err != nil {
		return errors.New("failed to close gzip")
	}
	return nil
}

func collectFile(tarWriter *tar.Writer, fileName, tarEntryPath string, checkFunc CheckFunc) error {
	fileStat, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	fileSize := fileStat.Size()
	if checkFunc != nil && checkFunc(fileName) != nil {
		return err
	}

	tarHeader := &tar.Header{
		Name: tarEntryPath,
		Mode: common.Mode400,
		Size: fileSize,
	}
	if err := tarWriter.WriteHeader(tarHeader); err != nil {
		return err
	}

	inputFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer func() {
		if err := inputFile.Close(); err != nil {
			hwlog.RunLog.Error("failed to close source file")
		}
	}()

	limitedReader := io.LimitReader(inputFile, fileSize)
	_, err = io.Copy(tarWriter, limitedReader)
	return err
}

func (g LogGroup) listFiles() ([]string, error) {
	absRoot, err := filepath.Abs(g.RootDir)
	if err != nil {
		return nil, err
	}
	if _, err := utils.RealDirChecker(absRoot, false, false); err != nil {
		return nil, err
	}
	fileNames, err := walkLogDir(absRoot, 0)
	if err != nil {
		return nil, err
	}
	relativeFileNames := make([]string, 0, len(fileNames))
	for _, fileName := range fileNames {
		relPath, err := filepath.Rel(absRoot, fileName)
		if err != nil {
			return nil, err
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
		return nil, err
	}
	entries, err := os.ReadDir(absRoot)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, err
	}
	if resolved != absRoot {
		return nil, errors.New("symlink is not allowed")
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
		} else {
			fileNames = append(fileNames, entryPath)
		}
	}
	return fileNames, nil
}
