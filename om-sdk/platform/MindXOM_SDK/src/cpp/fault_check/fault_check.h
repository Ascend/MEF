/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: 故障检测功能实现
 * Author: huawei
 * Create: 2020-11-12
 */
#ifndef _FAULT_CHECK_H_
#define _FAULT_CHECK_H_
#include <time.h>
#include "log_common.h"

#ifndef FALSE
#define FALSE (0)
#endif

#ifndef TRUE
#define TRUE (!FALSE)
#endif

#define MAX_PATH_LENGTH 256

typedef enum {
    FAULT_STORE_M2 = 0,         // M.2硬盘
    FAULT_STORE_HARD_DISK0 = 1, // 硬盘
    FAULT_STORE_HARD_DISK1 = 2, // 硬盘
    FAULT_STORE_EMMC = 3,       // emmc卡      mmcblk0
    FAULT_STORE_SD = 4,         // sd卡        mmcblk1
    FAULT_STORE_USB0 = 5,       // u盘 面板下面
    FAULT_STORE_USB1 = 6,       // u盘 面板上面
    FAULT_STORE_USB2 = 7,       // u盘 底板
    FAULT_STORE_BUTT
} FAULT_STORE_DEV_ENUM; // 存储设备类型

// OM故障检测需要检测的器件类型
typedef enum {
    FAULT_LV1_MNT_ERR = 0,
    FAULT_LV1_NFS_ERR = 1,
    FAULT_LV1_SPACE_FULL = 2,
    FAULT_LV1_CERT_ERR = 3,
    FAULT_LV1_SD_ERR = 4,
    FAULT_LV1_BUTT
} FAULT_LV1_ENUM; // 器件类型

// 告警id中一级,后续如果新增故障id不要跟驱动侧使用的冲突，修改前两边协商一下。
typedef enum {
    FAULT_LV1_ID_HARD_DISK_ERR = 0, // 硬盘, 驱动使用，OM侧废弃
    FAULT_LV1_ID_RESVER0 = 1,       // 保留, 驱动使用，OM侧废弃
    FAULT_LV1_ID_RESVER1 = 2,       // 保留, 驱动使用，OM侧废弃
    FAULT_LV1_ID_EMMC_ERR = 3,      // emmc, 驱动使用，OM侧废弃
    FAULT_LV1_ID_SD_ERR = 4,        // sd卡, 驱动使用，OM侧废弃
    FAULT_LV1_ID_USB_ERR = 5,       // u盘, 驱动使用，OM侧废弃
    FAULT_LV1_ID_RESVER2 = 6,       // 保留, 驱动使用，OM侧废弃
    FAULT_LV1_ID_RESVER3 = 7,       // 保留, 驱动使用，OM侧废弃
    FAULT_LV1_ID_DDR_ERR = 8,       // ddr, 驱动使用，OM侧废弃
    FAULT_LV1_ID_ETH_ERR = 9,       // 网口, 驱动使用，OM侧废弃
    FAULT_LV1_ID_RESVER4 = 10,      // 保留, 驱动使用，OM侧废弃
    FAULT_LV1_ID_PCIE_SW_ERR = 11,  // 驱动使用，OM侧废弃
    FAULT_LV1_ID_PCIE_ERR = 12,     // 驱动使用，OM侧废弃
    FAULT_LV1_ID_USB_HUB_ERR = 13,  // 驱动使用，OM侧废弃
    FAULT_LV1_ID_MINID_ERR = 14,    // 驱动使用，OM侧废弃
    FAULT_LV1_ID_MCU_ERR = 15,      // 驱动使用，OM侧废弃
    FAULT_LV1_ID_3559_ERR = 16,     // 驱动使用，OM侧废弃
    FAULT_LV1_ID_RTC_ERR = 17,      // 驱动使用，OM侧废弃
    FAULT_LV1_ID_MNT_ERR = 18,
    FAULT_LV1_ID_HA_ERR = 19,
    FAULT_LV1_ID_NFS_ERR = 20,
    FAULT_LV1_ID_LTE_ERR = 21, // 驱动使用，OM侧废弃
    FAULT_LV1_ID_SPACE_FULL = 22,
    FAULT_LV1_ID_FAN_SPEED_ERR = 23, // 驱动使用，OM侧废弃
    FAULT_LV1_ID_CERT_ERR = 24,
    FAULT_LV1_ID_WIFI_ERR = 25, // 驱动使用，OM侧废弃
    FAULT_LV1_ID_BUTT
} FAULT_LV1_ID_ENUM;

