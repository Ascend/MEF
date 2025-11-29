/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Log service.
 * Create 2020-11-07
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
