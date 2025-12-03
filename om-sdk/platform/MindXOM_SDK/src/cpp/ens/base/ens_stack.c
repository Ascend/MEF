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
 * Description: Stack operation.
 */

#include "ens_stack.h"
#include "ens_log.h"

ens_stack_t *ens_stack_create(void)
{
    ens_stack_t *stack = malloc(sizeof(ens_stack_t));
    if (stack == NULL) {
        ENS_LOG_ERR("malloc fail.");
        return NULL;
    }

    if (memset_s(stack, sizeof(ens_stack_t), 0, sizeof(ens_stack_t)) != 0) {
        ENS_LOG_ERR("memset_s error!");
        free(stack);
        stack = NULL;
        return NULL;
    }

    return stack;
}
int ens_stack_destroy(ens_stack_t *stack)
{
    if (stack == NULL) {
        ENS_LOG_ERR("check faild for inputs null.");
        return ENS_ERR;
    }
    ens_stack_node_t *node = stack->top;
    while (node) {
        free(node->data);
        node->data = NULL;
        stack->top = node->next;
        free(node);
        node = stack->top;
    }

    return ENS_OK;
}
int ens_stack_push(ens_stack_t *stack, void *data)
{
    ens_stack_node_t *new_node = NULL;

    if ((stack == NULL) || (data == NULL)) {
        ENS_LOG_ERR("check faild for inputs null.");
        return ENS_ERR;
    }

    new_node = malloc(sizeof(ens_stack_node_t));
    if (new_node) {
        new_node->data = data;
        new_node->next = stack->top;
        stack->top = new_node;
    } else {
        return ENS_ERR;
    }

    return ENS_OK;
}
void *ens_stack_pop(ens_stack_t *stack)
{
    void *data = NULL;

    if (stack == NULL) {
        ENS_LOG_ERR("check faild for inputs null.");
        return NULL;
    }

    ens_stack_node_t *node = stack->top;
    if (node) {
        data = node->data;
        if (data == NULL) {
            return NULL;
        }
        stack->top = node->next;
        node->next = NULL;
        free(node);
        node = NULL;
    }

    return data;
}
