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
 * 功能描述   : CMS验签和证书吊销比较
 */
#ifndef CMSCBB_USE_CALLBACK_PLT
#define CMSCBB_USE_CALLBACK_PLT
#ifdef _MSC_VER
#include <stdlib.h>
#include <string.h>
#else
#include <securec.h>
#endif /* _MSC_VER */
#include "openssl/evp.h"
#include "openssl/err.h"
#include "openssl/bn.h"
#include "openssl/rsa.h"
#include "openssl/core_names.h"
#include "openssl/param_build.h"
#include "openssl/ec.h"

#include "cmscbb_config.h"
#include "cmscbb_err_def.h"
#include "cmscbb_types.h"
#include "cmscbb_sdk.h"


#ifndef CVB_ERROR
#define CVB_ERROR (CMSCBB_ERROR_CODE)(-1)
#endif

static CVB_VOID CvbOpensslUninit(void);

CMSCBB_ERROR_CODE CmscbbMalloc(CVB_VOID** ppByte, CVB_SIZE_T size)
{
    if (ppByte == CVB_NULL || size == 0) {
        return CMSCBB_ERR_CONTEXT_INVALID_PARAM;
    }
    *ppByte = CVB_NULL;

    void* pByte = malloc(size);
    if (pByte == CVB_NULL) {
        return CMSCBB_ERR_SYS_MEM_ALLOC;
    }

    *ppByte = pByte;
    return CVB_SUCCESS;
}

CVB_VOID CmscbbFree(CVB_VOID* ptr)
{
    if (ptr == CVB_NULL) {
        return;
    }

    free(ptr);
    ptr = CVB_NULL;
}

CVB_INT CmscbbMemCmp(const CVB_VOID* s1, const CVB_VOID* s2, CVB_SIZE_T n)
{
    if (s1 == NULL || s2 == NULL) {
        return -1;
    }

    return (CVB_INT)memcmp(s1, s2, n);
}

CVB_INT CmscbbStrNCmp(const CVB_CHAR* s1, const CVB_CHAR* s2, CVB_SIZE_T n)
{
    if (s1 == NULL || s2 == NULL) {
        return -1;
    }

    return (CVB_INT)strncmp((const char*)s1, (const char*)s2, n);
}

const CVB_CHAR* CmscbbStrStr(const CVB_CHAR* haystack, const CVB_CHAR* needle)
{
    if (haystack == NULL || needle == NULL) {
        return NULL;
    }

    return (const CVB_CHAR*)strstr((const char*)haystack, (const char*)needle);
}

CVB_CHAR* CmscbbStrChr(const CVB_CHAR* s, CVB_CHAR c)
{
    if (s == NULL) {
        return NULL;
    }

    return (CVB_CHAR*)strchr((char*)s, (int)c);
}

CVB_UINT32 CmscbbStrlen(const CVB_CHAR* s)
{
    if (s == NULL) {
        return 0;
    }

    return (CVB_UINT32)strlen((const char*)s);
}

CVB_INT CmscbbStrCmp(const CVB_CHAR* s1, const CVB_CHAR* s2)
{
    if (s1 == NULL || s2 == NULL) {
        return -1;
    }
    if (s1 == s2) {
        return 0;
    }
    int len1 = (int)strlen((const char*)s1);
    int len2 = (int)strlen((const char*)s2);
    if (len1 != len2) {
        return -1;
    }
    return strncmp((const char*)s1, (const char*)s2, (size_t)len1);
}

#if defined(CMSCBB_SUPPORT_FILE) && CMSCBB_SUPPORT_FILE != 0
CVB_FILE_HANDLE CmscbbFileOpen(const CVB_CHAR* path, const CVB_CHAR* mode)
{
#ifdef _MSC_VER
    FILE* pfile = NULL;

    if (fopen_s(&pfile, (const char*)path, (const char*)mode) != 0)  {
        return NULL;
    }

#else
    if (strlen(path) > PATH_MAX - 1) {
        return NULL;
    }

    char cvbPath[PATH_MAX] = {0x00};
    if (realpath(path, cvbPath) == NULL) {
        return NULL;
    }
    FILE* pfile = fopen((const char*)cvbPath, (const char*)mode);
#endif

    return (CVB_FILE_HANDLE)pfile;
}