// 告警级别
typedef enum {
    FAULT_LEVEL_CRITICAL = 0, // 紧急告警
    FAULT_LEVEL_MAJOR = 1,    // 严重告警
    FAULT_LEVEL_MINOR = 2,    // 轻微告警
    FAULT_LEVEL_OK = 4,
    FAULT_LEVEL_BUTT
} FAULT_LEVEL_ENUM;

// 告警处理类型
typedef enum {
    FAULT_DEAL_NONE = 0,  // 什么也不处理
    FAULT_DEAL_LOG = 1,   // 记录日志
    FAULT_DEAL_ALARM = 2, // 上报告警，这个不用
    FAULT_DEAL_ALL = 3,   // 记录日志并上报告警
    FAULT_DEAL_BUTT
} FAULT_DEAL_ENUM;

// 告警处理类型
typedef enum {
    FAULT_SLIP_TYPE_SINGLE = 0, // 单次
    FAULT_SLIP_TYPE_CONTIN = 1, // 连续
    FAULT_SLIP_TYPE_BUTT
} FAULT_SLIP_TYPE_ENUM;

// 故障状态
typedef enum {
    FAULT_STATUS_OK = 0, // 正常无故障
    FAULT_STATUS_ERR = 1 // 有故障
} FAULT_STATUS_ENUM;

typedef enum {
    ALARM_ATTR_MARK_OM = 0,
    ALARM_ATTR_MARK_DRV = 1,
    ALARM_ATTR_MARK_CUSTOMIZED = 2,
    ALARM_ATTR_MARK_EXTEND_STORAGE = 3
} ALARM_ATTR_MARK;

typedef int (*FAULT_DATA_COLLECT)(unsigned int fault_id, unsigned int sub_id, unsigned short *returndata);

#define FAULT_CHECK_TASK_DELAY_MS 100
#define FAULT_CHECK_TASK_DELAY (100 * 1000) // 任务延时 最小100ms
#define FAULT_CHECK_START_DELAY (60)        // 初始化时延时1分钟
#define FAULT_RECORD_DELAY (10)             // 10轮才更新一次文件
#define FAULT_EVEN_DELAY (20)               // 事件检测周期2s
#define FAULT_SLOW_DELAY (50)               // 耗时检测周期5s

typedef struct {
    unsigned short sub_fault_id; // 子告警id
    unsigned short fault_level;  // 告警级别 FAULT_LEVEL_ENUM :紧急告警  严重告警 轻微告警
    char fault_name[64]; // 告警名称
    char resource[32];   // 告警实体

    unsigned int slip_type;        // 类型: 单次、滑窗型
    unsigned int slip_size;        // 滑窗大小
    short slip_occur_thresh; // 产生门限
    short slip_resum_thresh; // 恢复门限
    unsigned int slip_status;      // 滑窗，最大32，门限不能超过32

    unsigned int check_period;                // 检查周期，单位ms
    unsigned short fault_deal_type;             // FAULT_DEAL_ENUM,是上报告警还是只记录日志
    unsigned int clear_alarm_flag;            // 告警清除标记
    unsigned int shield_alarm_flag;           // 告警屏蔽标记
    FAULT_DATA_COLLECT funcdatacollect; // 注册函数
} FAULT_LV2_ITEM_CFG;                   // 每项检查项，配置数据

typedef struct {
    unsigned char fault_id;
    FAULT_LV2_ITEM_CFG *sub_fault;
} FAULT_LV1_CFG_STRU; // 单个器件

