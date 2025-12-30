// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	oneDaySeconds           = 24 * 60 * 60
	defaultCapacity         = 20
	timeFormat              = "2006-01-02T15-04-05.000"
	kilobytes               = 1024
	defaultDirPermission    = 0750
	defaultFilePermission   = 0600
	defaultBackupPermission = 0400
	maxCapacity             = 400
	minSaveVolume           = 1
	maxSaveVolume           = 30
	maxSaveTime             = 700
	minSaveTime             = 7
	gZipExt                 = ".gz"
)

// Logs is an io.WriteCloser.
type Logs struct {
	file   *os.File
	mutex  sync.Mutex
	rmOnce sync.Once

	// FileName is the file where logs are written.
	FileName string `json:"filename" yaml:"filename"`

	// Capacity is the maximum number of bytes before the log file
	// is rotated, and the default value is 128 megabytes.
	Capacity int `json:"capacity" yaml:"capacity"`

	// SaveTime is the maximum number of days for retaining old log
	// files. It calculates the retention time based on the timestamp
	// of the old log file name and the current time.
	SaveTime int `json:"savetime" yaml:"savetime"`

	// SaveVolume is the maximum number of old log files that can be
	// retained. It saves all old files by default.
	SaveVolume int `json:"savevolume" yaml:"savevolume"`

	// UTC determines whether to use the local time of the computer
	// or the UTC time as the timestamp in the formatted backup file.
	LocalOrUTC bool `json:"localorutc" yaml:"localorutc"`

	backupDir  string
	isCompress bool

	disableRotationIfSwitchUser bool

	length int64
	rmCh   chan bool
}

// logFile is a struct that is used to return filename and
// timestamp.
type logFile struct {
	fileInfo  os.FileInfo
	timeStamp time.Time
}

var (
	// mByte is used to convert capacity into bytes.
	mByte = kilobytes * kilobytes
)

