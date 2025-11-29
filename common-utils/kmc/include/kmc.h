/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
 * Description: kmc header file.
 * Create: 2022-05-20
 */

#ifndef __KMC_H__
#define __KMC_H__

#include <stddef.h>
#include <stdint.h>
#include <dlfcn.h>
#include <stdio.h>
#include <linux/limits.h>

static void *kmcextHandle = NULL;

#define KE_SUCCESS (0)
#define KE_FAILURE (1)
#define KE_ERROR_LIBRARY_NOT_FOUND  (997)
#define KE_ERROR_FUNCTION_NOT_FOUND  (998)
#define KE_ERROR_UNKNOWN  (999)
#define KMC_ALG_CNF_REVERSE (520)

typedef enum {
    LOG_DISABLE = 0,
    LOG_ERROR,
    LOG_WARN,
    LOG_INFO,
    LOG_DEBUG,
    LOG_TRACE
} LogLevel;

#define SEC_PATH_MAX PATH_MAX

typedef struct TagKmcConfig {
    char primaryKeyStoreFile[SEC_PATH_MAX];
    char standbyKeyStoreFile[SEC_PATH_MAX];
    int domainCount;
    int role;
    int procLockPerm;
    int sdpAlgId;
    int hmacAlgId;
    int semKey;
    int innerSymmAlgId;
    int innerHashAlgId;
    int innerHmacAlgId;
    int innerKdfAlgId;
    int workKeyIter;
    int rootKeyIter;
} KmcConfig;

typedef struct TagKeCallBackParam {
    void *notifyCbCtx;
    void *loggerCtx;
    void *hwCtx;
} KeCallBackParam;

typedef struct TagKmcHardWareParm {
    int len;
    char *hardParam;
} KmcHardWareParm;

typedef struct TagKmcConfigEx {
    int enableHw;
    KmcHardWareParm kmcHardWareParm;
    KeCallBackParam *keCbParam;
    KmcConfig kmcConfig;
} KmcConfigEx;

#pragma pack(1)
typedef struct TagKmcAlgCnfParam {
    uint32_t symmAlg;
    uint32_t kdfAlg;
    uint32_t hmacAlg;
    uint32_t hashAlg;
    uint32_t workKeyIter;
    uint32_t saltLen;
    unsigned char reserve[16];
} KmcAlgCnfParam;
#pragma pack()

typedef struct tagKmcContext {
    unsigned char reserved[KMC_ALG_CNF_REVERSE];
    KmcAlgCnfParam algCnf;
} KmcContext;

typedef struct TagExtContext {
    int flag; /* flag this context is single or multi */
    unsigned int sdpAlgId;
    unsigned int hmacAlgId;
    KmcContext *kmcContext;
    void *extUserData;
    int isInit;
} ExtConText;


typedef void (*LoggerCallbackEx)(const void *logRef, LogLevel level, const char *msg);

int (*KeInitializeExFunc)(KmcConfigEx *kmcConfig, void **ctx);
int (*KeFinalizeExFunc)(const void **ctx);
int (*KeSecureEraseKeystoreExFunc)(const void *ctxExt);
int (*KeActiveNewKeyExFunc)(const void *ctx, unsigned int domainID);
int (*KeRegisterByteKeyExFunc)(const void *ctx, unsigned int domainID,
    unsigned int keyID, unsigned char *key, int keyLen);
void (*KeSetLoggerCallbackExFunc)(LoggerCallbackEx logFn);
void (*KeSetLoggerLevelFunc)(LogLevel level);
int (*KeDecryptByDomainExFunc)(const void *ctx, unsigned int domainID, const char *cipherText, int cipherTextLen,
    char **plainText, int *plainTextLen);
int (*KeEncryptByDomainExFunc)(const void *ctx, unsigned int domainID, const char *plainText, int plainTextLen,
    char **cipherText, int *cipherTextLen);
int (*KeGetCipherDataLenExFunc)(const void *ctx, int plainTextLen, int *cipherTextLen);
int (*KeGetKeyByIDExFunc)(const void *ctx, unsigned int domainID, unsigned int keyID, char **key,
    int *keyLen, int base64Ind);
int (*KeCheckAndUpdateMkExFunc)(const void* ctx, unsigned int domainID, int advanceDay);
int (*KeRemoveKeyByIDExFunc)(const void* ctx, unsigned int domainID, unsigned int keyID);
int (*KeRefreshMkMaskExFunc)(const void *ctx);
int (*KeGetMaxMkIDExFunc)(const void *ctx, unsigned int domainID, unsigned int *maxKeyID);
int (*KeSetMkStatusExFunc)(const void *ctx, unsigned int domainId, unsigned int keyId, unsigned char status);
int (*UpdateRootKeyFunc)(const void *ctx);
int (*KeHmacByDomainV2ExFunc)(const void *ctx, unsigned int domainID, const char *plainText, int plainTextLen,
    char **hmacData, int *hmacDateLen);
