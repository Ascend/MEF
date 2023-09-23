// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base file utils used
package common

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
)

// CopyDir is used to copy dir and all files into it
func CopyDir(srcPath string, dstPath string, includeDir bool) error {
	if !includeDir {
		srcPath = srcPath + "/."
	}

	if _, err := envutils.RunCommand(CommandCopy, envutils.DefCmdTimeoutSec, "-r", srcPath, dstPath); err != nil {
		return err
	}
	return nil
}

// CreateSoftLink creates a softLink to dstPath on srcPath.
func CreateSoftLink(dstPath, srcPath string) error {
	return os.Symlink(dstPath, srcPath)
}

// ExtraTarGzFile extract tar.gz file
func ExtraTarGzFile(tarGzFile, extractPath string, allowLink bool) error {
	cleanExtractPath := filepath.Clean(extractPath)
	srcFile, err := os.Open(tarGzFile)
	if err != nil {
		return errors.New("open tar.gz file failed")
	}
	defer func() {
		if err = srcFile.Close(); err != nil {
			hwlog.RunLog.Error("close tar.gz file failed")
			return
		}
	}()

	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer func() {
		if err = gzReader.Close(); err != nil {
			hwlog.RunLog.Error("close gzip reader failed")
			return
		}
	}()

	tarReader := tar.NewReader(gzReader)
	var totalWrote int64
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("get next entry in tar file failed, error: %v", err)
		}

		if err = copyTarFile(cleanExtractPath, header, tarReader, allowLink); err != nil {
			return err
		}
		totalWrote += header.Size
	}
	return nil
}

func copyTarFile(extractPath string, header *tar.Header, tarReader *tar.Reader, allowLink bool) error {
	extraFilePath := filepath.Join(extractPath, header.Name)
	switch header.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(extraFilePath, header.FileInfo().Mode()); err != nil {
			return fmt.Errorf("create path [%s] failed, error: %v", extraFilePath, err)
		}
	case tar.TypeReg:
		targetFile, err := os.OpenFile(extraFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("open dst file [%s] failed, error: %v", header.Name, err)
		}

		defer func() {
			if err = targetFile.Close(); err != nil {
				hwlog.RunLog.Errorf("close dst file [%s] failed, error: %v", header.Name, err)
				return
			}
		}()

		if _, err = io.Copy(targetFile, tarReader); err != nil {
			return fmt.Errorf("copy src file [%s] failed, error: %v", header.Name, err)
		}
	case tar.TypeSymlink:
		if !allowLink {
			return fmt.Errorf("do not support symlink[%s]", header.Name)
		}
		if err := os.Symlink(header.Linkname, extraFilePath); err != nil {
			return fmt.Errorf("create symlink[%s] failed", header.Name)
		}
	default:
		return fmt.Errorf("do not support the type of [%c]", header.Typeflag)
	}
	return nil
}
