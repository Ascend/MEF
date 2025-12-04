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
 * Description: 证书生成
 */
#include "certproc.h"

#include <stdio.h>
#include <stdlib.h>
#include <openssl/x509v3.h>
#include <openssl/bn.h>
#include <openssl/asn1.h>
#include <openssl/x509.h>
#include <openssl/pem.h>
#include <openssl/bio.h>
#include <openssl/evp.h>
#include <openssl/rsa.h>
#include <openssl/ssl.h>
#include <openssl/pem.h>
#include <openssl/err.h>

#include "security_service.h"
#include "securec.h"
#include "log_common.h"
#include "file_checker.h"

/* Generates a 4096-bit RSA key. */
static EVP_PKEY *generate_key(void)
{
    EVP_PKEY_CTX *genctx = NULL;
    EVP_PKEY *pkey = NULL;

    /* Create context using RSA algorithm. "RSA-PSS" could also be used here. */
    genctx = EVP_PKEY_CTX_new_from_name(NULL, "RSA", NULL);
    if (genctx == NULL) {
        OM_LOG_ERROR("EVP_PKEY_CTX_new_from_name() failed.");
        return NULL;
    }

    /* Initialize context for key generation purposes. */
    if (EVP_PKEY_keygen_init(genctx) <= 0) {
        EVP_PKEY_CTX_free(genctx);
        OM_LOG_ERROR("EVP_PKEY_keygen_init() failed.");
        return NULL;
    }

    /*
     * Here we set the number of bits to use in the RSA key.
     * See comment at top of file for information on appropriate values.
     */
    if (EVP_PKEY_CTX_set_rsa_keygen_bits(genctx, MAX_PKEY_LEN) <= 0) {
        EVP_PKEY_CTX_free(genctx);
        OM_LOG_ERROR("setting RSA keygen bits failed.");
        return NULL;
    }

    /* Generate the RSA key and assign it to pkey. */
    if (EVP_PKEY_generate(genctx, &pkey) <= 0) {
        EVP_PKEY_CTX_free(genctx);
        OM_LOG_ERROR("Unable to generate 4096-bit RSA key.");
        return NULL;
    }

    EVP_PKEY_CTX_free(genctx);
    /* The key has been generated, return it. */
    return pkey;
}

/* add x509 cert extension. */
static int add_x509_extension(X509* x509, const char *type_name, const char *value)
{
    X509_EXTENSION *ext;
    X509V3_CTX ctx;
    X509V3_set_ctx_nodb(&ctx);
    X509V3_set_ctx(&ctx, x509, x509, NULL, NULL, 0);
    ext = X509V3_EXT_nconf_nid(NULL, &ctx, OBJ_txt2nid(type_name), value);
    if (X509_add_ext(x509, ext, -1) == 0) {
        OM_LOG_ERROR("failed to add extension!");
        X509_EXTENSION_free(ext);
        return COMMON_ERROR;
    }

    X509_EXTENSION_free(ext);
    OM_LOG_INFO("Extension added success!");
    return COMMON_SUCCESS;
}

