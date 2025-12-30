// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package x509 provides the capability of x509.
package x509

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/rand"
)

const (
	maxLen       = 2048
	byteToBit    = 8
	byteSize     = 32
	dayHours     = 24
	tenYearHours = 24 * 365 * 10

	// InvalidNum invalid num
	InvalidNum = -9999999
	x509v3     = 3

	initSize           = 4
	defaultWarningDays = 100

	// KeyStorePath KeyStorePath
	KeyStorePath = "KeyStorePath"
	// KeyStoreBackupPath KeyStoreBackupPath
	KeyStoreBackupPath = "KeyStoreBackupPath"
	// CertStorePath CertStorePath
	CertStorePath = "CertStorePath"
	// CertStoreBackupPath CertStoreBackupPath
	CertStoreBackupPath = "CertStoreBackupPath"
	// PassFilePath PassFilePath
	PassFilePath = "PassFilePath"
	// PassFileBackUpPath PassFileBackUpPath
	PassFileBackUpPath = "PassFileBackUpPath"
)

var (
	// certificateMap  using certificate information
	certificateMap = make(map[string]*CertStatus, initSize)
	// warningDays cert warning day ,unit days
	warningDays = defaultWarningDays
	// checkInterval  cert period check interval,unit days
	checkInterval = 1
)

// GetCertStatus return certificateMap
func GetCertStatus() map[string]*CertStatus {
	return certificateMap
}

// SetPeriodCheckParam set period parameter
func SetPeriodCheckParam(warningDaysFlag, checkIntervalFlag int) {
	warningDays = warningDaysFlag
	checkInterval = checkIntervalFlag
}

// PeriodCheck  period check certificate, need call SetPeriodCheckParam firstly if you
// want to change checking interval
func PeriodCheck() {
	ticker := time.NewTicker(time.Duration(checkInterval) * dayHours * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			now := time.Now()
			for _, v := range certificateMap {
				checkCertStatus(now, v)
			}
		}
	}
}

// LoadCertPairByte load and valid encrypted certificate and private key
func LoadCertPairByte(pathMap map[string]string, encryptAlgorithm int, mode os.FileMode) ([]byte, []byte, error) {
	if err := CheckCertFiles(pathMap); err != nil {
		return nil, nil, err
	}
	key := pathMap[KeyStorePath]
	keyBkp := pathMap[KeyStoreBackupPath]
	certInstance, err := NewBKPInstance(nil, pathMap[CertStorePath], pathMap[CertStoreBackupPath])
	if err != nil {
		return nil, nil, err
	}
	certBytes, err := certInstance.ReadFromDisk(mode, false)
	if err != nil {
		return nil, nil, fmt.Errorf("there is no certFile provided: %s", err.Error())
	}

	pdInstance, err := NewBKPInstance(nil, pathMap[PassFilePath], pathMap[PassFileBackUpPath])
	if err != nil {
		return nil, nil, err
	}
	encodedPd, err := pdInstance.ReadFromDisk(mode, true)
	if err != nil {
		return nil, nil, err
	}

	if err = kmc.Initialize(encryptAlgorithm, "", ""); err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := kmc.Finalize(); err != nil {
			hwlog.RunLog.Errorf("%s", err.Error())
		}
	}()
	pd, err := kmc.Decrypt(0, encodedPd)
	if err != nil {
		return nil, nil, errors.New("decrypt passwd failed")
	}
	defer PaddingAndCleanSlice(pd)
	hwlog.RunLog.Info("decrypt passwd successfully")
	err = checkWhetherEncrypted(key, keyBkp, err)
	if err != nil {
		return nil, nil, err
	}
	keyBlock, err := DecryptPrivateKeyWithPd(key, keyBkp, pd)
	if err != nil {
		return nil, nil, err
	}
	defer PaddingAndCleanSlice(keyBlock.Bytes)
	hwlog.RunLog.Info("decrypt success")
	return certBytes, pem.EncodeToMemory(keyBlock), nil
}

func checkWhetherEncrypted(mainKey string, keyBkp string, err error) error {
	checkPath := mainKey
	if !fileutils.IsExist(mainKey) {
		checkPath = keyBkp
	}
	isEncode, err := IsEncryptedKey(checkPath)
	if err != nil {
		return err
	}
	if !isEncode {
		return errors.New("mindx-dl don't support non-encrypted key ")
	}
	return nil
}