int (*KeHmacVerifyByDomainExFunc)(const void *ctx, unsigned int domainID, const char *plainText, int plainTextLen,
    const char *hmacData, int hmacDateLen);


#define RETURN_IF_NULL(FUNC_NAME) \
    if (kmcextHandle == NULL || (FUNC_NAME) == NULL) { \
        return KE_FAILURE; \
    }

static int KeInitializeEx(KmcConfigEx *kmcConfig, void **ctx, uint32_t saltLen)
{
    RETURN_IF_NULL(KeInitializeExFunc);
    int ret = KeInitializeExFunc(kmcConfig, ctx);
    if (ret != 0) {
        return ret;
    }
    if (ctx == NULL) {
        return -1;
    }
    ExtConText *extCtx = (ExtConText*)(*ctx);
    if (extCtx == NULL || extCtx->kmcContext == NULL) {
        return -1;
    }
    extCtx->kmcContext->algCnf.saltLen = saltLen;
    return ret;
}

static int KeFinalizeEx(const void **ctx)
{
    RETURN_IF_NULL(KeFinalizeExFunc);
    return KeFinalizeExFunc(ctx);
}

static int KeSecureEraseKeystoreEx(const void *ctxExt)
{
    RETURN_IF_NULL(KeSecureEraseKeystoreExFunc);
    return KeSecureEraseKeystoreExFunc(ctxExt);
}

static int KeActiveNewKeyEx(const void *ctx, unsigned int domainID)
{
    RETURN_IF_NULL(KeActiveNewKeyExFunc);
    return KeActiveNewKeyExFunc(ctx, domainID);
}

static int KeRegisterByteKeyEx(const void *ctx, unsigned int domainID,
    unsigned int keyID, unsigned char *key, int keyLen)
{
    RETURN_IF_NULL(KeRegisterByteKeyExFunc);
    return KeRegisterByteKeyExFunc(ctx, domainID, keyID, key, keyLen);
}

static int KeSetLoggerLevel(LogLevel level)
{
    RETURN_IF_NULL(KeSetLoggerLevelFunc);
    KeSetLoggerLevelFunc(level);
    return KE_SUCCESS;
}

static int KeSetLoggerCallbackEx(LoggerCallbackEx logFn)
{
    RETURN_IF_NULL(KeSetLoggerCallbackExFunc);
    KeSetLoggerCallbackExFunc(logFn);
    return KE_SUCCESS;
}

static int KeDecryptByDomainEx(const void *ctx, unsigned int domainID,
    const char *cipherText, int cipherTextLen, char **plainText, int *plainTextLen)
{
    RETURN_IF_NULL(KeDecryptByDomainExFunc);
    return KeDecryptByDomainExFunc(ctx, domainID, cipherText, cipherTextLen, plainText, plainTextLen);
}

static int KeEncryptByDomainEx(const void *ctx, unsigned int domainID, const char *plainText, int plainTextLen,
    char **cipherText, int *cipherTextLen)
{
    RETURN_IF_NULL(KeEncryptByDomainExFunc);
    return KeEncryptByDomainExFunc(ctx, domainID, plainText, plainTextLen, cipherText, cipherTextLen);
}

static int KeGetCipherDataLenEx(const void *ctx, int plainTextLen, int *cipherTextLen)
{
    RETURN_IF_NULL(KeGetCipherDataLenExFunc);
    return KeGetCipherDataLenExFunc(ctx, plainTextLen, cipherTextLen);
}

static int KeGetKeyByIDEx(const void *ctx, unsigned int domainID, unsigned int keyID,
    char **key, int *keyLen, int base64Ind)
{
    RETURN_IF_NULL(KeGetKeyByIDExFunc);
    return KeGetKeyByIDExFunc(ctx, domainID, keyID, key, keyLen, base64Ind);
}

static int KeCheckAndUpdateMkEx(const void* ctx, unsigned int domainID, int advanceDay)
{
    RETURN_IF_NULL(KeCheckAndUpdateMkExFunc);
    return KeCheckAndUpdateMkExFunc(ctx, domainID, advanceDay);
}

static int KeRemoveKeyByIDEx(const void *ctx, unsigned int domainID, unsigned int keyID)
{
    RETURN_IF_NULL(KeRemoveKeyByIDExFunc);
    return KeRemoveKeyByIDExFunc(ctx, domainID, keyID);
}

