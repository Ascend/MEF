/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Configuration file operations.
 * Create 2020-11-07
 */

#ifndef __ENS_CONF_FILE_H__
#define __ENS_CONF_FILE_H__


#include "ens_base.h"

static inline int is_space(const char c)
{
    if ((c == ' ') || (c == '\t') || (c == '\r') || (c == '\n')) {
        return ENS_OK;
    }
    return ENS_ERR;
}

static inline int is_valide_char(const char c)
{
    if (((c >= '0') && (c <= '9')) || ((c >= 'a') && (c <= 'z')) ||
        ((c >= 'A') && (c <= 'Z')) || (c == '_') || (c == '-')) {
        return ENS_OK;
    }
    return ENS_ERR;
}

#define ENS_CONF_FILE_MAX_SIZE   4096

int ens_conf_load(const char *filename);
int ens_conf_apply(void);
void ens_conf_initialize(void);

#endif