CVB_SIZE_T CmscbbFileRead(CVB_VOID* ptr, CVB_SIZE_T size, CVB_FILE_HANDLE hFile)
{
    if (ptr == NULL || size == 0) {
        return 0;
    }
    FILE* pfile = (FILE*)hFile;
#ifdef _MSC_VER
    return fread_s(ptr, size, 1, size, pfile);
#else
    return fread(ptr, 1, size, pfile);
#endif
}

CMSCBB_ERROR_CODE CmscbbFileClose(CVB_FILE_HANDLE hFile)
{
    FILE* pf = (FILE*)hFile;
    if (pf == CVB_NULL) {
        return CMSCBB_ERR_UNDEFINED;
    }

    return (CMSCBB_ERROR_CODE)fclose(pf);
}

CVB_UINT64 CmscbbFileGetSize(CVB_FILE_HANDLE hFile)
{
    FILE* pf = (FILE*)hFile;
    if (pf == CVB_NULL) {
        return 0;
    }

    long cur0 = (long)ftell(pf);
    if (cur0 < 0) {
        return 0;
    }

    (void)fseek(pf, 0, SEEK_END);
    long nFileContent = (long)ftell(pf);
    (void)fseek(pf, cur0, SEEK_SET);
    return (CVB_UINT64)(long long)nFileContent;
}
#endif

#if defined(CMSCBB_ENABLE_LOG) && CMSCBB_ENABLE_LOG != 0

CVB_VOID CmscbbLogPrint(CMSCBB_LOG_TYPE log_level, const CVB_CHAR* filename, CVB_INT line, const CVB_CHAR* function,
    CMSCBB_ERROR_CODE rc, const CVB_CHAR* msg)
{
    switch (log_level) {
        case CMSCBB_LOG_TYPE_ERROR: {
            (CVB_VOID)printf("[ERROR]");
            break;
        }
        case CMSCBB_LOG_TYPE_WARNING: {
            (CVB_VOID)printf("[WARNING]");
            break;
        }
        default:
            break;
    }
    (CVB_VOID)printf("%s(%d):(%s):%s", (const char*)filename, line, (const char*)function, (const char*)msg);
    if (rc != CVB_SUCCESS) {
        (CVB_VOID)printf("(err:%x)", (unsigned int)rc);
    }

    (CVB_VOID)printf("\r\n");
}
#endif

static CVB_VOID CvbOpensslUninit(void)
{
    EVP_cleanup();
    CRYPTO_cleanup_all_ex_data();
    ERR_free_strings();
}

typedef struct CryptoMdSt {
    void* hashHandler;
} CryptoMd;

CMSCBB_ERROR_CODE CmscbbMdCreateCtx(CMSCBB_CRYPTO_MD_CTX* mdCtx)
{
    CMSCBB_ERROR_CODE ret;
    CryptoMd* md = CVB_NULL;
    ret = CmscbbMalloc((void**)&md, sizeof(CryptoMd));
    if (ret != CVB_SUCCESS) {
        return ret;
    }

    errno_t state = memset_s(md, sizeof(CryptoMd), 0, sizeof(CryptoMd));
    if (state != CVB_SUCCESS) {
        CmscbbFree(md);
        return CMSCBB_ERR_SYS_MEM_SET;
    }

    *mdCtx = (CMSCBB_CRYPTO_MD_CTX)md;
    return ret;
}

