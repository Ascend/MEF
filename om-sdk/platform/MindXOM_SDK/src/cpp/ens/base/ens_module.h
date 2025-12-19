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
 * Description: Module operations.
 */

#ifndef __ENS_MODULE_H__
#define __ENS_MODULE_H__


#define ENS_MODULE_FULLPATH_MAXLEN  256
#define ENS_MODULE_INTF_NAME_MAXLEN 128


int ens_module_load(char *name);
int ens_module_assembly_all(void);
void ens_module_initialize(void);
#endif
