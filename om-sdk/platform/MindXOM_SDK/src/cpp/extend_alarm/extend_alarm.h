/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 * Description: Module operations.
 * Create 2023-08-14
 */

#ifndef MINDXOM_SDK_EXTEND_STORAGE_ALARM_H
#define MINDXOM_SDK_EXTEND_STORAGE_ALARM_H

#include <time.h>
#include "check_extend_alarm.h"

#define EXTEND_ALARM_CHECK_TASK_DELAY_MS 100
#define EXTEND_ALARM_CHECK_START_DELAY (60)        // 初始化时延时1分钟
#define EXTEND_ALARM_CHECK_TASK_DELAY (100 * 1000) // 任务延时 最小100ms
#define EXTEND_ALARM_RECORD_DELAY (10)             // 10轮才更新一次文件
#define EXTEND_ALARM_SLOW_DELAY (50)               // 耗时检测周期5s
#define MAX_FAULT_BUFF_LEN (64 * 1024)
#define NS_CONVERT2_US (1000)
#define EXTEND_ALARM_MARK 1
#define EXTEND_ALARM_MAX_NUM 73
#define BLOCK_CHECK_CYCLE 60
#define DCMI_DRIVER_LIB "/usr/lib64/libdcmi.so"


// 告警级别
typedef enum {
    FAULT_LEVEL_CRITICAL = 0, // 紧急告警
    FAULT_LEVEL_MAJOR = 1,    // 严重告警
    FAULT_LEVEL_MINOR = 2,    // 轻微告警
    FAULT_LEVEL_OK = 4,
    FAULT_LEVEL_BUTT
} ALARM_LEVEL_ENUM;


// 故障检测跟告警之间传输的数据格式
typedef struct {
    unsigned int data_len;            // 报文长度
    unsigned int owner;               // 上报的模块标识
    unsigned int item_num;            // 告警个数
} ALARM_MSG_INFO_HEAD;

typedef struct {
    unsigned short fault_id;            // 告警id
    unsigned short sub_fault_id;        // 子告警id
    unsigned short fault_level;         // 告警级别 ALARM_LEVEL_ENUM :紧急告警  严重告警 轻微告警
    unsigned short reserved;            // 4字节对齐
    time_t raise_time_stamp;            // 告警时间戳（元年到告警产生的秒数）
    char fault_name[64];                // 告警名称
    char resource[32];                  // 告警实体
} ALARM_MSG_INFO;

typedef int (*ALARM_DATA_COLLECT)(unsigned short fault_id, unsigned short sub_id, unsigned short *status);

// 检测项配置信息结构
typedef struct {
    unsigned short fault_id;            // 告警id
    unsigned short fault_out_id;        // 对外告警id
    unsigned short sub_fault_id;        // 子告警id
    unsigned short fault_status;        // 告警状态 0:无故障 1:有故障
    unsigned short fault_level;         // 告警级别
    time_t raise_time_stamp;            // 告警产生时间
    unsigned int run_time;              // 运行时间
    unsigned int check_period;          // 检测周期
    char fault_name[64];                // 告警名称
    char resource[32];                  // 告警主体
    ALARM_DATA_COLLECT funcdatacollect; // 检测函数
} CHECK_ITEM_CFG;

typedef int (*EXTEND_ALARM_DATA_REPORT)(const unsigned char *info);

typedef struct {
    EXTEND_ALARM_DATA_REPORT data_report_func;         // 故障上报函数
    unsigned int has_changed;                     // 告警状态是否变化
    CHECK_ITEM_CFG check_items[EXTEND_ALARM_MAX_NUM];                // 保存全局的故障状态
} EXTEND_ALARM_GLOBAL_INFO;

extern char g_store_dev_name[FAULT_STORE_BUTT][32];
extern char g_store_dev_block[FAULT_STORE_BUTT];
extern int g_store_dev_block_state[FAULT_STORE_BUTT];
extern char g_store_dev_persent[FAULT_STORE_BUTT];
extern char USB_HUB_ID[CONFIG_FIELD_LENGTH];
extern DCMI_INF_STRUCT g_dcmi_infs;

#endif // MINDXOM_SDK_EXTEND_STORAGE_ALARM_H
