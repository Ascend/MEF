//  Copyright(c) 2023. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package fileutils provides the util func to deal with file
package fileutils

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"huawei.com/mindx/common/rand"
)

var (
	// ErrFileTooLarge indicates file was failed to process because the file is too large
	ErrFileTooLarge = errors.New("file is too large")
)

const (
	maxSize          = 1024 * 1024 * 1024
	maxConfusionSize = 20 * 1024
	size512K         = 512 * 1024
	maxCopyCount     = 1000
)

// IsDir check whether the path is a directory.
func IsDir(path string) bool {
	if path == "" {
		return false
	}

	if !IsExist(path) {
		return path[len(path)-1:] == "/"
	}
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile check whether the path is a file
func IsFile(path string) bool {
	if path == "" {
		return false
	}
	return !IsDir(path)
}

// IsEmptyDir [method] check dir whether empty or not
func IsEmptyDir(path string) (bool, error) {
	reader, files, err := ReadDir(path)
	if err != nil {
		return false, err
	}

	CloseFile(reader)

	if len(files) == 0 {
		return true, nil
	}

	return false, nil
}

// EvalSymlinks get absolute and eval symlinks path
func EvalSymlinks(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("get the absolute path of [%s] failed, error: %v", filePath, err)
	}
	resoledPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", fmt.Errorf("get symlinks of [%s] failed, error: %v", filePath, err)
	}
	return resoledPath, nil
}

// ReadLink returns the abs path that a softlink towards
func ReadLink(filePath string) (string, error) {
	tgtPath, err := os.Readlink(filePath)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(tgtPath)
	if err != nil {
		return "", fmt.Errorf("get the absolute path of [%s] failed, error: %v", filePath, err)
	}

	return absPath, nil
}

// CloseFile is used to close a file handle
func CloseFile(file *os.File) {
	if file == nil {
		return
	}
	if err := file.Close(); err != nil {
		return
	}
	return
}

// GetRealPath returns the real path from a file handle
func GetRealPath(file *os.File, path string) (string, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("get file stat failed: %s", err.Error())
	}

	// files of the virtual type in the proc directory cannot eval symlinks through fd path
	if strings.HasPrefix(path, "/proc") && fileInfo.Size() == 0 {
		return EvalSymlinks(path)
	}

	realPath, err := EvalSymlinks(GetFdPath(file))
	if err != nil {
		return "", fmt.Errorf("get real path failed: %s", err.Error())
	}

	return realPath, nil
}

// GetFdPath returns the fd path from a file handle
func GetFdPath(file *os.File) string {
	fd := file.Fd()
	return filepath.Join(fdPath, strconv.Itoa(int(fd)))
}

func parseChecker(checkers ...FileChecker) FileChecker {
	if len(checkers) == 0 {
		return NewFileLinkChecker(true)
	}

	var ret FileChecker
	for _, checker := range checkers {
		if ret == nil {
			ret = checker
			continue
		}

		ret.SetNext(checker)
	}

	return ret
}

func check(path string, flag int, mode os.FileMode, checkerParam ...FileChecker) (*os.File, error) {
	// needs to close the file handle once the func being invoked
	file, err := os.OpenFile(path, flag, mode)
	if err != nil {
		return nil, err
	}

	checker := parseChecker(checkerParam...)
	if err = checker.Check(file, path); err != nil {
		CloseFile(file)
		return nil, err
	}

	return file, nil
}

func checkFile(path string, flag int, mode os.FileMode, checkerParam ...FileChecker) (*os.File, string, error) {
	// needs to close the file handle once the func being invoked
	file, err := check(path, flag, mode, checkerParam...)
	if err != nil {
		return nil, "", err
	}

	filePath, err := GetRealPath(file, path)
	if err != nil {
		CloseFile(file)
		return nil, "", err
	}

	return file, filePath, nil
}

func checkDir(path string, flag int, mode os.FileMode, checkerParam ...FileChecker) (*os.File, string, error) {
	// needs to close the file handle once the func being invoked
	file, err := check(path, flag, mode, checkerParam...)
	if err != nil {
		return nil, "", err
	}

	return file, GetFdPath(file), nil
}

func getExistsDir(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	for !IsLexist(path) {
		path = filepath.Dir(path)
		if path == "." {
			return "", os.ErrNotExist
		}
	}

	return path, nil
}

