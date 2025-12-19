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
 * Description: All header files.
 */

#ifndef __ENS_BASE_H__
#define __ENS_BASE_H__

#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dlfcn.h>
#include <fcntl.h>
#include <libgen.h>
#include <sys/types.h>
#include <sys/stat.h>
#include "securec.h"

#define ENS_FILE_DEFAULT_ACCESS  0640
#define ENS_MODULE_NAME_MAXLEN   32

#define ENS_OK  0
#define ENS_ERR (-1)

#define ENS_TRUE  1
#define ENS_FALSE 0

typedef u_int32_t ens_uint32_t;
typedef u_int8_t ens_uchar;
typedef struct ens_rbtree_node_s ens_rbtree_node_t;
typedef struct ens_rbtree_s ens_rbtree_t;
typedef struct ens_dlist_node_s ens_dlist_node_t;
typedef ens_dlist_node_t ens_dlist_head_t;
typedef struct ens_intf_s ens_intf_t;
typedef struct ens_dep_intf_s ens_dep_intf_t;
typedef struct ens_module_s ens_module_t;
typedef struct ens_context_s ens_context_t;
typedef struct ens_conf_s ens_conf_t;

typedef int (*ens_rbtree_compare_func_t)(const char *, const char *);
typedef int (*module_load_t)(void);
typedef int (*module_unload_t)(void);
typedef int (*module_start_t)(void);
typedef int (*module_stop_t)(void);

struct ens_rbtree_node_s {
    void *key;
    ens_rbtree_node_t *left;
    ens_rbtree_node_t *right;
    ens_rbtree_node_t *parent;
    ens_uchar color;
};

struct ens_rbtree_s {
    ens_rbtree_node_t *root;
    ens_rbtree_node_t sentinel;
    ens_rbtree_compare_func_t compare;
};

typedef struct ens_stack_node_s {
    void *data;
    void *next;
} ens_stack_node_t;

typedef struct ens_stack_s {
    ens_stack_node_t *top;
} ens_stack_t;

struct ens_dlist_node_s {
    ens_dlist_node_t *next;
    ens_dlist_node_t *prev;
};

typedef struct ens_conf_item_s {
    char *key;
    char *value;
    ens_dlist_node_t dl_node;
} ens_conf_item_t;

struct ens_intf_s {
    char *name;
    void *pfn;
    ens_module_t *provider;
    ens_rbtree_node_t rb_node;
};

struct ens_dep_intf_s {
    char *name;
    void **ppfn;
    ens_module_t *consumer;
    ens_dlist_node_t dlist_node;
};

struct ens_conf_s {
    char *file_name;
    ens_uint32_t item_n;
    ens_dlist_head_t items;
};

struct ens_context_s {
    char *home_path;
    ens_conf_t config;
    ens_module_t *curr_module;
    ens_uint32_t module_num;
    ens_rbtree_t modules;
    ens_uint32_t intf_num;
    ens_rbtree_t interfaces;
};

struct ens_module_s {
    char name[ENS_MODULE_NAME_MAXLEN];
    int             is_started;
    module_load_t   load;
    module_unload_t unload;
    module_start_t  start;
    module_stop_t   stop;
    ens_uint32_t    dep_intf_num;
    ens_dlist_head_t dep_intfs;
    ens_rbtree_node_t rb_node;
};


#endif
