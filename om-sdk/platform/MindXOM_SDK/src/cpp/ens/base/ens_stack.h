/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Stack operation.
 * Create 2020-11-07
 */

#ifndef __ENS_STACK_H__
#define __ENS_STACK_H__

#include "ens_base.h"

ens_stack_t *ens_stack_create(void);
int ens_stack_destroy(ens_stack_t *stack);
int ens_stack_push(ens_stack_t *stack, void *data);
void *ens_stack_pop(ens_stack_t *stack);

#endif
