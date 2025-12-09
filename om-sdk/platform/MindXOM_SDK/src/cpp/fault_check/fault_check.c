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
 * Description: 故障检测功能实现
 */
#include "fault_check.h"
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <fcntl.h>
#include <dlfcn.h>
#include <string.h>
#include <stddef.h>
#include <time.h>
#include <pthread.h>
#include <stdbool.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <sys/stat.h>
#include <sys/prctl.h>
#include <sys/mman.h>
#include "securec.h"
#include "ens_api.h"
#include "certproc.h"
#include "file_checker.h"

#define SD_DEVICE_PATH "/dev/mmcblk1"
#define CERT_CHECKTIME_CONFIGURE_FILE "/home/data/config/default/om_alg.json"
#define CERT_CHECK_SECURE_LEN 16
#define CERT_CHECK_PERIOD_TIME_RATE 1000
#define MAX_CERT_CHECK_PERIOD (24 * 3600 * 1000 * 7)
#define MIN_CERT_CHECK_PERIOD 10000
#define MAX_CERTFILE_SIZE (10 * 1024)
#define MAX_FAULT_BUFF_LEN (64 * 1024)
#define MAX_FAULT_DATA_LEN (1024 * 1024)
#define DRV_FAULT_CHECK_SO "/usr/local/lib/libdrv_fault_check.so"
#define CUSTOMIZED_ALARM_CHECK_SO "/usr/local/mindx/MindXOM/lib/libcustomized_alarm.so"

typedef enum {
    FAULT_LV2_MNT_CFG_ERR = 0,   // cfg error
    FAULT_LV2_MNT_PART_LOSE = 1, // partition lose
    FAULT_LV2_MNT_DEV_LOSE = 2,  // device lose
    FAULT_LV2_MNT_NOT_DO = 3,    // mount failed or not do
    FAULT_LV2_MNT_POINT_ERR = 4, // mount point error
    FAULT_LV2_MNT_BUTT
} FAULT_LV2_MNT_ENUM;

typedef enum {
    FAULT_LV2_NFS_MOUNT_ERR = 0, // mount failed error
    FAULT_LV2_NFS_BUTT
} FAULT_LV2_NFS_ENUM;

typedef enum {
    FAULT_LV2_SPACE_FULL_ERR = 0,
    FAULT_LV2_SPACE_FULL_BUTT
} FAULT_LV2_SPACE_FULL_ENUM;

typedef enum {
    FAULT_LV2_CERT_ERR = 0,
    FAULT_LV2_CERT_BUTT
} FAULT_LV2_CERT_ENUM;

typedef enum {
    FAULT_LV2_SD_PERSENT_ERR = 0,
    FAULT_LV2_SD_BUTT
} FAULT_LV2_SD_ENUM;

static FAULT_LV2_ITEM_CFG fault_lv2_mnt_cfg[FAULT_LV2_MNT_BUTT] = {
    {0, FAULT_LEVEL_MAJOR, "mnt cfg err", "MNT", FAULT_SLIP_TYPE_CONTIN, 3, 3, 0, 0, 1000,
     FAULT_DEAL_ALL, 0, 0, fault_check_mount_cfg_err},
    {1, FAULT_LEVEL_MAJOR, "mnt part lose", "MNT", FAULT_SLIP_TYPE_CONTIN, 3, 3, 0, 0, 1000,
     FAULT_DEAL_ALL, 0, 0, fault_check_mount_part_lose},
    {2, FAULT_LEVEL_MAJOR, "mnt dev lose", "MNT", FAULT_SLIP_TYPE_CONTIN, 3, 3, 0, 0, 1000,
     FAULT_DEAL_ALL, 0, 0, fault_check_mount_dev_lose},
    {3, FAULT_LEVEL_MAJOR, "mnt not do", "MNT", FAULT_SLIP_TYPE_CONTIN, 3, 3, 0, 0, 1000,
     FAULT_DEAL_ALL, 0, 0, fault_check_mount_not_do},
    {4, FAULT_LEVEL_MAJOR, "mnt point err", "MNT", FAULT_SLIP_TYPE_CONTIN, 3, 3, 0, 0, 1000,
     FAULT_DEAL_ALL, 0, 0, fault_check_mount_point_err},
};

static FAULT_LV2_ITEM_CFG fault_lv2_nfs_cfg[FAULT_LV2_NFS_BUTT] = {
    {0, FAULT_LEVEL_MAJOR, "nfs mount failed", "NFS", FAULT_SLIP_TYPE_SINGLE, 0, 0, 0, 0, 1000,
     FAULT_DEAL_ALL, 0, 0, fault_check_nfs_mount_err},
};