// ReadLimitBytes reads limit bytes from a file, it could optionally add FileChecker. On default,
// a link Checker will be used
func ReadLimitBytes(path string, limitLength int, checkerParam ...FileChecker) ([]byte, error) {
	if limitLength < 0 || limitLength > maxSize {
		return nil, errors.New("the limit length is not valid")
	}

	file, _, err := checkFile(path, os.O_RDONLY, Mode400, checkerParam...)
	if err != nil {
		return nil, fmt.Errorf("file check failed: %s", err.Error())
	}
	defer CloseFile(file)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("get file stat failed: %s", err.Error())
	}

	size := fileInfo.Size()
	// the file in /proc dir does not save in disk, therefore the size of all file there are zero
	if size == 0 {
		size = size512K
	}
	if size > int64(limitLength) {
		size = int64(limitLength)
	}
	buf := make([]byte, size, size)
	var (
		offset int
		eof    bool
	)
	for int64(offset) < size {
		l, err := file.Read(buf[offset:])
		if l < 0 {
			l = 0
		}
		offset += l
		if err != nil {
			if errors.Is(err, io.EOF) {
				eof = true
				break
			}
			return nil, fmt.Errorf("read file failed: %s", err.Error())
		}
		if l == 0 {
			return nil, errors.New("no byte was read")
		}
		eof = fileInfo.Size() == int64(offset)
	}

	if !eof {
		return buf[:offset], fmt.Errorf(
			"%w: %d bytes was read from [%s] but the file has more content", ErrFileTooLarge, offset, path)
	}
	return buf[:offset], nil
}

// LoadFile reads at most 10M content from a file, it could optionally add FileChecker. On default,
// a link Checker will be used
func LoadFile(path string, checkerParam ...FileChecker) ([]byte, error) {
	if path == "" {
		return nil, nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.New("the filePath is invalid")
	}

	if !IsExist(absPath) {
		return nil, errors.New("path does not exist")
	}

	return ReadLimitBytes(absPath, Size10M, checkerParam...)
}

// MakeSureDir create directory. The last element of path should end with slash, or it will be omitted.
// it could optionally add FileChecker. On default, a link Checker will be used
func MakeSureDir(path string, checkerParam ...FileChecker) error {
	dir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return err
	}

	if IsLexist(path) {
		return nil
	}

	existPath, err := getExistsDir(dir)
	if err != nil {
		return fmt.Errorf("get path %s's existing path failed: %s", path, err.Error())
	}
	relativePath, err := filepath.Rel(existPath, dir)
	if err != nil {
		return fmt.Errorf("get relative path failed: %s", err.Error())
	}

	file, existPath, err := checkDir(existPath, os.O_RDONLY, Mode400, checkerParam...)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(file)

	if err = os.MkdirAll(filepath.Join(existPath, relativePath), Mode700); err != nil {
		return fmt.Errorf("make dir all failed: %s", err.Error())
	}

	return nil
}

// CreateDir creates all dirs among a path
// it could optionally add FileChecker. On default, a link Checker will be used
func CreateDir(tgtPath string, mode os.FileMode, checkerParam ...FileChecker) error {
	tgtPath, err := filepath.Abs(tgtPath)
	if err != nil {
		return err
	}

	if IsLexist(tgtPath) {
		fileInfo, err := os.Stat(tgtPath)
		if err != nil {
			return fmt.Errorf("file %s exists but get its stat failed: %s", tgtPath, err.Error())
		}

		if !fileInfo.IsDir() {
			return fmt.Errorf("file %s exists but it is not a dir", tgtPath)
		}

		return nil
	}

	existPath, err := getExistsDir(tgtPath)
	if err != nil {
		return fmt.Errorf("get path %s's existing path failed: %s", tgtPath, err.Error())
	}
	relativePath, err := filepath.Rel(existPath, tgtPath)

	file, existPath, err := checkDir(existPath, os.O_RDONLY, Mode400, checkerParam...)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(file)

	if err = os.MkdirAll(filepath.Join(existPath, relativePath), mode); err != nil {
		return fmt.Errorf("make dir all failed: %s", err.Error())
	}

	return nil
}

