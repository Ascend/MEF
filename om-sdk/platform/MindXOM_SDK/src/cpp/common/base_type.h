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
 * Description: base_type.h
 * modify:
 *  1. 日期     :  2020年12月8号
 *      修改原因    : cleancode修改
 */
#ifndef _COMMON_BASETYPE_H
#define _COMMON_BASETYPE_H
#include <string.h>
#include <syslog.h>

typedef void   VOID;

typedef const void *gconstpointer;
typedef void* gpointer;

#define EDGE_OK  0
#define EDGE_ERR (-1)
#define EXIST  0
#define NOT_EXIST  (-1)
#define BUF_MAX_LENGTH 1024
#define CHECK_LINK_YES  0
#define CHECK_LINK_NO  (-1)

#define COMMON_TRUE  1
#define COMMON_FALSE 0

#define COMMON_SUCCESS  0
#define COMMON_ERROR (-1)
#define COMMON_INVALID_PARAM (-2)

#define COMMON_VALUE_LEN_32 32
#define COMMON_VALUE_LEN_64 64
#define COMMON_VALUE_LEN_128 128
#define COMMON_VALUE_LEN_256 256
#define COMMON_VALUE_LEN_512 512
#define COMMON_VALUE_LEN_1024 1024
#define COMMON_VALUE_LEN_4096 4096

#define MINDX_EDGE_UID 1024 /* MindXEdge用户用户uid定义 */
#define ROOT_UID 0          /* root用户uid定义 */

typedef enum {
    BOARD_TYPE_SEI = 0,
    BOARD_TYPE_HILENS,
    BOARD_TYPE_UNKNOWN,
}BOARD_TYPE;

#define MINID_FLAG_UNKNOWN   (-1)
#define MINID_FLAG_NOT_EXIST   0
#define MINID_FLAG_EXIST       1

typedef struct board_info {
    int board_id;
    int pcb_id;
    int bom_id;
}BOARD_INFO_S;

#define CODE_FILENAME (strrchr(__FILE__, '/') ? (strrchr(__FILE__, '/') + 1) : __FILE__)

#define SYS_LOG_ERROR(fmt, args...) do { \
        syslog(LOG_ERR, "[ERROR][COMMON][%s, %d, %s]: "fmt, CODE_FILENAME, __LINE__, __FUNCTION__, ##args); \
} while (0)
#endif