static FAULT_LV2_ITEM_CFG fault_lv2_space_full_cfg[FAULT_LV2_SPACE_FULL_BUTT] = {
    {0, FAULT_LEVEL_MINOR, "directory space full", "MEM OR STORAGE", FAULT_SLIP_TYPE_SINGLE, 0, 0, 0, 0, 5000,
     FAULT_DEAL_ALL, 0, 0, fault_check_space_full},
};

static FAULT_LV2_ITEM_CFG fault_lv2_cert_cfg[FAULT_LV2_CERT_BUTT] = {
    {0, FAULT_LEVEL_MAJOR, "cert warning", "CERT", FAULT_SLIP_TYPE_SINGLE, 0, 0, 0, 0, 86400000,
     FAULT_DEAL_ALL, 0, 0, fault_check_cert_warn},
};

static FAULT_LV2_ITEM_CFG fault_lv2_sd_cfg[FAULT_LV2_SD_BUTT] = {
    {0, FAULT_LEVEL_MINOR, "mount fail", "SD", FAULT_SLIP_TYPE_SINGLE, 0, 0, 0, 0, 1000,
     FAULT_DEAL_ALL, 0, 0, fault_get_sd_persent_temp},
};

/* 一级故障 */
static FAULT_LV1_CFG_STRU fault_lv1_cfg[FAULT_LV1_BUTT] = {
    {FAULT_LV1_ID_MNT_ERR, fault_lv2_mnt_cfg},
    {FAULT_LV1_ID_NFS_ERR, fault_lv2_nfs_cfg},
    {FAULT_LV1_ID_SPACE_FULL, fault_lv2_space_full_cfg},
    {FAULT_LV1_ID_CERT_ERR, fault_lv2_cert_cfg},
    {FAULT_LV1_ID_SD_ERR, fault_lv2_sd_cfg},
};

/* 每类二级故障总检测数  */
static unsigned char fault_sub_num_array[FAULT_LV1_BUTT] = {
    FAULT_LV2_MNT_BUTT,
    FAULT_LV2_NFS_BUTT,
    FAULT_LV2_SPACE_FULL_BUTT,
    FAULT_LV2_CERT_BUTT,
    FAULT_LV2_SD_BUTT,
};

#define FAULT_BLOCK_CHECK_START_DELAY (2) // 初始化时延时2s

// 全局变量
static FAULT_GLOBAL_INFO fault_global_info = { NULL, FALSE, TRUE, {} };

int check_cert_security(const char *filename)
{
    struct stat st;
    if (stat(filename, &st) < 0) {
        FAULT_LOG_ERR("certfile error");
        return VOS_ERR;
    }

    // 大小校验
    signed long fsize = st.st_size;
    if (fsize < 0 || fsize > MAX_CERTFILE_SIZE) {
        FAULT_LOG_ERR("certfile size:%ld invalid", fsize);
        return VOS_ERR;
    }

    // 软链接校验
    if (S_ISLNK(st.st_mode)) {
        FAULT_LOG_ERR("certfile is a link");
        return VOS_ERR;
    }

    return VOS_OK;
}

static int fault_append_item(unsigned char *buff, unsigned int buff_len,
                             unsigned int *offset, FAULT_LV2_MAPPING_STRU *fault_vl2_item)
{
    int ret;
    FAULT_INFO_STRU *item = NULL;

    if ((buff == NULL) || (offset == NULL) || (fault_vl2_item == NULL)) {
        FAULT_LOG_ERR("inputs null.");
        return EDGE_ERR;
    }
    if (*offset + sizeof(FAULT_INFO_STRU) > buff_len) {
        FAULT_LOG_ERR("size:%ld out of bufflen:%d.", (*offset + sizeof(FAULT_INFO_STRU)), buff_len);
        return EDGE_ERR;
    }

    item = (FAULT_INFO_STRU *)(buff + *offset);
    if (memset_s(item, sizeof(FAULT_INFO_STRU), 0x00, sizeof(FAULT_INFO_STRU)) != 0) {
        FAULT_LOG_ERR("memset_s error!");
        return EDGE_ERR;
    }

    item->fault_id = fault_vl2_item->fault_id_out;
    item->sub_fault_id = fault_vl2_item->item_cfg.sub_fault_id;
    item->fault_level = fault_vl2_item->item_cfg.fault_level;
    item->raise_time_stamp = fault_vl2_item->raise_time_stamp;
    ret = strncpy_s(item->fault_name, sizeof(item->fault_name), fault_vl2_item->item_cfg.fault_name,
        sizeof(item->fault_name) - 1);
    if (ret != 0) {
        FAULT_LOG_ERR("strncpy fault_name failed,ret:%d.", ret);
        return EDGE_ERR;
    }
    ret = strncpy_s(item->resource, sizeof(item->resource), fault_vl2_item->item_cfg.resource,
        sizeof(item->resource) - 1);
    if (ret != 0) {
        FAULT_LOG_ERR("strncpy resource failed,ret:%d.", ret);
        return EDGE_ERR;
    }
    *offset += sizeof(FAULT_INFO_STRU);

    return EDGE_OK;
}

