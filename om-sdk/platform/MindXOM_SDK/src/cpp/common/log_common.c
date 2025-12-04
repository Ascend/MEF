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
 * Description: 写日志接口函数
 */
#include "log_common.h"
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <sys/select.h>
#include <time.h>
#include <string.h>
#include <unistd.h>
#include "securec.h"
#include "file_checker.h"

static int GetTime(char *timestamp, int timestampLen)
{
    struct tm tmBuff;
    time_t seconds = time(NULL);
    if ((seconds == 0) || (timestamp == NULL) || (timestampLen < 0) || (timestampLen > TIMESTAMP_LEN)) {
        SYS_LOG_ERROR("invalid inputs");
        return COMMON_ERROR;
    }

    struct tm *tmRet = localtime_r(&seconds, &tmBuff);
    if (tmRet == NULL) {
        SYS_LOG_ERROR("localtime_r failed");
        return COMMON_ERROR;
    }

    int ret = strftime(timestamp, timestampLen, "%Y-%m-%d %H:%M:%S", &tmBuff);
    if (ret != 0) {
        return COMMON_SUCCESS;
    }

    SYS_LOG_ERROR("strftime failed");
    return COMMON_ERROR;
}

static FILE *GetLogFileFp(const char *strLogFile)
{
    FILE *logFileFp = fopen(strLogFile, "r");
    if (logFileFp == NULL) {                 /* 不存在 */
        logFileFp = fopen(strLogFile, "wb"); /* 创建 */
        if (logFileFp == NULL) {
            SYS_LOG_ERROR("fopen failed");
            return NULL;
        }
    } else { /* 存在 */
        (void)fclose(logFileFp);
        logFileFp = fopen(strLogFile, "ab"); /* 以追加方式重新打开 */
        if (logFileFp == NULL) {
            SYS_LOG_ERROR("fopen failed");
            return NULL;
        }
    }

    return logFileFp;
}

void OmsdkLogWrite(const char *pcFmt, ...)
{
    if (pcFmt == NULL) {
        SYS_LOG_ERROR("invalid inputs");
        return;
    }

    va_list pcArgs;
    char logBuf[OM_LOGMSG_MAX_LEN] = {0};
    char timestamp[TIMESTAMP_LEN] = {0};
    char log_file[TIMESTAMP_LEN] = {0};

    uid_t id = (uid_t)getuid();
    if (id == 0) {
        if (sprintf_s(log_file, TIMESTAMP_LEN - 1, "%s%s%s",
            OM_LOG_DIR, OM_LOG_NAME, OM_LOG_PROFIX) < 0) {
            SYS_LOG_ERROR("sprintf_s uid[%d] logfile failed.", id);
            return;
        }
    } else {
        if (sprintf_s(log_file, TIMESTAMP_LEN - 1, "%s%s%s",
            OM_REDFISH_LOG_DIR, REDFISH_OM_LOG_NAME, OM_LOG_PROFIX) < 0) {
            SYS_LOG_ERROR("sprintf_s uid[%d] logfile failed.", id);
            return;
        }
    }
    FILE *fileFp = GetLogFileFp(log_file);
    if (fileFp == NULL) {
        SYS_LOG_ERROR("get logfile pointer failed");
        return;
    }
    
    int logFd = fileno(fileFp);
    if (logFd > 0) {
        (void)fchmod(logFd, 0640); /* 更改文件权限0640 */
    }

    va_start(pcArgs, pcFmt);
    int logLen = vsnprintf_s(logBuf, OM_LOGMSG_MAX_LEN, OM_LOGMSG_MAX_LEN - 1, pcFmt, pcArgs);
    va_end(pcArgs);
    if (logLen < 0) {
        (void)fclose(fileFp);
        return;
    }

    if (logLen >= OM_LOGMSG_MAX_LEN) {
        logBuf[OM_LOGMSG_MAX_LEN - 1] = '\0';
    }

    int ret = GetTime(timestamp, sizeof(timestamp));
    if (ret != COMMON_SUCCESS) {
        (void)fclose(fileFp);
        return;
    }

    (void)fprintf(fileFp, "[%s]%s", timestamp, logBuf);
    (void)fflush(fileFp);
    (void)fclose(fileFp);
}
