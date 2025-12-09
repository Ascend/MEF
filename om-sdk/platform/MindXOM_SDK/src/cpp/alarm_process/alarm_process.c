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
#include "alarm_process.h"
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <time.h>
#include <sys/types.h>
#include <sys/stat.h>
#include "securec.h"
#include "ens_api.h"
#include "file_checker.h"

#define ALARM_LEVEL_DEFAULT 4
#define FAULT_ACTIVE_FILE_NAME "/run/all_active_alarm"
#define ALAM_FILE_MAX_SIZE (1024 * 1024)
#define ALARM_OFFSET_BITS 16
#define ALAM_FILE_HEAD_SIZE (64)

ALARM_GLOBAL_INFO g_alarm_global_info = { { 0 }, { NULL }, PTHREAD_MUTEX_INITIALIZER };

// 清除指定归属的故障状态
static void clear_fault_status(ALARM_ATTR_MARK flag)
{
    for (unsigned short i = 0; i < FAULT_LV1_MAX_SIZE; i++) {
        ALARM_LV1_STRU *cur = g_alarm_global_info.alarm_info_stru[i];
        while (cur != NULL) {
            if (cur->fault_attr_mark == flag) {
                cur->fault_status = ALARM_STATUS_OK;
            }
            cur = cur->next;
        }
    }
}

// 写入buff信息到文件
static int write_buff_to_file(const char *buf_block, size_t buf_len, const unsigned int active_num)
{
    char buf[ALAM_FILE_HEAD_SIZE + 1] = {0};
    if (sprintf_s(buf, ALAM_FILE_HEAD_SIZE, "0@%u@aabb\n", active_num) < 0) {
        ALARM_LOG_ERR("sprintf_s failed");
        return EDGE_ERR;
    }

    FILE *result_file = safety_fopen(FAULT_ACTIVE_FILE_NAME, "w+");
    if (result_file == NULL) {
        ALARM_LOG_ERR("failed open %s.", FAULT_ACTIVE_FILE_NAME);
        return EDGE_ERR;
    }

    if (safety_chmod_by_fd(result_file, S_IRUSR | S_IWUSR) != COMMON_SUCCESS) {
        ALARM_LOG_ERR("chmod result_file failed");
    }

    buf[ALAM_FILE_HEAD_SIZE] = '\0';
    if (fwrite(buf, sizeof(char), strlen(buf), result_file) < strlen(buf)) {
        ALARM_LOG_ERR("fwrite failed");
        (void)fclose(result_file);
        return EDGE_ERR;
    }

    if ((buf_block != NULL) && (buf_len == ALAM_FILE_MAX_SIZE) && (buf_block[buf_len - 1] == '\0')) {
        if (fwrite(buf_block, sizeof(char), strlen(buf_block), result_file) < strlen(buf_block)) {
            ALARM_LOG_ERR("fwrite failed");
        }
    } else {
        ALARM_LOG_ERR("fwrite failed, buf_len=%ld", buf_len);
    }

    if (fflush(result_file) != 0) {
        ALARM_LOG_ERR("fflush error");
        (void)fclose(result_file);
        return EDGE_ERR;
    }
    (void)fclose(result_file);
    return EDGE_OK;
}

