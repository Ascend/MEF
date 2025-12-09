// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package fileutils provides the util func to deal with file
package fileutils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	fdPath               = "/proc/self/fd"
	maxPathLen           = 4096
	defaultPathWhiteList = "-_./"
	parentDirFlag        = ".."
	currentPathFlag      = "./"
	maxDepth             = 99
)

func byteToMb(size int64) float32 {
	const byteToMbMultiplier = 1024 * 1024
	return float32(size) / byteToMbMultiplier
}

// FileChecker is the interface that realized the func needs in a file check chain
type FileChecker interface {
	Check(file *os.File, path string) error
	checkNext(file *os.File, path string) error
	SetNext(FileChecker)
}

// FileBaseChecker is the base struct of a file checker, it provides all basic method that its son struct needs
type FileBaseChecker struct {
	next FileChecker
}

// SetNext is the basic method in a file check chain. It sets the next checker to the deep of a check chain
func (c *FileBaseChecker) SetNext(checker FileChecker) {
	if c.next == nil {
		c.next = checker
		return
	}

	c.next.SetNext(checker)
}

// Check in BaseCheck does nothing but to call the check function of the next element in the check chain
func (c *FileBaseChecker) Check(file *os.File, path string) error {
	return c.checkNext(file, path)
}

func (c *FileBaseChecker) checkNext(file *os.File, path string) error {
	if c.next == nil {
		return nil
	}

	return c.next.Check(file, path)
}

func (c *FileBaseChecker) getDirFile(file *os.File, path string) (*os.File, error) {
	// needs to close the file handle once the func being invoked
	realPath, err := GetRealPath(file, path)
	if err != nil {
		return nil, err
	}

	if realPath == "/" {
		return nil, nil
	}

	dirPath := filepath.Dir(realPath)
	dirFile, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	return dirFile, nil
}

// FileLinkChecker is the struct to check if a file contains softlink
type FileLinkChecker struct {
	FileBaseChecker
	allowRel bool
}

// Check is the main func on a validation chain.
// It invokes the validation of the current checker and invokes the Check method of the next element in the chain.
func (c *FileLinkChecker) Check(file *os.File, path string) error {
	if file == nil {
		return errors.New("pointer file is nil")
	}

	realPath, err := GetRealPath(file, path)
	if err != nil {
		return err
	}

	var absPath string
	if c.allowRel {
		absPath, err = filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("get abs path failed: %s", err.Error())
		}
	} else {
		absPath = filepath.Clean(path)
	}

	if realPath != absPath {
		return errors.New("can't support symlinks")
	}

	return c.checkNext(file, path)
}

// FileSizeChecker is the struct to check a file's size, the unit of size is byte
type FileSizeChecker struct {
	FileBaseChecker
	size int64
}

// Check is the main func on a validation chain.
// It invokes the validation of the current checker and invokes the Check method of the next element in the chain.
func (c *FileSizeChecker) Check(file *os.File, path string) error {
	if file == nil {
		return errors.New("pointer file is nil")
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file stat failed: %s", err.Error())
	}

	if fileInfo.Size() > c.size {
		return fmt.Errorf("file size exceeds %.2f MB", byteToMb(c.size))
	}

	return c.checkNext(file, path)
}

// FileModeChecker is the struct to check a file's mode
type FileModeChecker struct {
	FileBaseChecker
	recursive  bool
	umask      fs.FileMode
	checkSetId bool
	checkType  bool
	depth      int
}

// Check is the main func on a validation chain.
// It invokes the validation of the current checker and invokes the Check method of the next element in the chain.
func (c *FileModeChecker) Check(file *os.File, path string) error {
	if err := c.singleCheck(file, path); err != nil {
		return err
	}

	if !c.recursive {
		return c.checkNext(file, path)
	}

	c.depth = 0
	if err := c.recursiveCheck(file, path); err != nil {
		return err
	}

	return c.checkNext(file, path)
}

func (c *FileModeChecker) recursiveCheck(file *os.File, path string) error {
	if c.depth > maxDepth {
		return fmt.Errorf("over maxDepth %d", maxDepth)
	}

	dirFile, err := c.getDirFile(file, path)
	if err != nil {
		return err
	}
	defer CloseFile(dirFile)

	if dirFile == nil {
		return nil
	}

	if err = c.singleCheck(dirFile, path); err != nil {
		return err
	}

	c.depth += 1
	return c.recursiveCheck(dirFile, path)
}

