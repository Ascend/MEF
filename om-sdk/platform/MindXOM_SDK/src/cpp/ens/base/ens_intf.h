/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Interface operations.
 * Create 2020-11-07
 */

#ifndef __ENS_INTF_H__
#define __ENS_INTF_H__

#include "ens_base.h"

#define ENS_INTF_NAME_MAX_LEN    256


ens_intf_t *ens_intf_get_by_name(const char *name);
void ens_intf_initialize(void);

#endif
