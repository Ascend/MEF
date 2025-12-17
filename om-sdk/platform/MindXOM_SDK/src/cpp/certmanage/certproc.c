/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */
#include "certproc.h"

#include <openssl/x509v3.h>
#include <openssl/bn.h>
#include <openssl/asn1.h>
#include <openssl/x509.h>
#include <openssl/pem.h>
#include <openssl/bio.h>

#include <openssl/rsa.h>
#include <openssl/ssl.h>
#include <openssl/pem.h>
#include <openssl/err.h>

#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>
#include <mntent.h>
#include <sys/wait.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <regex.h>

#include "securec.h"
#include "log_common.h"
#include "file_checker.h"

/* 厂家信息结构体 */
struct csr_entry {
    char *item;
    char *value;
};

/* temp certificate macro */
#define TEMPDIR_CERTCONF "/home/data/config/tmp/"
#define TEMPFILEPATH_CSR (TEMPDIR_CERTCONF "custom_csr.pem")
#define TEMPFILEPATH_UPLOAD_PUBLIC_PEM (TEMPDIR_CERTCONF "public_tmp.pem")
#define TEMPFILEPATH_UPLOAD_CUSTOM_CERT (TEMPDIR_CERTCONF "custom_cert_tmp.pem")
#define TEMPFILEPATH_UPLOAD_CUSTOM_PRIV (TEMPDIR_CERTCONF "custom_priv_tmp.pem")
#define TEMPFILEPATH_UPLOAD_CUSTOM_PSD (TEMPDIR_CERTCONF "custom_tmp.psd")

#define RSA_KEY_SIZE (4096)
#define CSR_ENTRIES (6)
#define MAX_CERTFILE_SIZE (10 * 1024)
#define WORKKEY_LEN (24)
#define MAX_CERTFILE_PATHLEN (1024)
#define MAX_INT_VAL (2147483647)

#define MAX_LINE_LENGTH 256
#define CERT_VALIDITY_THRESHOLD 90
#define VOS_VFS_MAX_PATH_LENGTH 256
#define MAX_ENC_LEN 1024
#define MAX_DEC_DATA 1024
#define MAX_ARRAY_LENGTH 1024
#define SLINE_BEGIN_LENGTH 2
#define DICT_SET_LENGTH 50
#define ONE_LINE_LENGTH 1024
#define BASE_DECIMAL_SYSTEM 10
#define RANDOM_SEQUENCE_LEN 12
#define MAX_CHAIN_NUMS 10

/* 内部数据定义 */
#define INNER_NUMBER_ZERO 0
#define INNER_NUMBER_ONE 1
#define INNER_NUMBER_TWO 2
#define INNER_NUMBER_THREE 3

/* Begin: 错误码定义 */
#define ERR_CERT (0x20140000)                            // offset
#define ERR_CERT_GENERATECSR_FAILED (ERR_CERT + 0x01)    // 产生CSR文件失败
#define ERR_CERT_LOADCERT_FAILED (ERR_CERT + 0x02)       // 加载证书文件失败
#define ERR_CERT_NOT_MATCHING (ERR_CERT + 0x03)          // 证书和私钥不匹配
#define ERR_CERT_FILELEN_ERROR (ERR_CERT + 0x04)         // 证书文件过大
#define ERR_CERT_PFX2PEM_FAILED (ERR_CERT + 0x05)        // PFX格式证书转换成PEM格式失败
#define ERR_CERT_NO_PRIVKEY_EXIST (ERR_CERT + 0x06)      // 私钥证书不存在
#define ERR_CERT_DECRTYPEDPRIVK_FAILED (ERR_CERT + 0x07) // 私钥证书解密失败
#define ERR_CERT_DECRYPT_FAILED (ERR_CERT + 0x08)        // 证书解密失败，用于解密自定义证书失败场景
#define ERR_CERT_VALIDITY_WARN (ERR_CERT + 0x09)         // 证书有效期告警
#define ERR_CERT_EARLY_USE_WARN (ERR_CERT + 0x10)        // 证书过早使用
#define ERR_CERT_PASSWD_TOO_LONG (ERR_CERT + 0x11)       // 证书密码过长
#define ERR_CERT_CHECK_PARAM_ERR (ERR_CERT + 0x12)       // 参数校验失败


#define MAX_CERT_WARN_DAY 180
#define MIN_CERT_WARN_DAY 7
#define CERT_FAULTTIME_CONFIG_FILE "/home/data/ies/certWarnTime.ini"
#define CERT_FAULTTIME_CONFIG_FILE_NAME "certWarnTime.ini"


const int DEFAULT_PROC_SEM_KEY = 0;
const int DEFAULT_DOMAIN_COUNT = 2;
const int DEFAULT_PERMISSION = 0600;
const int DEFAULT_ROLE = 1;
const unsigned int DEFAULT_DOMAIN_ID = 0;
const unsigned int DEFAULT_SDP_ALG_ID = 9;
const char *SDP_ALG_ID = "sdp_alg_id";