// CheckCertFiles check cert related files exist or not.
func CheckCertFiles(pathMap map[string]string) error {
	cert, ok := pathMap[CertStorePath]
	if !ok {
		return fmt.Errorf("%s is empty", CertStorePath)
	}
	certBkp, ok := pathMap[CertStoreBackupPath]
	if !ok {
		return fmt.Errorf("%s is empty", CertStoreBackupPath)
	}
	key, ok := pathMap[KeyStorePath]
	if !ok {
		return fmt.Errorf("%s is empty", KeyStorePath)
	}
	keyBkp, ok := pathMap[KeyStoreBackupPath]
	if !ok {
		return fmt.Errorf("%s is empty", KeyStoreBackupPath)
	}
	psFile, ok := pathMap[PassFilePath]
	if !ok {
		return fmt.Errorf("%s is empty", PassFilePath)
	}
	psFileBk, ok := pathMap[PassFileBackUpPath]
	if !ok {
		return fmt.Errorf("%s is empty", PassFileBackUpPath)
	}

	// if password file not exists, should remove privateKey and regenerate
	if !fileutils.IsExist(psFile) && !fileutils.IsExist(psFileBk) {
		hwlog.RunLog.Error("psFile and psFileBk file is not exist")
		return os.ErrNotExist
	}
	if !fileutils.IsExist(key) && !fileutils.IsExist(keyBkp) {
		hwlog.RunLog.Error("keyFile and keyBkp file is not exist")
		return os.ErrNotExist
	}
	if !fileutils.IsExist(cert) && !fileutils.IsExist(certBkp) {
		hwlog.RunLog.Error("certFile and certBkp file is not exist")
		return os.ErrNotExist
	}
	for _, v := range pathMap {
		_, err := fileutils.CheckOriginPath(v)
		if err == nil {
			continue
		}
		if err == os.ErrNotExist {
			continue
		}
		return err
	}
	return nil
}

func checkCertStatus(now time.Time, v *CertStatus) {
	if now.After(v.NotAfter) || now.Before(v.NotBefore) {
		hwlog.RunLog.Warnf("the certificate: %s is already overdue", v.FingerprintSHA256)
	}
	gapHours := v.NotAfter.Sub(now).Hours()
	overdueDays := gapHours / dayHours
	if overdueDays > math.MaxInt64 {
		overdueDays = math.MaxInt64
	}
	if overdueDays < float64(warningDays) && overdueDays >= 0 {
		hwlog.RunLog.Warnf("the certificate: %s will overdue after %d days later",
			v.FingerprintSHA256, int64(overdueDays))
	}
}

func write(path string, overrideByte []byte, mode os.FileMode) error {
	if err := ioutil.WriteFile(path, overrideByte, mode); err != nil {
		return errors.New("override password file failed")
	}
	return nil
}

// CheckCaCert check the import ca cert version, only used when import form user provide file which no backup file
func CheckCaCert(caFile string, overdueTime int) ([]byte, error) {
	caBytes, err := fileutils.LoadFile(caFile)
	if err != nil {
		return nil, err
	}
	if caBytes == nil {
		return nil, nil
	}
	return caBytes, VerifyCaCert(caBytes, overdueTime)
}

// VerifyCaCert  used when load ca from imported path
func VerifyCaCert(caBytes []byte, overdueTime int) error {
	caCrt, err := LoadCertsFromPEM(caBytes)
	if err != nil {
		return errors.New("convert ca certificate failed")
	}
	if !caCrt.IsCA {
		return errors.New("this is not ca certificate")
	}
	if err = CheckExtension(caCrt); err != nil {
		return err
	}
	switch overdueTime {
	case InvalidNum:
		err = CheckValidityPeriod(caCrt, false)
	default:
		err = CheckValidityPeriodWithError(caCrt, overdueTime)
	}
	if err != nil {
		return err
	}
	if err = caCrt.CheckSignature(caCrt.SignatureAlgorithm, caCrt.RawTBSCertificate, caCrt.Signature); err != nil {
		return errors.New("check ca certificate signature failed")
	}
	if err = AddToCertStatusTrace(caCrt); err != nil {
		return err
	}
	hwlog.RunLog.Infof("ca certificate signature check pass")
	return nil
}

