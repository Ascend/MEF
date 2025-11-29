/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 * Description: Module operations.
 * Create 2023-08-14
 */

#ifndef MINDXOM_SDK_CHECK_EXTEND_STORAGE_H
#define MINDXOM_SDK_CHECK_EXTEND_STORAGE_H

#include "log_common.h"

#define EXTEND_ALARM_ERROR (-1)
#define SUCCESS (0)
#define LINE_INFO_SIZE (256)
#define CONFIG_FIELD_LENGTH (64)
#define CMD_SIZE (256)
#define DEV_PATH_SIZE (1024)
#define FAULT_HARD_DISK_STATUS_NORMAL (1) /* 正常状态 */
#define FAULT_HARD_DISK_STATUS_BLOCK (2)  /* 阻塞状态 */
#define FAULT_HARD_DISK_BLOCK_CNT (10)    /* 等待10s */
#define MAX_LINE_COUNT (256)    /* 文件最大行数 */
#define DECIMAL (10) // 十进制
#define EXT_CSD_DEVICE_LIFE_TIME_EST_TYP_B  (269)     /* RO */
#define EXT_CSD_DEVICE_LIFE_TIME_EST_TYP_A  (268)     /* RO */
#define EXT_CSD_PRE_EOL_INFO (267)     /* RO */
#define EXT_CSD_REG_NUM (512)
#define MAX_PE_NUM (9)
#define MMC_SEND_EXT_CSD (8)       /* adtc R1 */
#define MMC_CMD_ADTC (1 << 5)
#define MMC_RSP_SPI_R1 (1 << 7)
#define MMC_RSP_PRESENT (1 << 0)
#define MMC_RSP_CRC (1 << 2)        /* expect valid crc */
#define MMC_RSP_OPCODE (1 << 4)
#define MMC_RSP_R1 (MMC_RSP_PRESENT | MMC_RSP_CRC | MMC_RSP_OPCODE)
#define MINID_EXIST (1)
#define FAULT_MINID_ERR_CODE_NUM (128)
#define FAULT_MINID_BUTT (59)
#define FAULT_MINID_TEMP_THERSH (95)         // Dmini高温告警门限
#define FAULT_MINID_TEMP_RECOVER_THERSH (90) // Dmini告警恢复门限

typedef struct {
    char platform[64];
    char device[64];
    char location[64];
    char name[64];
} EXTEND_STORAGE_CONFIG; // 扩展存储配置信息

// 故障状态
typedef enum {
    FAULT_STATUS_OK = 0, // 正常无故障
    FAULT_STATUS_ERR = 1 // 有故障
} FAULT_STATUS_ENUM;

typedef enum {
    FAULT_STORE_M2 = 0,  // M.2硬盘
    FAULT_STORE_HARD_DISK0 = 1,  // 硬盘
    FAULT_STORE_HARD_DISK1 = 2,  // 硬盘
    FAULT_STORE_EMMC = 3, // emmc卡
    FAULT_STORE_SD = 4,   // sd卡
    FAULT_STORE_USB0 = 5,  // u盘 面板下面
    FAULT_STORE_USB1 = 6,  // u盘 面板上面
    FAULT_STORE_USB2 = 7,  // u盘 底板
    FAULT_STORE_BUTT
} FAULT_STORE_DEV_ENUM; // 存储设备类型

typedef int (*FN_DCMI_INIT)(void);
typedef int (*FN_GET_DEVICE_NUM_IN_CARD)(int, int *);
typedef int (*FN_GET_DEVICE_HEALTH)(int, int, unsigned int *);
typedef int (*FN_GET_DEVICE_TEMPERATURE)(int, int, int *);
typedef int (*FN_GET_DEVICE_ERROR_CODE)(int, int, int *, unsigned int *, unsigned int);

typedef struct dcmi_inf_struct {
    FN_DCMI_INIT pfn_dcmi_init;
    FN_GET_DEVICE_NUM_IN_CARD pfn_get_device_num_in_card;
    FN_GET_DEVICE_HEALTH pfn_get_device_health;
    FN_GET_DEVICE_TEMPERATURE pfn_get_device_temperature;
    FN_GET_DEVICE_ERROR_CODE pfn_get_device_errorcode;
} DCMI_INF_STRUCT;

int get_disk_persent(unsigned short fault_id, unsigned short sub_id, unsigned short *status);
int get_hard_disk_temp(unsigned short fault_id, unsigned short sub_id, unsigned short *status);
int get_hard_disk_age(unsigned short fault_id, unsigned short sub_id, unsigned short *status);
int get_usb_hub_alarm(unsigned short fault_id, unsigned short sub_id, unsigned short *status);
int get_hard_disk_block(unsigned short fault_id, unsigned short sub_id, unsigned short *status);
int check_extcsd_info(unsigned short fault_id, unsigned short sub_id, unsigned short *status);
int get_minid_alarm(unsigned short fault_id, unsigned short sub_id, unsigned short *status);
int check_minid_temperature(unsigned short fault_id, unsigned short sub_id, unsigned short *status);

#define EXTEND_ALARM_LOG_ERR(fmt, args...) OM_LOG_ERROR(fmt, ##args)
#define EXTEND_ALARM_LOG_INFO(fmt, args...) OM_LOG_INFO(fmt, ##args)
#define EXTEND_ALARM_LOG_WARN(fmt, args...) OM_LOG_WARN(fmt, ##args)
#define FAULT_HARD_DISK_UP_TEMP_THERSH 70 /* 硬盘底温告警门限 */
#endif // MINDXOM_SDK_CHECK_EXTEND_STORAGE_H
