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
 * Description: Bidirectional linked list.
 */

#ifndef __ENS_DLIST_H__
#define __ENS_DLIST_H__

#include "ens_base.h"


static inline void ens_dlist_init_head(ens_dlist_head_t *head)
{
    head->next = head->prev = head;
}

static inline void ens_dlist_init_node(ens_dlist_head_t *node)
{
    node->next = node->prev = node;
}

static inline void ens_dlist_add(ens_dlist_node_t *node, ens_dlist_head_t *where)
{
    node->next = where->next;
    node->prev = where;
    where->next = node;
    node->next->prev = node;
}

static inline void ens_dlist_add_before(ens_dlist_node_t *node, ens_dlist_head_t *where)
{
    ens_dlist_add(node, where->prev);
}

static inline void ens_dlist_remove(ens_dlist_node_t *node)
{
    node->prev->next = node->next;
    node->next->prev = node->prev;
}


#endif
