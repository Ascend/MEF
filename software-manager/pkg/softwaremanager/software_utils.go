// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package softwaremanager

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"software-manager/pkg/restfulservice"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
)

var dbCtlInstance SoftwareDbCtl
var softwareDbServiceSingleton sync.Once

// SoftwareDbCtlInstance returns the singleton instance of software database control
func SoftwareDbCtlInstance() SoftwareDbCtl {
	softwareDbServiceSingleton.Do(func() {
		dbCtlInstance = &softwareDbCtlImpl{
			db: getDb()}
	})
	return dbCtlInstance
}

func checkFields(contentType string, version string) error {
	if err := checkContentType(contentType); err != nil {
		return err
	}
	if err := checkVersion(version); err != nil {
		return err
	}
	return nil
}

func checkContentType(contentType string) error {
	if contentType != mefEdge {
		return fmt.Errorf("%s is a incorrect field", contentType)
	}
	return nil
}

func checkVersion(version string) error {
	match, err := regexp.MatchString(regexExp, version)
	if err != nil {
		return errors.New("regexp match error")
	}
	if !match {
		return fmt.Errorf("%s is a incorrect field", version)
	}
	return nil
}

func checkFile(file *multipart.FileHeader) (bool, error) {
	fileInf, err := file.Open()
	if err != nil {
		hwlog.RunLog.Error("file open error")
		return false, err
	}
	ok, err := checkZipType(fileInf)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return false, err
	}
	if !ok {
		hwlog.RunLog.Error("wrong file format")
		return false, nil
	}
	if float64(file.Size)/kbToMB > maximumSize {
		return false, nil
	}
	return true, nil
}

func checkSoftwareExist(contentType string, version string) (bool, error) {
	record, err := SoftwareDbCtlInstance().querySoftware(contentType, version)
	if err != nil {
		return false, err
	}
	if record == nil {
		return false, nil
	}
	return true, nil
}

func softwarePathJoin(contentType string, version string, fileName string) string {
	return filepath.Join(RepositoryFilesPath, contentType,
		fmt.Sprintf("%s%s%s", contentType, "_", version), fileName)
}

func returnLatestVer(contentType string) (string, error) {
	result, err := SoftwareDbCtlInstance().queryLaSoftware(contentType)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return "", err
	}
	if result == nil {
		hwlog.RunLog.Errorf("%s does not exist. Need to import one first", contentType)
		return "", fmt.Errorf("%s does not exist. Need to import one first", contentType)
	}
	return result.Version, nil
}

func creatDir(contentType string, version string) (string, error) {
	dst := filepath.Join(RepositoryFilesPath, contentType, contentType+"_"+version)
	if !utils.IsExist(dst) {
		err := os.MkdirAll(dst, os.ModePerm)
		if err != nil {
			hwlog.RunLog.Error("create directory error")
			return "", err
		}
	}
	return dst, nil
}

// extraZipFile extract zip file
func extraZipFile(zipFile, extractPath string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		hwlog.RunLog.Errorf("open zip reader failed, error: %v", err)
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			hwlog.RunLog.Errorf("close zip reader failed, error: %v", err)
			return
		}
	}()

	if len(reader.File) > maxExtractFileCount {
		hwlog.RunLog.Error("too many files will be unzip")
		return errors.New("too many files will be unzip")
	}

	var totalWrote uint64
	for _, file := range reader.File {
		if err = copyZipFile(extractPath, file); err != nil {
			return err
		}
		totalWrote += file.UncompressedSize64
	}
	return nil
}

func copyZipFile(extractPath string, file *zip.File) error {
	extraFilePath := filepath.Join(extractPath, file.Name)
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(extraFilePath, file.Mode()); err != nil {
			hwlog.RunLog.Errorf("create path [%s] failed, error: %v", extraFilePath, err)
			return err
		}
		return nil
	}

	fileReader, err := file.Open()
	if err != nil {
		hwlog.RunLog.Errorf("open src file [%s] failed, error: %v", file.Name, err)
		return err
	}
	defer func() {
		if err = fileReader.Close(); err != nil {
			hwlog.RunLog.Errorf("close src file [%s] failed, error: %v", file.Name, err)
			return
		}
	}()

	targetFile, err := os.OpenFile(extraFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		hwlog.RunLog.Errorf("open dst file [%s] failed", file.Name)
		return err
	}
	defer func() {
		if err = targetFile.Close(); err != nil {
			hwlog.RunLog.Errorf("close dst file [%s] failed, error: %v", file.Name, err)
			return
		}
	}()

	if _, err = io.CopyN(targetFile, fileReader, int64(file.UncompressedSize64)); err != nil {
		hwlog.RunLog.Errorf("copy src file [%s] failed, error: %v", file.Name, err)
		return err
	}
	return nil
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		hwlog.RunLog.Error("open file error")
		return err
	}
	defer src.Close()
	out, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, dBFileMode)
	if err != nil {
		hwlog.RunLog.Error("open file error")
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, src); err != nil {
		hwlog.RunLog.Error("copy file error")
		return err
	}
	return nil
}

