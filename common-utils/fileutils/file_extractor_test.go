// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package fileutils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestCopyTarFileForLinkName(t *testing.T) {
	convey.Convey("test link name within tar gz", t, func() {
		defer os.RemoveAll("/tmp/TestExtraTarGzFile")

		testcases := []struct {
			name string
			pass bool
		}{
			{name: "/abs_link", pass: false},
			{name: "../relative_link", pass: false},
			{name: "normal_link", pass: true},
		}
		for _, testcase := range testcases {
			fmt.Printf("test %s\n", testcase.name)
			tarHeader := &tar.Header{
				Linkname: testcase.name,
				Mode:     0777,
				Typeflag: tar.TypeSymlink,
			}
			err := copyTarFile("/tmp/TestExtraTarGzFile", tarHeader, tar.NewReader(bytes.NewReader(nil)), true)
			if testcase.pass {
				convey.So(err, convey.ShouldBeNil)
			} else {
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "invalid link name")
			}
		}
	})
}

func TestCopyTarFileForLink(t *testing.T) {
	convey.Convey("test link name within tar gz", t, func() {
		defer os.RemoveAll("/tmp/TestExtraTarGzFile")

		err := os.MkdirAll("/tmp/TestExtraTarGzFile/dir", mode755)
		convey.So(err, convey.ShouldBeNil)
		err = os.Symlink("/tmp/TestExtraTarGzFile/dir", "/tmp/TestExtraTarGzFile/dir_link")
		convey.So(err, convey.ShouldBeNil)
		testcases := []struct {
			name     string
			typeFlag byte
			pass     bool
		}{
			{name: "dir_link/dir", typeFlag: tar.TypeDir, pass: false},
			{name: "dir_link/file", typeFlag: tar.TypeReg, pass: false},
			{name: "dir/file", typeFlag: tar.TypeReg, pass: true},
		}
		for _, testcase := range testcases {
			fmt.Printf("test %s\n", testcase.name)
			tarHeader := &tar.Header{
				Name:     testcase.name,
				Mode:     mode755,
				Typeflag: testcase.typeFlag,
			}
			err := copyTarFile("/tmp/TestExtraTarGzFile", tarHeader, tar.NewReader(bytes.NewReader(nil)), true)
			if testcase.pass {
				convey.So(err, convey.ShouldBeNil)
			} else {
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "can't support symlinks")
			}
		}
	})
}

func TestExtraTarGzFile(t *testing.T) {
	// 准备数据
	tarGzFile := CreateTarGzFile()
	defer func() {
		if err := os.Remove(tarGzFile.Name()); err != nil {
			panic(err)
		}
	}()
	extractPath := "/tmp/TestExtraTarGzFile"
	defer func() {
		if err := os.RemoveAll(extractPath); err != nil {
			panic(err)
		}
	}()
	patches1 := gomonkey.ApplyFunc(RealFileCheck, MockRealFileChecker)
	defer patches1.Reset()

	type args struct {
		tarGzFiles   string
		extractPaths string
		allowLink    bool
	}

	tests := []struct {
		name   string
		args   args
		expect error
	}{
		{
			name: "Case1 :normal",
			args: args{
				tarGzFiles:   tarGzFile.Name(),
				extractPaths: extractPath,
				allowLink:    false,
			},
			expect: nil,
		},
	}
	convey.Convey("TestExtraTarGzFile", t, func() {
		for _, tt := range tests {
			convey.Convey(tt.name, func() {
				convey.So(ExtraTarGzFile(tt.args.tarGzFiles, tt.args.extractPaths, tt.args.allowLink),
					convey.ShouldResemble, tt.expect)
			})
		}

	})
}

func CreateTarGzFile() *os.File {
	rand.Seed(time.Now().UnixNano())

	file, err := os.CreateTemp("", "tempTarGzFile")
	if err != nil {
		panic(err)
	}

	// 创建一个gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer func() {
		if err := gzipWriter.Close(); err != nil {
			panic(err)
		}
	}()

	// 创建一个tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if err := tarWriter.Close(); err != nil {
			panic(err)
		}
	}()

	// 写入文件到tar writer
	data := []byte("Hello, world!")
	header := &tar.Header{
		Name: "example.txt",
		Mode: 0644,
		Size: int64(len(data)),
	}
	err = tarWriter.WriteHeader(header)
	if err != nil {
		panic(err)
	}
	_, err = tarWriter.Write(data)
	if err != nil {
		panic(err)
	}

	return file
}

func MockRealFileChecker(file string, _, _ bool, _ int64) (string, error) {
	return file, nil
}