static char *RedefineStrtok(char *str, const char *delim)
{
    char *saveptr = NULL;

    if ((str == NULL) || (delim == NULL)) {
        OM_LOG_ERROR("inputs null");
        return NULL;
    }
    return strtok_r(str, delim, &saveptr);
}

int GetJsonKeyString(const char *title, const char *filename, char *buf, unsigned int bufSize)
{
    if ((title == NULL) || (filename == NULL) || (buf == NULL)) {
        OM_LOG_ERROR("input null.");
        return VOS_ERR;
    }

    FILE *fp;
    char *wTmp = NULL;
    char sLine[ONE_LINE_LENGTH] = {0};
    char *bufTmp = NULL;

    if ((fp = safety_fopen(filename, "r")) == NULL) {
        OM_LOG_ERROR("open config file failed");
        return VOS_ERR;
    }

    for (int i = 0; i < MAX_LINE_LENGTH; i++) {
        if (fgets(sLine, ONE_LINE_LENGTH, fp) == NULL) {
            break;
        }

        if (strncmp("//", sLine, INNER_NUMBER_TWO) == 0) {
            continue;
        }
        if (sLine[0] == '#') {
            continue;
        }
        if (sLine[0] == '{') {
            continue;
        }
        wTmp = strstr(sLine, title);
        if ((wTmp != NULL)) {
            (void)fclose(fp);
            wTmp = strchr(sLine, ':');
            if (wTmp == NULL) {
                return VOS_ERR;
            }
            while (*(wTmp + 1) == ' ') {
                wTmp++;
            }
            bufTmp = RedefineStrtok(wTmp + 1, "\"");
            if (bufTmp == NULL) {
                return VOS_ERR;
            }
            int ret = strcpy_s(buf, bufSize, bufTmp);
            if (ret != 0) {
                return VOS_ERR;
            }
            return VOS_OK;
        }
    }
    (void)fclose(fp);
    return VOS_ERR;
}

static unsigned int GetAlgIdFromConfig(const char *sectionName, const char *alg_json_file)
{
    char algIdStr[SECURE_LEN] = {0};
    unsigned int algId = 0;
    int tmpAlgId = 0;
    if (check_file_path_valid(alg_json_file) != COMMON_SUCCESS) {
        OM_LOG_ERROR("alg conf load fail for filename not valid");
        return 0;
    }
    if (GetJsonKeyString(sectionName, alg_json_file, algIdStr, SECURE_LEN) == VOS_OK) {
        tmpAlgId = StrToInt(algIdStr, SECURE_LEN);
    }
    if (tmpAlgId > 0) {
        algId = tmpAlgId;
    }
    return algId;
}

signed long cert_manage_init(const char *primary_file, const char *standby_file, const char *alg_json_file)
{
    if (check_file_path_valid(primary_file) != COMMON_SUCCESS) {
        OM_LOG_ERROR("KmcConfig primary path not valid.");
        return VOS_ERR;
    }
    if (check_file_path_valid(standby_file) != COMMON_SUCCESS) {
        OM_LOG_ERROR("KmcConfig standby path not valid.");
        return VOS_ERR;
    }

    // replace the following code to initialize cert management library
    FILE* fp;
    struct stat stat_result;
    int i;
    const char* filename;
    const char* keystore_files[] = {primary_file, standby_file, NULL};
    for (i = 0; ; i++) {
        filename = keystore_files[i];
        if (filename == NULL) {
            break;
        }

        if (stat(filename, &stat_result) == 0 && stat_result.st_size > 0) {
            continue;
        }

        fp = safety_fopen(filename, "w");
        if (fp == NULL) {
            OM_LOG_ERROR("open keystore file %s fail", filename);
            return VOS_ERR;
        }
        fwrite("\n", 1, 1, fp);
        (void)fclose(fp);
    }
    return VOS_OK;
}

signed long cert_manage_finalize(void)
{
    // Insert code here to finalize cert management library
    return VOS_OK;
}

static int SecurityEncrypt(const char *buf,
                           unsigned int bufLen,
                           char *bufEnc,
                           unsigned int *bufEncLen,
                           const char *alg_json_file)
{
    unsigned int sdpAlgId = GetAlgIdFromConfig(SDP_ALG_ID, alg_json_file);
    if (sdpAlgId == 0) {
        sdpAlgId = DEFAULT_SDP_ALG_ID;
    }
    // Replace the following 2 lines to encrypt buf
    *bufEncLen = bufLen;
    int ret = memcpy_s(bufEnc, *bufEncLen, buf, bufLen);

    if (ret != 0) {
        OM_LOG_ERROR("Encrypt failed, result %d", (int)ret);
        return VOS_ERR;
    }
    return VOS_OK;
}