static int KeRefreshMkMaskEx(const void *ctx)
{
    RETURN_IF_NULL(KeRefreshMkMaskExFunc);
    return KeRefreshMkMaskExFunc(ctx);
}

static int KeGetMaxMkIDEx(const void *ctx, unsigned int domainID, unsigned int *maxKeyID)
{
    RETURN_IF_NULL(KeGetMaxMkIDExFunc);
    return KeGetMaxMkIDExFunc(ctx, domainID, maxKeyID);
}

static int KeSetMkStatusEx(const void *ctx, unsigned int domainId, unsigned int keyId, unsigned char status)
{
    RETURN_IF_NULL(KeSetMkStatusExFunc);
    return KeSetMkStatusExFunc(ctx, domainId, keyId, status);
}

static int KeUpdateRootKeyEx(const void *ctx)
{
    RETURN_IF_NULL(UpdateRootKeyFunc);
    return UpdateRootKeyFunc(ctx);
}

static int KeHmacByDomainV2Ex(const void *ctx, unsigned int domainID, const char *plainText, int plainTextLen,
    char **hmacData, int *hmacDateLen)
{
    RETURN_IF_NULL(KeHmacByDomainV2ExFunc);
    return KeHmacByDomainV2ExFunc(ctx, domainID, plainText, plainTextLen, hmacData, hmacDateLen);
}

static int KeHmacVerifyByDomainEx(const void *ctx, unsigned int domainID, const char *plainText, int plainTextLen,
    const char *hmacData, int hmacDateLen)
{
    RETURN_IF_NULL(KeHmacVerifyByDomainExFunc);
    return KeHmacVerifyByDomainExFunc(ctx, domainID, plainText, plainTextLen, hmacData, hmacDateLen);
}

static int InitKMC(void)
{
    if (kmcextHandle != NULL) {
        return KE_SUCCESS;
    }
    kmcextHandle = dlopen("libkmcext.so", RTLD_LAZY);
    if (kmcextHandle == NULL) {
        fprintf(stderr, "[error] dlopen libkmcext.so - %s\n", dlerror());
        return KE_ERROR_LIBRARY_NOT_FOUND;
    }
    // Clear any existing error
    dlerror();

    KeInitializeExFunc          = dlsym(kmcextHandle, "KeInitializeEx");
    KeFinalizeExFunc            = dlsym(kmcextHandle, "KeFinalizeEx");
    KeSecureEraseKeystoreExFunc = dlsym(kmcextHandle, "KeSecureEraseKeystoreEx");
    KeDecryptByDomainExFunc     = dlsym(kmcextHandle, "KeDecryptByDomainEx");
    KeEncryptByDomainExFunc     = dlsym(kmcextHandle, "KeEncryptByDomainEx");
    KeActiveNewKeyExFunc        = dlsym(kmcextHandle, "KeActiveNewKeyEx");
    KeRegisterByteKeyExFunc     = dlsym(kmcextHandle, "KeRegisterByteKeyEx");
    KeGetCipherDataLenExFunc    = dlsym(kmcextHandle, "KeGetCipherDataLenEx");
    KeRefreshMkMaskExFunc       = dlsym(kmcextHandle, "KeRefreshMkMaskEx");
    KeGetKeyByIDExFunc          = dlsym(kmcextHandle, "KeGetKeyByIDEx");
    KeRemoveKeyByIDExFunc       = dlsym(kmcextHandle, "KeRemoveKeyByIDEx");
    KeSetLoggerLevelFunc        = dlsym(kmcextHandle, "KeSetLoggerLevel");
    KeSetLoggerCallbackExFunc     = dlsym(kmcextHandle, "KeSetLoggerCallbackEx");
    KeCheckAndUpdateMkExFunc    = dlsym(kmcextHandle, "KeCheckAndUpdateMkEx");
    KeGetMaxMkIDExFunc          = dlsym(kmcextHandle, "KeGetMaxMkIDEx");
    KeSetMkStatusExFunc         = dlsym(kmcextHandle, "KeSetMkStatusEx");
    UpdateRootKeyFunc           = dlsym(kmcextHandle, "KeUpdateRootKeyEx");
    KeHmacByDomainV2ExFunc      = dlsym(kmcextHandle, "KeHmacByDomainV2Ex");
    KeHmacVerifyByDomainExFunc    = dlsym(kmcextHandle, "KeHmacVerifyByDomainEx");

    char *errorMessage;
    if ((errorMessage = dlerror()) != NULL) {
        fprintf(stderr, "[error] dlsym libkmcext.so function - %s\n", errorMessage);
        return KE_ERROR_FUNCTION_NOT_FOUND;
    }

    return KE_SUCCESS;
}
#endif