// Write implements io.Writer. If a write would not cause the size of
// the log file to exceed Capacity, the log file is written normally.
// If a write would cause the size of the log file to exceed Capacity,
// but the write length is less than Capacity, the log file is closed,
// renamed to include a timestamp of the current time, and a new log
// is created using the original log file name. If the length of a write
// is greater than the Capacity, an error is returned.
func (l *Logs) Write(d []byte) (int, error) {
	if l == nil {
		return 0, fmt.Errorf("logs pointer does not exist")
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	writeLenth := int64(len(d))
	if writeLenth > l.maxLenth() {
		return 0, fmt.Errorf("the write lenth %d is greater than the maximum file size %d",
			writeLenth, l.maxLenth(),
		)
	}

	if l.file == nil {
		if err := l.openOrCreateFile(writeLenth); err != nil {
			return 0, err
		}
	}
	fileInfo, err := l.file.Stat()
	if err != nil {
		return 0, err
	}
	l.length = fileInfo.Size()
	if writeLenth+l.length > l.maxLenth() {
		if err := l.roll(); err != nil {
			return 0, err
		}
	}

	n, err := l.file.Write(d)
	if err != nil {
		return 0, err
	}
	l.length += int64(n)
	return n, err
}

// Roll causes Logs to close the existing log file and create a new log
// file immediately. The purpose of this function is to provide rotation
// outside the normal rotation rule, e.g. in response to SIGHUP. After
// rotation, the deletion of the old log files is initiated.
func (l *Logs) Roll() error {
	if l == nil {
		return fmt.Errorf("logs pointer does not exist")
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.roll()
}

// Close implements io.Closer. It closses the current log file.
func (l *Logs) Close() error {
	if l == nil {
		return fmt.Errorf("logs pointer does not exist")
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.close()
}

// Flush persist the contents of the current memory.
func (l *Logs) Flush() error {
	if l == nil {
		return fmt.Errorf("logs pointer does not exist")
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.file == nil {
		return nil
	}
	return l.file.Sync()
}

// maxLenth return the number of bytes of the maximum log size
// before rotating.
func (l *Logs) maxLenth() int64 {
	if l.Capacity > 0 && l.Capacity < maxCapacity {
		return int64(l.Capacity) * int64(mByte)
	}
	return int64(defaultCapacity * mByte)
}

// fileName return the name of the log file.
func (l *Logs) fileName() string {
	if l.FileName != "" {
		return l.FileName
	}
	logName := filepath.Base(os.Args[0]) + "-mindx-dl.log"
	return filepath.Join(os.TempDir(), logName)
}

// openOrCreateFile opens the log file if it exists and the
// current write would not exceed the Capacity. It will create
// a new file if there is no such file or the write would exceed
// the Capacity.
func (l *Logs) openOrCreateFile(writeLen int64) error {
	if l.disableRotationIfSwitchUser && checkSwitchUser() {
		return errors.New("rotation was disabled")
	}
	l.remove()

	name := l.fileName()
	message, err := os.Stat(name)
	if os.IsNotExist(err) {
		return l.create()
	}

	if err != nil {
		return fmt.Errorf("failed to get log file message: %v", err)
	}

	if writeLen+message.Size() >= l.maxLenth() {
		return l.roll()
	}

	f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, defaultFilePermission)
	if err != nil {
		return l.create()
	}
	l.file = f
	l.length = message.Size()
	return nil
}

// create creates a new log file for writing, and backs up the
// old log file. The file is closed when this method is invoked
// by default.
func (l *Logs) create() error {
	if err := os.MkdirAll(l.getDir(), defaultDirPermission); err != nil {
		return fmt.Errorf("unable to create directory for new log file: %v", err)
	}

	fileName, fileMode := l.fileName(), os.FileMode(defaultFilePermission)
	if message, err := os.Stat(fileName); err == nil {
		fileMode = message.Mode()
		backupName := l.backup()
		if err := l.backupFile(fileName, backupName); err != nil {
			return fmt.Errorf("failed to backup the log file: %v", err)
		}
		if err := os.Chmod(backupName, defaultBackupPermission); err != nil {
			return fmt.Errorf("failed to change backup log file permission: %v", err)
		}
	}
	newFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileMode)
	if err != nil {
		return fmt.Errorf("unable to open new log file: %v", err)
	}
	l.length, l.file = 0, newFile
	return nil
}

// backup generates a backup file name based on the original file
// name and inserts a timestamp between the file name and extension.
// The timestamp uses the UTC time by default.
func (l *Logs) backup() string {
	prefix, extension := l.getPreAndExt()
	return filepath.Join(l.getBackupDir(), fmt.Sprintf("%s%s%s", prefix, l.getTimestamp(), extension))
}

// getDir returns the directory for the current filename.
func (l *Logs) getDir() string {
	return filepath.Dir(l.fileName())
}

// getBackupDir returns the directory for the backup files.
func (l *Logs) getBackupDir() string {
	if l.backupDir == "" {
		return l.getDir()
	}
	return l.backupDir
}

// getPreAndExt returns the prefix name and extension name
// from Logs's filename.
func (l *Logs) getPreAndExt() (string, string) {
	name := filepath.Base(l.fileName())
	extension := filepath.Ext(name)
	prefix := name[:len(name)-len(extension)] + "-"
	if l.isCompress {
		extension += gZipExt
	}
	return prefix, extension
}

// getTimestamp returns the timestamp of current time, and
// uses UTC time by default.
func (l *Logs) getTimestamp() string {
	t := time.Now()
	if !l.LocalOrUTC {
		t = t.UTC()
	}
	return t.Format(timeFormat)
}

// getDiskFree is used to get the free disk space of a path
func (l *Logs) getDiskFree(path string) (uint64, error) {
	fileStat := syscall.Statfs_t{}
	if err := syscall.Statfs(path, &fileStat); err != nil {
		return 0, err
	}
	diskFree := fileStat.Bavail * uint64(fileStat.Bsize)
	if fileStat.Bavail != 0 && diskFree/fileStat.Bavail != uint64(fileStat.Bsize) {
		return 0, errors.New("unsigned number will be wrap")
	}
	return diskFree, nil
}

// checkDiskSpace is used to check whether the disk space on a path is enough
func (l *Logs) checkDiskSpace(path string, limit uint64) error {
	availSpace, err := l.getDiskFree(path)
	if err != nil {
		return fmt.Errorf("get path [%s]'s disk available space failed: %v", path, err)
	}

	if availSpace < limit {
		return errors.New("no enough space")
	}

	return nil
}

// makeSureBackupSpaceAvailable is used to make sure disk space is available for backup
func (l *Logs) makeSureBackupSpaceAvailable() error {
	fileName := l.fileName()
	fi, err := os.Stat(fileName)
	if err != nil {
		return fmt.Errorf("get backup file stat failed: %v", err)
	}
	backupFileSize := uint64(fi.Size())
	oldFiles, err := l.oldFilesList()
	if err != nil {
		return err
	}
	for i := len(oldFiles) - 1; i >= 0; i-- {
		if l.checkDiskSpace(l.getBackupDir(), backupFileSize) == nil {
			return nil
		}
		rmError := os.Remove(filepath.Join(l.getBackupDir(), oldFiles[i].fileInfo.Name()))
		if rmError != nil {
			fmt.Println("delete backup file failed:", rmError)
			continue
		}
	}
	return l.checkDiskSpace(l.getBackupDir(), backupFileSize)
}

// roll rotates the log file, close the existing log file and
// create a new one immediately. After rotating, this method
// deletes the old log files according to the configuration.
func (l *Logs) roll() error {
	if l.disableRotationIfSwitchUser && checkSwitchUser() {
		return errors.New("rotation was disabled")
	}
	if err := l.close(); err != nil {
		return err
	}
	if err := l.makeSureBackupSpaceAvailable(); err != nil {
		return err
	}
	if err := l.create(); err != nil {
		return err
	}
	l.remove()
	return nil
}

// backupFile backup the log file. If BackupDir is not
// configured and IsCompress is false, this method will rename
// the source to the destination. Otherwise, this method will
// copy or compress the source to the destination. After
// writing, the source file will be removed.
func (l *Logs) backupFile(src, dst string) error {
	if l.backupDir == "" {
		if err := os.Rename(src, dst); err != nil {
			fmt.Println("backup: unable to rename source file")
			return err
		}
		return nil
	}

	if err := l.copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

// copyFile copys file from src to dst.
func (l *Logs) copyFile(src, dst string) (firstErr error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	var dstFileCreated bool
	defer func() {
		if err := srcFile.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Println("backup: unable to close source file")
		}
		if firstErr != nil && dstFileCreated {
			if err := os.Remove(dst); err != nil {
				fmt.Println("backup: unable to clean destination file")
			}
		}
	}()
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(defaultFilePermission))
	if err != nil {
		firstErr = err
		return firstErr
	}
	dstFileCreated = true
	defer func() {
		if err := dstFile.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Println("backup: unable to close destination file")
		}
	}()

	if _, firstErr = l.copyStream(dstFile, srcFile); firstErr != nil {
		return firstErr
	}
	return dstFile.Sync()
}

