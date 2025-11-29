// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package envutils for run command
package envutils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

const (
	// DefCmdTimeoutSec represent the default timeout time to exec cmd
	DefCmdTimeoutSec = 180

	maxFileSizeInMb = 100
	illegalChars    = "\n!|\\; &$<>`"
)

// CommandOptions command options
type CommandOptions struct {
	RunAsUser       uint32
	RunAsGroup      uint32
	SwitchUser      bool
	PipeStdout      bool
	PipeStderr      bool
	Stdin           []byte
	WaitTimeSeconds int
}

// CommandResult command result
type CommandResult struct {
	Err      error
	Done     bool
	Exited   bool
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

type command struct {
	name string
	args []string
	opts CommandOptions
}

// allowLinkWhiteList some executable named files in the directories
// named by the PATH environment variable are symlinks in some operating system
var allowLinkWhiteList = []string{"/usr/bin/docker", "/usr/bin/awk", "/usr/sbin/iptables"}

// CheckCommandAllowedSugid check command which allows suid and sgid
// If the system command is invoked by a common user after the check function, security risks exist.
// Advised to invoke the function and the system command as root user.
func CheckCommandAllowedSugid(cmdName string) error {
	const oneMegabytes = 1024 * 1024
	cmdPath, err := exec.LookPath(cmdName)
	if err != nil {
		return err
	}

	cmdFile, err := os.OpenFile(cmdPath, os.O_RDONLY, fileutils.Mode400)
	if err != nil {
		return fmt.Errorf("open cmd %s file failed, error: %v", cmdName, err)
	}
	defer fileutils.CloseFile(cmdFile)

	ownerChecker := fileutils.NewFileOwnerChecker(true, false, fileutils.RootUid, fileutils.RootGid)
	modeChecker := fileutils.NewFileModeChecker(true, fileutils.DefaultWriteFileMode, false, true)
	sizeChecker := fileutils.NewFileSizeChecker(maxFileSizeInMb * oneMegabytes)
	ownerChecker.SetNext(modeChecker)
	ownerChecker.SetNext(sizeChecker)

	if !inAllowLinkWhiteList(cmdPath) {
		linkChecker := fileutils.NewFileLinkChecker(false)
		ownerChecker.SetNext(linkChecker)
	}

	if err = ownerChecker.Check(cmdFile, cmdPath); err != nil {
		return fmt.Errorf("check cmd %s file failed, error: %v", cmdName, err)
	}

	return nil
}

// RunCommand run command by current user, return the output of command
func RunCommand(name string, waitTime int, arg ...string) (string, error) {
	cmd := command{
		name: name,
		args: arg,
		opts: CommandOptions{WaitTimeSeconds: waitTime},
	}
	result := cmd.run()
	return string(result.Stdout), result.Err
}

// RunInteractCommand run interact command by current user, return the output of command
func RunInteractCommand(name string, interactArg string, waitTime int, arg ...string) (string, error) {
	cmd := command{
		name: name,
		args: arg,
		opts: CommandOptions{
			WaitTimeSeconds: waitTime,
			Stdin:           []byte(interactArg),
		},
	}
	result := cmd.run()
	return string(result.Stdout), result.Err
}

// RunCommandWithUser run command by specified user, return the output of command
func RunCommandWithUser(name string, waitTime int, uid uint32, gid uint32, arg ...string) (string, error) {
	cmd := command{
		name: name,
		args: arg,
		opts: CommandOptions{
			WaitTimeSeconds: waitTime,
			SwitchUser:      true,
			RunAsUser:       uid,
			RunAsGroup:      gid,
		},
	}
	result := cmd.run()
	return string(result.Stdout), result.Err
}

// RunCommandWithOsStdout run command by current user and print the output to terminal directly
func RunCommandWithOsStdout(name string, waitTime int, arg ...string) error {
	cmd := command{
		name: name,
		args: arg,
		opts: CommandOptions{
			WaitTimeSeconds: waitTime,
			PipeStdout:      true,
		},
	}
	return cmd.run().Err
}

// RunCommandWithOptions runs command with options
func RunCommandWithOptions(opts *CommandOptions, name string, arg ...string) CommandResult {
	if opts == nil {
		opts = &CommandOptions{}
	}
	if opts.WaitTimeSeconds == 0 {
		const defaultWaitTimeSeconds = 30
		opts.WaitTimeSeconds = defaultWaitTimeSeconds
	}
	cmd := command{
		name: name,
		args: arg,
		opts: *opts,
	}
	return cmd.run()
}

// RunResidentCmd starts to run a persistent command and returns the process of the command and errors
func RunResidentCmd(name string, arg ...string) (int, error) {
	cmd := command{
		name: name,
		args: arg,
	}
	return cmd.start()
}

func shellDefenceCheck(commands ...string) bool {
	for _, cmd := range commands {
		if strings.ContainsAny(cmd, illegalChars) {
			return true
		}
	}
	return false
}

func checkPersistentCmdArgs(realPath string, allowLink bool, arg ...string) error {
	if _, err := fileutils.RealFileCheck(realPath, true, allowLink, maxFileSizeInMb); err != nil {
		return err
	}
	if shellDefenceCheck(arg...) {
		return errors.New("exec command check failed: contain illegal chars")
	}

	return nil
}

func inAllowLinkWhiteList(name string) bool {
	for _, allowFile := range allowLinkWhiteList {
		if name == allowFile {
			return true
		}
	}
	return false
}

func (c *command) createOsCommand(stdout, stderr *bytes.Buffer, persistent bool) (*exec.Cmd, error) {
	// check command binary and arguments
	cmdPath, err := c.check(persistent)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cmdPath, c.args...)
	// setup user/group
	if c.opts.SwitchUser {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{Uid: c.opts.RunAsUser, Gid: c.opts.RunAsGroup}}
	}

	// setup stdin/stdout/stderr
	if c.opts.PipeStdout {
		cmd.Stdout = os.Stdout
	} else if stdout != nil {
		cmd.Stdout = stdout
	}
	if c.opts.PipeStderr {
		cmd.Stderr = os.Stderr
	} else if stderr != nil {
		cmd.Stderr = stderr
	}
	if len(c.opts.Stdin) > 0 {
		cmd.Stdin = bytes.NewReader(c.opts.Stdin)
	} else {
		cmd.Stdin = os.Stdin
	}

	return cmd, nil
}