static int get_cert_validity_threshold()
{
    char time_info[MAX_BUF_LEN_128] = {0};
    int day = CERT_VALIDITY_THRESHOLD;
    FILE *fp = NULL;
    size_t read_len;
    int t_day;

    if (access(CERT_FAULTTIME_CONFIG_FILE, F_OK) == 0) {
        fp = safety_fopen(CERT_FAULTTIME_CONFIG_FILE, "r");
        if (fp == NULL) {
            OM_LOG_ERROR("open file %s fail", CERT_FAULTTIME_CONFIG_FILE_NAME);
            return day;
        }

        read_len = fread(time_info, 1, sizeof(time_info) - 1, fp);
        if (read_len <= 0) {
            OM_LOG_ERROR("read cert valid threshold fail");
            (void)fclose(fp);
            return day;
        }
        (void)fclose(fp);
        fp = NULL;

        t_day = StrToInt(time_info, MAX_BUF_LEN_128);
        if ((t_day >= MIN_CERT_WARN_DAY) && (t_day <= MAX_CERT_WARN_DAY)) {
            day = t_day;
        } else {
            OM_LOG_ERROR("cert valid threshold: %d invalid.", t_day);
        }
    }

    return day;
}

static FILE *certOpenCertfile(const char *certinfo_file)
{
    if (certinfo_file == NULL) {
        OM_LOG_ERROR("inputs null");
        return NULL;
    }

    FILE *fp = safety_fopen(certinfo_file, "r");
    if (fp == NULL) {
        OM_LOG_ERROR("fopen failed");
    }

    return fp;
}

signed long certPeriodicalCheck(const char *certinfo_file, int *day)
{
    ASN1_TIME *not_after = NULL;
    ASN1_TIME *current_asn1 = NULL;
    X509 *cert = NULL;
    time_t current_time;
    time_t cert_warning_time;
    int sec = 0;
    if ((certinfo_file == NULL) || (day == NULL)) {
        return VOS_ERR;
    }

    FILE *fp = certOpenCertfile(certinfo_file);
    if (fp == NULL) {
        return VOS_ERR;
    }

    /* read cert */
    cert = PEM_read_X509(fp, NULL, NULL, NULL);
    if (cert == NULL) {
        OM_LOG_ERROR("unable to parse certificate");
        (void)fclose(fp);
        return VOS_ERR;
    }

    if (fclose(fp) != 0) {
        goto __error;
    }
    /* read the local time */
    current_time = time((time_t *)NULL);
    int cert_validity_threshold = get_cert_validity_threshold();
    cert_warning_time = current_time + cert_validity_threshold * 24 * 60 * 60; // one day 24 * 60 * 60

    /* read not after time */
    not_after = X509_get_notAfter(cert);

    /* read current time to ASN1_TIME */
    current_asn1 = ASN1_TIME_set(current_asn1, current_time);
    if (current_asn1 == NULL) {
        goto __error;
    }

    /* caculate diff time */
    int ret = ASN1_TIME_diff(day, &sec, (const ASN1_TIME *)current_asn1, (const ASN1_TIME *)not_after);
    if (ret != 1) {
        goto __error;
    }

    /* earlier than current time */
    if (X509_cmp_time(not_after, &cert_warning_time) < 0) {
        OM_LOG_ERROR("cert not_after time is earlier than cert_warning_time");
        X509_free(cert);
        ASN1_TIME_free(current_asn1);
        return ERR_CERT_VALIDITY_WARN;
    }

    X509_free(cert);
    ASN1_TIME_free(current_asn1);
    return VOS_OK;
__error:
    OM_LOG_ERROR("certPeriodicalCheck error");
    X509_free(cert);
    ASN1_TIME_free(current_asn1);
    return VOS_ERR;
}


/* ****************************************************************************
 功能描述  : 转换ASN1_TIME的输出格式
**************************************************************************** */
static int convert_ASN1TIME(ASN1_TIME *t, char *buf, int len)
{
    int rc;
    BIO *b = BIO_new(BIO_s_mem());
    rc = ASN1_TIME_print(b, t);
    if (rc <= 0) {
        OM_LOG_ERROR("ASN1_TIME_print failed or wrote no data.");
        BIO_free(b);
        return VOS_ERR;
    }
    rc = BIO_gets(b, buf, len);
    if (rc <= 0) {
        OM_LOG_ERROR("BIO_gets call failed to transfer contents to buf");
        BIO_free(b);
        return VOS_ERR;
    }
    BIO_free(b);
    return VOS_OK;
}

