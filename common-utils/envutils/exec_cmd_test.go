// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package envutils
package envutils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestCheckCommandAllowedSugid(t *testing.T) {
	convey.Convey("test check command allowed sugid", t, func() {
		err := CheckCommandAllowedSugid("cat")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestRunCommand(t *testing.T) {
	convey.Convey("test run command", t, func() {
		output, err := RunCommand("echo", 1, "123")
		convey.So(err, convey.ShouldBeNil)
		convey.So(output, convey.ShouldEqual, "123")
	})
}

func TestRunInteractCommand(t *testing.T) {
	convey.Convey("test run interact command", t, func() {
		output, err := RunInteractCommand("tee", "123", 1, "test_run_interact_command.txt")
		convey.So(err, convey.ShouldBeNil)
		convey.So(output, convey.ShouldEqual, "123")
	})
}

func TestRunCommandWithOsStdout(t *testing.T) {
	convey.Convey("test run command with os stdout", t, func() {
		data, err := runEchoCommandWithOsStdout("123")
		convey.So(err, convey.ShouldBeNil)
		convey.So(strings.TrimSpace(string(data)), convey.ShouldEqual, "123")
	})
}

func TestRunCommandWithOptions(t *testing.T) {
	convey.Convey("test run command wit options", t, func() {
		result := RunCommandWithOptions(nil, "echo", "123")
		convey.So(result.Err, convey.ShouldBeNil)
		convey.So(string(result.Stdout), convey.ShouldEqual, "123")
		convey.So(result.ExitCode, convey.ShouldEqual, 0)
		convey.So(result.Done, convey.ShouldBeTrue)
		convey.So(result.Exited, convey.ShouldBeTrue)
	})
}

func TestRunResidentCmd(t *testing.T) {
	convey.Convey("test run resident command", t, func() {
		pid, err := RunResidentCmd("ping", "127.0.0.1")
		convey.So(err, convey.ShouldBeNil)

		process, err := os.FindProcess(pid)
		convey.So(err, convey.ShouldBeNil)
		// the results of signal 0 checks whether process exists
		err = process.Signal(syscall.Signal(0))
		convey.So(err, convey.ShouldBeNil)

		err = process.Signal(syscall.SIGTERM)
		convey.So(err, convey.ShouldBeNil)

		waitResult := make(chan error)
		go func() {
			_, err := process.Wait()
			if err != nil {
				fmt.Printf("exec error %v\n", err)
			}
			waitResult <- err
		}()
		select {
		case <-time.After(time.Second):
			convey.So(errors.New("wait time"), convey.ShouldBeNil)
		case <-waitResult:
			convey.So(err, convey.ShouldBeNil)
		}
	})
}

func runEchoCommandWithOsStdout(data string) ([]byte, error) {
	const rwMode = 0600
	file, err := os.OpenFile("test_run_command_with_os_stdout.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, rwMode)
	if err != nil {
		return nil, fmt.Errorf("open file error, %v", err)
	}
	originalStdout := os.Stdout
	os.Stdout = file
	defer func() {
		os.Stdout = originalStdout
		if err := file.Close(); err != nil {
			fmt.Printf("close error, %v\n", err)
		}
	}()

	if err = RunCommandWithOsStdout("echo", 1, data); err != nil {
		return nil, fmt.Errorf("run command error, %v", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek error, %v", err)
	}
	return io.ReadAll(file)
}