func (c *command) run() CommandResult {
	var stdout, stderr bytes.Buffer
	cmd, err := c.createOsCommand(&stdout, &stderr, false)
	if err != nil {
		return CommandResult{Err: err}
	}

	done := make(chan error)
	go func() { done <- cmd.Run() }()

	timer := time.NewTimer(time.Duration(c.opts.WaitTimeSeconds) * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		if cmd.Process == nil {
			hwlog.RunLog.Warnf("execute command %s timeout but command process is not created yet", c.name)
			return CommandResult{Err: errors.New("exec command timeout")}
		}
		if err = cmd.Process.Kill(); err != nil {
			hwlog.RunLog.Warnf("exec command %s timeout and stop it failed. %v", c.name, err)
		}
		return CommandResult{Err: errors.New("exec command timeout")}
	case err = <-done:
		stderrOutput := bytes.Trim(stderr.Bytes(), "\n")
		stdoutOutput := bytes.Trim(stdout.Bytes(), "\n")
		result := CommandResult{
			Done:   true,
			Stdout: stdoutOutput,
			Stderr: stderrOutput,
		}
		if err == nil {
			result.Exited = true
			return result
		}

		result.Err = err
		if len(stderrOutput) > 0 {
			result.Err = errors.New(string(stderrOutput))
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
			result.Exited = exitError.Exited()
		}
		return result
	}
}

func (c *command) start() (int, error) {
	cmd, err := c.createOsCommand(nil, nil, true)
	if err != nil {
		return 0, err
	}

	err = cmd.Start()
	if err != nil {
		hwlog.RunLog.Errorf("exec cmd %s failed: %v", c.name, err)
		return 0, fmt.Errorf("exec cmd %s failed", c.name)
	}

	if cmd.Process == nil {
		return 0, fmt.Errorf("exec cmd %s faild: start process failed", c.name)
	}

	// to release any resource once the cmd ends
	go cmd.Wait()

	return cmd.Process.Pid, nil
}

func (c *command) check(persistent bool) (string, error) {
	realPath, err := exec.LookPath(c.name)
	if err != nil {
		return "", err
	}
	allowLink := inAllowLinkWhiteList(realPath)
	if err := checkPersistentCmdArgs(realPath, allowLink, c.args...); err != nil {
		return "", err
	}
	if !persistent && c.opts.WaitTimeSeconds <= 0 {
		return "", errors.New("waitTime invalid")
	}
	return realPath, nil
}