static int get_fault_buff_data(unsigned char **ppbuff, unsigned int buff_len)
{
    if ((ppbuff == NULL) || (buff_len > MAX_FAULT_BUFF_LEN)) {
        return EDGE_ERR;
    }

    unsigned char *pbuff = (unsigned char *)malloc(buff_len);
    if (pbuff == NULL) {
        FAULT_LOG_ERR("malloc failed,len:%u.", buff_len);
        return EDGE_ERR;
    }
    if (memset_s(pbuff, buff_len, 0, buff_len) != 0) {
        FAULT_LOG_ERR("memset_s error!");
        free(pbuff);
        pbuff = NULL;
        return EDGE_ERR;
    }
    *ppbuff = pbuff;
    return EDGE_OK;
}

// 上报故障到alarm模块
static int report_fault_record_to_alarm(void)
{
    int ret;
    unsigned int active_num = 0;
    const unsigned int buff_len = MAX_FAULT_BUFF_LEN;
    unsigned int offset = sizeof(ALARM_INFO_HEAD);
    unsigned char *fault_buff = NULL;

    ret = get_fault_buff_data(&fault_buff, buff_len);
    if (ret != EDGE_OK) {
        return EDGE_ERR;
    }

    for (unsigned int index = 0; index < FAULT_LV1_BUTT; index++) {
        FAULT_LV1_MAPPING_STRU *fault_lv1_item = &fault_global_info.fault_info_stru[index];
        for (unsigned int sub_index = 0; sub_index < fault_lv1_item->sub_num; sub_index++) {
            FAULT_LV2_MAPPING_STRU *fault_vl2_item = fault_lv1_item->sub_fault + sub_index;
            if (fault_vl2_item == NULL) {
                FAULT_LOG_ERR("fault_vl2_item is null(%d-%d).", index, sub_index);
                break;
            }

            if (fault_vl2_item->fault_status == FAULT_STATUS_OK) {
                continue;
            }
            ret = fault_append_item(fault_buff, buff_len, &offset, fault_vl2_item);
            if (ret != EDGE_OK) {
                FAULT_LOG_ERR("call fault_append_item fail(%d-%d).", index, sub_index);
            } else {
                active_num++;
            }
        }
    }
    ALARM_INFO_HEAD *msg_head = (ALARM_INFO_HEAD *)fault_buff;
    msg_head->data_len = offset;
    msg_head->item_num = active_num;
    msg_head->owner = ALARM_ATTR_MARK_OM;

    if (fault_global_info.data_report_func) {
        ret = (*fault_global_info.data_report_func)((unsigned char *)fault_buff);
    } else {
        FAULT_LOG_ERR("data_report_func null.");
        ret = EDGE_ERR;
    }

    free(fault_buff);
    fault_buff = NULL;
    msg_head = NULL;
    if (ret != EDGE_OK) {
        FAULT_LOG_ERR("report_fault_info fail,active_num:%d.", active_num);
    }

    return ret;
}

// 更新告警状态
static void update_sub_item_fault_status(FAULT_LV2_MAPPING_STRU *sub_item_fault, const unsigned short status)
{
    time_t time_curr = { 0 };

    if (sub_item_fault == NULL) {
        FAULT_LOG_ERR("Input args is NULL.");
        return;
    }

    if (status != sub_item_fault->fault_status) {
        FAULT_LOG_ERR("fault_flush_status:(%d-%d)status from(%d) to (%d).", sub_item_fault->fault_id,
            sub_item_fault->item_cfg.sub_fault_id, sub_item_fault->fault_status, status);
        sub_item_fault->fault_status = status;
        (void)time(&time_curr);
        sub_item_fault->raise_time_stamp = time_curr;
        (void)localtime_r(&time_curr, &sub_item_fault->raise_time);
        fault_global_info.fault_has_change = TRUE;
    }
    return;
}