static int add_x509_name_entry(X509_NAME *name, const char *domain_name)
{
    if (name == NULL) {
        OM_LOG_ERROR("X509_NAME name is NULL.");
        return COMMON_ERROR;
    }

    if (domain_name == NULL) {
        OM_LOG_ERROR("domain_name is NULL.");
        return COMMON_ERROR;
    }

    const char *country_name = "CN";
    const char *org_name = "Huawei";
    const char *org_unit = "CPL Ascend";

    if (X509_NAME_add_entry_by_txt(name, "C", MBSTRING_ASC, (const unsigned char*)country_name, -1, -1, 0) == 0) {
        OM_LOG_ERROR("X509_NAME_add_entry_by_txt set C name Error.");
        return COMMON_ERROR;
    }

    if (X509_NAME_add_entry_by_txt(name, "O", MBSTRING_ASC, (const unsigned char*)org_name, -1, -1, 0) == 0) {
        OM_LOG_ERROR("X509_NAME_add_entry_by_txt set O name Error.");
        return COMMON_ERROR;
    }

    if (X509_NAME_add_entry_by_txt(name, "OU", MBSTRING_ASC, (const unsigned char*)org_unit, -1, -1, 0) == 0) {
        OM_LOG_ERROR("X509_NAME_add_entry_by_txt set OU name Error.");
        return COMMON_ERROR;
    }

    if (X509_NAME_add_entry_by_txt(name, "CN", MBSTRING_ASC, (const unsigned char*)domain_name, -1, -1, 0) == 0) {
        OM_LOG_ERROR("X509_NAME_add_entry_by_txt set CN name Error.");
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* Generates root certificate. */
static X509* generate_rootCert(EVP_PKEY *pkey, const char *domain_name)
{
    if (domain_name == NULL) {
        OM_LOG_ERROR("domain name is NULL.");
        return NULL;
    }

    const int  version3 = 2;
    /* Allocate memory for the X509 structure. */
    X509* x509 = X509_new();
    if (x509 == NULL) {
        OM_LOG_ERROR("Unable to create X509 structure.");
        return NULL;
    }

    /* Set the serial number. */
    ASN1_INTEGER_set(X509_get_serialNumber(x509), 1);
    /* Set the cert vesion. */
    X509_set_version(x509, version3);

    /* This certificate is valid from now until exactly twenty year from now. */
    X509_gmtime_adj(X509_get_notBefore(x509), -31536000L);
    X509_gmtime_adj(X509_get_notAfter(x509), 599184000L);

    /* We want to copy the subject name to the issuer name. */
    X509_NAME *name = X509_get_subject_name(x509);
    /* Set the country code and common name. */
    if (add_x509_name_entry(name, domain_name) != COMMON_SUCCESS) {
        OM_LOG_ERROR("add_x509_name_entry failed.");
        X509_free(x509);
        return NULL;
    }
    X509_set_issuer_name(x509, name);

    /* Set the public key for our certificate. */
    if (X509_set_pubkey(x509, pkey) == 0) {
        OM_LOG_ERROR("Unable to create X509 structure.");
        X509_free(x509);
        return NULL;
    }

    if (add_x509_extension(x509, "basicConstraints", "critical,CA:TRUE,pathlen:3") != COMMON_SUCCESS) {
        OM_LOG_ERROR("failed to add extension!");
        X509_free(x509);
        return NULL;
    }

    if (add_x509_extension(x509, "keyUsage", "critical,digitalSignature,keyCertSign") != COMMON_SUCCESS) {
        OM_LOG_ERROR("failed to add extension!");
        X509_free(x509);
        return NULL;
    }

    /* Actually sign the certificate with our key. */
    if (X509_sign(x509, pkey, EVP_sha256()) == 0) {
        OM_LOG_ERROR("Error signing certificate.");
        X509_free(x509);
        return NULL;
    }

    return x509;
}

/* Generates a self-signed x509 certificate. */
static X509* generate_x509Cert(EVP_PKEY *pkey, X509 *root_cert, X509_REQ *csr)
{
    const int  version3 = 2;
    const int  serial_number = 1;
    /* Allocate memory for the X509 structure. */
    X509* x509 = X509_new();
    if (x509 == NULL) {
        OM_LOG_ERROR("Unable to create X509 structure.");
        return NULL;
    }

    /* Set the serial number. */
    ASN1_INTEGER_set(X509_get_serialNumber(x509), serial_number);
    /* Set the cert vesion. */
    X509_set_version(x509, version3);
    /* This certificate is valid from now until exactly twenty year from now. */
    X509_gmtime_adj(X509_get_notBefore(x509), -31536000L);
    X509_gmtime_adj(X509_get_notAfter(x509), 599184000L);

    EVP_PKEY *public_key = X509_REQ_get_pubkey(csr);
    if (public_key == NULL) {
        OM_LOG_ERROR("X509_REQ_get_pubkey failed.");
        X509_free(x509);
        return NULL;
    }
    /* Set the public key for our certificate. */
    if (X509_set_pubkey(x509, public_key) == 0) {
        OM_LOG_ERROR("Unable to create X509 structure.");
        X509_free(x509);
        return NULL;
    }

    X509_NAME *subject_name = X509_REQ_get_subject_name(csr);
    if (subject_name == NULL) {
        OM_LOG_ERROR("X509_REQ_get_subject_name failed.");
        X509_free(x509);
        return NULL;
    }
    X509_set_subject_name(x509, subject_name);

    X509_NAME *issuer_name = X509_get_subject_name(root_cert);
    if (issuer_name == NULL) {
        OM_LOG_ERROR("X509 get root cert subject name failed.");
        X509_free(x509);
        return NULL;
    }
    X509_set_issuer_name(x509, issuer_name);

    if (add_x509_extension(x509, "basicConstraints", "critical,CA:FALSE") != COMMON_SUCCESS) {
        OM_LOG_ERROR("failed to add extension basicConstraints!");
        X509_free(x509);
        return NULL;
    }
    if (add_x509_extension(x509, "subjectAltName", "critical,IP:127.0.0.1") != COMMON_SUCCESS) {
        OM_LOG_ERROR("failed to add extension subjectAltName!");
        X509_free(x509);
        return NULL;
    }
    
    /* Actually sign the certificate with our key. */
    if (X509_sign(x509, pkey, EVP_sha256()) == 0) {
        OM_LOG_ERROR("Error signing certificate.");
        X509_free(x509);
        return NULL;
    }

    return x509;
}

static int set_x509_req_subject(X509_REQ* x509_req, const char *domain_name)
{
    if (x509_req == NULL) {
        OM_LOG_ERROR("parameter x509_req is NULL.");
        return COMMON_ERROR;
    }

    X509_NAME *x509_name = X509_REQ_get_subject_name(x509_req);
    if (add_x509_name_entry(x509_name, domain_name) != COMMON_SUCCESS) {
        OM_LOG_ERROR("add_x509_name_entry failed.");
        return COMMON_ERROR;
    }

    if (X509_REQ_set_subject_name(x509_req, x509_name) == 0) {
        OM_LOG_ERROR("X509_REQ_set_subject_name Error.");
        return COMMON_ERROR;
    }
    return COMMON_SUCCESS;
}

/* Generates a cert-sign-request. */
static X509_REQ *generate_X509Request(const char *domain_name, EVP_PKEY *pkey)
{
    if (domain_name == NULL) {
        OM_LOG_ERROR("domain name is NULL.");
        return NULL;
    }

    int             ret = 0;
    const int       version3 = 2;
    X509_REQ        *x509_req = NULL;

    // 1. set version of x509 req
    x509_req = X509_REQ_new();
    if (X509_REQ_set_version(x509_req, version3) == 0) {
        OM_LOG_ERROR("X509_REQ_set_version Error.");
        X509_REQ_free(x509_req);
        return NULL;
    }

    // 2. set subject of x509 req
    if (set_x509_req_subject(x509_req, domain_name) != COMMON_SUCCESS) {
        OM_LOG_ERROR("set_x509_req_subject Error.");
        X509_REQ_free(x509_req);
        return NULL;
    }

    // 3. set public key
    if (X509_REQ_set_pubkey(x509_req, pkey) == 0) {
        OM_LOG_ERROR("X509_REQ_set_pubkey set pubkey Error.");
        X509_REQ_free(x509_req);
        return NULL;
    }

    // 4. set sign key of x509 req, return x509_req->signature->length
    ret = X509_REQ_sign(x509_req, pkey, EVP_sha256());
    if (ret <= 0) {
        OM_LOG_ERROR("X509_REQ_sign sign by pkey Error.");
        X509_REQ_free(x509_req);
        return NULL;
    }

    return x509_req;
}

static int save_server_cert(const char *cert_path, const char *file_name, X509 *x509_cert)
{
    char cert_path_buff[MAX_DEC_LEN] = {0};
    /* Open the PEM file for writing the key. */
    if (sprintf_s(cert_path_buff, MAX_DEC_LEN - 1, "%s/%s%s", cert_path, file_name, ".cert") < 0) {
        OM_LOG_ERROR("sprintf_s fail cert path.");
        return COMMON_ERROR;
    }

    /* Open the PEM file for writing the certificate. */
    FILE *cert_fd = safety_fopen(cert_path_buff, "wb");
    if (cert_fd == NULL) {
        OM_LOG_ERROR("fail to open [%s.cert] for writing.", file_name);
        return COMMON_ERROR;
    }

    /* Write the certificate. */
    if (PEM_write_X509(cert_fd, x509_cert) == 0) {
        OM_LOG_ERROR("fail to write certificate.");
        (void)fclose(cert_fd);
        return COMMON_ERROR;
    }

    if (safety_chmod_by_fd(cert_fd, USER_READ_ONLY_PERMISSION) != COMMON_SUCCESS) {
        OM_LOG_ERROR("chmod cert file failed");
    }
    (void)fclose(cert_fd);
    return COMMON_SUCCESS;
}

static int save_pwd_key(const char *cert_path, const char *file_name, char *pwd_buf, int pwd_buf_len)
{
    char pwd_key_path[MAX_DEC_LEN] = {0};
    /* Open the PEM file for writing the key. */
    if (sprintf_s(pwd_key_path, MAX_DEC_LEN - 1, "%s/%s%s", cert_path, file_name, ".psd") < 0) {
        OM_LOG_ERROR("sprintf_s fail pwd file path.");
        return COMMON_ERROR;
    }
    char alg_json_path[MAX_DEC_LEN] = {0};
    /* Open the PEM file for writing the key. */
    if (sprintf_s(alg_json_path, MAX_DEC_LEN - 1, "%s/%s", cert_path, "om_alg.json") < 0) {
        OM_LOG_ERROR("sprintf_s fail algorithm json file path.");
        return COMMON_ERROR;
    }
    /* 替换口令 */
    int ret = replace_server_kmc_psd(pwd_key_path, pwd_buf, pwd_buf_len, alg_json_path);
    if (ret != VOS_OK) {
        OM_LOG_ERROR("save kmc psd failed.");
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

static int save_server_private_key(const char *cert_path, const char *file_name, EVP_PKEY *pkey)
{
    char pwd_buf[MAX_DEC_LEN] = {0};
    char priv_key_path[MAX_DEC_LEN] = {0};
    /* 生成随机口令 */
    int ret = generate_pwd(pwd_buf, MAX_DEC_LEN);
    if (ret != VOS_OK) {
        OM_LOG_ERROR("generate pwd error");
        goto err;
    }
    /* Open the PEM file for writing the key. */
    if (sprintf_s(priv_key_path, MAX_DEC_LEN - 1, "%s/%s%s", cert_path, file_name, ".priv") < 0) {
        OM_LOG_ERROR("sprintf_s fail priv path.");
        goto err;
    }

    OM_LOG_INFO("priv_key_path is: %s", priv_key_path);
    FILE *priv_key_fd = safety_fopen(priv_key_path, "wb");
    if (priv_key_fd == NULL) {
        OM_LOG_ERROR("fail to open [server_kmc.priv] for writing.");
        goto err;
    }

    /* Write the key to file. */
    if (PEM_write_PrivateKey(priv_key_fd, pkey, EVP_aes_256_cbc(),
        (unsigned char *)pwd_buf, SECURE_LEN, NULL, NULL) == 0) {
        OM_LOG_ERROR("fail to write private key.");
        (void)fclose(priv_key_fd);
        goto err;
    }

    if (safety_chmod_by_fd(priv_key_fd, USER_READ_ONLY_PERMISSION) != COMMON_SUCCESS) {
        OM_LOG_ERROR("chmod private key failed");
    }
    (void)fclose(priv_key_fd);

    ret = save_pwd_key(cert_path, file_name, pwd_buf, MAX_DEC_LEN);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("fail to save pwd file.");
        goto err;
    }

    if (memset_s(pwd_buf, MAX_DEC_LEN, 0, MAX_DEC_LEN) != 0) {
        OM_LOG_ERROR("memory set failed");
    }

    return COMMON_SUCCESS;
err:
    if (memset_s(pwd_buf, MAX_DEC_LEN, 0, MAX_DEC_LEN) != 0) {
        OM_LOG_ERROR("memory set failed");
    }
    return COMMON_ERROR;
}

static int generate_server_cert_by_root(const char *cert_path, EVP_PKEY *root_private_key,
                                        X509 *root_cert, const char *server_cert_name, const char *domain_name)
{
    OM_LOG_INFO("Generating server cert start");
    EVP_PKEY *private_key = generate_key();
    if (private_key == NULL) {
        OM_LOG_ERROR("Generating server RSA key error!");
        return COMMON_ERROR;
    }

    /* Generate the server certificate. */
    OM_LOG_INFO("Generating server cert start");
    X509_REQ *csr = generate_X509Request(domain_name, private_key);
    if (csr == NULL) {
        OM_LOG_ERROR("Generating cert sign request error!");
        EVP_PKEY_free(private_key);
        return COMMON_ERROR;
    }
    
    X509 *server_cert = generate_x509Cert(root_private_key, root_cert, csr);
    if (server_cert == NULL) {
        OM_LOG_ERROR("Generating x509 certificate error!");
        X509_REQ_free(csr);
        EVP_PKEY_free(private_key);
        return COMMON_ERROR;
    }
    X509_REQ_free(csr);

    int ret = save_server_cert(cert_path, server_cert_name, server_cert);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("save cert failed!");
        EVP_PKEY_free(private_key);
        X509_free(server_cert);
        return COMMON_ERROR;
    }
    X509_free(server_cert);

    ret = save_server_private_key(cert_path, server_cert_name, private_key);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("save private key failed!");
        EVP_PKEY_free(private_key);
        return COMMON_ERROR;
    }

    EVP_PKEY_free(private_key);
    OM_LOG_INFO("create server cert success!");
    return COMMON_SUCCESS;
}

int create_server_cert(const char *cert_path, unsigned long is_client, const char *ca_domain_name,
                       const char *server_domain_name, const char *client_domain_name)
{
    int ret = check_file_path_valid(cert_path);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("check_file_path_valid ret: %d!", ret);
        return COMMON_ERROR;
    }

    OM_LOG_INFO("Generating root CA start");
    const char *root_cert_name = "root_ca";
    EVP_PKEY *root_private_key = generate_key();
    if (root_private_key == NULL) {
        OM_LOG_ERROR("Generating root RSA key error!");
        return COMMON_ERROR;
    }

    /* Generate the certificate. */
    OM_LOG_INFO("Generating root certificate start");
    X509 *root_cert = generate_rootCert(root_private_key, ca_domain_name);
    if (root_cert == NULL) {
        OM_LOG_ERROR("Generating root certificate error!");
        EVP_PKEY_free(root_private_key);
        return COMMON_ERROR;
    }

    ret = save_server_private_key(cert_path, root_cert_name, root_private_key);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("save root private key failed!");
        EVP_PKEY_free(root_private_key);
        X509_free(root_cert);
        return COMMON_ERROR;
    }

    ret = save_server_cert(cert_path, root_cert_name, root_cert);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("save root cert failed!");
        EVP_PKEY_free(root_private_key);
        X509_free(root_cert);
        return COMMON_ERROR;
    }

    ret = generate_server_cert_by_root(cert_path, root_private_key, root_cert, "server_kmc", server_domain_name);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("generate server cert by root failed!");
        EVP_PKEY_free(root_private_key);
        X509_free(root_cert);
        return COMMON_ERROR;
    }

    if (is_client == 1) {
        ret = generate_server_cert_by_root(cert_path, root_private_key, root_cert, "client_kmc", client_domain_name);
        if (ret != COMMON_SUCCESS) {
            OM_LOG_ERROR("generate client cert by root failed!");
            EVP_PKEY_free(root_private_key);
            X509_free(root_cert);
            return COMMON_ERROR;
        }
    }
    
    EVP_PKEY_free(root_private_key);
    X509_free(root_cert);
    OM_LOG_INFO("create server cert success!");
    return COMMON_SUCCESS;
}

