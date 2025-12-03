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
 * Description: Configuration file operations.
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
