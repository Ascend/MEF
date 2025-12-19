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
 * Description: Module operations.
 */

#include <sys/prctl.h>
#include <pthread.h>
#include <unistd.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <sys/prctl.h>
#include <sys/time.h>
#include <errno.h>
#include "securec.h"
#include "string.h"
#include "time.h"
#include "ens_api.h"
#include "file_checker.h"
#include "extend_alarm.h"

static CHECK_ITEM_CFG extend_alarm_check_cfg[EXTEND_ALARM_MAX_NUM] = {
    {0, 0, 2, 0, FAULT_LEVEL_MINOR, 0, 0, 1000, "persent", "M.2", get_disk_persent},
    {1, 0, 2, 0, FAULT_LEVEL_MINOR, 0, 0, 1000, "persent", "HARD_DISK0", get_disk_persent},
    {2, 0, 2, 0, FAULT_LEVEL_MINOR, 0, 0, 1000, "persent", "HARD_DISK1", get_disk_persent},
    {0, 0, 0, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "temperature up anomaly", "M.2", get_hard_disk_temp},
    {1, 0, 0, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "temperature up anomaly", "HARD_DISK0", get_hard_disk_temp},
    {2, 0, 0, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "temperature up anomaly", "HARD_DISK1", get_hard_disk_temp},
    {0, 0, 1, 0, FAULT_LEVEL_MINOR, 0, 0, 3600000, "life expiration", "M.2", get_hard_disk_age},
    {1, 0, 1, 0, FAULT_LEVEL_MINOR, 0, 0, 3600000, "life expiration", "HARD_DISK0", get_hard_disk_age},
    {2, 0, 1, 0, FAULT_LEVEL_MINOR, 0, 0, 3600000, "life expiration", "HARD_DISK1", get_hard_disk_age},
    {3, 3, 0, 0, FAULT_LEVEL_CRITICAL, 0, 0, 6000, "life expiration", "eMMC", check_extcsd_info},
    {13, 13, 0, 0, FAULT_LEVEL_MINOR, 0, 0, 8000, "usb hub fault", "USB_HUB", get_usb_hub_alarm},
    {0, 0, 3, 0, FAULT_LEVEL_CRITICAL, 0, 0, 1000, "operate block", "M.2", get_hard_disk_block},
    {1, 0, 3, 0, FAULT_LEVEL_CRITICAL, 0, 0, 1000, "operate block", "HARD_DISK0", get_hard_disk_block},
    {2, 0, 3, 0, FAULT_LEVEL_CRITICAL, 0, 0, 1000, "operate block", "HARD_DISK1", get_hard_disk_block},
    {14, 14, 0, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "TEEDrv Hardware Module Error", "NPU", get_minid_alarm},
    {14, 14, 1, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000,
     "Multi-bit ECC Error in the TEEDrv Hardware", "NPU", get_minid_alarm},
    {14, 14, 2, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "TS Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 3, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "TS Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 4, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "slogd Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 5, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "dmp_daemon Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 6, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "log-daemon Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 7, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "sklogd Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 8, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "Iammgr Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 9, 0, FAULT_LEVEL_MINOR, 0, 0, 2000,
     "ProcLauncher Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 10, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "ProcMgr Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 11, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "Non-fatal High Temperature", "NPU", get_minid_alarm},
    {14, 14, 12, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000,
     "LPM Sub-system Heartbeat Detection Abnormal", "NPU", get_minid_alarm},
    {14, 14, 13, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000,
     "Abnormal Voltage Adjustment Detected by the LPM", "NPU", get_minid_alarm},
    {14, 14, 14, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000,
     "Abnormal Frequency Adjustment Detected by the LPM", "NPU", get_minid_alarm},
    {14, 14, 15, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000,
     "Chip Current Abnormal Detected by the LPM", "NPU", get_minid_alarm},
    {14, 14, 16, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000,
     "Tsensor Module Exception Detected by the LPM", "NPU", get_minid_alarm},
    {14, 14, 17, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000,
     "PMBUS Module Exception Detected by the LPM", "NPU", get_minid_alarm},
    {14, 14, 18, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "AIC Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 19, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "External Input Error Detected by the AIC", "NPU", get_minid_alarm},
    {14, 14, 20, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "AIC Bus Access Error", "NPU", get_minid_alarm},
    {14, 14, 21, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "AIC Dispatch Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 22, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "AIC Dispatch Input Error", "NPU", get_minid_alarm},
    {14, 14, 23, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "AO Dispatch Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 24, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "AO Dispatch Input Error", "NPU", get_minid_alarm},
    {14, 14, 25, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "TaishanCore Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 26, 0, FAULT_LEVEL_CRITICAL, 0, 0, 2000, "DDR Memory Chip Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 27, 0, FAULT_LEVEL_CRITICAL, 0, 0, 2000, "DDRA Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 28, 0, FAULT_LEVEL_CRITICAL, 0, 0, 2000, "Error Response from DDRC", "NPU", get_minid_alarm},
    {14, 14, 29, 0, FAULT_LEVEL_CRITICAL, 0, 0, 2000, "DDRC Hardware Internal Logic Exception", "NPU", get_minid_alarm},
    {14, 14, 30, 0, FAULT_LEVEL_CRITICAL, 0, 0, 2000, "DDRC Bus Access Error", "NPU", get_minid_alarm},
    {14, 14, 31, 0, FAULT_LEVEL_CRITICAL, 0, 0, 2000, "DDRC Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 32, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000,
     "DDR Chip High Temperature: Non-fatal High Temperature", "NPU", get_minid_alarm},
    {14, 14, 33, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "DVPP Dispatch Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 34, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "DVPP Dispatch Input Error", "NPU", get_minid_alarm},
    {14, 14, 35, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "HSM Key Management Module Error", "NPU", get_minid_alarm},
    {14, 14, 36, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "HSM Cryptographic Algorithm Module Error", "NPU", get_minid_alarm},
    {14, 14, 37, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "HWTS Bus Access Error", "NPU", get_minid_alarm},
    {14, 14, 38, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "HWTS Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 39, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "JPEGD Bus Access Error", "NPU", get_minid_alarm},
    {14, 14, 40, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "JPEGE Hardware Encoding Exception", "NPU", get_minid_alarm},
    {14, 14, 41, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "JPEGE Bus Access Error", "NPU", get_minid_alarm},
    {14, 14, 42, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "L2BUFF Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 43, 0, FAULT_LEVEL_MINOR, 0, 0, 2000,
     "L2BUFF Internal Software Configuration Error", "NPU", get_minid_alarm},
    {14, 14, 44, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "L3D Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 45, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "L3T Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 46, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "NIC Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 47, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "NIC Module Abnormal", "NPU", get_minid_alarm},
    {14, 14, 48, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "PERI Dispatch Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 49, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "PERI Dispatch Input Error", "NPU", get_minid_alarm},
    {14, 14, 50, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "SDMA Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 51, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "SDMA Module Bus Error", "NPU", get_minid_alarm},
    {14, 14, 52, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "VDEC Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 53, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "VENC Hardware Encoding Timeout", "NPU", get_minid_alarm},
    {14, 14, 54, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "VENC Hardware Encoding Exception", "NPU", get_minid_alarm},
    {14, 14, 55, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "VPC Image Processing Hardware Exception", "NPU", get_minid_alarm},
    {14, 14, 56, 0, FAULT_LEVEL_MINOR, 0, 0, 2000,
     "VPC Image Processing Configuration Exception", "NPU", get_minid_alarm},
    {14, 14, 57, 0, FAULT_LEVEL_MINOR, 0, 0, 2000, "VPC Multi-bit ECC Error", "NPU", get_minid_alarm},
    {14, 14, 58, 0, FAULT_LEVEL_MAJOR, 0, 0, 2000, "NPU Overtemperature", "NPU", check_minid_temperature},
    };

EXTEND_STORAGE_CONFIG* extend_storage_configs[FAULT_STORE_BUTT];

// 全局变量
static EXTEND_ALARM_GLOBAL_INFO extend_alarm_global_info = {NULL, 0, {{0}}};

static void *execute_check_cmd(void * cmd)
{
    system((char*)cmd);
    return NULL;
}

static void execute_hard_disk_block_check_cmd(unsigned short fault_id)
{
    char cmd_temp[] = "smartctl -H %s";
    char cmd_info[CMD_SIZE] = {0};
    struct timespec out_time;
    struct timeval now;
    int ret;
    pthread_t tid;
    if (fault_id >= FAULT_STORE_EMMC) {
        return;
    }

    if (strlen(g_store_dev_name[fault_id]) == 0) {
        return;
    }

    // 只开始检测一次，中间故障不处理
    if (g_store_dev_block[fault_id] == FAULT_HARD_DISK_STATUS_BLOCK) {
        return;
    }
    if (sprintf_s(cmd_info, CMD_SIZE, cmd_temp, g_store_dev_name[fault_id]) < 0) {
        EXTEND_ALARM_LOG_ERR("call sprintf_s failed!");
        return;
    }

    gettimeofday(&now, NULL);
    out_time.tv_sec = now.tv_sec + FAULT_HARD_DISK_BLOCK_CNT;
    out_time.tv_nsec = now.tv_usec * NS_CONVERT2_US;
    ret = pthread_create(&tid, NULL, execute_check_cmd, (void*)cmd_info);
    if (ret != 0) {
        EXTEND_ALARM_LOG_ERR("create execute check block cmd pthread error");
        return;
    }
    ret = pthread_timedjoin_np(tid, NULL, &out_time);
    if (ret == ETIMEDOUT) {
        pthread_cancel(tid);
        g_store_dev_block_state[fault_id] = g_store_dev_block_state[fault_id] + 1;
        if (g_store_dev_block_state[fault_id] >= 5) { // 连续5次访问阻塞，标记此设备为阻塞状态
            g_store_dev_block[fault_id] = FAULT_HARD_DISK_STATUS_BLOCK;
            g_store_dev_block_state[fault_id] = 5; // 连续5次访问阻塞，标记此设备为阻塞状态
            EXTEND_ALARM_LOG_ERR("device %s has blocked.", g_store_dev_name[fault_id]);
        }
        return;
    }
    g_store_dev_block[fault_id] = FAULT_HARD_DISK_STATUS_NORMAL;
    if (g_store_dev_block_state[fault_id] > 0) {
        g_store_dev_block_state[fault_id] = g_store_dev_block_state[fault_id] - 1;
    }
    return;
}

// 硬盘阻塞检测
void *extend_storage_block_check_thread()
{
    int ret;
    unsigned short index;

    prctl(PR_SET_NAME, "extend_storage_block_check_thread", 0, 0, 0);

    ret = pthread_detach(pthread_self());
    if (ret != 0) {
        EXTEND_ALARM_LOG_ERR("extend_storage_block_check_thread:thread_detach failed , ret = %d", ret);
    }
    // 初始化时延时2s
    sleep(2);

    do {
        for (index = 0; index < FAULT_STORE_EMMC; index++) {
            execute_hard_disk_block_check_cmd(index);
        }
        sleep(BLOCK_CHECK_CYCLE);
    } while (1);
    return NULL;
}

static void extend_alarm_check_task_process(void)
{
    int ret = 0;
    unsigned short status_curr = FAULT_STATUS_OK;

    for (unsigned int index = 0; index < EXTEND_ALARM_MAX_NUM; index++) {
        CHECK_ITEM_CFG *check_item = &extend_alarm_global_info.check_items[index];
        check_item->run_time += EXTEND_ALARM_CHECK_TASK_DELAY_MS;
        if (check_item->run_time < check_item->check_period) {
            continue;
        }
        if (check_item->funcdatacollect == NULL) {
            continue;
        }
        ret = check_item->funcdatacollect(check_item->fault_id, check_item->sub_fault_id, &status_curr);
        check_item->run_time = 0;
        if (ret != 0) {
            EXTEND_ALARM_LOG_ERR("(%d-%d)check failed.", check_item->fault_id, check_item->sub_fault_id);
            continue;
        }

        if (status_curr != check_item->fault_status) {
            check_item->fault_status = status_curr;
            (&extend_alarm_global_info)->has_changed = 1;
        }
        status_curr = FAULT_STATUS_OK;
    }
}

static int append_item(unsigned char *buff, unsigned int buff_len,
                       unsigned int *offset, CHECK_ITEM_CFG *checkItemCfg)
{
    int ret;
    ALARM_MSG_INFO *item = NULL;

    if ((buff == NULL) || (offset == NULL) || (checkItemCfg == NULL)) {
        EXTEND_ALARM_LOG_ERR("inputs param null.");
        return EXTEND_ALARM_ERROR;
    }
    if (*offset + sizeof(ALARM_MSG_INFO) > buff_len) {
        EXTEND_ALARM_LOG_ERR("size:%ld out of bufflen:%d.", (*offset + sizeof(ALARM_MSG_INFO)), buff_len);
        return EXTEND_ALARM_ERROR;
    }

    item = (ALARM_MSG_INFO *)(buff + *offset);
    if (memset_s(item, sizeof(ALARM_MSG_INFO), 0, sizeof(ALARM_MSG_INFO)) != 0) {
        EXTEND_ALARM_LOG_ERR("memset_s error!");
        return EXTEND_ALARM_ERROR;
    }

    item->fault_id = checkItemCfg->fault_out_id;
    item->sub_fault_id = checkItemCfg->sub_fault_id;
    item->fault_level = checkItemCfg->fault_level;
    item->raise_time_stamp = time(NULL);
    ret = strncpy_s(item->fault_name, sizeof(item->fault_name), checkItemCfg->fault_name,
                    sizeof(item->fault_name) - 1);
    if (ret != 0) {
        EXTEND_ALARM_LOG_ERR("strncpy fault_name failed,ret:%d.", ret);
        return EXTEND_ALARM_ERROR;
    }
    ret = strncpy_s(item->resource, sizeof(item->resource), checkItemCfg->resource,
                    sizeof(item->resource) - 1);
    if (ret != 0) {
        EXTEND_ALARM_LOG_ERR("call strncpy_s failed,ret:%d.", ret);
        return EXTEND_ALARM_ERROR;
    }
    *offset += sizeof(ALARM_MSG_INFO);

    return SUCCESS;
}

// 上报故障到alarm模块
static int report_record_to_alarm(void)
{
    int ret;
    unsigned int active_num = 0;
    const unsigned int buff_len = MAX_FAULT_BUFF_LEN;
    unsigned int offset = sizeof(ALARM_MSG_INFO_HEAD);
    unsigned int items_num = 0;

    unsigned char *fault_buff = (unsigned char *)malloc(buff_len);
    if (fault_buff == NULL) {
        EXTEND_ALARM_LOG_ERR("call malloc failed,len:%u.", buff_len);
        return EXTEND_ALARM_ERROR;
    }
    if (memset_s(fault_buff, buff_len, 0, buff_len) != 0) {
        EXTEND_ALARM_LOG_ERR("call memset_s error!");
        free(fault_buff);
        fault_buff = NULL;
        return EDGE_ERR;
    }

    items_num = sizeof(extend_alarm_global_info.check_items) / sizeof(CHECK_ITEM_CFG);
    for (unsigned int index = 0; index < items_num; index++) {
        CHECK_ITEM_CFG check_item = extend_alarm_global_info.check_items[index];
        if (check_item.fault_status == FAULT_STATUS_OK) {
            continue;
        }
        ret = append_item(fault_buff, buff_len, &offset, &check_item);
        if (ret != EDGE_OK) {
            EXTEND_ALARM_LOG_ERR("call append_item fail(%d).", index);
        } else {
            active_num++;
        }
    }

    ALARM_MSG_INFO_HEAD *msg_head = (ALARM_MSG_INFO_HEAD *)fault_buff;
    msg_head->data_len = offset;
    msg_head->item_num = active_num;
    msg_head->owner = EXTEND_ALARM_MARK;

    if (extend_alarm_global_info.data_report_func) {
        ret = (extend_alarm_global_info.data_report_func)(fault_buff);
    } else {
        EXTEND_ALARM_LOG_ERR("data_report_func is null.");
        ret = EXTEND_ALARM_ERROR;
    }

    free(fault_buff);
    fault_buff = NULL;
    msg_head = NULL;
    if (ret != EDGE_OK) {
        EXTEND_ALARM_LOG_ERR("report_fault_info failed, active_num:%u.", active_num);
    }

    return ret;
}

static void get_dev_name(char *path, char *dev_name, int dev_name_len)
{
    int ret;
    int len = (int)strlen(path);
    int i;
    if (path == NULL || dev_name == NULL) {
        EXTEND_ALARM_LOG_ERR("Input args are null.");
        return;
    }

    for (i = len - 1; i >= 0; i--) {
        if (path[i] == '/') {
            ret = strncpy_s(dev_name, dev_name_len, path + i + 1, strlen(path + i + 1));
            if (ret != 0) {
                EXTEND_ALARM_LOG_ERR("call strncpy_s failed, ret = %d, dev_name = %s.", ret, dev_name);
            }
            return;
        }
    }
}

static int get_block_dev_info(const char *platform, const char *device, char *dev_info, int dev_info_len)
{
    FILE *fp = NULL;
    char line_info[LINE_INFO_SIZE] = {0};
    char dev_path[DEV_PATH_SIZE] = {0};
    int ret;

    if (platform == NULL || device == NULL || dev_info == NULL) {
        EXTEND_ALARM_LOG_ERR("Input arg is NULL.");
        return EXTEND_ALARM_ERROR;
    }

    fp = safety_fopen("/run/block_dev_temp", "r");
    if (fp == NULL) {
        EXTEND_ALARM_LOG_ERR("safety_fopen block temp file failed, fp is NULL.");
        return EXTEND_ALARM_ERROR;
    }

    while (fgets(line_info, sizeof(line_info) - 1, fp) != NULL) {
        ret = sscanf_s(line_info, "%*[^>]> ..%[^'\n']", dev_path, sizeof(dev_path) - 1);
        if (ret <= 0) {
            EXTEND_ALARM_LOG_ERR("call sscanf_s failed, line_info = %s", line_info);
            continue;
        }
        if ((strstr(dev_path, platform) != NULL) && (strstr(dev_path, device) != NULL)) {
            get_dev_name(dev_path, dev_info, dev_info_len);
            fclose(fp);
            return SUCCESS;
        }
    }
    (void)fclose(fp);
    return SUCCESS;
}

static int fault_get_block_mem_dev(void)
{
    int ret;
    unsigned int index;
    char file_buff[32] = {0};
    size_t result_len;
    char block_info[] = "ls -l /sys/block/* > /run/block_dev_temp";

    if (check_file_path_valid("/run/block_dev_temp") != EDGE_OK) {
        EXTEND_ALARM_LOG_ERR("block_dev_temp path is not valid");
        return EXTEND_ALARM_ERROR;
    }
    (void)system(block_info);
    for (index = 0; index < FAULT_STORE_BUTT; index++) {
        if (index == FAULT_STORE_EMMC || index == FAULT_STORE_SD || extend_storage_configs[index] == NULL) {
            continue;
        }
        (void)get_block_dev_info(extend_storage_configs[index]->platform, extend_storage_configs[index]->device,
                                 file_buff, sizeof(file_buff) - 1);
        result_len = strlen(file_buff);
        if (result_len == 0) {
            ret = memset_s(g_store_dev_name[index], sizeof(g_store_dev_name[index]), 0,
                           sizeof(g_store_dev_name[index]));
            if (ret != 0) {
                EXTEND_ALARM_LOG_ERR("call memset_s failed.");
                continue;
            }
        } else {
            file_buff[result_len - 1] = (file_buff[result_len - 1] == '\n') ? '\0' : file_buff[result_len - 1];
            ret = sprintf_s(g_store_dev_name[index], sizeof(g_store_dev_name[index]) - 1, "/dev/%s", file_buff);
            if (ret < 0) {
                EXTEND_ALARM_LOG_ERR("call sprintf_s failed, ret=%d.", ret);
            }
        }
        ret = memset_s(file_buff, sizeof(file_buff), 0, sizeof(file_buff));
        if (ret != 0) {
            EXTEND_ALARM_LOG_ERR("call memset_s failed. ret=%d", ret);
        }
    }

    (void)unlink("/run/block_dev_temp");
    return SUCCESS;
}

static void *extend_alarm_check_thread(void *p_void)
{
    int ret;
    unsigned int record_delay = 0;
    unsigned int slow_delay = 0;
    (void)p_void;

    prctl(PR_SET_NAME, "extend_alarm_check_thread", 0, 0, 0);
    ret = pthread_detach(pthread_self());
    if (ret != 0) {
        EXTEND_ALARM_LOG_ERR("extend_alarm_check_thread thread_detach failed,ret:%d", ret);
    }
    sleep(EXTEND_ALARM_CHECK_START_DELAY);

    do {
        extend_alarm_check_task_process();
        record_delay++;
        if (record_delay >= EXTEND_ALARM_RECORD_DELAY) {
            record_delay = 0;
            if (extend_alarm_global_info.has_changed == 1) {
                extend_alarm_global_info.has_changed = 0;
                (void)report_record_to_alarm();
            }
        }
        // 查询存储设备信息
        slow_delay++;
        if (slow_delay >= EXTEND_ALARM_SLOW_DELAY) {
            slow_delay = 0;
            (void)fault_get_block_mem_dev();
        }
        (void)usleep(EXTEND_ALARM_CHECK_TASK_DELAY);
    } while (1);
    return NULL;
}

static int get_index_by_location(const char *location)
{
    if (strcmp(location, "PCIE-0") == 0) {
        return FAULT_STORE_M2;
    } else if (strcmp(location, "PCIE-1") == 0) {
        return FAULT_STORE_HARD_DISK0;
    } else if (strcmp(location, "PCIE-2") == 0) {
        return FAULT_STORE_HARD_DISK1;
    } else if (strcmp(location, "eMMC1") == 0) {
        return FAULT_STORE_EMMC;
    } else if (strcmp(location, "SDIO1") == 0) {
        return FAULT_STORE_SD;
    } else if (strcmp(location, "usb0") == 0) {
        return FAULT_STORE_USB0;
    } else if (strcmp(location, "usb1") == 0) {
        return FAULT_STORE_USB1;
    } else if (strcmp(location, "usb2") == 0) {
        return FAULT_STORE_USB2;
    } else {
        return EXTEND_ALARM_ERROR;
    }
}

static int parse_storage_config(char *info, EXTEND_STORAGE_CONFIG *config)
{
    size_t field_len = 0;
    char *save_ptr = NULL;
    char *key = NULL;
    char *value = NULL;

    if (info == NULL || config == NULL) {
        EXTEND_ALARM_LOG_ERR("param is null.");
        return EXTEND_ALARM_ERROR;
    }

    key = strtok_r(info, "=", &save_ptr);
    value = strtok_r(NULL, "=", &save_ptr);
    if (key == NULL || value == NULL) {
        EXTEND_ALARM_LOG_ERR("key_value is null.");
        return EXTEND_ALARM_ERROR;
    }
    field_len = strlen(value);
    if (field_len > 0 && value[field_len - 1] == '\n') {
        value[field_len - 1] = '\0';
    }

    if (strcmp(key, "platform") == 0) {
        if (strcpy_s(config->platform, CONFIG_FIELD_LENGTH - 1, value) != 0) {
            return EXTEND_ALARM_ERROR;
        }
    } else if (strcmp(key, "device") == 0) {
        if (strcpy_s(config->device, CONFIG_FIELD_LENGTH - 1, value) != 0) {
            return EXTEND_ALARM_ERROR;
        }
    } else if (strcmp(key, "location") == 0) {
        if (strcpy_s(config->location, CONFIG_FIELD_LENGTH - 1, value) != 0) {
            return EXTEND_ALARM_ERROR;
        }
    } else if (strcmp(key, "name") == 0) {
        if (strcpy_s(config->name, CONFIG_FIELD_LENGTH - 1, value) != 0) {
            return EXTEND_ALARM_ERROR;
        }
    }
    return SUCCESS;
}

static int generate_extend_storage_config(char* line_info, int line_len)
{
    char line_temp[LINE_INFO_SIZE] = {0};
    char *temp = NULL;
    char *save_ptr = NULL;
    char *usb_hub_id = NULL;
    if (line_info == NULL || line_len > LINE_INFO_SIZE || line_len < 0) {
        EXTEND_ALARM_LOG_ERR("param of line length input failed.");
        return EXTEND_ALARM_ERROR;
    }
    if (strstr(line_info, "usb_hub_id")) {
        (void)strtok_r(line_info, "=", &save_ptr);
        usb_hub_id = strtok_r(NULL, "=", &save_ptr);
        if (usb_hub_id == NULL) {
            EXTEND_ALARM_LOG_WARN("usb_hub_id is null.");
            return SUCCESS;
        }
        size_t result_len = strlen(usb_hub_id);
        if (result_len > 0) {
            usb_hub_id[result_len - 1] = (usb_hub_id[result_len - 1] == '\n') ? '\0' : usb_hub_id[result_len - 1];
        }

        if (strcpy_s(USB_HUB_ID, sizeof(USB_HUB_ID) - 1, usb_hub_id) != 0) {
            EXTEND_ALARM_LOG_ERR("strcpy_s usb_hub_id failed.");
            return EXTEND_ALARM_ERROR;
        }
        return SUCCESS;
    }

    if (strcpy_s(line_temp, LINE_INFO_SIZE, line_info) != 0) {
        EXTEND_ALARM_LOG_ERR("strcpy_s line_info failed.");
        return EXTEND_ALARM_ERROR;
    }
    EXTEND_STORAGE_CONFIG* config = (EXTEND_STORAGE_CONFIG*)malloc(sizeof(EXTEND_STORAGE_CONFIG));
    if (config == NULL) {
        EXTEND_ALARM_LOG_ERR("malloc config failed");
        return EXTEND_ALARM_ERROR;
    }
    if (memset_s(config, sizeof(EXTEND_STORAGE_CONFIG), 0, sizeof(EXTEND_STORAGE_CONFIG)) != 0) {
        EXTEND_ALARM_LOG_ERR("memset_s failed.");
        free(config);
        config = NULL;
        return EXTEND_ALARM_ERROR;
    }

    temp = strtok_r(line_info, ",", &save_ptr);
    while (temp != NULL) {
        if (parse_storage_config(temp, config) != SUCCESS) {
            EXTEND_ALARM_LOG_ERR("parse_storage_config failed.");
            free(config);
            config = NULL;
            return EXTEND_ALARM_ERROR;
        }
        temp = strtok_r(NULL, ",", &save_ptr);
    }

    int index = get_index_by_location(config->location);
    if (index == EXTEND_ALARM_ERROR) {
        EXTEND_ALARM_LOG_ERR("location is wrong.");
        free(config);
        config = NULL;
        return EXTEND_ALARM_ERROR;
    }

    if (extend_storage_configs[index] != NULL) {
        EXTEND_ALARM_LOG_ERR("location is exists.");
        free(config);
        config = NULL;
        return EXTEND_ALARM_ERROR;
    }
    extend_storage_configs[index] = config;
    EXTEND_ALARM_LOG_INFO("generate extend storage config success.");
    return SUCCESS;
}

static int init_extend_storage_config()
{
    FILE *fp = NULL;
    char line_info[LINE_INFO_SIZE] = {0};
    int count = 0;

    fp = safety_fopen("/run/formated_hw.info", "r");
    if (fp == NULL) {
        EXTEND_ALARM_LOG_ERR("open /run/formated_hw.info failed, fp is NULL.");
        return EXTEND_ALARM_ERROR;
    }

    while (fgets(line_info, sizeof(line_info) - 1, fp) != NULL) {
        if (count > MAX_LINE_COUNT) {
            EXTEND_ALARM_LOG_ERR("/run/formated_hw.info content is invalid.");
            (void)fclose(fp);
            return EXTEND_ALARM_ERROR;
        }
        if (generate_extend_storage_config(line_info, LINE_INFO_SIZE) != SUCCESS) {
            EXTEND_ALARM_LOG_ERR("generate extend storage failed.");
            fclose(fp);
            return EXTEND_ALARM_ERROR;
        }
        count++;
    }

    (void)fclose(fp);
    return SUCCESS;
}

static int clean_extend_storage_config(void)
{
    if (extend_storage_configs == NULL) {
        return SUCCESS;
    }

    for (int i = 0; i < FAULT_STORE_BUTT; ++i) {
        if (extend_storage_configs[i] != NULL) {
            free(extend_storage_configs[i]);
            extend_storage_configs[i] = NULL;
        }
    }
    return SUCCESS;
}


int hal_init_dcmi_api(void)
{
    int ret;
    void *p_dcmi_lib = safety_dlopen(DCMI_DRIVER_LIB, RTLD_GLOBAL | RTLD_LAZY, COMMON_TRUE, ROOT_UID);
    if (p_dcmi_lib == NULL) {
        EXTEND_ALARM_LOG_ERR("failed to dlopen %s, dcmi APIs will not be available, err: %s.",
                             DCMI_DRIVER_LIB, dlerror());
        return EXTEND_ALARM_ERROR;
    }

    g_dcmi_infs.pfn_dcmi_init = dlsym(p_dcmi_lib, "dcmi_init");
    if (g_dcmi_infs.pfn_dcmi_init == NULL) {
        EXTEND_ALARM_LOG_ERR("failed to load init function, err: %s.", dlerror());
        goto LABEL_ERROR;
    }

    ret = g_dcmi_infs.pfn_dcmi_init();
    if (ret != 0) {
        EXTEND_ALARM_LOG_ERR("dcmi_init failed. ret is %d", ret);
        goto LABEL_ERROR;
    }

    g_dcmi_infs.pfn_get_device_num_in_card = dlsym(p_dcmi_lib, "dcmi_get_device_num_in_card");
    if (g_dcmi_infs.pfn_get_device_num_in_card == NULL) {
        EXTEND_ALARM_LOG_ERR("failed to load dcmi_get_device_num_in_card function, err: %s.", dlerror());
        goto LABEL_ERROR;
    }

    g_dcmi_infs.pfn_get_device_health = dlsym(p_dcmi_lib, "dcmi_get_device_health");
    if (g_dcmi_infs.pfn_get_device_health == NULL) {
        EXTEND_ALARM_LOG_ERR("failed to load dcmi_get_device_health function, err: %s.", dlerror());
        goto LABEL_ERROR;
    }

    g_dcmi_infs.pfn_get_device_temperature = dlsym(p_dcmi_lib, "dcmi_get_device_temperature");
    if (g_dcmi_infs.pfn_get_device_temperature == NULL) {
        EXTEND_ALARM_LOG_ERR("failed to load dcmi_get_device_temperature function, err: %s.", dlerror());
        goto LABEL_ERROR;
    }

    g_dcmi_infs.pfn_get_device_errorcode = dlsym(p_dcmi_lib, "dcmi_get_device_errorcode_v2");
    if (g_dcmi_infs.pfn_get_device_errorcode == NULL) {
        EXTEND_ALARM_LOG_ERR("failed to load dcmi_get_device_errorcode function, err: %s.", dlerror());
        goto LABEL_ERROR;
    }

    EXTEND_ALARM_LOG_INFO("hal_init_dcmi_api success!");
    return SUCCESS;

LABEL_ERROR:
    dlclose(p_dcmi_lib);
    return EXTEND_ALARM_ERROR;
}

// 故障检测初始化主函数
static int extend_alarm_check_init(void)
{
    int ret;
    unsigned long int ext_storage_check = 0;
    unsigned long int ext_storage_block_check = 0;

    // 初始化扩展存储配置
    ret = init_extend_storage_config();
    if (ret != SUCCESS) {
        EXTEND_ALARM_LOG_ERR("init_extend_storage_config failed, ret = %d.", ret);
        goto ERROR_RETURN;
    }

    // 初始化dcmi
    ret = hal_init_dcmi_api();
    if (ret != 0) {
        EXTEND_ALARM_LOG_ERR("hal_init_dcmi_api failed, ret = %d.", ret);
        goto ERROR_RETURN;
    }

    // 初始化全局变量
    ret = memcpy_s(extend_alarm_global_info.check_items, sizeof(extend_alarm_global_info.check_items),
                   extend_alarm_check_cfg, sizeof(extend_alarm_global_info.check_items));
    if (ret != SUCCESS) {
        EXTEND_ALARM_LOG_ERR("init global info failed, reason: memcpy_s failed, ret = %d", ret);
        goto ERROR_RETURN;
    }
    // 更新硬盘阻塞状态
    ret = pthread_create(&ext_storage_block_check, NULL, extend_storage_block_check_thread, NULL);
    if (ret != SUCCESS) {
        EXTEND_ALARM_LOG_ERR("create thread of disk block check:  pthread create failed, ret = %d.", ret);
        goto ERROR_RETURN;
    }
    // 检测主任务
    ret = pthread_create(&ext_storage_check, NULL, extend_alarm_check_thread, NULL);
    if (ret != SUCCESS) {
        EXTEND_ALARM_LOG_ERR("create disk check main task thread failed,ret:%d.", ret);
        goto ERROR_RETURN;
    }
    return SUCCESS;
ERROR_RETURN:
    clean_extend_storage_config();
    return EXTEND_ALARM_ERROR;
}

// 框架相关
int extend_alarm_load(void)
{
    ens_intf_import("alarm_report", (void **)&extend_alarm_global_info.data_report_func);
    return SUCCESS;
}

int extend_alarm_unload(void)
{
    return SUCCESS;
}

int extend_alarm_start(void)
{
    return extend_alarm_check_init();
}

int extend_alarm_stop(void)
{
    return SUCCESS;
}