static const EVP_MD* CvbGetOpensslEvpmd(CVB_UINT32 hashId)
{
    const EVP_MD* handler = NULL;
    switch (hashId) {
        case CMSCBB_HASH_SHA256:
            handler = EVP_sha256();
            break;
        case CMSCBB_HASH_SHA384:
            handler = EVP_sha384();
            break;
        case CMSCBB_HASH_SHA512:
            handler = EVP_sha512();
            break;
        case CMSCBB_HASH_SM3:
            handler = EVP_sm3();
            break;
        default:
            break;
    }
    return handler;
}

CMSCBB_ERROR_CODE CmscbbMdInit(CMSCBB_CRYPTO_MD_CTX mdCtx, CVB_UINT32 hashId)
{
    CMSCBB_ERROR_CODE ret = 0;
    const EVP_MD* evpMd = NULL;
    void *handle = NULL;
    CryptoMd* md = (CryptoMd*)mdCtx;
    if (md == CVB_NULL) {
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }
    handle = EVP_MD_CTX_create();
    if (handle == NULL) {
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }
    if (EVP_MD_CTX_init(handle) != 1) {
        EVP_MD_CTX_destroy(handle);
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }
    evpMd = CvbGetOpensslEvpmd(hashId);
    if (evpMd == NULL) {
        EVP_MD_CTX_destroy(handle);
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }
    if (EVP_DigestInit_ex(handle, evpMd, NULL) != 1) {
        EVP_MD_CTX_destroy(handle);
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }
    md->hashHandler = handle;
    return ret;
}

CMSCBB_ERROR_CODE CmscbbMdUpdate(CMSCBB_CRYPTO_MD_CTX mdCtx, const CVB_BYTE* data, CVB_UINT32 len)
{
    CryptoMd *md = (CryptoMd*)mdCtx;
    if (md == CVB_NULL || md->hashHandler == CVB_NULL) {
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_UPDATE;
    }
    if (data == CVB_NULL || len == 0) {
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_UPDATE;
    }
    if (EVP_DigestUpdate(md->hashHandler, (const void*)data, len) == 0) {
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_UPDATE;
    }
    return CVB_SUCCESS;
}

CMSCBB_ERROR_CODE CmscbbMdFinal(CMSCBB_CRYPTO_MD_CTX mdCtx, CVB_BYTE* digest,
                                CVB_UINT32* len, const CVB_UINT32* digestMaxLen)
{
    CryptoMd *md = (CryptoMd*)mdCtx;

    if (md == CVB_NULL || digest == CVB_NULL || len == CVB_NULL || digestMaxLen == CVB_NULL) {
        return CMSCBB_ERR_CONTEXT_INVALID_PARAM;
    }

    if (EVP_DigestFinal(md->hashHandler, digest, (unsigned int*)len) == 0) {
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_FINAL;
    }
    return CVB_SUCCESS;
}

CVB_VOID CmscbbMdDestoryCtx(CMSCBB_CRYPTO_MD_CTX mdCtx)
{
    CryptoMd *md = (CryptoMd*)mdCtx;
    if (md == CVB_NULL) {
        return;
    }
    if (md->hashHandler != CVB_NULL) {
        (void)EVP_MD_CTX_reset(md->hashHandler);
        EVP_MD_CTX_destroy(md->hashHandler);
    }
    CmscbbFree(md);
    CvbOpensslUninit();
}

typedef struct CryptoVrfSt {
    void* mdCtx;
    EVP_PKEY* pub;
    CVB_UINT32 encAlg;
    EVP_PKEY_CTX* pctx;
} CryptoVrf;

CMSCBB_ERROR_CODE CmscbbCryptoVerifyCreateCtx(CMSCBB_CRYPTO_VRF_CTX* ctx)
{
    CMSCBB_ERROR_CODE ret;
    CryptoVrf* vrf = CVB_NULL;
    if (ctx == CVB_NULL) {
        return CMSCBB_ERR_CONTEXT_INVALID_PARAM;
    }
    ret = CmscbbMalloc((CVB_VOID**)&vrf, sizeof(CryptoVrf));
    if (ret != CVB_SUCCESS) {
        return CMSCBB_ERR_SYS_MEM_ALLOC;
    }
    errno_t state = memset_s((void*)vrf, sizeof(CryptoVrf), 0, sizeof(CryptoVrf));
    if (state != CVB_SUCCESS) {
        CmscbbFree(vrf);
        return CMSCBB_ERR_SYS_MEM_SET;
    }
    *ctx = (CMSCBB_CRYPTO_VRF_CTX)vrf;
    return ret;
}

