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
 * Description: Interface operations.
 */

#include "ens_intf.h"
#include "ens_log.h"
#include "ens_context.h"
#include "ens_rbtree.h"
#include "ens_dlist.h"

int ens_intf_export(char *fn_name, void *fn)
{
    ens_intf_t *intf = NULL;
    ens_intf_t *temp_intf = NULL;
    ens_rbtree_node_t *node = NULL;

    if (fn_name != NULL && fn != NULL && strlen(fn_name) < ENS_INTF_NAME_MAX_LEN) {
        intf = malloc(sizeof(ens_intf_t));
        if (intf) {
            if (memset_s(intf, sizeof(ens_intf_t), 0, sizeof(ens_intf_t)) != 0) {
                ENS_LOG_ERR("memset_s error!");
                free(intf);
                intf = NULL;
                return ENS_ERR;
            }
            intf->name = fn_name;
            intf->pfn = fn;
            intf->provider = ens_ctx.curr_module;

            ens_rbtree_init_node(&ens_ctx.interfaces, &intf->rb_node, intf->name);

            node = ens_rbtree_insert(&ens_ctx.interfaces, &intf->rb_node);
            if (node) {
                temp_intf = ((ens_intf_t *)((void *)(node) - (size_t)(&((ens_intf_t *)0)->rb_node)));
                ENS_LOG_FATAL("Module[%s] export an duplicated interface[%s] with module[%s].",
                    intf->provider->name, intf->name, temp_intf->provider->name);
                free(intf);
                intf = NULL;
                return ENS_ERR;
            }

            ens_ctx.intf_num++;
        } else {
            ENS_LOG_FATAL("malloc fail when exporting interface [%s].", fn_name);
        }
    } else {
        ENS_LOG_ERR("Invalid input parameter.");
        return ENS_ERR;
    }

    return ENS_OK;
}

int ens_intf_import(char *fn_name, void **ppfn)
{
    ens_dep_intf_t *dep_intf = NULL;
    ens_module_t *curr_module = ens_ctx.curr_module;

    if (curr_module == NULL) {
        ENS_LOG_FATAL("Invalid module context.");
        return ENS_ERR;
    }

    if (fn_name != NULL && ppfn != NULL && strlen(fn_name) < ENS_INTF_NAME_MAX_LEN) {
        dep_intf = malloc(sizeof(ens_dep_intf_t));
        if (dep_intf) {
            if (memset_s(dep_intf, sizeof(ens_dep_intf_t), 0, sizeof(ens_dep_intf_t)) != 0) {
                ENS_LOG_ERR("memset_s error!");
                free(dep_intf);
                dep_intf = NULL;
                return ENS_ERR;
            }
            dep_intf->name = fn_name;
            dep_intf->ppfn = ppfn;
            dep_intf->consumer = ens_ctx.curr_module;
            ens_dlist_init_node(&dep_intf->dlist_node);

            ens_dlist_add_before(&dep_intf->dlist_node, &curr_module->dep_intfs);
            curr_module->dep_intf_num++;
        } else {
            ENS_LOG_FATAL("malloc fail when importing interface:%s", fn_name);
            return ENS_ERR;
        }
    }

    return ENS_OK;
}

ens_intf_t *ens_intf_get_by_name(const char *name)
{
    if (name == NULL) {
        ENS_LOG_FATAL("inputs null.");
        return NULL;
    }

    ens_rbtree_node_t *rb_node = ens_rbtree_search(&ens_ctx.interfaces, name);
    if (rb_node) {
        return ((ens_intf_t *)((void *)(rb_node) - (size_t)(&((ens_intf_t *)0)->rb_node)));
    }
    return NULL;
}

static int ens_intf_compare_key(const char *key1, const char *key2)
{
    if ((key1 == NULL) || (key2 == NULL)) {
        ENS_LOG_FATAL("inputs null.");
        return ENS_ERR;
    }
    return strcmp(key1, key2);
}

void ens_intf_initialize(void)
{
    ens_rbtree_init(&ens_ctx.interfaces, ens_intf_compare_key);
    return;
}