// copyStream copies data form reader stream to writer stream
func (l *Logs) copyStream(writer io.Writer, reader io.Reader) (int64, error) {
	if l.isCompress {
		return l.compress(writer, reader)
	} else {
		return io.Copy(writer, reader)
	}
}

// compress reads data from source stream and compress data as gzip.
// This method will return size of uncompressed written data
func (l *Logs) compress(dst io.Writer, src io.Reader) (int64, error) {
	gzipWriter := gzip.NewWriter(dst)
	nWrites, firstErr := io.Copy(gzipWriter, src)
	if err := gzipWriter.Close(); err != nil {
		if firstErr == nil {
			firstErr = err
		} else {
			fmt.Println("backup: unable to close gzip stream")
		}
	}
	return nWrites, firstErr
}

// close closes the file if it is open.
func (l *Logs) close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Sync()
	if err != nil {
		return err
	}
	err = l.file.Close()
	l.file = nil
	return err
}

// remove delete outdated log files, starting the remove
// goroutine if necessary.
func (l *Logs) remove() {
	l.rmOnce.Do(func() {
		l.rmCh = make(chan bool, 1)
		go l.removeRun()
	})
	select {
	case l.rmCh <- true:
	default:
	}
}

// removeRun manages the deletion of the old log files after
// rotating, which runs in a goroutine.
func (l *Logs) removeRun() {
	for range l.rmCh {
		if err := l.removeRunOnce(); err != nil {
			fmt.Println("failed to remove runonce: ", err)
		}
	}
}