static OSSL_PARAM_BLD *CvbObtainRsaBld(CmscbbBigInt* nva, CmscbbBigInt* eva, BIGNUM **resN, BIGNUM **resE)
{
    CVB_INT32 ret;
    BIGNUM *n = BN_bin2bn(nva->aVal, (CVB_INT32)nva->uiLength, NULL);
    BIGNUM *e = BN_bin2bn(eva->aVal, (CVB_INT32)eva->uiLength, NULL);
    OSSL_PARAM_BLD *bld = OSSL_PARAM_BLD_new();
    if (bld == NULL) {
        printf("OSSL_PARAM_BLD_new failed.");
        return NULL;
    }
    if (n != NULL && e != NULL) {
        ret = OSSL_PARAM_BLD_push_BN(bld, OSSL_PKEY_PARAM_RSA_N, n);
        if (ret <= 0) {
            printf("OSSL_PARAM_BLD_push_BN pub key n is null(%d).", ret);
            OSSL_PARAM_BLD_free(bld);
            return NULL;
        }
        ret = OSSL_PARAM_BLD_push_BN(bld, OSSL_PKEY_PARAM_RSA_E, e);
        if (ret <= 0) {
            printf("OSSL_PARAM_BLD_push_BN pub key e is null(%d).", ret);
            OSSL_PARAM_BLD_free(bld);
            return NULL;
        }
    } else {
        OSSL_PARAM_BLD_free(bld);
        return NULL;
    }
    *resN = n;
    *resE = e;
    return bld;
}

static OSSL_PARAM_BLD *InObtainRsaBld(CmscbbKeyAndAlgInfo* info, BIGNUM **n, BIGNUM **e)
{
    return CvbObtainRsaBld(info->n, info->e, n, e);
}


static const char *InGetAlgName(CVB_UINT32 alg)
{
    switch (alg) {
        case CMSCBB_ENC_RSA_PSS:
            return "RSA-PSS";
        default:
            return "RSA-PSS";
    }
}

static EVP_PKEY *InGetEvpkeyObj(CmscbbKeyAndAlgInfo* info)
{
    EVP_PKEY_CTX *ctx = NULL;
    EVP_PKEY *evpKey = NULL;
    OSSL_PARAM_BLD *bld = NULL;
    OSSL_PARAM *params = NULL;
    BIGNUM *numN = NULL;
    BIGNUM *numE = NULL;
    ctx = EVP_PKEY_CTX_new_from_name(NULL, InGetAlgName(info->encAlg), NULL);
    if (ctx == NULL) {
        printf("EVP_PKEY_CTX_new_from_name failed.");
        return NULL;
    }
    do {
        if (info->encAlg == CMSCBB_ENC_RSA || info->encAlg == CMSCBB_ENC_RSA_PSS) {
            bld = InObtainRsaBld(info, &numN, &numE);
        }

        if (bld == NULL) {
            break;
        }
        if (EVP_PKEY_fromdata_init(ctx) <= 0) {
            break;
        }
        params = OSSL_PARAM_BLD_to_param(bld);
        if (params == NULL) {
            break;
        }
        if (EVP_PKEY_fromdata(ctx, &evpKey, EVP_PKEY_PUBLIC_KEY, params) <= 0) {
            break;
        }
    } while (0);
    BN_free(numN);
    BN_free(numE);
    OSSL_PARAM_free(params);
    OSSL_PARAM_BLD_free(bld);
    EVP_PKEY_CTX_free(ctx);
    return evpKey;
}