// WriteData is used to write data with path check
// it could optionally add FileChecker. On default, a link Checker will be used
func WriteData(filePath string, fileData []byte, checkerParam ...FileChecker) error {
	err := MakeSureDir(filePath)
	if err != nil {
		return err
	}

	if !IsLexist(filePath) {
		dirFile, dirPath, err := checkFile(filepath.Dir(filePath), os.O_RDONLY, Mode400, checkerParam...)
		if err != nil {
			return fmt.Errorf("check file failed: %s", err.Error())
		}
		defer CloseFile(dirFile)
		filePath = filepath.Join(dirPath, filepath.Base(filePath))
	}

	writer, _, err := checkFile(filePath, os.O_WRONLY|os.O_CREATE, Mode600, checkerParam...)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(writer)

	if err := syscall.Flock(int(writer.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("accquire file lock failed, %v", err)
	}
	defer func() {
		if err := syscall.Flock(int(writer.Fd()), syscall.LOCK_UN); err != nil {
			fmt.Printf("release file lock failed, %v\n", err)
		}
	}()

	if err := writer.Truncate(0); err != nil {
		return fmt.Errorf("truncate file failed, %v", err)
	}
	if _, err := writer.Write(fileData); err != nil {
		return err
	}

	// make sure all data was written to disk before we release file lock
	if err := writer.Sync(); err != nil {
		return fmt.Errorf("sync file failed, %v", err)
	}
	return nil
}

// DeleteFile delete a single file
// it could optionally add FileChecker. On default, a link Checker will be used
func DeleteFile(path string, checkerParam ...FileChecker) error {
	if !IsLexist(path) {
		return nil
	}

	dirPath := filepath.Dir(path)
	baseName := filepath.Base(path)

	file, dirPath, err := checkDir(dirPath, os.O_RDONLY, Mode400, checkerParam...)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(file)

	return os.Remove(filepath.Join(dirPath, baseName))
}

// ReadDir realized read the file list into a dir
// the file handle needs to be return out since the fd saved in the os.DirEntry will be closed
// once the handle being closed
// it could optionally add FileChecker. On default, a link Checker will be used
func ReadDir(path string, checkerParam ...FileChecker) (*os.File, []os.DirEntry, error) {
	// needs to close the file handle once the func being invoked
	checker := parseChecker(checkerParam...)
	dirChecker := NewIsDirChecker(true)
	dirChecker.SetNext(checker)

	file, realPath, err := checkDir(path, os.O_RDONLY, Mode400, dirChecker)
	if err != nil {
		return nil, nil, fmt.Errorf("check file failed: %s", err.Error())
	}
	dirs, err := os.ReadDir(realPath)

	return file, dirs, err
}

// CreateFile creates a file
// it could optionally add FileChecker. On default, a link Checker will be used
func CreateFile(filePath string, mode os.FileMode, checkerParam ...FileChecker) error {
	if IsLexist(filePath) {
		return nil
	}

	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)

	dirFile, dirFdPath, err := checkDir(dir, os.O_RDONLY, Mode400, checkerParam...)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(dirFile)

	file, err := os.OpenFile(filepath.Join(dirFdPath, baseName), os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("open file %s failed: %s", filepath.Join(dirFdPath, baseName), err.Error())
	}
	defer CloseFile(file)

	return nil
}

// RenameFile renames a file
// it could optionally add FileChecker. On default, a link Checker will be used
func RenameFile(oldPath, newPath string, checkerParam ...FileChecker) error {
	oldFile, oldPath, err := checkFile(oldPath, os.O_RDONLY, Mode400, checkerParam...)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(oldFile)

	if IsLexist(newPath) {
		newFile, newPath, err := checkFile(newPath, os.O_RDONLY, Mode400, checkerParam...)
		if err != nil {
			return fmt.Errorf("check file failed: %s", err.Error())
		}
		defer CloseFile(newFile)

		return os.Rename(oldPath, newPath)
	}

	newDirFile, newDirPath, err := checkDir(filepath.Dir(newPath), os.O_RDONLY, Mode400)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(newDirFile)

	return os.Rename(oldPath, filepath.Join(newDirPath, filepath.Base(newPath)))
}