/* ****************************************************************************
 功能描述  : 转换subject 和 issuer 的输出格式
**************************************************************************** */
static signed long transfer_format(char *tmpBuf, char *dest, unsigned long dest_length)
{
    /*
    for example:
    from:  /C=US/ST=California/L=Mountain View/O=Google Inc/CN=*.google.com
    to:    C=US, ST=Texas, L=Austin, O=Polycom Inc., OU=Video Division, CN=a.digitalnetbr.net
    */
    int i;
    int curr_spot = 0;
    char *s = tmpBuf + 1; /* skip the first slash */
    char *c = s;
    int ret;
    while (1) {
        if (((*s == '/') &&
            ((s[1] >= 'A') && (s[1] <= 'Z') &&  // 1为字符数组索引1
            ((s[2] == '=') || ((s[2] >= 'A') && (s[2] <= 'Z') &&  // 2为字符数组索引2
            (s[3] == '='))))) ||  //  3为字符数组索引3
            (*s == '\0')) {
            i = s - c;
            if (curr_spot + i >= dest_length) {
                return VOS_ERR;
            }
            ret = strncpy_s(dest + curr_spot, dest_length - curr_spot, c, i);
            if (ret != 0) {
                return VOS_ERR;
            }
            curr_spot += i;

            if (curr_spot + INNER_NUMBER_TWO >= dest_length) {
                return VOS_ERR;
            }
            c = s + 1; /* skip following slash */
            if ((*s != '\0') && (strncpy_s(dest + curr_spot, dest_length - curr_spot, ", ", INNER_NUMBER_TWO) != 0)) {
                return VOS_ERR;
            }
            if (*s != '\0') {
                curr_spot += INNER_NUMBER_TWO;
            }
        }
        if (*s == '\0') {
            break;
        }
        ++s;
    }

    return VOS_OK;
}

/* ****************************************************************************
 功能描述  : 获取证书信息
**************************************************************************** */
static X509 *cert_readcertfile(const char *certinfo_file)
{
    FILE *fp = safety_fopen(certinfo_file, "r");
    if (fp == NULL) {
        return NULL;
    }

    /* read cert */
    X509 *cert = PEM_read_X509(fp, NULL, NULL, NULL);
    if (cert == NULL) {
        OM_LOG_ERROR("unable to parse certificate");
        (void)fclose(fp);
        return NULL;
    }

    if (fclose(fp) != 0) {
        OM_LOG_ERROR("unable to parse certificate");
        X509_free(cert);
        return NULL;
    }

    return cert;
}

/* ****************************************************************************
 功能描述  : 获取吊销列表信息
**************************************************************************** */
static X509_CRL *read_crl_file(const char *crl_file)
{
    X509_CRL *crl = NULL;
    FILE *fp = safety_fopen(crl_file, "r");
    if (fp == NULL) {
        OM_LOG_ERROR("open crl file failed.");
        return NULL;
    }
    // 调用PEM_read_X509_CRL函数，从fp中解析出CRL
    crl = PEM_read_X509_CRL(fp, NULL, NULL, NULL);
    if (crl == NULL) {
        OM_LOG_ERROR("PEM_read_X509_CRL failed.");
        fclose(fp);
        return NULL;
    }

    if (fclose(fp) != 0) {
        OM_LOG_ERROR("unable to close crl file.");
        X509_CRL_free(crl);
        return NULL;
    }
    return crl;
}

/* ****************************************************************************
 功能描述  : 读取证书信息中的subjec和issuer
**************************************************************************** */
static void cert_read_subject_and_issuer(const X509 *cert, char subj[],
                                         unsigned long subj_len, char issuer[],
                                         unsigned long issuer_len)
{
    if (subj == NULL || issuer == NULL || cert == NULL) {
        OM_LOG_ERROR("invalid inputs");
        return;
    }

    char tmpBuf[MAX_LINE_LENGTH + 1] = {0};

    /* read subject */
    X509_NAME_oneline(X509_get_subject_name(cert), tmpBuf, MAX_LINE_LENGTH);
    if (transfer_format(tmpBuf, subj, subj_len) != VOS_OK) {
        OM_LOG_WARN("transfer_format failed to subj,len:%lu", subj_len);
    }

    /* read issuer */
    X509_NAME_oneline(X509_get_issuer_name(cert), tmpBuf, MAX_LINE_LENGTH);
    if (transfer_format(tmpBuf, issuer, issuer_len) != VOS_OK) {
        OM_LOG_WARN("transfer_format failed to issuer,len:%lu", issuer_len);
    }

    return;
}

/* ****************************************************************************
 功能描述  : 读取CA证书链数量，超过10个时则认为数量过多
**************************************************************************** */
static int get_cert_chain_nums(const char *cert_file, int *chain_nums)
{
    FILE *fp = safety_fopen(cert_file, "r");
    if (fp == NULL) {
        OM_LOG_ERROR("open cert file pointer is NULL.");
        return 0;
    }

    STACK_OF(X509_INFO) *cert_stack = PEM_X509_INFO_read(fp, NULL, NULL, NULL);
    if (cert_stack == NULL) {
        OM_LOG_ERROR("read x509 cert info is NULL.");
        (void)fclose(fp);
        return 0;
    }

    *chain_nums = sk_X509_INFO_num(cert_stack);
    OM_LOG_INFO("get CA certificate chain nums is [%d]", *chain_nums);
    sk_X509_INFO_pop_free(cert_stack, X509_INFO_free);
    (void)fclose(fp);
    return 1;
}