// 计算滑窗值
static unsigned int calc_sliding_window_value(FAULT_LV2_MAPPING_STRU *sub_item_fault, unsigned int status)
{
    unsigned int ulnum;
    unsigned int slip_status = 0;

    if (sub_item_fault == NULL) {
        FAULT_LOG_ERR("Input args is NULL.");
        return 0;
    }

    // 刷新滑窗
    sub_item_fault->item_cfg.slip_status = (sub_item_fault->item_cfg.slip_status << 1);

    if (status == FAULT_STATUS_ERR) {
        sub_item_fault->item_cfg.slip_status |= 0x1;
    }

    // 计算滑窗值
    ulnum = 0;
    slip_status = sub_item_fault->item_cfg.slip_status;
    for (unsigned int index = 0; index < sub_item_fault->item_cfg.slip_size; index++) {
        if (((slip_status >> index) & 0x1) != 0) {
            ulnum++;
        }
    }
    return ulnum;
}

static unsigned int calc_slip_index(FAULT_LV2_MAPPING_STRU *sub_item_fault, unsigned int status)
{
    unsigned int slip_index_0;

    if (sub_item_fault == NULL) {
        FAULT_LOG_ERR("Input args is NULL.");
        return -1;
    }

    if (status == FAULT_STATUS_ERR) {
        if (sub_item_fault->slip_long_index >= FAULT_SLIP_STATUS_LONG_SIZE) {
            FAULT_LOG_ERR("invalid slip_long_index:%u.", sub_item_fault->slip_long_index);
            return -1;
        }
        sub_item_fault->slip_status_long[sub_item_fault->slip_long_index] = sub_item_fault->slip_seq_id;
        if ((sub_item_fault->slip_long_index + 1) >= (unsigned int)sub_item_fault->item_cfg.slip_occur_thresh) {
            sub_item_fault->slip_long_index = 0;
        } else {
            sub_item_fault->slip_long_index += 1;
        }

        slip_index_0 = sub_item_fault->slip_long_index;
        return slip_index_0;
    }

    if (sub_item_fault->slip_long_index >= 1) {
        slip_index_0 = sub_item_fault->slip_long_index - 1;
    } else {
        slip_index_0 = 0;
        if (sub_item_fault->item_cfg.slip_occur_thresh > 1) {
            slip_index_0 = sub_item_fault->item_cfg.slip_occur_thresh - 1;
        }
    }
    return slip_index_0;
}

// 由当前状态来更新告警信息
static int fault_flush_status(FAULT_LV2_MAPPING_STRU *sub_item_fault, unsigned short status)
{
    unsigned int ulnum;
    unsigned short status_curr = 0;
    unsigned int slip_index_0;

    if (sub_item_fault == NULL) {
        FAULT_LOG_ERR("Input args is NULL.");
        return EDGE_ERR;
    }

    if (sub_item_fault->item_cfg.slip_type == FAULT_SLIP_TYPE_SINGLE) {
        update_sub_item_fault_status(sub_item_fault, status);
        return EDGE_OK;
    }

    if (sub_item_fault->item_cfg.slip_size > FAULT_SLIP_MAX_SIZE) {
        sub_item_fault->slip_seq_id++;
        if (status == FAULT_STATUS_OK && sub_item_fault->fault_status == FAULT_STATUS_OK) {
            return EDGE_OK;
        }

        slip_index_0 = calc_slip_index(sub_item_fault, status);
        if (slip_index_0 >= FAULT_SLIP_STATUS_LONG_SIZE) {
            FAULT_LOG_ERR("invalid slip_index_0:%u.", slip_index_0);
            return EDGE_ERR;
        }

        if (status == FAULT_STATUS_ERR) {
            if (sub_item_fault->slip_status_long[slip_index_0] == 0) { // 初始化为0, 表示还没有赋过值
                return EDGE_OK;
            }
            if ((sub_item_fault->slip_seq_id - sub_item_fault->slip_status_long[slip_index_0]) <=
                sub_item_fault->item_cfg.slip_size) {
                status_curr = FAULT_STATUS_ERR;
            }
        } else {
            if ((sub_item_fault->slip_seq_id - sub_item_fault->slip_status_long[slip_index_0]) >
                sub_item_fault->item_cfg.slip_size) {
                status_curr = FAULT_STATUS_OK;
            }
        }
    } else {
        ulnum = calc_sliding_window_value(sub_item_fault, status);
        if (ulnum >= (unsigned int)sub_item_fault->item_cfg.slip_occur_thresh) {
            status_curr = FAULT_STATUS_ERR;
        } else if (ulnum <= (unsigned int)sub_item_fault->item_cfg.slip_resum_thresh) {
            status_curr = FAULT_STATUS_OK;
        } else {
            return EDGE_OK;
        }
    }

    update_sub_item_fault_status(sub_item_fault, status_curr);
    return EDGE_OK;
}