/* support RSA pkcs1_v1.5 and RSA PSS */
CMSCBB_ERROR_CODE CmscbbCryptoVerifyInit(CMSCBB_CRYPTO_VRF_CTX vrfCtx, CmscbbKeyAndAlgInfo* info)
{
    CryptoVrf* vrf = (CryptoVrf*)vrfCtx;
    const EVP_MD* mdHandler = NULL;
    EVP_PKEY* pubKey = NULL;
    EVP_PKEY_CTX *pctx = NULL;

    if (vrf == CVB_NULL || info == CVB_NULL) {
        return CMSCBB_ERR_CONTEXT_INVALID_PARAM;
    }

    pubKey = InGetEvpkeyObj(info);
    if (pubKey == NULL) {
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }
    mdHandler = CvbGetOpensslEvpmd(info->hashAlg);
    if (mdHandler == NULL) {
        EVP_PKEY_free(pubKey);
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }

    vrf->mdCtx = EVP_MD_CTX_create();
    if (vrf->mdCtx == NULL) {
        EVP_PKEY_free(pubKey);
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }

    if (EVP_DigestVerifyInit(vrf->mdCtx, NULL, mdHandler, NULL, pubKey) == 0) {
        if (pctx != CVB_NULL) {
            EVP_PKEY_CTX_free(pctx);
        }
        EVP_PKEY_free(pubKey);
        EVP_MD_CTX_free(vrf->mdCtx);
        vrf->mdCtx = NULL;
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_INIT;
    }
    vrf->pub = pubKey;
    vrf->encAlg = info->encAlg;
    return CVB_SUCCESS;
}

CMSCBB_ERROR_CODE CmscbbCryptoVerifyUpdate(CMSCBB_CRYPTO_VRF_CTX vrfCtx, const CVB_BYTE* data, CVB_UINT32 len)
{
    CryptoVrf* vrf = (CryptoVrf*)vrfCtx;
    if (vrf == CVB_NULL || vrf->mdCtx == CVB_NULL) {
        return CMSCBB_ERR_CONTEXT_INVALID_PARAM;
    }
    if (EVP_DigestVerifyUpdate(vrf->mdCtx, data, len) != 1) {
        return CMSCBB_ERR_PKI_CRYPTO_DIGEST_UPDATE;
    }
    return CVB_SUCCESS;
}

CMSCBB_ERROR_CODE CmscbbCryptoVerifyFinal(CMSCBB_CRYPTO_VRF_CTX vrfCtx, const CVB_BYTE* signature,
                                          CVB_UINT32 len, CVB_INT* result)
{
    int rc = 0;
    CryptoVrf* vrf = (CryptoVrf*)vrfCtx;
    if (vrf == CVB_NULL || vrf->mdCtx == CVB_NULL) {
        return CMSCBB_ERR_CONTEXT_INVALID_PARAM;
    }
    rc = EVP_DigestVerifyFinal(vrf->mdCtx, (const unsigned char*)signature, len);
    if (rc <= 0) {
        ERR_print_errors_fp(stdout);
    }
    *result = (rc == 1) ? 1 : 0;
    return CVB_SUCCESS;
}

CVB_VOID CmscbbCryptoVerifyDestroyCtx(CMSCBB_CRYPTO_VRF_CTX vrfCtx)
{
    CryptoVrf* vrf = (CryptoVrf*)vrfCtx;
    if (vrf == CVB_NULL) {
        return;
    }
    if (vrf->mdCtx != CVB_NULL) {
        (void)EVP_MD_CTX_reset(vrf->mdCtx);
        EVP_MD_CTX_destroy(vrf->mdCtx);
    }
    if (vrf->pctx != CVB_NULL) {
        EVP_PKEY_CTX_free(vrf->pctx);
    }
    EVP_PKEY_free(vrf->pub);
    CmscbbFree(vrf);
    CvbOpensslUninit();
    return;
}

#endif /* CMSCBB_USE_CALLBACK_PLT */