/* ****************************************************************************
 功能描述  : 获取证书信息
**************************************************************************** */
signed long cert_getcertinfo(const char *certinfo_file, char subj[], unsigned long subj_len, char issuer[],
    unsigned long issuer_len, char notbegin[], int notbegin_len, char notafter[], int notafter_len,
    char serialnum[], unsigned long serialnum_len)
{
    if (subj == NULL || issuer == NULL || notbegin == NULL || notafter == NULL || serialnum == NULL
        || certinfo_file == NULL) {
        return VOS_ERR;
    }

    X509 *cert = cert_readcertfile(certinfo_file);
    if (cert == NULL) {
        return VOS_ERR;
    }

    cert_read_subject_and_issuer(cert, subj, subj_len, issuer, issuer_len);

    /* read serial num */
    ASN1_INTEGER *serial = X509_get_serialNumber(cert);
    BIGNUM *bn = ASN1_INTEGER_to_BN(serial, NULL);
    if (bn == NULL) {
        OM_LOG_ERROR("unable to convert ASN1INTEGER to BN");
        X509_free(cert);
        return VOS_ERR;
    }

    char *tmp = BN_bn2dec(bn);
    if (tmp == NULL) {
        OM_LOG_ERROR("unable to convert BN to decimal string.");
        BN_free(bn);
        X509_free(cert);
        return VOS_ERR;
    }

    if (strlen(tmp) >= serialnum_len) {
        OM_LOG_ERROR("buffer length shorter than serial number");
        BN_free(bn);
        OPENSSL_free(tmp);
        X509_free(cert);
        return VOS_ERR;
    }

    int ret = strncpy_s(serialnum, serialnum_len, tmp, strlen(tmp));
    if (ret != 0) {
        ret = VOS_ERR;
    } else {
        /* read not after time */
        ASN1_TIME *not_after = X509_get_notAfter(cert);
        (void)convert_ASN1TIME(not_after, notafter, notafter_len);
        OM_LOG_INFO("notafter: %s", notafter);

        /* read not before time */
        ASN1_TIME *not_before = X509_get_notBefore(cert);
        (void)convert_ASN1TIME(not_before, notbegin, notbegin_len);
        OM_LOG_INFO("notbegin: %s", notbegin);
        ret = VOS_OK;
    }

    BN_free(bn);
    OPENSSL_free(tmp);
    X509_free(cert);
    return ret;
}


/* ****************************************************************************
 功能描述  : 读取CA证书扩展信息
**************************************************************************** */
static void get_cert_extension(const char *certinfo_file, const X509 *cert, int *kusage, int *ca, int *chain_nums)
{
    // 获取key_cert_sign扩展字段
    int crit = 0;
    *kusage = 0;
    *ca = 0;
    *chain_nums = MAX_CHAIN_NUMS + 1;
    ASN1_BIT_STRING *usage = X509_get_ext_d2i(cert, NID_key_usage, &crit, NULL);
    if (usage != NULL && usage->length > 0) {
        // 解码扩展字段的值
        *kusage = usage->data[0] & 0x04;
        OM_LOG_INFO("get keyCertSign(%d)", *kusage);
        ASN1_BIT_STRING_free(usage);
    }
    // 获取ca扩展字段
    BASIC_CONSTRAINTS *bc = X509_get_ext_d2i(cert, NID_basic_constraints, &crit, NULL);
    if (bc != NULL) {
        // 判断是否可以用于签发其他证书
        *ca = bc->ca;
        OM_LOG_INFO("CA(%d), length(%d).", bc->ca, ASN1_INTEGER_get(bc->pathlen));
        // 该证书为根CA证书时，需要获取其证书链数量
        if (*ca) {
            if (get_cert_chain_nums(certinfo_file, chain_nums) == 0) {
                OM_LOG_ERROR("get cert chain nums failed.");
            }
        }
        // 释放内存
        BASIC_CONSTRAINTS_free(bc);
    }

    return;
}

/* ****************************************************************************
 功能描述  : 获取证书扩展信息
**************************************************************************** */
signed long get_extend_certinfo(const char *certinfo_file, char signature_algorithm[],
                                unsigned long signature_algorithm_len, char fingerprint[],
                                int fingerprint_len, int *pubkey_len, int *pubkey_type,
                                int *cert_version, int *kusage, int *ca, int *chain_nums)
{
    int ret = VOS_OK;
    if (certinfo_file == NULL || signature_algorithm == NULL || pubkey_len == NULL ||
        pubkey_type == NULL || cert_version == NULL || kusage == NULL || ca == NULL
        || chain_nums == NULL) {
        OM_LOG_ERROR("check parameter failed.");
        return VOS_ERR;
    }

    X509 *cert = cert_readcertfile(certinfo_file);
    if (cert == NULL) {
        return VOS_ERR;
    }

    get_cert_extension(certinfo_file, cert, kusage, ca, chain_nums);
    const X509_ALGOR *algor = X509_get0_tbs_sigalg(cert);
    int nid = OBJ_obj2nid(algor->algorithm);
    if (strcpy_s(signature_algorithm, signature_algorithm_len, OBJ_nid2ln(nid)) != 0) {
        OM_LOG_ERROR("strcpy_s signature failed.");
        ret = VOS_ERR;
    }

    EVP_PKEY *pubKey = X509_get_pubkey(cert);
    *pubkey_len = EVP_PKEY_bits(pubKey);
    *pubkey_type = EVP_PKEY_id(pubKey);
    *cert_version =  X509_get_version(cert);

    // 计算摘要
    unsigned char md[EVP_MAX_MD_SIZE];
    unsigned int md_len;
    X509_digest(cert, EVP_sha256(), md, &md_len);
    for (int i = 0; i < md_len && i < EVP_MAX_MD_SIZE; i++) {
        char tmp[EVP_MAX_MD_SIZE] = {0};
        if (sprintf_s(tmp, EVP_MAX_MD_SIZE - 1, "%02X:", md[i]) < 0) {
            OM_LOG_ERROR("sprintf_s md failed.");
            ret = VOS_ERR;
            break;
        }
        if (strcat_s(fingerprint, fingerprint_len, tmp) != EOK) {
            OM_LOG_ERROR("strcat_s fingerprint error!");
            ret = VOS_ERR;
            break;
        }
    }
    EVP_PKEY_free(pubKey);
    X509_free(cert);
    return ret;
}