static void traversal_fault_lv1_item(ALARM_LV1_STRU *fault_lv1_item, char **buf_curr,
                                     unsigned int *active_num_tmp, int *alarm_level)
{
    int size = 0;
    int ret;

    if ((buf_curr == NULL) || (*buf_curr == NULL) || (active_num_tmp == NULL) || (alarm_level == NULL)) {
        ALARM_LOG_ERR("invalid inputs.");
        return;
    }

    while (fault_lv1_item != NULL) {
        if (fault_lv1_item->shield_alarm_flag == SHILED_FLAG_UNSHIELDED &&
            fault_lv1_item->fault_status == ALARM_STATUS_ERR) {
            ret = sprintf_s(*buf_curr, ALAM_FILE_MAX_SIZE - size, "%08X@%s@%s@%lu@%d@aabb\n",
                ((fault_lv1_item->fault_info.fault_id << ALARM_OFFSET_BITS) +
                fault_lv1_item->fault_info.sub_fault_id),
                fault_lv1_item->fault_info.fault_name, fault_lv1_item->fault_info.resource,
                fault_lv1_item->fault_info.raise_time_stamp, fault_lv1_item->fault_info.fault_level);
            if (ret < 0) {
                ALARM_LOG_ERR("call sprintf_s fail.");
            } else if (*alarm_level > fault_lv1_item->fault_info.fault_level) {
                *buf_curr += ret;
                size += ret;
                (*active_num_tmp)++;
                *alarm_level = fault_lv1_item->fault_info.fault_level;
            } else {
                *buf_curr += ret;
                size += ret;
                (*active_num_tmp)++;
            }
        }
        fault_lv1_item = fault_lv1_item->next;
    }
    return;
}
// 根据故障状态生成告警文件
int fault_record_to_file(unsigned int *active_num, int *alarm_level)
{
    int ret;
    char *buf_block = NULL;
    char *buf_curr = NULL;
    unsigned int active_num_tmp = 0;
    int alarm_level_tmp = ALARM_LEVEL_DEFAULT;
    ALARM_LV1_STRU *fault_lv1_item = NULL;

    if (alarm_level == NULL) {
        ALARM_LOG_ERR("input alarm_level null.");
        return EDGE_ERR;
    }
    if (active_num == NULL) {
        ALARM_LOG_ERR("input active_num null.");
        return EDGE_ERR;
    }
    *active_num = 0;

    buf_block = malloc(ALAM_FILE_MAX_SIZE);
    if (buf_block == NULL) {
        ALARM_LOG_ERR("call malloc fail.");
        return EDGE_ERR;
    }
    if (memset_s(buf_block, ALAM_FILE_MAX_SIZE, 0, ALAM_FILE_MAX_SIZE) != 0) {
        free(buf_block);
        ALARM_LOG_ERR("memset_s fail.");
        return EDGE_ERR;
    }
    buf_curr = buf_block;
    for (unsigned int index = 0; index < FAULT_LV1_MAX_SIZE; index++) {
        fault_lv1_item = g_alarm_global_info.alarm_info_stru[index];
        traversal_fault_lv1_item(fault_lv1_item, &buf_curr, &active_num_tmp, &alarm_level_tmp);
    }
    ret = write_buff_to_file(buf_block, ALAM_FILE_MAX_SIZE, active_num_tmp);
    free(buf_block);
    buf_block = NULL;
    buf_curr = NULL;
    *active_num = active_num_tmp;
    *alarm_level = alarm_level_tmp;
    return ret;
}
// 创建新的告警节点
ALARM_LV1_STRU *create_new_alarm()
{
    ALARM_LV1_STRU *cur_fault_stru = malloc(sizeof(ALARM_LV1_STRU));
    if (cur_fault_stru == NULL) {
        ALARM_LOG_ERR("report_fault_info:cur_fault_stru malloc failed.");
        return cur_fault_stru;
    }
    if (memset_s(cur_fault_stru, sizeof(ALARM_LV1_STRU), 0, sizeof(ALARM_LV1_STRU)) != 0) {
        ALARM_LOG_ERR("memset_s error!");
        free(cur_fault_stru);
        cur_fault_stru = NULL;
        return cur_fault_stru;
    }
    cur_fault_stru->next = NULL;
    return cur_fault_stru;
}

// 根据cur_fault_info的信息到g_fault_stru中找到对应的节点，没有找到就新增，失败返回NULL
ALARM_LV1_STRU *search_item_by_info(const ALARM_MSG_FAULT_INFO *cur_fault_info, int is_shield)
{
    ALARM_LV1_STRU *cur_fault_stru = NULL;
    ALARM_LV1_STRU *pre_fault_stru = NULL;
    int is_find = 0;
    int count = 0;

    if (cur_fault_info == NULL) {
        ALARM_LOG_ERR("input cur_fault_info null.");
        return NULL;
    }
    if (cur_fault_info->fault_id >= FAULT_LV1_MAX_SIZE) {
        ALARM_LOG_ERR("fault_id(%d) out of range.", cur_fault_info->fault_id);
        return NULL;
    }

    if (cur_fault_info->sub_fault_id >= FAULT_LV2_MAX_SIZE) {
        ALARM_LOG_ERR("sub_fault_id(%d) out of range.", cur_fault_info->sub_fault_id);
        return NULL;
    }

    cur_fault_stru = g_alarm_global_info.alarm_info_stru[cur_fault_info->fault_id];
    while (cur_fault_stru != NULL) {
        if (count >= FAULT_LV2_MAX_SIZE) {
            ALARM_LOG_ERR("subItems len is too big.");
            return NULL;
        }
        char *p = cur_fault_stru->fault_info.resource;
        if (cur_fault_stru->fault_info.sub_fault_id == cur_fault_info->sub_fault_id &&
            (strncmp(p, cur_fault_info->resource, strlen(cur_fault_info->resource)) == 0)) {
            is_find = 1;
            break;
        }
        pre_fault_stru = cur_fault_stru;
        cur_fault_stru = cur_fault_stru->next;
        count++;
    }

    // 没找到节点需要新增
    if (cur_fault_stru == NULL) {
        cur_fault_stru = create_new_alarm();
        if (cur_fault_stru == NULL) {
            return cur_fault_stru;
        }
    }

    // cur_fault_stru->fault_info和cur_fault_info都是ALARM_MSG_FAULT_INFO类型，无需判断返回值
    // 屏蔽告警找到对应告警实体, 不需要copy
    if (is_find != 1 || is_shield != 1) {
        (void)memcpy_s(&cur_fault_stru->fault_info, sizeof(ALARM_MSG_FAULT_INFO),
                       cur_fault_info, sizeof(ALARM_MSG_FAULT_INFO));
    }

    if (pre_fault_stru == NULL) {
        g_alarm_global_info.alarm_info_stru[cur_fault_info->fault_id] = cur_fault_stru;
    } else {
        pre_fault_stru->next = cur_fault_stru;
    }
    return cur_fault_stru;
}