static int fault_data_report(unsigned char *alarm_buff, unsigned int len)
{
    (void)len;
    if (alarm_buff == NULL) {
        FAULT_LOG_ERR("input null.");
        return EDGE_ERR;
    }

    if (fault_global_info.data_report_func == NULL) {
        FAULT_LOG_ERR("data_report_func is null,no need to call.");
        return EDGE_ERR;
    }

    int ret = fault_global_info.data_report_func(alarm_buff);
    if (ret != EDGE_OK) {
        FAULT_LOG_ERR("data_report_func fail ret:%d.", ret);
        return EDGE_ERR;
    }

    return EDGE_OK;
}

// 处理驱动上报的故障信息并上报给告警模块
static int fault_check_proc(unsigned char *data, unsigned int data_len)
{
    int ret;
    unsigned int buff_len;

    if (data == NULL) {
        FAULT_LOG_ERR("input NULL.");
        return EDGE_ERR;
    }

    if ((data_len < sizeof(FAULT_INFO_HEAD)) || (data_len > MAX_FAULT_DATA_LEN)) {
        FAULT_LOG_ERR("invalid input data_len:%d.", data_len);
        return EDGE_ERR;
    }

    buff_len = data_len + sizeof(ALARM_INFO_HEAD) - sizeof(FAULT_INFO_HEAD);
    if ((buff_len < sizeof(ALARM_INFO_HEAD)) || (buff_len > MAX_FAULT_DATA_LEN)) {
        FAULT_LOG_ERR("invalid len buf_len:%d", buff_len);
        return EDGE_ERR;
    }

    unsigned char *alarm_buff = (unsigned char *)malloc(buff_len);
    if (alarm_buff == NULL) {
        FAULT_LOG_ERR("malloc failed len:%d", buff_len);
        return EDGE_ERR;
    }
    if (memset_s(alarm_buff, buff_len, 0, buff_len) != 0) {
        FAULT_LOG_ERR("memset_s error!");
        free(alarm_buff);
        alarm_buff = NULL;
        return EDGE_ERR;
    }

    FAULT_INFO_HEAD *msg_head = (FAULT_INFO_HEAD *)data;
    ALARM_INFO_HEAD *alarm_msg_head = (ALARM_INFO_HEAD *)(alarm_buff);
    alarm_msg_head->data_len = buff_len;
    alarm_msg_head->item_num = msg_head->item_num;
    alarm_msg_head->owner = msg_head->cmd;

    ret = memcpy_s(alarm_buff + sizeof(ALARM_INFO_HEAD), buff_len - sizeof(ALARM_INFO_HEAD),
        data + sizeof(FAULT_INFO_HEAD), data_len - sizeof(FAULT_INFO_HEAD));
    /* 解决msg_head->item_num为0时(该场景下仍然上报)执行memcpy_s返回非0问题 */
    if ((msg_head->item_num != 0) && (ret != 0)) {
        FAULT_LOG_ERR("memcpy failed ret:%d item_num:%d data_len:%d buff_len:%d.", ret, msg_head->item_num, data_len,
            buff_len);
        free(alarm_buff);
        alarm_buff = NULL;
        return EDGE_ERR;
    }

    ret = fault_data_report(alarm_buff, buff_len);
    if (ret != EDGE_OK) {
        FAULT_LOG_ERR("data_len:%d,item_num:%d.", data_len, msg_head->item_num);
    }
    free(alarm_buff);
    alarm_buff = NULL;

    return ret;
}

typedef int (*DRV_FAULT_CHECK_PROC)(unsigned char *data, unsigned int data_len);