func (c *FileModeChecker) singleCheck(file *os.File, path string) error {
	if file == nil {
		return errors.New("pointer file is nil")
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file stat failed: %s", err.Error())
	}

	filePath, err := GetRealPath(file, path)
	if err != nil {
		return fmt.Errorf("file mode %s unsupported and get cur path failed: %s", fileInfo.Mode(), err.Error())
	}

	fileMode := fileInfo.Mode()
	if fileMode&c.umask != 0 {
		return fmt.Errorf("path %s's file mode %s unsupported", filePath, fileInfo.Mode())
	}

	if c.checkType && !fileMode.IsRegular() && !fileMode.IsDir() {
		return fmt.Errorf("path %s's file is not a regular file", filePath)
	}
	if c.checkSetId && fileMode&os.ModeSetuid != 0 {
		return fmt.Errorf("path %s's file has set uid which is not allowed", filePath)
	}
	if c.checkSetId && fileMode&os.ModeSetgid != 0 {
		return fmt.Errorf("path %s's file has set gid which is not allowed", filePath)
	}

	return nil
}

// FileOwnerChecker is the struct to check a file's owner
type FileOwnerChecker struct {
	FileBaseChecker
	recursive        bool
	allowCurrentUser bool
	owner            uint32
	group            uint32
	depth            int
	meetRootUid      bool
	meetRootGid      bool
}

func (c *FileOwnerChecker) init() {
	c.depth = 0
	c.meetRootUid = false
	c.meetRootGid = false
}

// Check is the main func on a validation chain.
// It invokes the validation of the current checker and invokes the Check method of the next element in the chain.
func (c *FileOwnerChecker) Check(file *os.File, path string) error {
	c.init()

	if err := c.singleCheck(file, path); err != nil {
		return err
	}

	if !c.recursive {
		return c.checkNext(file, path)
	}

	if err := c.recursiveCheck(file, path); err != nil {
		return err
	}

	return c.checkNext(file, path)
}

func (c *FileOwnerChecker) recursiveCheck(file *os.File, path string) error {
	if c.depth > maxDepth {
		return fmt.Errorf("over maxDepth %d", maxDepth)
	}
	dirFile, err := c.getDirFile(file, path)
	if err != nil {
		return err
	}
	defer CloseFile(dirFile)

	if dirFile == nil {
		return nil
	}

	if err = c.singleCheck(dirFile, path); err != nil {
		return err
	}

	c.depth += 1
	return c.recursiveCheck(dirFile, path)
}

func (c *FileOwnerChecker) singleCheck(file *os.File, path string) error {
	if file == nil {
		return errors.New("pointer file is nil")
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file stat failed: %s", err.Error())
	}

	st := fileInfo.Sys()
	fileUid := st.(*syscall.Stat_t).Uid
	fileGid := st.(*syscall.Stat_t).Gid
	if err = c.checkOwner(fileUid); err != nil {
		filePath, pathErr := GetRealPath(file, path)
		if pathErr != nil {
			return fmt.Errorf("file owner [%d] unsupported and get cur path failed: %v", fileUid, pathErr)
		}
		return fmt.Errorf("the owner of file [%s] %v", filePath, err)
	}

	if err = c.checkGroup(fileGid); err != nil {
		filePath, pathErr := GetRealPath(file, path)
		if pathErr != nil {
			return fmt.Errorf("file group [%d] unsupported and get cur path failed: %v", fileGid, pathErr)
		}
		return fmt.Errorf("the group of file [%s] %v", filePath, err)
	}

	return nil
}

func (c *FileOwnerChecker) checkOwner(uid uint32) error {
	if c.isSetUser(uid) || (c.allowCurrentUser && c.isCurrentUser(uid)) {
		if uid != RootUid && c.meetRootUid {
			return fmt.Errorf("[uid=%d] is not supported since root file is inside the non-root dir", uid)
		}
		if uid == RootUid {
			c.meetRootUid = true
		}

		return nil
	}

	return fmt.Errorf("[uid=%d] is not supported", uid)
}

