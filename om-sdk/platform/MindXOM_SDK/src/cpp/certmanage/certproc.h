/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   OMSDK is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */
#ifndef _CERTPROC_H_
#define _CERTPROC_H_

#include <stdlib.h>

#ifndef VOS_OK
#define VOS_OK (0)
#endif

#ifndef VOS_ERR
#define VOS_ERR (-1)
#endif

#define SECURE_LEN                24
#define MAX_BUF_LEN_128           128
#define MAX_ENTRY_COUNT           64
#define MAX_PKEY_LEN              4096
#define MAX_DEC_LEN               1024
#define USER_READ_ONLY_PERMISSION (S_IRUSR | S_IWUSR)

enum CERT_RETURN_PARAM {
    CERT_CREAT_SSL_CTX_FAILED = 1,
    CERT_LOADCERT_FAILED = 2,
    CERT_LOADPRIVATEKEY_FAILED = 3,
    CERT_NOT_MATCHING = 4,
};

int GetJsonKeyString(const char *title, const char *filename, char *buf, unsigned int bufSize);
signed long certPeriodicalCheck(const char *certinfo_file, int *day);
signed long cert_getcertinfo(const char *certinfo_file, char subj[], unsigned long subj_len, char issuer[],
    unsigned long issuer_len, char notbegin[], int notbegin_len, char notafter[], int notafter_len,
    char serialnum[], unsigned long serialnum_len);

signed long get_extend_certinfo(const char *certinfo_file, char signature_algorithm[],
                                unsigned long signature_algorithm_len, char fingerprint[],
                                int fingerprint_len, int *pubkey_len, int *pubkey_type,
                                int *cert_version, int *kusage, int *ca, int *chain_nums);
signed long get_crl_info(const char *crl_file,  int* issuer_len, char issuer_obj_name[], char issuer_value_name[],
                         int issuer_value_length[], unsigned long row_len, char last_update_time[],
                         int last_update_len, char next_update_time[], unsigned long next_update_len);

int verify_ca_cert_sign(const char *ca_file);
signed long verify_cert_by_crl(const char *cert_file,  const char *crl_file);

signed long cert_manage_init(const char *primary_file, const char *standby_file, const char *alg_json_file);
signed long  cert_manage_finalize(void);

int StrToInt(char *buf, int buf_size);
int generate_pwd(char *pwd_buf, size_t buf_len);
int replace_server_kmc_psd(const char *path, char *pwd_buf, int pwd_buf_len, const char *alg_json_file);
int create_server_cert(const char *cert_path, unsigned long is_client, const char *ca_domain_name,
                       const char *server_domain_name, const char *client_domain_name);
int create_cert_sign_request(const char *cert_path, const char *domain_name);

#endif