/* ****************************************************************************
 功能描述  : 获取吊销列表基本信息
**************************************************************************** */
signed long get_crl_info(const char *crl_file,  int* issuer_len, char issuer_obj_name[], char issuer_value_name[],
                         int issuer_value_length[], unsigned long row_len, char last_update_time[],
                         int last_update_len, char next_update_time[], unsigned long next_update_len)
{
    if (crl_file == NULL || issuer_len == NULL || issuer_obj_name == NULL || issuer_value_name == NULL ||
        issuer_value_length == NULL || last_update_time == NULL || next_update_time == NULL) {
        OM_LOG_ERROR("check parameter failed.");
        return VOS_ERR;
    }

    X509_CRL *crl = read_crl_file(crl_file);
    if (crl == NULL) {
        OM_LOG_ERROR("read_crl_file failed.");
        return VOS_ERR;
    }

    X509_NAME *issuer = X509_CRL_get_issuer(crl);
    *issuer_len = X509_NAME_entry_count(issuer);
    for (int i = 0; i < X509_NAME_entry_count(issuer) && i < MAX_ENTRY_COUNT; i++) {
        X509_NAME_ENTRY *entry = X509_NAME_get_entry(issuer, i);
        ASN1_STRING *value = X509_NAME_ENTRY_get_data(entry);
        const char *obj_name = OBJ_nid2sn(OBJ_obj2nid(X509_NAME_ENTRY_get_object(entry)));
        issuer_value_length[i] = ASN1_STRING_length(value);
        if (strcat_s(issuer_obj_name, row_len, obj_name) != 0) {
            OM_LOG_ERROR("copy obj_name data faild!");
            X509_CRL_free(crl);
            return VOS_ERR;
        }
        if (strcat_s(issuer_obj_name, row_len, ":") != 0) {
            OM_LOG_ERROR("copy obj_name split faild!");
            X509_CRL_free(crl);
            return VOS_ERR;
        }
        if (strcat_s(issuer_value_name, row_len, (const char *)ASN1_STRING_get0_data(value)) != 0) {
            OM_LOG_ERROR("copy issuer value data faild!");
            X509_CRL_free(crl);
            return VOS_ERR;
        }
        if (strcat_s(issuer_value_name, row_len, ":") != 0) {
            OM_LOG_ERROR("copy issuer value split faild!");
            X509_CRL_free(crl);
            return VOS_ERR;
        }
    }
    ASN1_TIME *last_update = (ASN1_TIME*)X509_CRL_get0_lastUpdate(crl);
    ASN1_TIME *next_update = (ASN1_TIME*)X509_CRL_get0_nextUpdate(crl);

    (void)convert_ASN1TIME(last_update, last_update_time, last_update_len);
    (void)convert_ASN1TIME(next_update, next_update_time, next_update_len);

    X509_CRL_free(crl);
    return VOS_OK;
}

static int check_cert_is_ca(X509 *cert)
{
    if (cert == NULL) {
        OM_LOG_ERROR("input parameter is NULL");
        return 0;
    }

    BASIC_CONSTRAINTS *bc = X509_get_ext_d2i(cert, NID_basic_constraints, NULL, NULL);
    if (bc == NULL) {
        OM_LOG_ERROR("getting basic constraints failed");
        return 0;
    }

    int ca = bc->ca;
    BASIC_CONSTRAINTS_free(bc);
    return ca;
}

/* ****************************************************************************
 功能描述  : 只用于校验根CA证书的签名信息是否正确
******************************************************************************/
int verify_ca_cert_sign(const char *ca_file)
{
    if (ca_file == NULL) {
        OM_LOG_ERROR("invalid input ca file path");
        return VOS_ERR;
    }

    X509 *cert = cert_readcertfile(ca_file);
    if (cert == NULL) {
        OM_LOG_ERROR("read cert content from file failed");
        return VOS_ERR;
    }

    if (check_cert_is_ca(cert) == 0) {
        OM_LOG_ERROR("input cert is not a ca cert");
        X509_free(cert);
        return VOS_ERR;
    }

    EVP_PKEY *pubkey = X509_get_pubkey(cert);
    if (pubkey == NULL) {
        OM_LOG_ERROR("X509_get_pubkey failed");
        X509_free(cert);
        return VOS_ERR;
    }

    int ret = X509_verify(cert, pubkey);
    if (ret != 1) {
        OM_LOG_ERROR("X509_verify failed");
        EVP_PKEY_free(pubkey);
        X509_free(cert);
        return VOS_ERR;
    }

    EVP_PKEY_free(pubkey);
    X509_free(cert);
    return VOS_OK;
}