func (c *FileOwnerChecker) checkGroup(gid uint32) error {
	if c.isSetGroup(gid) || (c.allowCurrentUser && c.isCurrentGroup(gid)) {
		if gid != RootGid && c.meetRootGid {
			return fmt.Errorf("[gid=%d] is not supported since root file is inside the non-root dir", gid)
		}
		if gid == RootGid {
			c.meetRootGid = true
		}

		return nil
	}

	return fmt.Errorf("[gid=%d] is not supported", gid)
}

func (c *FileOwnerChecker) isCurrentUser(uid uint32) bool {
	return uid == uint32(os.Geteuid())
}

func (c *FileOwnerChecker) isCurrentGroup(gid uint32) bool {
	return gid == uint32(os.Getegid())
}

func (c *FileOwnerChecker) isSetUser(uid uint32) bool {
	return uid == c.owner
}

func (c *FileOwnerChecker) isSetGroup(gid uint32) bool {
	return gid == c.group
}

// FilePathChecker is the struct to check a file's path
type FilePathChecker struct {
	FileBaseChecker
}

// Check is the main func on a validation chain.
// It invokes the validation of the current checker and invokes the Check method of the next element in the chain.
// FilePathChecker only supports absolute paths,path with parent[..] current[./]
// or not start with '/' will not be supported
func (c *FilePathChecker) Check(file *os.File, path string) error {
	if len(path) > maxPathLen {
		return errors.New("path length exceeds the limitation")
	}

	if strings.Contains(path, parentDirFlag) || strings.Contains(path, currentPathFlag) ||
		(len(path) > 0 && path[0] != '/') {
		return errors.New("the input path is not a valid absolute path")
	}

	for _, char := range path {
		if !c.isValidCode(char) && !c.isInWhiteList(char) {
			return errors.New("path has unsupported character")
		}
	}

	return c.checkNext(file, path)
}

func (*FilePathChecker) isValidCode(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')
}

func (*FilePathChecker) isInWhiteList(c rune) bool {
	return strings.Contains(defaultPathWhiteList, string(c))
}

// IsDirChecker is the struct to check if a file is a dir
type IsDirChecker struct {
	FileBaseChecker
	expectedDir bool
}

// Check is the main func on a validation chain.
// It invokes the validation of the current checker and invokes the Check method of the next element in the chain.
func (c *IsDirChecker) Check(file *os.File, path string) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file %s's stat failed: %s", path, err.Error())
	}

	if c.expectedDir && !fileInfo.IsDir() {
		return fmt.Errorf("path %s is not dir", path)
	}
	if !c.expectedDir && fileInfo.IsDir() {
		return fmt.Errorf("path %s is a dir", path)
	}

	return c.checkNext(file, path)
}

// NewFileLinkChecker creates a FileLinkChecker
func NewFileLinkChecker(allowRel bool) *FileLinkChecker {
	return &FileLinkChecker{
		FileBaseChecker: FileBaseChecker{},
		allowRel:        allowRel,
	}
}

// NewFileSizeChecker creates a FileSizeChecker
func NewFileSizeChecker(size int64) *FileSizeChecker {
	return &FileSizeChecker{
		FileBaseChecker: FileBaseChecker{},
		size:            size,
	}
}

// NewFileModeChecker creates a FileModeChecker
func NewFileModeChecker(recursive bool, umask fs.FileMode, checkSetId, checkFileType bool) *FileModeChecker {
	return &FileModeChecker{
		FileBaseChecker: FileBaseChecker{},
		recursive:       recursive,
		umask:           umask,
		checkSetId:      checkSetId,
		checkType:       checkFileType,
	}
}

// NewFileOwnerChecker creates a FileOwnerChecker
func NewFileOwnerChecker(recursive, allowCurrentUser bool, owner, group uint32) *FileOwnerChecker {
	return &FileOwnerChecker{
		FileBaseChecker:  FileBaseChecker{},
		recursive:        recursive,
		allowCurrentUser: allowCurrentUser,
		owner:            owner,
		group:            group,
	}
}

// NewFilePathChecker creates FilePathChecker
func NewFilePathChecker() *FilePathChecker {
	return &FilePathChecker{FileBaseChecker: FileBaseChecker{}}
}

// NewIsDirChecker creates IsDirChecker
func NewIsDirChecker(expectedDir bool) *IsDirChecker {
	return &IsDirChecker{
		FileBaseChecker: FileBaseChecker{},
		expectedDir:     expectedDir,
	}
}
