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
 * Description: Log service.
 */

#ifndef __ENS_LOG_H__
#define __ENS_LOG_H__
#include "log_common.h"


#define ENS_LOG_INFO(fmt, args...) OM_LOG_INFO(fmt, ##args)

#define ENS_LOG_WARN(fmt, args...) OM_LOG_WARN(fmt, ##args)

#define ENS_LOG_ERR(fmt, args...) OM_LOG_ERROR(fmt, ##args)

#define ENS_LOG_FATAL(fmt, args...) OM_LOG_ERROR(fmt, ##args)

#define ENS_LOG_TRACE(fmt, args...) OM_LOG_ERROR(fmt, ##args)

#endif
