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
 * Description: 写日志接口
 */
#ifndef _LOG_COMMON_H_
#define _LOG_COMMON_H_

#include <string.h>
#include "base_type.h"

/* 一条日志的最大长度 */
#define OM_LOGMSG_MAX_LEN 4096
#define TIMESTAMP_LEN 64
/* 日志文件的大小 */
#define OM_LOG_DIR "/var/plog/ibma_edge/"
#define OM_REDFISH_LOG_DIR "/var/plog/redfish/"
#define OM_LOG_PROFIX ".log"
#define OM_LOG_NAME "om_platform_run"
#define REDFISH_OM_LOG_NAME "redfish_platform_run"

void OmsdkLogWrite(const char *pcFmt, ...);

#define CODE_FILENAME (strrchr(__FILE__, '/') ? (strrchr(__FILE__, '/') + 1) : __FILE__)

#define OM_MOD_LOG(logLevel, fmt, args...) \
    do {                                                           \
        OmsdkLogWrite("[%s][%s,%s,%d]:" fmt "\n", logLevel,      \
            CODE_FILENAME,                                         \
            __FUNCTION__,                                          \
            __LINE__,                                              \
            ##args);                                               \
    } while (0)

#define OM_LOG_INFO(fmt, args...)                      \
    OM_MOD_LOG("INFO",                                 \
        fmt,                                              \
        ##args)

#define OM_LOG_WARN(fmt, args...)                      \
    OM_MOD_LOG("WARN",                                 \
        fmt,                                              \
        ##args)

#define OM_LOG_ERROR(fmt, args...)                     \
    OM_MOD_LOG("ERROR",                                \
        fmt,                                              \
        ##args)

#endif /* _LOG_COMMON_H_ */
