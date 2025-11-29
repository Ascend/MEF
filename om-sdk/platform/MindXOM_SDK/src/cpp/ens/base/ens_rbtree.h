/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Red-black tree.
 * Create 2020-11-07
 */

#ifndef __ENS_RBTREE_H__
#define __ENS_RBTREE_H__

#include <stdbool.h>
#include "ens_base.h"


#define ENS_RBTREE_BLACK 0
#define ENS_RBTREE_RED   1


static inline void ens_rbtree_set_black(ens_rbtree_node_t *node)
{
    node->color = ENS_RBTREE_BLACK;
}

static inline void ens_rbtree_set_red(ens_rbtree_node_t *node)
{
    node->color = ENS_RBTREE_RED;
}

static inline bool  ens_rbtree_is_red(ens_rbtree_node_t *node)
{
    return (node->color == ENS_RBTREE_RED);
}

static inline bool ens_rbtree_is_sentinel(ens_rbtree_t *tree, ens_rbtree_node_t *node)
{
    return (node == &tree->sentinel);
}

static inline void ens_rbtree_init(ens_rbtree_t *tree, ens_rbtree_compare_func_t compare_fn)
{
    ens_rbtree_set_black(&tree->sentinel);
    tree->sentinel.left = 0;
    tree->sentinel.right = 0;
    tree->sentinel.parent = 0;
    tree->root = &tree->sentinel;
    tree->compare = compare_fn;
}

static inline void ens_rbtree_init_node(ens_rbtree_t *tree, ens_rbtree_node_t *node, char *node_key)
{
    node->key = node_key;
    node->left = node->right = node->parent = &tree->sentinel;
    node->color = ENS_RBTREE_BLACK;
}


ens_rbtree_node_t *ens_rbtree_insert(ens_rbtree_t *tree, ens_rbtree_node_t *node);
ens_rbtree_node_t *ens_rbtree_search(ens_rbtree_t *tree, const char *key);
ens_rbtree_node_t *ens_rbtree_next(ens_rbtree_t *tree, ens_rbtree_node_t *node);
ens_rbtree_node_t *ens_rbtree_first(ens_rbtree_t *tree);

#endif