/* ****************************************************************************
 功能描述  : 使用吊销列表校验cert
******************************************************************************/
signed long verify_cert_by_crl(const char *cert_file,  const char *crl_file)
{
    if (crl_file == NULL || cert_file == NULL) {
        OM_LOG_ERROR("check parameter failed.");
        return VOS_ERR;
    }
    OM_LOG_INFO("load cert path: %s, crl path: %s.", cert_file, crl_file);
    X509_CRL *crl = read_crl_file(crl_file);
    if (crl == NULL) {
        OM_LOG_ERROR("read_crl_file failed.");
        return VOS_ERR;
    }

    X509 *cert = cert_readcertfile(cert_file);
    if (cert == NULL) {
        OM_LOG_ERROR("readcertfile failed.");
        X509_CRL_free(crl);
        return VOS_ERR;
    }
    EVP_PKEY *pkey = X509_get_pubkey(cert);
    int ret = X509_CRL_verify(crl, pkey);
    if (ret != 1) {
        OM_LOG_WARN("Error verifying CRL signature.");
        EVP_PKEY_free(pkey);
        X509_CRL_free(crl);
        X509_free(cert);
        return VOS_ERR;
    }

    OM_LOG_INFO("certificate ok!");
    EVP_PKEY_free(pkey);
    X509_CRL_free(crl);
    X509_free(cert);
    return VOS_OK;
}

/* ****************************************************************
函 数 名  :  check_dict_data
返回值： -1 同一种类型数据
         0  不同类型数据
**************************************************************** */
static signed long check_dict_data(const char *buf)
{
    char dict_set_a[DICT_SET_LENGTH] = "`~!@#$%^&*()-_=+\\|[{}];:'\",<.>/? ";

    for (unsigned long i = 0; i < SECURE_LEN && *(buf + i) != '\0'; i++) {
        for (unsigned long k = 0; k < strlen(dict_set_a); k++) {
            if (*(buf + i) == dict_set_a[k]) {
                break;
            } else if (k == strlen(dict_set_a) - 1) {
                return 0;
            }
        }
    }
    return -1;
}

/* ****************************************************************
函 数 名  :  is_data_check_finish
返回值： -1 处理结束
         0 未结束
**************************************************************** */
static signed long is_data_check_finish(char begin, char end, const char *buf)
{
    unsigned long i;

    for (i = 0; i < SECURE_LEN && buf[i] != '\0'; i++) {
        if (buf[i] >= begin && buf[i] <= end) {
            if (i == SECURE_LEN - 1) {
                return -1;
            }
        } else {
            break;
        }
    }

    return 0;
}

/* ****************************************************************
函 数 名  :  ispasscontaincomb
返回值： -1 同一种类型数据
         0  不同类型数据
**************************************************************** */
static signed long ispasscontaincomb(const char *buf, size_t len)
{
    unsigned long j;
    char begin = 0;
    char end = 0;
    (void)len;

    if (buf == NULL) {
        return -1;
    }

    for (j = 0; j <= INNER_NUMBER_THREE; j++) {
        if (j == INNER_NUMBER_ZERO) {
            begin = 'a';
            end = 'z';
        } else if (j == INNER_NUMBER_ONE) {
            begin = 'A';
            end = 'Z';
        } else if (j == INNER_NUMBER_TWO) {
            begin = '0';
            end = '9';
        } else if (j == INNER_NUMBER_THREE) {
            return check_dict_data(buf);
        }

        if (is_data_check_finish(begin, end, buf) != 0) {
            return -1;
        }
    }
    return 0;
}

static unsigned long GenerateSingleRandNum(void)
{
    unsigned long ret = 0xffffffff;
    ssize_t randomLen = 0;
    ssize_t result;
    int fd = -1;
    fd = open("/dev/random", 0);
    if (fd < 0) {
        OM_LOG_ERROR("open file failed");
        return ret;
    }

    while (randomLen < sizeof(unsigned long)) {
        result = read(fd, ((char *)&ret) + randomLen, (sizeof(unsigned long) - randomLen));
        if (result < 0) {
            OM_LOG_ERROR("read failed");
            break;
        }
        randomLen += result;
    }

    close(fd);
    return ret;
}

static char GetSelectedChar(unsigned long secure_num)
{
    char dict_set[128] =
        "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`~!@#$%^&*()-_=+\\|[{}];:'\",<.>/? ";
    size_t dict_len = strlen(dict_set);
    if (dict_len == 0) {
        return VOS_ERR;
    }
    return dict_set[secure_num % dict_len];
}