// CopyFile copies a file to dst loc
// it could optionally add FileChecker. On default, a link Checker will be used
func CopyFile(src, dst string, checkerParam ...FileChecker) error {
	srcFile, _, err := checkFile(src, os.O_RDONLY, Mode400, checkerParam...)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(srcFile)

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("get file status failed: %s", err.Error())
	}

	if !IsLexist(dst) {
		dstDirFile, dstDirPath, err := checkFile(filepath.Dir(dst), os.O_RDONLY, Mode400, checkerParam...)
		if err != nil {
			return fmt.Errorf("check file failed: %s", err.Error())
		}
		defer CloseFile(dstDirFile)
		dst = filepath.Join(dstDirPath, filepath.Base(dst))
	}

	dstFile, _, err := checkFile(dst, os.O_WRONLY|os.O_CREATE, Mode600, checkerParam...)
	if err != nil {
		return fmt.Errorf("check file failed: %s", err.Error())
	}
	defer CloseFile(dstFile)

	if err := syscall.Flock(int(dstFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("accquire file lock failed, %v", err)
	}
	defer func() {
		if err := syscall.Flock(int(dstFile.Fd()), syscall.LOCK_UN); err != nil {
			fmt.Printf("release file lock failed, %v\n", err)
		}
	}()

	if err := dstFile.Truncate(0); err != nil {
		return fmt.Errorf("truncate file failed, %v", err)
	}

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file failed: %s", err.Error())
	}

	// make sure all data was written to disk before we release file lock
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("sync file failed, %v", err)
	}
	return dstFile.Chmod(srcInfo.Mode())
}

// CopyDir copies a dir and all contents into it to dst loc
// a soft link will be replaced by its entity file during the copy process
// it could optionally add FileChecker. On default, a link Checker will be used
func CopyDir(src, dst string, checkerParam ...FileChecker) error {
	return copyDir(src, dst, 0, checkerParam...)
}

