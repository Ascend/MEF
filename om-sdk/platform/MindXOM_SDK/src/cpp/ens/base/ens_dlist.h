/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Bidirectional linked list.
 * Create 2020-11-07
 */

#ifndef __ENS_DLIST_H__
#define __ENS_DLIST_H__

#include "ens_base.h"
#include "ens_log.h"


static inline void ens_dlist_init_head(ens_dlist_head_t *head)
{
    if (head == NULL) {
        ENS_LOG_ERR("input error!");
        return;
    }
    head->next = head->prev = head;
}

static inline void ens_dlist_init_node(ens_dlist_head_t *node)
{
    if (node == NULL) {
        ENS_LOG_ERR("input error!");
        return;
    }
    node->next = node->prev = node;
}

static inline void ens_dlist_add(ens_dlist_node_t *node, ens_dlist_head_t *where)
{
    if (node == NULL || where == NULL) {
        ENS_LOG_ERR("input error!");
        return;
    }
    node->next = where->next;
    node->prev = where;
    where->next = node;
    node->next->prev = node;
}

static inline void ens_dlist_add_before(ens_dlist_node_t *node, ens_dlist_head_t *where)
{
    if (node == NULL || where == NULL) {
        ENS_LOG_ERR("input error!");
        return;
    }
    ens_dlist_add(node, where->prev);
}

static inline void ens_dlist_remove(ens_dlist_node_t *node)
{
    if (node == NULL) {
        ENS_LOG_ERR("input error!");
        return;
    }
    node->prev->next = node->next;
    node->next->prev = node->prev;
}


#endif