static void remove_first_str(char *secure, size_t secure_len)
{
    size_t i;
    for (i = 1; i < secure_len; i++) {
        secure[i - 1] = secure[i];
    }
    secure[i - 1] = '\0';
}

/* ****************************************************************************
 功能描述  : 生成24个字符的密码
**************************************************************************** */
int generate_pwd(char *pwd_buf, size_t buf_len)
{
    char secure_p1[SECURE_LEN * 2] = {0};
    char secure_p2[SECURE_LEN * 2] = {0};
    unsigned long secure_num;
    unsigned long tryCount = 0;
    unsigned long invalidRandomValue = 0xffffffff;
    int ret = 0;

    while (strlen(secure_p1) < SECURE_LEN) {
        secure_num = GenerateSingleRandNum();
        if (secure_num == invalidRandomValue) {
            tryCount++;
            if (tryCount >= SECURE_LEN) {
                OM_LOG_ERROR("Generate single random num failed, retry count up to limit");
                goto err;
            }
            OM_LOG_WARN("Generate single random num failed, retry continue");
            continue;
        }
        char select_char = GetSelectedChar(secure_num);
        ret = sprintf_s(secure_p2, sizeof(secure_p2), "%s%c", secure_p1, select_char);
        if (ret < 0) {
            goto err;
        }
        size_t pwd_2_len = strlen(secure_p2);
        if ((pwd_2_len == SECURE_LEN) && (ispasscontaincomb(secure_p2, pwd_2_len) != 0)) {
            // 如果口令字符满足了长度的要求，但是只有一种字符，则去掉最前面的字符，继续生成
            remove_first_str(secure_p2, pwd_2_len);
        }

        ret = strcpy_s(secure_p1, sizeof(secure_p1), secure_p2);
        if (ret != 0) {
            goto err;
        }
    }

    ret = strcpy_s(pwd_buf, buf_len, secure_p1);
    if (ret != 0) {
        goto err;
    }
    if (memset_s(secure_p1, sizeof(secure_p1), 0, sizeof(secure_p1)) != 0 ||
        memset_s(secure_p2, sizeof(secure_p2), 0, sizeof(secure_p2)) != 0) {
        OM_LOG_ERROR("memset_s error");
        return VOS_ERR;
    }
    return VOS_OK;

 err:
    if (memset_s(secure_p1, sizeof(secure_p1), 0, sizeof(secure_p1)) != 0 ||
        memset_s(secure_p2, sizeof(secure_p2), 0, sizeof(secure_p2)) != 0) {
        OM_LOG_ERROR("memset_s error");
        ret = VOS_ERR;
    }
    return ret;
}

int replace_server_kmc_psd(const char *path, char *pwd_buf, int pwd_buf_len, const char *alg_json_file)
{
    char pwd_buf_enc[MAX_ENC_LEN] = {0};

    /* 加密保存口令 */
    size_t plainLen = strlen(pwd_buf);
    size_t cipherLen = sizeof(pwd_buf_enc);
    int ret = SecurityEncrypt(pwd_buf, (unsigned int)plainLen, pwd_buf_enc, (unsigned int *)&cipherLen, alg_json_file);
    if (ret != 0) { // 如果加密失败
        OM_LOG_ERROR("encrypt psd error");
        return VOS_ERR;
    }

    if (memset_s(pwd_buf, pwd_buf_len, 0, pwd_buf_len) != 0) {
        OM_LOG_ERROR("memset_s error");
    }

    /* 替换口令 */
    FILE *fp = safety_fopen(path, "wb");
    if (fp == NULL) {
        OM_LOG_ERROR("creat cert file error");
        if (memset_s(pwd_buf_enc, sizeof(pwd_buf_enc), 0, sizeof(pwd_buf_enc)) != EOK) {
            OM_LOG_ERROR("memset_s error");
        }
        return VOS_ERR;
    }

    if (safety_chmod_by_fd(fp, USER_READ_ONLY_PERMISSION) != COMMON_SUCCESS) {
        OM_LOG_ERROR("chmod failed");
    }

    size_t write_ret = fwrite(pwd_buf_enc, strlen(pwd_buf_enc), 1, fp);
    (void)fclose(fp);

    if (memset_s(pwd_buf_enc, sizeof(pwd_buf_enc), 0, sizeof(pwd_buf_enc)) != EOK) {
        OM_LOG_ERROR("memset_s error.");
    }

    if (write_ret != 1) {
        OM_LOG_ERROR("write pwd_buf_enc error");
        return VOS_ERR;
    }

    return VOS_OK;
}

int StrToInt(char *buf, int buf_size)
{
    char *end = NULL;

    if ((buf == NULL) || (buf_size <= 0)) {
        OM_LOG_ERROR("invalid inputs");
        return 0;
    }

    buf[buf_size - 1] = '\0';
    return (int)strtol(buf, &end, BASE_DECIMAL_SYSTEM);
}