// LoadCertsFromPEM load the certification from pem
func LoadCertsFromPEM(pemCerts []byte) (*x509.Certificate, error) {
	if len(pemCerts) <= 0 {
		return nil, errors.New("wrong input")
	}
	var block *pem.Block
	block, pemCerts = pem.Decode(pemCerts)
	if block == nil {
		return nil, errors.New("parse cert failed")
	}
	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, errors.New("invalid cert bytes")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("parse cert failed")
	}
	return cert, nil
}

// CheckExtension check the certificate extensions, the cert version must be x509v3 and if the cert is ca,
//
// need check keyUsage, the keyUsage must include keyCertSign.
//
// detail information refer to https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.3
func CheckExtension(cert *x509.Certificate) error {
	if cert.Version != x509v3 {
		return errors.New("the certificate must be x509v3")
	}
	if !cert.IsCA {
		return nil
	}
	// ca cert need check whether the keyUsage include CertSign
	if (cert.KeyUsage & x509.KeyUsageCertSign) != x509.KeyUsageCertSign {
		msg := "CA certificate keyUsage didn't include keyCertSign"
		return errors.New(msg)
	}
	return nil
}

// CheckCaExtension check the cert extensions,
// the cert version must be x509v3 and the cert should be ca contains keyCertSign keyUsage
func CheckCaExtension(cert *x509.Certificate) error {
	if cert.Version != x509v3 {
		return errors.New("the certificate must be x509v3")
	}
	if !cert.IsCA || !cert.BasicConstraintsValid {
		return errors.New("the cert is not a ca cert")
	}
	// ca cert need check whether the keyUsage include CertSign
	if (cert.KeyUsage & x509.KeyUsageCertSign) != x509.KeyUsageCertSign {
		msg := "CA certificate keyUsage didn't include keyCertSign"
		return errors.New(msg)
	}
	return nil
}

// CheckValidityPeriod check certification validity period
// options is used to indicate whether allow the certificate has one day grace period before "NotBefore"
func CheckValidityPeriod(cert *x509.Certificate, allowFutureEffective bool) error {
	overdueDays, err := GetValidityPeriod(cert, allowFutureEffective)
	if err != nil {
		return err
	}
	if overdueDays < float64(warningDays) && overdueDays > 0 {
		hwlog.RunLog.Warnf("the certificate will overdue after %d days later", int64(overdueDays))
	}

	return nil
}

// CheckValidityPeriodWithError if the time expires, an error is reported
func CheckValidityPeriodWithError(cert *x509.Certificate, overdueTime int) error {
	overdueDays, err := GetValidityPeriod(cert, false)
	if err != nil {
		return err
	}
	if overdueDays <= float64(overdueTime) {
		return fmt.Errorf("overdueDayes is (%#v) need to update certification", overdueDays)
	}
	return nil
}

// AddToCertStatusTrace  add cert status to trace map
func AddToCertStatusTrace(cert *x509.Certificate) error {
	if cert == nil {
		return errors.New("cert is nil")
	}
	sh256 := sha256.New()
	_, err := sh256.Write(cert.Raw)
	if err != nil {
		return err
	}
	fpsha256 := hex.EncodeToString(sh256.Sum(nil))

	cs := &CertStatus{
		NotBefore:         cert.NotBefore,
		NotAfter:          cert.NotAfter,
		IsCA:              cert.IsCA,
		FingerprintSHA256: fpsha256,
	}
	certificateMap[fpsha256] = cs
	return nil
}

// GetValidityPeriod get certification validity period
func GetValidityPeriod(cert *x509.Certificate, allowFutureEffective bool) (float64, error) {
	now := time.Now().In(time.UTC)
	notBefore := cert.NotBefore
	if allowFutureEffective {
		notBefore = notBefore.Add(-dayHours * time.Hour)
	}

	if now.After(cert.NotAfter) || now.Before(notBefore) {
		return 0, errors.New("the certificate overdue ")
	}
	if cert.NotAfter.Sub(cert.NotBefore).Hours() > tenYearHours {
		hwlog.RunLog.Warn("the certificate valid period is more than 10 years")
	}
	gapHours := cert.NotAfter.Sub(now).Hours()
	overdueDays := gapHours / dayHours
	if overdueDays > math.MaxInt64 {
		overdueDays = math.MaxInt64
	}
	return overdueDays, nil
}

