/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Module operations.
 * Create 2020-11-07
 */

#ifndef __ENS_MODULE_H__
#define __ENS_MODULE_H__


#define ENS_MODULE_FULLPATH_MAXLEN  256
#define ENS_MODULE_INTF_NAME_MAXLEN 128


int ens_module_load(char *name);
int ens_module_assembly_all(void);
void ens_module_initialize(void);
#endif
