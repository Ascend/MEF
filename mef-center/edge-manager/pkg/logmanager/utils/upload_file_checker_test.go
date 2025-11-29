// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindxedge/base/common"
)

func TestUploadFileSize(t *testing.T) {
	convey.Convey("test upload file size", t, func() {
		testcases := []struct {
			name     string
			tarFiles []tar.Header
			success  bool
		}{
			{
				name:     "oversize log",
				tarFiles: []tar.Header{{Name: "a.log", Size: maxFileSize + 1}},
			},
			{
				name:     "maxsize log",
				tarFiles: []tar.Header{{Name: "a.log", Size: maxFileSize}},
				success:  true,
			},
			{
				name:     "oversize gz",
				tarFiles: []tar.Header{{Name: "a.gz", Size: maxFileSize + 1}},
			},
			{
				name:     "maxsize gz",
				tarFiles: []tar.Header{{Name: "a.gz", Size: maxFileSize}},
				success:  true,
			},
		}
		for _, tc := range testcases {
			assertion := convey.ShouldNotBeNil
			if tc.success {
				assertion = convey.ShouldBeNil
			}
			createTarGz(tc.tarFiles, func(checker *UploadFileChecker) {
				fmt.Printf("test %s\n", tc.name)
				convey.So(checker.Check(), assertion)
			})
		}
	})
}

func TestUploadFileTooManyFiles(t *testing.T) {
	convey.Convey("test upload file count", t, func() {
		var maxFileTar []tar.Header
		for i := 0; i < maxFileCount; i++ {
			maxFileTar = append(maxFileTar, tar.Header{
				Name: fmt.Sprintf("%d.log", i),
			})
		}
		var tooManyFileTar []tar.Header
		for i := 0; i < maxFileCount+1; i++ {
			tooManyFileTar = append(tooManyFileTar, tar.Header{
				Name: fmt.Sprintf("%d.log", i),
			})
		}
		testcases := []struct {
			name     string
			tarFiles []tar.Header
			success  bool
		}{
			{
				name:     "max file number",
				tarFiles: maxFileTar,
				success:  true,
			},
			{
				name:     "too file number",
				tarFiles: tooManyFileTar,
			},
		}
		for _, tc := range testcases {
			assertion := convey.ShouldNotBeNil
			if tc.success {
				assertion = convey.ShouldBeNil
			}
			createTarGz(tc.tarFiles, func(checker *UploadFileChecker) {
				fmt.Printf("test %s\n", tc.name)
				convey.So(checker.Check(), assertion)
			})
		}
	})
}

func TestUploadFileName(t *testing.T) {
	convey.Convey("test upload file name", t, func() {
		testcases := []struct {
			name     string
			tarFiles []tar.Header
			success  bool
		}{
			{
				name:     "path traversal",
				tarFiles: []tar.Header{{Name: "../a.log"}},
			},
			{
				name:     "absolute path",
				tarFiles: []tar.Header{{Name: "/a.log"}},
			},
			{
				name:     "bad ext",
				tarFiles: []tar.Header{{Name: "a.exe"}},
			},
		}
		for _, tc := range testcases {
			assertion := convey.ShouldNotBeNil
			if tc.success {
				assertion = convey.ShouldBeNil
			}
			createTarGz(tc.tarFiles, func(checker *UploadFileChecker) {
				fmt.Printf("test %s\n", tc.name)
				convey.So(checker.Check(), assertion)
			})
		}
	})
}

func TestUploadFileMode(t *testing.T) {
	convey.Convey("test upload file mode", t, func() {
		const mode440 = 0440
		testcases := []struct {
			name     string
			tarFiles []tar.Header
			success  bool
		}{
			{
				name:     "normal mode",
				tarFiles: []tar.Header{{Name: "a.log", Mode: mode440}},
				success:  true,
			},
			{
				name:     "executable mode",
				tarFiles: []tar.Header{{Name: "a.log", Mode: common.Mode500}},
			},
			{
				name:     "symlink",
				tarFiles: []tar.Header{{Name: "a.log", Typeflag: tar.TypeSymlink}},
			},
			{
				name:     "regular file",
				tarFiles: []tar.Header{{Name: "a.log", Typeflag: tar.TypeReg}},
				success:  true,
			},
		}
		for _, tc := range testcases {
			assertion := convey.ShouldNotBeNil
			if tc.success {
				assertion = convey.ShouldBeNil
			}
			createTarGz(tc.tarFiles, func(checker *UploadFileChecker) {
				fmt.Printf("test %s\n", tc.name)
				convey.So(checker.Check(), assertion)
			})
		}
	})
}

func createTarGz(entries []tar.Header, fn func(checker *UploadFileChecker)) {
	file, err := os.OpenFile("temp.tgz", os.O_RDWR|os.O_CREATE|os.O_TRUNC, common.Mode600)
	convey.So(err, convey.ShouldBeNil)
	defer file.Close()

	checker := &UploadFileChecker{
		File: file,
	}
	patch := gomonkey.ApplyFuncReturn(checkSha256sum, nil)
	defer patch.Reset()

	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, entry := range entries {
		if entry.Mode == 0 {
			entry.Mode = common.Mode400
		}
		if strings.HasSuffix(entry.Name, ".gz") {
			var buffer bytes.Buffer
			gzWriter := gzip.NewWriter(&buffer)
			convey.So(fillWriter(gzWriter, entry.Size), convey.ShouldBeNil)
			convey.So(gzWriter.Close(), convey.ShouldBeNil)
			entry.Size = int64(buffer.Len())
			convey.So(tarWriter.WriteHeader(&entry), convey.ShouldBeNil)
			_, err := tarWriter.Write(buffer.Bytes())
			convey.So(err, convey.ShouldBeNil)
			continue
		}
		convey.So(tarWriter.WriteHeader(&entry), convey.ShouldBeNil)
		convey.So(fillWriter(tarWriter, entry.Size), convey.ShouldBeNil)
	}

	convey.So(tarWriter.Close(), convey.ShouldBeNil)
	convey.So(gzipWriter.Close(), convey.ShouldBeNil)
	fn(checker)
}

func fillWriter(writer io.Writer, size int64) error {
	const bufferSize = 1024
	buffer := make([]byte, bufferSize)
	remainBytes := size
	for remainBytes > 0 {
		var bufferLen int64 = bufferSize
		if bufferLen > remainBytes {
			bufferLen = remainBytes
		}
		nWrote, err := writer.Write(buffer[:bufferLen])
		if err != nil {
			return err
		}
		remainBytes -= int64(nWrote)
	}
	return nil
}
