// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base file utils used
package common

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/utils"
)

const (
	fileMode = 0600
	maxSize  = 100 * KB
)

// WriteData write data with path check
func WriteData(filePath string, fileData []byte) error {
	filePath, err := utils.CheckPath(filePath)
	if err != nil {
		return err
	}

	err = utils.MakeSureDir(filePath)
	if err != nil {
		return err
	}

	writer, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer func() {
		err := writer.Close()
		if err != nil {
			return
		}
	}()
	_, err = writer.Write(fileData)
	if err != nil {
		return err
	}
	return nil
}

func isKeyFile(path string) bool {
	sufList := []string{
		".key",
		".ks",
	}

	for _, suf := range sufList {
		if strings.HasSuffix(path, suf) {
			return true
		}
	}

	return false
}

func confusionFile(path string) error {
	// Override with zero
	overrideByte := make([]byte, maxSize, maxSize)
	if err := WriteData(path, overrideByte); err != nil {
		return fmt.Errorf("confusion file with 0 failed: %s", err.Error())
	}

	for i := range overrideByte {
		overrideByte[i] = 1
	}
	if err := WriteData(path, overrideByte); err != nil {
		return fmt.Errorf("confusion file with 1 failed: %s", err.Error())
	}

	if _, err := rand.Read(overrideByte); err != nil {
		return errors.New("get random words failed")
	}
	if err := WriteData(path, overrideByte); err != nil {
		return fmt.Errorf("confusion file with random num failed: %s", err.Error())
	}

	return nil
}

func recursiveConfusionFile(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	if !isKeyFile(path) {
		return nil
	}

	// if a file larger than maxSize, it should be considered as a malicious file so that we do not confuse it
	if info.Size() > maxSize {
		return nil
	}

	if err = confusionFile(path); err != nil {
		return err
	}

	return nil
}

// DeleteAllFile is used to delete all files into a path
func DeleteAllFile(filePath string) error {
	if !utils.IsLexist(filePath) {
		return nil
	}

	if err := filepath.Walk(filePath, recursiveConfusionFile); err != nil {
		return fmt.Errorf("confusion path %s failed: %s", filePath, err.Error())
	}

	return os.RemoveAll(filePath)
}

// DeleteFile is used to delete one file into a path
func DeleteFile(filePath string) error {
	if utils.IsLexist(filePath) {
		return os.Remove(filePath)
	}

	return nil
}

// MakeSurePath is used to make sure a path exists by creating it if not
func MakeSurePath(tgtPath string) error {
	if utils.IsExist(tgtPath) {
		return nil
	}

	if err := os.MkdirAll(tgtPath, Mode700); err != nil {
		return errors.New("create directory failed")
	}

	return nil
}

// CreateFile creates the named file with mode
func CreateFile(filePath string, mode os.FileMode) error {
	file, err := os.OpenFile(filePath, os.O_CREATE, mode)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			hwlog.RunLog.Error("close file failed")
			return
		}
	}()
	return nil
}

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

// ReadDir func is the func to return the file list in a dir
func ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

// RenameFile renames (moves) old path to new path.
func RenameFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
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

// IsSoftLink is the func to check if a path is soft link
func IsSoftLink(path string) error {
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("get real path failed: %s", err.Error())
	}

	if !(path == realPath) {
		return fmt.Errorf("path [%s] is a softlink", path)
	}

	return nil
}
