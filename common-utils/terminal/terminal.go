//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package terminal provide a safe reader for password
package terminal

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

const ccLen = 19

// dummy pointer  for use when we need a valid pointer to 0 bytes.
var _zero uintptr

// termios struct is one of the standard interface of POSIX, more info search linux termios
type termios struct {
	iflag  uint32       // input model flag
	oflag  uint32       // output model flag
	cflag  uint32       // control model flag
	lflag  uint32       // local model flag
	cc     [ccLen]uint8 // control character
	ispeed uint32       // input baud speed
	ospeed uint32       // output baud spead
}

// ReadPassword  safe read password, will not echo input on console
func ReadPassword(fd, maxReadLength int) ([]byte, error) {
	// get the original config of syscallIOctl
	tmios, err := ioctlGetTermios(fd, syscall.TCGETS)
	if err != nil {
		return nil, err
	}

	newConfig := *tmios
	newConfig.lflag &^= syscall.ECHO // close echo
	newConfig.lflag |= syscall.ISIG | syscall.ICANON
	newConfig.iflag |= syscall.ICRNL // change enter to line
	// set the console don't echo the input
	if err = ioctlSetTermios(fd, syscall.TCSETS, &newConfig); err != nil {
		return nil, err
	}
	// recover the original config of syscallIOctl
	defer func() {
		err = ioctlSetTermios(fd, syscall.TCSETS, tmios)
		if err != nil {
			fmt.Println("error recover ioctl config")
		}
	}()
	return readPasswordLine(localReader(fd), maxReadLength)
}

// ReadPasswordWithTimeout command readpassword function add time out time, (attention better not use in circulation)
func ReadPasswordWithTimeout(fd, maxReadLength int, timeout time.Duration) ([]byte, error) {
	tmios, err := ioctlGetTermios(fd, syscall.TCGETS)
	if err != nil {
		return nil, err
	}

	var pwdChannel = make(chan []byte, 1)
	go getPwd(pwdChannel, fd, maxReadLength)

	var pwd []byte
	var ok bool
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case pwd, ok = <-pwdChannel:
		if !ok {
			return nil, errors.New("get input content from channel failed")
		}
	case <-timer.C:
		if err = ioctlSetTermios(fd, syscall.TCSETS, tmios); err != nil {
			return nil, errors.New("error recover ioctl config, when time out")
		}
		return nil, errors.New("wait for the content to be entered time out")
	}

	return pwd, nil
}

func getPwd(pwdChannel chan<- []byte, fd, maxReadLength int) {
	if pwdChannel == nil {
		fmt.Println("channel is nil")
		return
	}

	pwd, err := ReadPassword(fd, maxReadLength)
	if err != nil {
		fmt.Printf("get content failed %v\n", err)
		return
	}

	pwdChannel <- pwd
}

// readPasswordLine reads from reader until it finds \n(linux) or \r (windows) or io.EOF.
// The slice returned does not include the \n or \r .
// liunx user \n as end of line
// Windows uses \r as end of line
func readPasswordLine(reader io.Reader, maxReadLength int) ([]byte, error) {
	var buf [1]byte
	var res []byte
	for {
		n, err := reader.Read(buf[:])
		if n <= 0 {
			if err == io.EOF && len(res) > 0 {
				return res, nil
			}
			if err != nil {
				return res, err
			}
		}
		switch buf[0] {
		case '\r':
			if runtime.GOOS == "windows" {
				return res, nil
			}
		case '\n':
			if runtime.GOOS != "windows" {
				return res, nil
			}
		default:
			// if the input extend the max length， discard it
			if len(res) >= maxReadLength {
				continue
			}
			res = append(res, buf[0])
		}
	}
}

func syscallIOctl(fd int, req uint, arg uintptr) error {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(req), uintptr(arg))
	return errnoConvert(e1)
}

func ioctlGetTermios(fd int, req uint) (*termios, error) {
	var value termios
	err := syscallIOctl(fd, req, uintptr(unsafe.Pointer(&value)))
	return &value, err
}

func ioctlSetTermios(fd int, req uint, value *termios) error {
	err := syscallIOctl(fd, req, uintptr(unsafe.Pointer(value)))
	runtime.KeepAlive(value)
	return err
}