int create_cert_sign_request(const char *cert_path, const char *domain_name)
{
    if (domain_name == NULL) {
        OM_LOG_ERROR("domain name is empty!");
        return COMMON_ERROR;
    }

    int ret = check_file_path_valid(cert_path);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("check_file_path_valid ret: %d!", ret);
        return COMMON_ERROR;
    }

    OM_LOG_INFO("Generating RSA key start");
    EVP_PKEY *private_key = generate_key();
    if (private_key == NULL) {
        OM_LOG_ERROR("Generating RSA key error!");
        return COMMON_ERROR;
    }

    char path_buff[MAX_DEC_LEN] = {0};
    /* Open the PEM file for writing the key. */
    if (sprintf_s(path_buff, MAX_DEC_LEN - 1, "%s/%s", cert_path, "x509Req.pem") < 0) {
        OM_LOG_ERROR("sprintf_s fail to csr path.");
        EVP_PKEY_free(private_key);
        return COMMON_ERROR;
    }

    /* Generate the certificate. */
    OM_LOG_INFO("Generating cert sign request start");
    X509_REQ *csr = generate_X509Request(domain_name, private_key);
    if (csr == NULL) {
        OM_LOG_ERROR("Generating cert sign request error!");
        EVP_PKEY_free(private_key);
        return COMMON_ERROR;
    }

    BIO *out = BIO_new_file(path_buff, "w");
    if (!PEM_write_bio_X509_REQ(out, csr)) {
        OM_LOG_ERROR("unable to write X509 request.");
        BIO_free(out);
        X509_REQ_free(csr);
        EVP_PKEY_free(private_key);
        return COMMON_ERROR;
    }
    BIO_free(out);
    X509_REQ_free(csr);

    ret = save_server_private_key(cert_path, "server_kmc", private_key);
    if (ret != COMMON_SUCCESS) {
        OM_LOG_ERROR("save private key failed!");
        EVP_PKEY_free(private_key);
        return COMMON_ERROR;
    }

    EVP_PKEY_free(private_key);
    OM_LOG_INFO("create server cert success!");
    return COMMON_SUCCESS;
}
