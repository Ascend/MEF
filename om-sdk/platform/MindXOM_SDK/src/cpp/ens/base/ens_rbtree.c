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
 * Description: Red-black tree.
 */

#include "ens_rbtree.h"
#include <stdbool.h>
#include "ens_log.h"


static void ens_rbtree_left_rotate(ens_rbtree_t *tree, ens_rbtree_node_t *node)
{
    ens_rbtree_node_t *new_node = NULL;

    new_node = node->right;
    new_node->parent = node->parent;

    if (node == tree->root) {
        tree->root = new_node;
    } else if (node == node->parent->left) {
        node->parent->left = new_node;
    } else {
        node->parent->right = new_node;
    }

    node->right = new_node->left;
    if (!ens_rbtree_is_sentinel(tree, new_node->left)) {
        new_node->left->parent = node;
    }

    new_node->left = node;
    node->parent = new_node;

    return;
}

static void ens_rbtree_right_rotate(ens_rbtree_t *tree, ens_rbtree_node_t *node)
{
    ens_rbtree_node_t *new_node = NULL;

    new_node = node->left;
    new_node->parent = node->parent;

    if (node == tree->root) {
        tree->root = new_node;
    } else if (node == node->parent->right) {
        node->parent->right = new_node;
    } else {
        node->parent->left = new_node;
    }

    node->left = new_node->right;
    if (!ens_rbtree_is_sentinel(tree, new_node->right)) {
        new_node->right->parent = node;
    }

    new_node->right = node;
    node->parent = new_node;

    return;
}


static void ens_rbtree_insert_fixup(ens_rbtree_t *tree, ens_rbtree_node_t *fixup_node)
{
    if ((tree == NULL) || (fixup_node == NULL)) {
        ENS_LOG_ERR("check faild for inputs null.");
        return;
    }

    ens_rbtree_node_t *node = fixup_node;
    while (node != tree->root && ens_rbtree_is_red(node->parent)) {
        if (node->parent == node->parent->parent->left) {
            if (ens_rbtree_is_red(node->parent->parent->right)) {
                ens_rbtree_set_black(node->parent->parent->right);
                ens_rbtree_set_black(node->parent);
                ens_rbtree_set_red(node->parent->parent);
                node = node->parent->parent;
            } else {
                if (node == node->parent->right) {
                    node = node->parent;
                    ens_rbtree_left_rotate(tree, node);
                }
                ens_rbtree_set_black(node->parent);
                ens_rbtree_set_red(node->parent->parent);
                ens_rbtree_right_rotate(tree, node->parent->parent);
            }
        } else {
            if (ens_rbtree_is_red(node->parent->parent->left)) {
                ens_rbtree_set_black(node->parent->parent->left);
                ens_rbtree_set_black(node->parent);
                ens_rbtree_set_red(node->parent->parent);
                node = node->parent->parent;
            } else {
                if (node == node->parent->left) {
                    node = node->parent;
                    ens_rbtree_right_rotate(tree, node);
                }
                ens_rbtree_set_black(node->parent);
                ens_rbtree_set_red(node->parent->parent);
                ens_rbtree_left_rotate(tree, node->parent->parent);
            }
        }
    }

    ens_rbtree_set_black(tree->root);
    fixup_node = node;

    return;
}

ens_rbtree_node_t *ens_rbtree_insert(ens_rbtree_t *tree, ens_rbtree_node_t *node)
{
    int ret;
    ens_rbtree_node_t **curr_node = NULL;
    ens_rbtree_node_t *parent = NULL;

    if ((tree == NULL) || (node == NULL)) {
        ENS_LOG_ERR("check faild for inputs null.");
        return NULL;
    }

    if (ens_rbtree_is_sentinel(tree, tree->root)) {
        node->parent = &tree->sentinel;
        node->left = &tree->sentinel;
        node->right = &tree->sentinel;
        ens_rbtree_set_black(node);
        tree->root = node;

        return NULL;
    }

    curr_node = &tree->root;
    parent = &tree->sentinel;
    while (!ens_rbtree_is_sentinel(tree, *curr_node)) {
        parent = *curr_node;
        ret = tree->compare(node->key, (*curr_node)->key);
        if (ret < 0) {
            curr_node = &(*curr_node)->left;
        } else if (ret > 0) {
            curr_node = &(*curr_node)->right;
        } else {
            return *curr_node;
        }
    }

    *curr_node = node;
    node->parent = parent;
    node->left = node->right = &tree->sentinel;
    ens_rbtree_set_red(node);

    ens_rbtree_insert_fixup(tree, node);

    return NULL;
}

ens_rbtree_node_t *ens_rbtree_search(ens_rbtree_t *tree, const char *key)
{
    int ret;
    ens_rbtree_node_t *curr_node = NULL;

    if ((tree == NULL) || (key == NULL)) {
        ENS_LOG_ERR("check faild for inputs null.");
        return NULL;
    }

    curr_node = tree->root;
    while (!ens_rbtree_is_sentinel(tree, curr_node)) {
        ret = tree->compare(key, curr_node->key);
        if (ret < 0) {
            curr_node = curr_node->left;
        } else if (ret > 0) {
            curr_node = curr_node->right;
        } else {
            return curr_node;
        }
    }

    return NULL;
}

ens_rbtree_node_t *ens_rbtree_next(ens_rbtree_t *tree, ens_rbtree_node_t *node)
{
    if ((tree == NULL) || (node == NULL)) {
        ENS_LOG_ERR("check faild for inputs null.");
        return NULL;
    }

    ens_rbtree_node_t *tmp_node = node;

    if (tmp_node->right != &tree->sentinel) {
        tmp_node = tmp_node->right;

        while (tmp_node->left != &tree->sentinel) {
            tmp_node = tmp_node->left;
        }

        return tmp_node;
    }

    do {
        if ((tmp_node == NULL) || (tmp_node == tree->root)) {
            return NULL;
        }

        if (tmp_node == tmp_node->parent->left) {
            return tmp_node->parent;
        }

        tmp_node = tmp_node->parent;
    } while (true);
}

ens_rbtree_node_t *ens_rbtree_first(ens_rbtree_t *tree)
{
    if (tree == NULL) {
        ENS_LOG_ERR("check faild for inputs null.");
        return NULL;
    }
    ens_rbtree_node_t *node = tree->root;

    while ((node != NULL) && (node->left != &tree->sentinel)) {
        node = node->left;
    }

    if (node == &tree->sentinel) {
        node = NULL;
    }

    return node;
}