// IsEncryptedKey check key is encrypted or not
func IsEncryptedKey(keyFile string) (bool, error) {
	keyBytes, err := fileutils.ReadLimitBytes(keyFile, fileutils.Size10M)
	if err != nil {
		return false, err
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return false, errors.New("decode key file failed")
	}
	return x509.IsEncryptedPEMBlock(block), nil
}

// CheckSignatureAlgorithm check signature algorithm of the certification
func CheckSignatureAlgorithm(cert *x509.Certificate) error {
	var signAl = cert.SignatureAlgorithm.String()
	if strings.Contains(signAl, "MD2") || strings.Contains(signAl, "MD5") ||
		strings.Contains(signAl, "SHA1") || signAl == "0" {
		return errors.New("the signature algorithm is unsafe,please use safe algorithm ")
	}
	hwlog.RunLog.Info("signature algorithm validation passed")
	return nil
}

// CheckPubKeyLength checks the length of the pub key length of a cert.
// Argument allowMiddleStrengthRsaPublicKey indicates whether we accept a 2048-bit rsa public key.
func CheckPubKeyLength(cert *x509.Certificate, allowMiddleStrengthRsaPublicKey ...bool) error {
	publicKey := cert.PublicKey
	pubKeyAlgo := cert.PublicKeyAlgorithm

	const (
		MinRsaPubKeyLen            = 2048
		MinRecommendedRsaPubKeyLen = 3072
		MinEcdsaPubKenLen          = 256
	)

	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		if pubKeyAlgo != x509.RSA {
			return errors.New("the public key mismatches the public key algo")
		}
		minRasPubKeyLen := MinRecommendedRsaPubKeyLen
		if len(allowMiddleStrengthRsaPublicKey) > 0 && allowMiddleStrengthRsaPublicKey[0] {
			minRasPubKeyLen = MinRsaPubKeyLen
		}

		pubKeyLen := pub.N.BitLen()
		if pubKeyLen < minRasPubKeyLen {
			return fmt.Errorf("the length of RSA public key %d less than %d", pubKeyLen, minRasPubKeyLen)
		}

		if pubKeyLen < MinRecommendedRsaPubKeyLen {
			hwlog.RunLog.Warnf("The length of RSA public key %d for cert [%s] less than %d, which is not recommended",
				pubKeyLen, cert.Subject.String(), MinRecommendedRsaPubKeyLen)
		}
	case *ecdsa.PublicKey:
		if pubKeyAlgo != x509.ECDSA {
			return errors.New("the public key mismatches the publick key algo")
		}

		if params := pub.Params(); params.BitSize < MinEcdsaPubKenLen {
			return errors.New("pub key length is not enough")
		}
	default:
		return errors.New("unsupported public key type")
	}

	return nil
}

// GetPrivateKeyLength  return the length and type of private key
func GetPrivateKeyLength(cert *x509.Certificate, certificate *tls.Certificate) (int, string, error) {
	if certificate == nil {
		return 0, "", errors.New("certificate is nil")
	}
	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		priv, ok := certificate.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return 0, "RSA", errors.New("get rsa key length failed")
		}
		return priv.N.BitLen(), "RSA", nil
	case *ecdsa.PublicKey:
		priv, ok := certificate.PrivateKey.(*ecdsa.PrivateKey)
		if !ok {
			return 0, "ECC", errors.New("get ecdsa key length failed")
		}
		return len(priv.X.Bytes()) * byteToBit, "ECC", nil
	case ed25519.PublicKey:
		priv, ok := certificate.PrivateKey.(ed25519.PrivateKey)
		if !ok {
			return 0, "ED25519", errors.New("get ed25519 key length failed")
		}
		return len(priv.Public().(ed25519.PublicKey)), "ED25519", nil
	default:
		return 0, "", errors.New("get key length failed")
	}
}

