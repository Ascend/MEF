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
 * Description: 告警处理
 */
#ifndef ALARM_PROCESS_H
#define ALARM_PROCESS_H
#include <time.h>
#include <pthread.h>
#include "log_common.h"

#ifndef FALSE
#define FALSE (0)
#endif

#ifndef TRUE
#define TRUE (!FALSE)
#endif

#define RESOURCE_MAX_LEN 32

// 故障检测跟告警之间传输的数据格式，修改时需要两天同时修改
typedef struct {
    unsigned int data_len;            // 报文长度
    unsigned int owner;               // 上报的模块标识
    unsigned int item_num;            // 告警个数
} ALARM_MSG_INFO_HEAD;

typedef struct {
    unsigned short fault_id;            // 为LV1中告警id, 对应FAULT_LV2_MAPPING_STRU的fault_id_out
    unsigned short sub_fault_id;        // 子告警id
    unsigned short fault_level;         // 告警级别 FAULT_LEVEL_ENUM :紧急告警  严重告警 轻微告警
    unsigned short reserved;            // 4字节对齐
    time_t raise_time_stamp;    // 告警时间戳（元年到告警产生的秒数）
    char fault_name[64];        // 告警名称
    char resource[RESOURCE_MAX_LEN]; // 告警实体
} ALARM_MSG_FAULT_INFO;

typedef enum {
    ALARM_ATTR_MARK_OM = 0,
    ALARM_ATTR_MARK_DRV = 1,
    ALARM_ATTR_MARK_CUSTOMIZED = 2,
    ALARM_ATTR_MARK_EXTEND_STORAGE = 3
} ALARM_ATTR_MARK;

// 内存中保存的结构
typedef struct ALARM_LV1_NODE {
    ALARM_MSG_FAULT_INFO fault_info;  // 故障信息
    unsigned short fault_status;              // 告警状态
    unsigned char  fault_attr_mark;           // 故障归属标记 0:非硬件 1:硬件
    unsigned int clear_alarm_flag;          // 告警清除标记
    unsigned int shield_alarm_flag;         // 告警屏蔽标记
    struct ALARM_LV1_NODE *next;
} ALARM_LV1_STRU;

typedef enum {
    SHILED_FLAG_UNSHIELDED = 0,  // 未屏蔽
    SHILED_FLAG_SHIELDED = 1     // 已屏蔽
} SHILED_FLAG_ENUM;

typedef enum {
    ALARM_STATUS_OK = 0,  // 无故障，无需告警
    ALARM_STATUS_ERR = 1, // 有故障，需要告警
} ALARM_STATUS_ENUM;

typedef int (*PROCESS_ALARM_EVENT_FUNC)(int alarm_num);
#define FAULT_LV1_MAX_SIZE 50
#define FAULT_LV2_MAX_SIZE 100
#define MAX_REGISTER_FUN_NUM 10
typedef struct {
    ALARM_LV1_STRU* alarm_info_stru[FAULT_LV1_MAX_SIZE];
    PROCESS_ALARM_EVENT_FUNC alarm_event_process_funcs[MAX_REGISTER_FUN_NUM];
    pthread_mutex_t mutex;
} ALARM_GLOBAL_INFO;

extern ALARM_GLOBAL_INFO g_alarm_global_info;

#define CLEAR_ALARM_FILE "/run/clear_alarm.ini"
#define SHIELD_ALARM_FILE "/home/data/ies/shield_alarm.ini"
#define CLEAR_ALARM_TYPE 0
#define SHIELD_ALARM_TYPE 1

int set_clear_shield_alarm_flag(const char *filename, int alarm_option_type);
void *fault_manager_shield_alarm_thread(void *p_void);
int fault_record_to_file(unsigned int *active_num, int *alarm_level);
ALARM_LV1_STRU *create_new_alarm();
ALARM_LV1_STRU *search_item_by_info(const ALARM_MSG_FAULT_INFO *cur_fault_info, int is_shield);
void publish_alarm_event(int alarm_level);
int alarm_report(const unsigned char *info);
int alarm_process_load(void);
int alarm_process_unload(void);
int alarm_process_start();

#define ALARM_LOG_ERR(fmt, args...) OM_LOG_ERROR(fmt, ##args)

#endif  // ALARM_PROCESS_H
