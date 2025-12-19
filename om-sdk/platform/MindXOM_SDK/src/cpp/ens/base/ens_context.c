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
 * Description: Context operations.
 */

#include "ens_context.h"
#include "ens_log.h"
#include "ens_conf_file.h"
#include "ens_module.h"
#include "ens_intf.h"

ens_context_t ens_ctx = {0};

int ens_ctx_initialize(void)
{
    ENS_LOG_WARN("[step1]ens_conf_initialize start.");
    ens_conf_initialize();

    ENS_LOG_WARN("[step2]ens_module_initialize start.");
    ens_module_initialize();

    ENS_LOG_WARN("[step3]ens_intf_initialize start.");
    ens_intf_initialize();

    ENS_LOG_WARN("[step4]ens_conf_load start.");
    ens_conf_load(ens_ctx.config.file_name);

    ENS_LOG_WARN("[step5]ens_conf_apply start.");
    ens_conf_apply();

    ENS_LOG_WARN("[step6]end all init success!");
    return 0;
}
