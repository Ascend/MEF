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
 * Description: Stack operation.
 */

#ifndef __ENS_STACK_H__
#define __ENS_STACK_H__

#include "ens_base.h"

ens_stack_t *ens_stack_create(void);
int ens_stack_destroy(ens_stack_t *stack);
int ens_stack_push(ens_stack_t *stack, void *data);
void *ens_stack_pop(ens_stack_t *stack);

#endif