func copyDir(src, dst string, count int, checkerParam ...FileChecker) error {
	var (
		err     error
		dirs    []os.DirEntry
		fds     []os.FileInfo
		dstInfo os.FileInfo
	)
	if count > maxCopyCount {
		return errors.New("the file inside the dir exceed the max limitation")
	}

	if subFolder(src, dst) {
		return errors.New("the destination directory is a subdirectory of the source directory")
	}

	if dstInfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = CreateDir(dst, dstInfo.Mode(), checkerParam...); err != nil {
		return err
	}

	file, dirs, err := ReadDir(src, checkerParam...)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	for _, dir := range dirs {
		fd, err := dir.Info()
		if err == nil {
			fds = append(fds, fd)
		}
	}

	for _, fd := range fds {
		srcFile := filepath.Join(src, fd.Name())
		dstFile := filepath.Join(dst, fd.Name())
		if fd.IsDir() {
			if err = copyDir(srcFile, dstFile, count+1, checkerParam...); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcFile, dstFile, checkerParam...); err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyDirWithSoftlink is used to copy a dir and all contents into it to dst location,
// a soft link will be recreated in the new dir and link to the origin target file
func CopyDirWithSoftlink(src, dst string, checkerParam ...FileChecker) error {
	if len(checkerParam) == 0 {
		checkerParam = []FileChecker{
			&FileBaseChecker{},
		}
	}
	if err := filepath.Walk(src, func(path string, _ fs.FileInfo, err error) error {
		relativePath, walkErr := filepath.Rel(src, path)
		if walkErr != nil {
			return fmt.Errorf("get path %s's relative path failed: %v", path, err)
		}
		dstPath := filepath.Join(dst, relativePath)
		walkErr = prepareOneLib(path, dstPath, checkerParam...)
		if walkErr != nil {
			return walkErr
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func prepareOneLib(srcFile, dstDir string, checkerParam ...FileChecker) error {
	srcStat, err := os.Stat(srcFile)
	if err != nil {
		return fmt.Errorf("get srcFile stat failed: %v", err)
	}
	if srcStat.IsDir() {
		if err = CreateDir(dstDir, srcStat.Mode(), checkerParam...); err != nil {
			return fmt.Errorf("create dir %s failed: %v", dstDir, err)
		}
		return nil
	}

	absSrcFile, err := filepath.Abs(srcFile)
	if err != nil {
		return fmt.Errorf("get src file abs path failed: %v", err)
	}

	realSrcPath, err := filepath.EvalSymlinks(absSrcFile)
	if err != nil {
		return fmt.Errorf("get real src path path failed: %v", err)
	}

	if absSrcFile == realSrcPath {
		if err = MakeSureDir(dstDir, checkerParam...); err != nil {
			return fmt.Errorf("make sure lib dir failed: %v", err)
		}

		if err = CopyFile(srcFile, dstDir, checkerParam...); err != nil {
			return fmt.Errorf("copy file [%s] failed, error: %v", srcFile, err)
		}

		return nil
	}

	if err = os.Symlink(filepath.Base(realSrcPath), dstDir); err != nil {
		return fmt.Errorf("create softlink failed: %s", err.Error())
	}

	return nil
}

func subFolder(src, dst string) bool {
	if src == dst {
		return true
	}
	srcReal, err := EvalSymlinks(src)
	if err != nil {
		return false
	}
	dstReal, err := EvalSymlinks(dst)
	if err != nil {
		return false
	}
	srcList := strings.Split(srcReal, string(os.PathSeparator))
	dstList := strings.Split(dstReal, string(os.PathSeparator))
	if len(srcList) > len(dstList) {
		return false
	}
	return reflect.DeepEqual(srcList, dstList[:len(srcList)])
}

// GetFileSha256 is used to get the sha256sum value of a file
func GetFileSha256(path string) (string, error) {
	const maxAllowFileSize = 1024 * 100

	modeChecker := NewFileModeChecker(false, DefaultWriteFileMode, true, true)
	ownerChecker := NewFileOwnerChecker(false, true, RootUid, RootGid)
	linkChecker := NewFileLinkChecker(true)
	pathChecker := NewFilePathChecker()
	sizeChecker := NewFileSizeChecker(maxAllowFileSize)

	modeChecker.SetNext(ownerChecker)
	modeChecker.SetNext(linkChecker)
	modeChecker.SetNext(pathChecker)
	modeChecker.SetNext(sizeChecker)

	fileData, err := LoadFile(path, modeChecker)
	if err != nil {
		return "", fmt.Errorf("load file failed: %s", err.Error())
	}

	hash := sha256.New()
	if _, err := hash.Write(fileData); err != nil {
		return "", fmt.Errorf("get file sha256sum failed: %s", err.Error())
	}

	// The returned sha256 value should be a hexadecimal number.
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetSha256Bytes return the sha256 hash bytes of a bytes data
func GetSha256Bytes(data []byte) ([]byte, error) {
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		return nil, fmt.Errorf("get sha256sum failed: %s", err.Error())
	}
	return hash.Sum(nil), nil
}

func recursiveConfusionFile(path string, info fs.FileInfo, err error, checkerParam ...FileChecker) error {
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
	if info.Size() > maxConfusionSize {
		return nil
	}

	if err = IsSoftLink(path); err != nil {
		return nil
	}

	if err = confusionFile(path, info.Size(), checkerParam...); err != nil {
		return err
	}

	return nil
}

func isKeyFile(path string) bool {
	sufList := []string{
		".key",
		".ks",
		".key.backup",
	}

	for _, suf := range sufList {
		if strings.HasSuffix(path, suf) {
			return true
		}
	}

	return false
}

func confusionFile(path string, size int64, checkerParam ...FileChecker) error {
	if size > maxConfusionSize {
		size = maxConfusionSize
	}
	if err := SetPathPermission(path, Mode600, false, false); err != nil {
		return fmt.Errorf("set path permission failed: %s", err.Error())
	}
	// Override with zero
	overrideByte := make([]byte, size, size)
	if err := WriteData(path, overrideByte, checkerParam...); err != nil {
		return fmt.Errorf("confusion file with 0 failed: %s", err.Error())
	}

	for i := range overrideByte {
		overrideByte[i] = 0xff
	}
	if err := WriteData(path, overrideByte, checkerParam...); err != nil {
		return fmt.Errorf("confusion file with 1 failed: %s", err.Error())
	}

	if _, err := rand.Read(overrideByte); err != nil {
		return errors.New("get random words failed")
	}
	if err := WriteData(path, overrideByte, checkerParam...); err != nil {
		return fmt.Errorf("confusion file with random num failed: %s", err.Error())
	}

	return nil
}

// DeleteAllFileWithConfusion is used to delete all files with confusion
func DeleteAllFileWithConfusion(filePath string, checkerParam ...FileChecker) error {
	if !IsLexist(filePath) {
		return nil
	}

	dirPath := filepath.Dir(filePath)
	if err := IsSoftLink(dirPath); err != nil {
		return fmt.Errorf("check path failed: %v", err)
	}

	const maxFileCount = 1000
	fileCount := 0
	if err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		fileCount++
		if fileCount > maxFileCount {
			return fmt.Errorf("file count exceeds the limitation")
		}
		return recursiveConfusionFile(path, info, err, checkerParam...)
	}); err != nil {
		return fmt.Errorf("confusion path %s failed: %s", filePath, err.Error())
	}

	return os.RemoveAll(filePath)
}