static void process_fault_info(ALARM_MSG_FAULT_INFO *fault_info, unsigned int owner)
{
    ALARM_LV1_STRU *fault_info_node = search_item_by_info(fault_info, 0);
    if (fault_info_node == NULL) {
        ALARM_LOG_ERR("search item by info faild.owner:%u.", owner);
        return;
    }
    fault_info_node->fault_attr_mark = owner;
    fault_info_node->fault_status = ALARM_STATUS_ERR;
    return;
}

void publish_alarm_event(int alarm_level)
{
    for (int i = 0; i < MAX_REGISTER_FUN_NUM; i++) {
        if (g_alarm_global_info.alarm_event_process_funcs[i] != NULL) {
            g_alarm_global_info.alarm_event_process_funcs[i](alarm_level);
        }
    }
    return;
}

int subscribe_alarm_event(PROCESS_ALARM_EVENT_FUNC p_func)
{
    int id_x = -1;
    for (int i = 0; i < MAX_REGISTER_FUN_NUM; i++) {
        if (g_alarm_global_info.alarm_event_process_funcs[i] == p_func) {
            return EDGE_OK;
        }
        if (g_alarm_global_info.alarm_event_process_funcs[i] == NULL) {
            id_x = i;
            g_alarm_global_info.alarm_event_process_funcs[i] = p_func;
            break;
        }
    }
    if (id_x == -1) {
        ALARM_LOG_ERR("subscribe_alarm_event, Maximum number has been reached.");
        return EDGE_ERR;
    }
    return EDGE_OK;
}


// 处理上报的故障
int alarm_report(const unsigned char *info)
{
    int ret;
    unsigned int active_num = 0;
    unsigned int item_num = 0;
    int alarm_level = ALARM_LEVEL_DEFAULT;
    ALARM_MSG_INFO_HEAD *msg_head = NULL;
    ALARM_MSG_FAULT_INFO *cur_fault_info = NULL;
    if (info == NULL) {
        ALARM_LOG_ERR("invalid info null.");
        return EDGE_ERR;
    }
    msg_head = (ALARM_MSG_INFO_HEAD *)info;
    item_num = msg_head->item_num;

    if (msg_head->data_len != sizeof(ALARM_MSG_INFO_HEAD) + item_num * sizeof(ALARM_MSG_FAULT_INFO)) {
        ALARM_LOG_ERR("input data error,input data_len=%d,len=%lu.",
                      msg_head->data_len,
                      sizeof(ALARM_MSG_INFO_HEAD) + item_num * sizeof(ALARM_MSG_FAULT_INFO));
        return EDGE_ERR;
    }

    pthread_mutex_lock(&g_alarm_global_info.mutex);
    clear_fault_status(msg_head->owner);
    // 循环处理每个故障信息,更新状态
    for (unsigned int i = 0; i < item_num; i++) {
        cur_fault_info =
            (ALARM_MSG_FAULT_INFO *)(info + sizeof(ALARM_MSG_INFO_HEAD) + i * sizeof(ALARM_MSG_FAULT_INFO));
        process_fault_info(cur_fault_info, msg_head->owner);
    }
    // 根据故障状态写文件
    ret = fault_record_to_file(&active_num, &alarm_level);
    if (ret != EDGE_OK) {
        pthread_mutex_unlock(&g_alarm_global_info.mutex);
        ALARM_LOG_ERR("record to file failed.");
        return EDGE_ERR;
    }
    publish_alarm_event(alarm_level);
    pthread_mutex_unlock(&g_alarm_global_info.mutex);
    return EDGE_OK;
}


// ens框架相关
int alarm_process_load(void)
{
    ens_intf_export("alarm_report", (void *)alarm_report);
    ens_intf_export("subscribe_alarm_event", (void *)subscribe_alarm_event);
    return EDGE_OK;
}

int alarm_process_unload(void)
{
    return EDGE_OK;
}

int alarm_process_start()
{
    static int binit = 0;
    int ret;
    unsigned long int fault_shield_alarm_thread = 0;
    if (binit == 1) {
        ALARM_LOG_ERR("alarm process already start.");
        return EDGE_OK;
    }

    pthread_mutex_init(&g_alarm_global_info.mutex, NULL);
    (void)set_clear_shield_alarm_flag(SHIELD_ALARM_FILE, SHIELD_ALARM_TYPE);
    ret = pthread_create(&fault_shield_alarm_thread, NULL, fault_manager_shield_alarm_thread, NULL);
    if (ret != 0) {
        pthread_mutex_destroy(&g_alarm_global_info.mutex);
        ALARM_LOG_ERR("pthread create failed,ret=%d.", ret);
        return EDGE_ERR;
    }

    binit = 1;
    return EDGE_OK;
}

int alarm_process_stop(void)
{
    return EDGE_OK;
}