// 初始化故障检测动态库
static int fault_check_so_init(const char *fault_check_so_path)
{
    int ret;
    void *fault_check_handle = NULL;
    int (*fault_check_client_init)(VOID);
    int (*fault_check_callback)(const DRV_FAULT_CHECK_PROC);
    // 存在文件就加载
    if (access(fault_check_so_path, F_OK) == 0) {
        fault_check_handle = safety_dlopen(fault_check_so_path, RTLD_LAZY, COMMON_TRUE, 0);
        if (fault_check_handle == NULL) {
            FAULT_LOG_ERR("load fault_check_so failed,err:%s.", dlerror());
            return EDGE_ERR;
        }
        fault_check_client_init = (int (*)(VOID))dlsym(fault_check_handle, "drv_fault_check_init");
        if (fault_check_client_init == NULL) {
            FAULT_LOG_ERR("fault_check_client_init failed,err:%s.", dlerror());
            goto ERROR_RETURN;
        }
        fault_check_callback =
            (int (*)(const DRV_FAULT_CHECK_PROC))dlsym(fault_check_handle, "drv_fault_check_register_callback");
        if (fault_check_callback == NULL) {
            FAULT_LOG_ERR("drv_fault_check_register_callback failed,err:%s.", dlerror());
            goto ERROR_RETURN;
        }
        ret = fault_check_client_init();
        if (ret != EDGE_OK) {
            FAULT_LOG_ERR("call fault_check_client_init failed.");
            goto ERROR_RETURN;
        }
        ret = fault_check_callback(fault_check_proc);
        if (ret != EDGE_OK) {
            FAULT_LOG_ERR("call fault_check_callback failed.");
            goto ERROR_RETURN;
        }
    } else {
        FAULT_LOG_ERR("fault check lib %s not exist.", fault_check_so_path);
    }
    return EDGE_OK;
ERROR_RETURN:
    dlclose(fault_check_handle);
    return EDGE_ERR;
}

static void fault_check_task_process(void)
{
    int ret;
    unsigned short status_curr = FAULT_STATUS_OK;
    FAULT_LV1_MAPPING_STRU *fault_lv1_item = NULL;
    FAULT_LV2_MAPPING_STRU *sub_item_fault = NULL;

    for (unsigned int index = 0; index < FAULT_LV1_BUTT; index++) {
        fault_lv1_item = &fault_global_info.fault_info_stru[index];
        for (unsigned int sub_index = 0; sub_index < fault_lv1_item->sub_num; sub_index++) {
            sub_item_fault = (FAULT_LV2_MAPPING_STRU *)(fault_lv1_item->sub_fault + sub_index);
            sub_item_fault->run_time += FAULT_CHECK_TASK_DELAY_MS;
            if (sub_item_fault->run_time < sub_item_fault->item_cfg.check_period &&
                (index != FAULT_LV1_CERT_ERR || check_cert_update_flag() != 1)) {
                continue;
            }
            sub_item_fault->run_time = 0;
            if (sub_item_fault->item_cfg.fault_deal_type == FAULT_DEAL_NONE) {
                continue;
            }

            if (sub_item_fault->simulate_flag == TRUE) {
                fault_flush_status(sub_item_fault, sub_item_fault->simulate_value);
                continue;
            }

            if (sub_item_fault->item_cfg.funcdatacollect == NULL) {
                continue;
            }
            ret = sub_item_fault->item_cfg.funcdatacollect(index, sub_item_fault->item_cfg.sub_fault_id, &status_curr);
            if (ret == EDGE_OK) {
                fault_flush_status(sub_item_fault, status_curr);
            }
        }
    }
}

// 告警主任务
static void *fault_manage_check_thread(void *p_void)
{
    int ret;
    unsigned int record_delay = 0;
    (void)p_void;

    prctl(PR_SET_NAME, "fault_manage_check_thread", 0, 0, 0);
    ret = pthread_detach(pthread_self());
    if (ret != 0) {
        FAULT_LOG_ERR("thread_detach fail,ret:%d", ret);
    }

    sleep(FAULT_CHECK_START_DELAY);

    do {
        fault_check_task_process();
        record_delay++;
        if (record_delay >= FAULT_RECORD_DELAY) {
            record_delay = 0;
            if (fault_global_info.fault_has_change == TRUE) {
                fault_global_info.fault_has_change = FALSE;
                (void)report_fault_record_to_alarm();
            }
        }
        (void)usleep(FAULT_CHECK_TASK_DELAY);
    } while (true);
    return NULL;
}

// 磁盘挂载检测任务
static void *fault_manage_mount_check_thread(void *p_void)
{
    int ret;

    (void)p_void;

    prctl(PR_SET_NAME, "mount_check_thread", 0, 0, 0);

    ret = pthread_detach(pthread_self());
    if (ret != 0) {
        FAULT_LOG_ERR("thread_detach fail,ret:%d", ret);
    }

    sleep(FAULT_CHECK_START_DELAY);

    do {
        (void)fault_get_mount_status();
        (void)usleep(FAULT_CHECK_TASK_DELAY * FAULT_SLOW_DELAY);
    } while (true);
    return NULL;
}