// removeRunOnce performs removal of outdated log files.
// Old log files are removed if the number of old files
// exceed the Capacity or the retention time of old files
// is greater than SaveTime.
func (l *Logs) removeRunOnce() error {
	if l.SaveVolume == 0 && l.SaveTime == 0 {
		return nil
	}

	if err := checkParam(l.SaveVolume, l.SaveTime); err != nil {
		return err
	}

	oldFiles, err := l.oldFilesList()
	if err != nil {
		return err
	}

	var removeFiles []logFile
	if l.SaveTime > 0 {
		delTime := time.Now().Unix() - int64(l.SaveTime)*oneDaySeconds
		var remainingFiles []logFile
		for _, f := range oldFiles {
			if f.timeStamp.Unix() <= delTime {
				removeFiles = append(removeFiles, f)
				continue
			}
			remainingFiles = append(remainingFiles, f)
		}
		oldFiles = remainingFiles
	}

	if l.SaveVolume > 0 && l.SaveVolume < len(oldFiles) {
		saved := make(map[string]struct{}, len(oldFiles))
		var remainingFiles []logFile
		for _, f := range oldFiles {
			saved[f.fileInfo.Name()] = struct{}{}
			if l.SaveVolume >= len(saved) {
				remainingFiles = append(remainingFiles, f)
				continue
			}
			removeFiles = append(removeFiles, f)
		}
		oldFiles = remainingFiles
	}

	for _, f := range removeFiles {
		rmError := os.Remove(filepath.Join(l.getBackupDir(), f.fileInfo.Name()))
		if rmError != nil {
			err = rmError
		}
	}
	return err
}

// oldFilesList returns the list of backup log files sorted
// by ModTime. These backup log files are stored in the same
// directory as the current log file.
func (l *Logs) oldFilesList() ([]logFile, error) {
	logFiles, err := ioutil.ReadDir(l.getBackupDir())
	if err != nil {
		return nil, fmt.Errorf("unable to open the log file directory: %v", err)
	}

	prefix, extension := l.getPreAndExt()

	var oldFiles []logFile

	for _, file := range logFiles {
		if file.IsDir() {
			continue
		}
		if timeStamp, err := l.extractTime(file.Name(), prefix, extension); err == nil {
			oldFiles = append(oldFiles, logFile{fileInfo: file, timeStamp: timeStamp})
			continue
		}
	}
	sort.Slice(oldFiles, func(i, j int) bool {
		if i < 0 || i > len(oldFiles) || j < 0 || j > len(oldFiles) {
			return false
		}
		return oldFiles[i].timeStamp.After(oldFiles[j].timeStamp)
	})

	return oldFiles, nil
}

// extractTime extracts the formatted time from file name by
// stripping the prefix and extension of the file name. This
// prevents fileName from being confused with time.parse.
func (l *Logs) extractTime(name, prefix, extension string) (time.Time, error) {
	if !strings.HasSuffix(name, extension) {
		return time.Time{}, errors.New("unmatched extension")
	}

	if !strings.HasPrefix(name, prefix) {
		return time.Time{}, errors.New("unmatched prefix")
	}

	timeStamp := name[len(prefix) : len(name)-len(extension)]
	return time.Parse(timeFormat, timeStamp)
}

// checkParam checks whether the parameters are correct
func checkParam(volume int, time int) error {
	if volume != 0 {
		if volume < minSaveVolume || volume > maxSaveVolume {
			return fmt.Errorf("the value of savevolume is incorrect")
		}
	}
	if time != 0 {
		if time < minSaveTime || time > maxSaveTime {
			return fmt.Errorf("the value of savetime is incorrect")
		}
	}
	return nil
}

// hasUserSwitched checks whether the euid/egid has been changed
func checkSwitchUser() bool {
	return syscall.Getuid() != syscall.Geteuid() || syscall.Getgid() != syscall.Getegid()
}
