// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logcollect provides utils for log collection
package logcollect

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

const (
	metaBytes             = 1 << 20
	randCount             = 64
	defaultFileSize       = metaBytes
	defaultDirPermission  = 0600
	defaultFilePermission = 0600

	ten = 10
)

func setup() error {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		return err
	}
	if err := hwlog.InitOperateLogger(logConfig, context.Background()); err != nil {
		return err
	}
	return nil
}

// TestMain setups environment
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Printf("failed to init hwlog, %v\n", err)
		return
	}
	exitCode := m.Run()
	fmt.Printf("test complete, exitCode=%d\n", exitCode)
}

// TestPackFile functional test for log collection
func TestPackFile(t *testing.T) {
	convey.Convey("pack logs functional test", t, func() {
		testDir, err := filepath.Abs("testPackFile")
		convey.So(err, convey.ShouldBeNil)
		dst := filepath.Join(testDir, "pack.tar.gz")
		groups := []LogGroup{
			{
				RootDir: filepath.Join(testDir, "a"),
				BaseDir: "aa",
			},
			{
				RootDir: filepath.Join(testDir, "b"),
				BaseDir: "bb",
			},
		}
		err = fillData(filepath.Join(testDir, "a/a.log"))
		convey.So(err, convey.ShouldBeNil)
		err = fillData(filepath.Join(testDir, "a/b.log"))
		convey.So(err, convey.ShouldBeNil)
		err = fillData(filepath.Join(testDir, "b/a.log"))
		convey.So(err, convey.ShouldBeNil)

		collector := NewCollector(dst, groups, metaBytes*ten, []string{dst})
		packed, err := collector.Collect()
		convey.So(err, convey.ShouldBeNil)
		err = checkPack(packed, groups)
		convey.So(err, convey.ShouldBeNil)
	})
}

func fillData(filename string) error {
	if err := os.MkdirAll(filepath.Dir(filename), defaultDirPermission); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultFilePermission)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close file, %s\n", err.Error())
		}
	}()

	data := make([]byte, defaultFileSize)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < randCount; i++ {
		index := random.Intn(defaultFileSize)
		if index > len(data) || index < 0 {
			continue
		}
		data[index] = 1
	}

	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func checkPack(packFile string, groups []LogGroup) error {
	packedFiles := make(map[string]string)
	for _, group := range groups {
		childFiles, err := group.listFiles()
		if err != nil {
			return err
		}
		for _, childFile := range childFiles {
			tarEntryPath := filepath.Join(group.BaseDir, childFile)
			if _, ok := packedFiles[tarEntryPath]; ok {
				return errors.New("duplicate tar entry")
			}
			packedFiles[tarEntryPath] = filepath.Join(group.RootDir, childFile)
		}
	}
	resultFile, err := os.Open(packFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := resultFile.Close(); err != nil {
			fmt.Printf("failed to close tar.gz: %v\n", err)
		}
	}()
	gzipReader, err := gzip.NewReader(resultFile)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzipReader)
	for {
		tarHdr, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		originalFile, ok := packedFiles[tarHdr.Name]
		if !ok {
			return errors.New("unexpected tar entry")
		}
		if err := checkFileContent(originalFile, io.LimitReader(tarReader, tarHdr.Size)); err != nil {
			return err
		}
	}
}

func checkFileContent(original string, actual io.Reader) error {
	actualBytes, err := io.ReadAll(actual)
	if err != nil {
		return err
	}
	originalBytes, err := fileutils.LoadFile(original)
	if err != nil {
		return err
	}
	if bytes.Compare(actualBytes, originalBytes) != 0 {
		return errors.New("file content not equal")
	}
	return nil
}