// 初始化失败，释放资源
static void fault_manage_release(void)
{
    FAULT_LV1_MAPPING_STRU *fault_lv1_item = NULL;

    for (unsigned int index = 0; index < FAULT_LV1_BUTT; index++) {
        fault_lv1_item = &fault_global_info.fault_info_stru[index];
        if (fault_lv1_item->sub_fault != NULL) {
            free(fault_lv1_item->sub_fault);
            fault_lv1_item->sub_fault = NULL;
        }
    }
}

// 配置证书检测周期
static int get_cert_check_period_configure(const char *sectionName)
{
    int certCheckPeriod = 0;
    if (sectionName == NULL) {
        FAULT_LOG_ERR("inputs null.");
        return certCheckPeriod;
    }

    char certCheckStr[CERT_CHECK_SECURE_LEN] = {0};
    if (access(CERT_CHECKTIME_CONFIGURE_FILE, F_OK) == 0) {
        if (check_cert_security(CERT_CHECKTIME_CONFIGURE_FILE) == VOS_ERR) {
            return certCheckPeriod;
        }
        if (GetJsonKeyString(sectionName, CERT_CHECKTIME_CONFIGURE_FILE, certCheckStr, CERT_CHECK_SECURE_LEN) == VOS_ERR) {
            return certCheckPeriod;
        }
        certCheckPeriod = StrToInt(certCheckStr, CERT_CHECK_SECURE_LEN) * CERT_CHECK_PERIOD_TIME_RATE;
    }
    return certCheckPeriod;
}

/* 对滑窗长度超过32情况，产生门限必现小于6且恢复门限为0 */
static int judge_slipe_status(unsigned short index, unsigned int sub_num, unsigned short fault_id_out,
    const FAULT_LV2_ITEM_CFG *sub_fault, FAULT_LV2_MAPPING_STRU *fault_vl2_item)
{
    if ((fault_vl2_item == NULL) || (sub_fault == NULL)) {
        FAULT_LOG_ERR("inputs null.");
        return EDGE_ERR;
    }

    for (unsigned int sub_index = 0; sub_index < sub_num; sub_index++) {
        if (sub_fault + sub_index == NULL) {
            FAULT_LOG_ERR("malloc fail.sub_index:%d", sub_index);
            return EDGE_ERR;
        }

        /* 对滑窗长度超过32情况，产生门限必现小于6且恢复门限为0 */
        if ((sub_fault->slip_size > FAULT_SLIP_MAX_SIZE) &&
            ((sub_fault->slip_occur_thresh > FAULT_SLIP_STATUS_LONG_SIZE) || (sub_fault->slip_resum_thresh != 0))) {
            FAULT_LOG_ERR("fault(%d-%d) slip error(%d-%d-%d).", index, sub_index, sub_fault->slip_size,
                sub_fault->slip_occur_thresh, sub_fault->slip_resum_thresh);
            return EDGE_ERR;
        }

        FAULT_LV2_MAPPING_STRU *fault_vl2_item_curr = fault_vl2_item + sub_index;
        int ret = memcpy_s(&fault_vl2_item_curr->item_cfg, sizeof(FAULT_LV2_ITEM_CFG), sub_fault + sub_index,
            sizeof(FAULT_LV2_ITEM_CFG));
        if (ret != 0) {
            FAULT_LOG_ERR("memcpy fail.sub_index:%d", sub_index);
            return EDGE_ERR;
        }

        fault_vl2_item_curr->fault_id = index;
        fault_vl2_item_curr->fault_id_out = fault_id_out;
    }
    return EDGE_OK;
}