func checkDownloadRight(userName string, password []byte, nodeID string) bool {
	userInfo, exist := restfulservice.QueryUserInfo(nodeID)
	if !exist {
		return false
	}
	if userName != (*userInfo).UserName {
		hwlog.RunLog.Error("wrong userName")
		return false
	}
	if !checkByteArr(password, (*userInfo).Password) {
		hwlog.RunLog.Error("wrong password")
		return false
	}
	return true
}

func geneRandStr(source string, length int) (string, error) {
	if length > maxLength {
		hwlog.RunLog.Error("user length is too long")
		return "", errors.New("user length is too long")
	}
	rand.Seed(time.Now().UnixNano())
	randomByteArr := make([]byte, length, length)
	for i := 0; i < length; i++ {
		index := rand.Intn(len(source))
		randomByteArr[i] = source[index]
	}
	return string(randomByteArr), nil
}

func geneUsrPsw(nodeID string) (string, *[]byte, error) {
	userName, err := geneRandStr(randomSet, userLength)
	if err != nil {
		return "", nil, err
	}
	password, err := x509.GetRandomPass()
	if err != nil {
		return "", nil, err
	}
	userInfo := &restfulservice.UserPriInfo{
		UserName: userName,
		Password: password,
	}
	err = restfulservice.AddUserInfo(nodeID, userInfo)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return "", nil, err
	}
	return userName, &password, nil
}

func checkNodeID(nodeID string) bool {
	if nodeID == "" {
		hwlog.RunLog.Info("incorrect node_id")
	}
	return nodeID != ""
}

func checkByteArr(arr1 []byte, arr2 []byte) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for index, val := range arr1 {
		if val != arr2[index] {
			return false
		}
	}
	return true
}

// DownloadInfo [struct] to software download info
type DownloadInfo struct {
	Package  string  `json:"package"`
	SignFile string  `json:"signFile,omitempty"`
	CrlFile  string  `json:"crlFile,omitempty"`
	UserName string  `json:"username"`
	Password *[]byte `json:"password"`
}

func findPackageFileName(dir string) string {
	files, _ := os.ReadDir(dir)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar.gz") {
			return file.Name()
		}
	}
	return ""
}

func findSignFileName(dir string) string {
	files, _ := os.ReadDir(dir)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar.gz.cms") {
			return file.Name()
		}
	}
	return ""
}

func findCrlFileName(dir string) string {
	files, _ := os.ReadDir(dir)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar.gz.crl") {
			return file.Name()
		}
	}
	return ""
}

func fillDownloadData(downloadInfo *DownloadInfo, urlReq *URLReq) error {
	url := "GET " + "https://" + IP + ":" + strconv.Itoa(Port) + "/softwaremanager/v1/?" +
		"contentType=" + urlReq.ContentType + "&version=" + urlReq.Version
	ps := "Atlas12#$"
	pwd := []byte(ps)

	pkgDir := filepath.Join(RepositoryFilesPath, urlReq.ContentType,
		fmt.Sprintf("%s%s%s", urlReq.ContentType, "_", urlReq.Version))
	downloadInfo.UserName = "admin"
	downloadInfo.Password = &pwd
	downloadInfo.Package = url + "&fileName=" + findPackageFileName(pkgDir)
	downloadInfo.SignFile = url + "&fileName=" + findSignFileName(pkgDir)
	downloadInfo.CrlFile = url + "&fileName=" + findCrlFileName(pkgDir)
	return nil
}

func deleteSoftware(id int, notDeleteID *[]int) error {
	result, err := SoftwareDbCtlInstance().querySoftwareByID(id)
	if err != nil {
		*notDeleteID = append(*notDeleteID, id)
		hwlog.RunLog.Error(err.Error())
		return err
	}
	if result == nil {
		*notDeleteID = append(*notDeleteID, id)
		hwlog.RunLog.Errorf("software(ID=%d) dose not exist", id)
		return fmt.Errorf("software(ID=%d) dose not exist", id)
	}
	dst := filepath.Join(RepositoryFilesPath, result.ContentType,
		fmt.Sprintf("%s%s%s", result.ContentType, "_", result.Version))
	if err := os.RemoveAll(dst); err != nil {
		hwlog.RunLog.Errorf("delete software(ID=%d) error：%s", id, err.Error())
		*notDeleteID = append(*notDeleteID, id)
		return err
	}
	if err := SoftwareDbCtlInstance().deleteSoftware(id, notDeleteID); err != nil {
		hwlog.RunLog.Error(fmt.Sprintf("database delete id=%d error in dbDelete", id))
		return err
	}
	return nil
}

const readBufSize int = 20

func checkZipType(reader io.Reader) (bool, error) {
	buf := make([]byte, readBufSize)
	n, err := reader.Read(buf)
	if err != nil {
		return false, err
	}
	fileCode := bytesToHexString(buf[:n])
	return strings.HasPrefix(fileCode, zipFileHeader), nil
}

func bytesToHexString(src []byte) string {
	res := bytes.Buffer{}
	if src == nil || len(src) <= 0 {
		return ""
	}
	var temp []byte
	i, length := maxByteLength, len(src)
	if length < i {
		i = length
	}
	for j := 0; j < i; j++ {
		sub := src[j] & hexTag
		hv := hex.EncodeToString(append(temp, sub))
		if len(hv) < stringLength {
			res.WriteString(strconv.FormatInt(int64(0), common.BaseHex))
		}
		res.WriteString(hv)
	}
	return res.String()

}