#define FAULT_SLIP_STATUS_LONG_SIZE 6
#define FAULT_SLIP_MAX_SIZE 32
typedef struct {
    FAULT_LV2_ITEM_CFG item_cfg; // 配置信息

    unsigned short fault_id;         // 为LV1中告警id
    unsigned short fault_id_out;     // 为LV1中告警id
    unsigned short fault_status;     // 告警状态
    unsigned short fault_status_old; // 上一次告警状态
    time_t raise_time_stamp; // 告警时间戳（元年到告警产生的秒数）
    struct tm raise_time;    // 告警产生时间
    unsigned int run_time;         // 当前时间

    unsigned int slip_status_long[FAULT_SLIP_STATUS_LONG_SIZE];
    unsigned int slip_long_index;
    unsigned int slip_seq_id;

    unsigned short simulate_flag;  // 模拟开关
    unsigned short simulate_value; // 模拟值
} FAULT_LV2_MAPPING_STRU;  // 单个检查项信息

typedef struct {
    unsigned short fault_id;     // 内部使用
    unsigned short fault_id_out; // 告警id使用
    unsigned short sub_num;
    FAULT_LV2_MAPPING_STRU *sub_fault;
} FAULT_LV1_MAPPING_STRU; // 单个器件

// 上报给alarm模块的消息头
typedef struct {
    unsigned int data_len; // 报文长度
    unsigned int owner;    // 上报的模块标识，OM：0，驱动：1
    unsigned int item_num; // 告警个数
} ALARM_INFO_HEAD;

// fault_check跟驱动之间的接口消息的消息头
typedef struct {
    unsigned int data_len; // 报文长度
    unsigned int cmd;      // 命令字，当前仅故障上报一种命令，设定为0
    unsigned int item_num; // 告警个数
} FAULT_INFO_HEAD;

typedef struct {
    unsigned short fault_id;         // 为LV1中告警id, 对应FAULT_LV2_MAPPING_STRU的fault_id_out
    unsigned short sub_fault_id;     // 子告警id
    unsigned short fault_level;      // 告警级别 FAULT_LEVEL_ENUM :紧急告警  严重告警 轻微告警
    unsigned short reserved;         // 4字节对齐
    time_t raise_time_stamp; // 告警时间戳（元年到告警产生的秒数）
    char fault_name[64];     // 告警名称
    char resource[32];       // 告警实体
} FAULT_INFO_STRU;

typedef int (*FAULT_DATA_REPORT)(const unsigned char *info);
typedef struct {
    FAULT_DATA_REPORT data_report_func;                     // 故障上报函数
    unsigned int fault_init;                                      // 是否初始化
    unsigned int fault_has_change;                                // 是否有故障
    FAULT_LV1_MAPPING_STRU fault_info_stru[FAULT_LV1_BUTT]; // 保存全局的故障状态
} FAULT_GLOBAL_INFO;

#define FAULT_LOG_ERR(fmt, args...) OM_LOG_ERROR(fmt, ##args)

// 证书相关检查
int check_cert_update_flag(void);
int fault_check_cert_warn(unsigned int fault_id, unsigned int sub_id, unsigned short *value);

// 磁盘挂载相关检查
int fault_get_mount_status(void);
int fault_check_mount_cfg_err(unsigned int fault_id, unsigned int sub_id, unsigned short *value);
int fault_check_mount_part_lose(unsigned int fault_id, unsigned int sub_id, unsigned short *value);
int fault_check_mount_dev_lose(unsigned int fault_id, unsigned int sub_id, unsigned short *value);
int fault_check_mount_not_do(unsigned int fault_id, unsigned int sub_id, unsigned short *value);
int fault_check_mount_point_err(unsigned int fault_id, unsigned int sub_id, unsigned short *value);
int fault_check_nfs_mount_err(unsigned int fault_id, unsigned int sub_id, unsigned short *value);

// 磁盘空间检查
int fault_check_space_full(unsigned int fault_id, unsigned int sub_id, unsigned short *value);
int fault_get_sd_persent_temp(unsigned int fault_id, unsigned int sub_id, unsigned short *temp);

int check_cert_security(const char *filename);
#endif