static int global_fault_info_init(void)
{
    const char *CERT_CHECK_PERIOD = "cert_check_period";
    int ret;
    int cert_period;
    unsigned char sub_num;
    FAULT_LV1_MAPPING_STRU *fault_lv1_item = NULL;
    FAULT_LV1_CFG_STRU *fault_vl1_cfg = NULL;

    for (unsigned short index = 0; index < FAULT_LV1_BUTT; index++) {
        fault_lv1_item = &fault_global_info.fault_info_stru[index];
        fault_vl1_cfg = &fault_lv1_cfg[index];
        sub_num = fault_sub_num_array[index];
        if (index == FAULT_LV1_CERT_ERR) {
            cert_period = get_cert_check_period_configure(CERT_CHECK_PERIOD); // 配置证书检测周期
            if ((cert_period >= MIN_CERT_CHECK_PERIOD) && (cert_period <= MAX_CERT_CHECK_PERIOD)) {
                fault_vl1_cfg->sub_fault->check_period = (unsigned int)cert_period;
            }
        }

        fault_lv1_item->fault_id = index;
        fault_lv1_item->fault_id_out = fault_vl1_cfg->fault_id;
        if ((fault_lv1_item->sub_num = sub_num) == 0) {
            fault_lv1_item->sub_fault = NULL;
            continue;
        }

        FAULT_LV2_MAPPING_STRU *fault_vl2_item = malloc(sizeof(FAULT_LV2_MAPPING_STRU) * sub_num);
        if (fault_vl2_item == NULL) {
            FAULT_LOG_ERR("malloc fail index:%d sub_num:%d ", index, sub_num);
            fault_lv1_item->sub_num = 0;
            fault_lv1_item->sub_fault = NULL;
            return EDGE_ERR;
        }

        if (memset_s(fault_vl2_item, sizeof(FAULT_LV2_MAPPING_STRU) * sub_num, 0,
            sizeof(FAULT_LV2_MAPPING_STRU) * sub_num) != 0) {
            FAULT_LOG_ERR("memset_s error!");
            fault_lv1_item->sub_num = 0;
            fault_lv1_item->sub_fault = NULL;
            free(fault_vl2_item);
            fault_vl2_item = NULL;
            return EDGE_ERR;
        }

        fault_lv1_item->sub_fault = fault_vl2_item;
        ret = judge_slipe_status(index, sub_num, fault_lv1_item->fault_id_out,
            fault_vl1_cfg->sub_fault, fault_vl2_item);
        if (ret == EDGE_ERR) {
            return EDGE_ERR;
        }
    }
    return EDGE_OK;
}

// 故障检测初始化主函数
static int fault_check_init(void)
{
    int ret;
    unsigned long int fault_check_thread = 0;
    unsigned long int mount_check_thread = 0;

    if (fault_global_info.fault_init == TRUE) {
        FAULT_LOG_ERR("fault_init already finished.");
        return EOK;
    }

    // 初始化全局变量g_fault_info_stru
    ret = global_fault_info_init();
    if (ret != EDGE_OK) {
        FAULT_LOG_ERR("global fault info init failed.");
        goto ERROR_RETURN;
    }

    // 磁盘挂载检测任务
    ret = pthread_create(&mount_check_thread, NULL, fault_manage_mount_check_thread, NULL);
    if (ret != 0) {
        FAULT_LOG_ERR("create mount_check thr failed,ret:%d.", ret);
        goto ERROR_RETURN;
    }

    ret = pthread_create(&fault_check_thread, NULL, fault_manage_check_thread, NULL);
    if (ret != 0) {
        FAULT_LOG_ERR("create fault_manage_check failed,ret:%d.", ret);
        goto ERROR_RETURN;
    }

    // 初始化硬件故障检测模块,初始化失败不会影响其他故障检测
    if (fault_check_so_init(DRV_FAULT_CHECK_SO) != EDGE_OK) {
        FAULT_LOG_ERR("drv fault check init failed.");
    }

    // 初始化自定义故障检测模块,初始化失败不会影响其他故障检测
    if (fault_check_so_init(CUSTOMIZED_ALARM_CHECK_SO) != EDGE_OK) {
        FAULT_LOG_ERR("customized alarm check init failed.");
    }

    fault_global_info.fault_init = TRUE;
    return EDGE_OK;
ERROR_RETURN:
    fault_manage_release();
    return EDGE_ERR;
}

// 框架相关
int fault_check_load(void)
{
    ens_intf_import("alarm_report", (void **)&fault_global_info.data_report_func);
    return EDGE_OK;
}

int fault_check_unload(void)
{
    return EDGE_OK;
}

int fault_check_start(void)
{
    return fault_check_init();
}

int fault_check_stop(void)
{
    return EDGE_OK;
}

// 检测sd在位
int fault_get_sd_persent_temp(unsigned int fault_id, unsigned int sub_id, unsigned short *temp)
{
    static char store_dev_persent[FAULT_STORE_BUTT] = {0};

    (void)sub_id;
    (void)fault_id;

    if (temp == NULL) {
        FAULT_LOG_ERR("fault_get_sd_persent_temp value is NULL!");
        return EDGE_ERR;
    }
    int ret;
    ret = access(SD_DEVICE_PATH, F_OK);
    if (ret != 0) {
        /* 不在位,且没有告警 */
        if (store_dev_persent[FAULT_STORE_SD] == 1) {
            /* 之前有在位过，现在不在位了，告警 */
            *temp = 1;
        } else {
            *temp = 0;
        }
    } else {
        /* 在位 */
        if (store_dev_persent[FAULT_STORE_SD] != 1) {
            /* 置在位标识 */
            store_dev_persent[FAULT_STORE_SD] = 1;
        }
        *temp = 0;
    }
    return 0;
}