// CheckRevokedCert check the revoked certification
func CheckRevokedCert(r *http.Request, crlcerList *pkix.CertificateList) bool {
	if crlcerList == nil || r.TLS == nil {
		hwlog.RunLog.Warnf("certificate or revokelist is nil")
		return false
	}
	revokedCertificates := crlcerList.TBSCertList.RevokedCertificates
	if len(revokedCertificates) == 0 {
		hwlog.RunLog.Warnf("revoked certificate length is 0")
		return false
	}
	// r.TLS.VerifiedChains [][]*x509.Certificate ,certificateChain[0] : current chain
	// certificateChain[0][0] : current certificate, certificateChain[0][1] :  certificate's issuer
	certificateChain := r.TLS.VerifiedChains
	if len(certificateChain) == 0 || len(certificateChain[0]) <= 1 {
		hwlog.RunLog.Warnf("VerifiedChains length is 0,or certificate is Cafile cannot revoke")
		return false
	}
	hwlog.RunLog.Infof("VerifiedChains length: %d,CertificatesChains length %d",
		len(certificateChain), len(certificateChain[0]))
	// CheckCRLSignature check CRL's issuer is certificate's issuer
	if err := certificateChain[0][1].CheckCRLSignature(crlcerList); err != nil {
		hwlog.RunLog.Warnf("CRL's issuer is not certificate's issuer")
		return false
	}
	for _, revokeCert := range revokedCertificates {
		for _, cert := range r.TLS.PeerCertificates {
			if cert.SerialNumber.Cmp(revokeCert.SerialNumber) == 0 {
				hwlog.RunLog.Warnf("revoked certificate SN: %s", cert.SerialNumber)
				return true
			}
		}
	}
	return false
}

// Interceptor Interceptor ensure https
func Interceptor(h http.Handler, crlCertList *pkix.CertificateList) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if crlCertList != nil && CheckRevokedCert(r, crlCertList) {
			return
		}
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		h.ServeHTTP(w, r)
	})
}

// EncryptPrivateKeyAgain encrypt PrivateKey with local password again, and encrypted save password into files
func EncryptPrivateKeyAgain(key *pem.Block, psFile, psBkFile string, encrypt int) (*pem.Block, error) {
	return EncryptPrivateKeyAgainWithMode(key, psFile, psBkFile, encrypt, fileutils.Mode600)
}

// EncryptPrivateKeyAgainWithMode encrypt private key again with mode
func EncryptPrivateKeyAgainWithMode(key *pem.Block, psFile, psBkFile string, encrypt int, mode os.FileMode) (*pem.Block,
	error) {
	// generate new passwd for private key
	pd, err := GetRandomPass()
	if err != nil {
		return nil, errors.New("generate passwd failed")
	}
	if err := kmc.Initialize(encrypt, "", ""); err != nil {
		return nil, err
	}
	encryptedPd, err := kmc.Encrypt(0, pd)
	if err != nil {
		return nil, errors.New("encrypt passwd failed")
	}
	hwlog.RunLog.Info("encrypt new passwd successfully")
	pwInstance, err := NewBKPInstance(encryptedPd, psFile, psBkFile)
	if err != nil {
		return nil, err
	}
	if err := pwInstance.WriteToDisk(mode, true); err != nil {
		return nil, errors.New("create or update  passwd backup file failed")
	}
	hwlog.RunLog.Info("create or update passwd backup file successfully")
	encryptedBlock, err := x509.EncryptPEMBlock(rand.Reader, key.Type, key.Bytes, pd, x509.PEMCipherAES256)
	if err != nil {
		return nil, errors.New("encrypted private key failed")
	}
	hwlog.RunLog.Info("encrypt private key by new passwd successfully")
	// clean password
	PaddingAndCleanSlice(pd)

	// wait certificate verify passed and then write key to file together
	if err := kmc.Finalize(); err != nil {
		hwlog.RunLog.Errorf("%s", err.Error())
	}
	return encryptedBlock, nil
}

// ValidateCRL ValidateCRL
func ValidateCRL(crlBytes []byte) (*pkix.CertificateList, error) {
	crlList, err := x509.ParseCRL(crlBytes)
	if err != nil {
		return nil, errors.New("parse crlFile failed")
	}
	if time.Now().Before(crlList.TBSCertList.ThisUpdate) || time.Now().After(crlList.TBSCertList.NextUpdate) {
		return nil, errors.New("crlFile update time not match")
	}
	return crlList, nil
}

// CheckCRL validate crl file
func CheckCRL(crlFile string) ([]byte, error) {
	crlBytes, err := fileutils.LoadFile(crlFile)
	if err != nil {
		return nil, err
	}
	if crlBytes == nil {
		return nil, nil
	}
	_, err = ValidateCRL(crlBytes)
	if err != nil {
		return nil, err
	}
	return crlBytes, nil
}
