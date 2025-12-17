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
 * Description: 告警屏蔽功能实现
 */
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <errno.h>
#include <time.h>
#include <pthread.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/inotify.h>
#include <sys/prctl.h>
#include "securec.h"
#include "alarm_process.h"
#include "file_checker.h"

#define ALARM_INFO_SIZE 256
#define INOTIFY_BUFF_MAX_LEN 4096
#define ALARM_INFO_NUMS 5
#define ALARM_LEVEL_DEFAULT 4
#define FAULT_ID_OUT_MASK 0xffff0000
#define MOVE_STEPS 16
#define TIME_INTERVAL 2
#define MAX_LINE_LENGTH 1024

int g_shield_alarm_flag = 0;

/* ****************************************************************************
 函 数 名  : set_clear_shield_by_fault_info
 功能描述  : 按照告警屏蔽配置文件，配置告警屏蔽位。
**************************************************************************** */
static int set_clear_shield_by_fault_info(const ALARM_MSG_FAULT_INFO *p_fault_info, int alarm_option_type)
{
    if (p_fault_info == NULL) {
        ALARM_LOG_ERR("check failed for p_fault_info null");
        return EDGE_ERR;
    }

    ALARM_LV1_STRU *fault_lv1_item = search_item_by_info(p_fault_info, 1);
    if (fault_lv1_item == NULL) {
        ALARM_LOG_ERR("search item failed.");
        return EDGE_ERR;
    }

    if (alarm_option_type == CLEAR_ALARM_TYPE) {
        fault_lv1_item->clear_alarm_flag = 1;
    } else if (alarm_option_type == SHIELD_ALARM_TYPE) {
        fault_lv1_item->shield_alarm_flag = SHILED_FLAG_SHIELDED;
    } else {
        ALARM_LOG_ERR("invalid alarm type:%d.", alarm_option_type);
    }
    return EDGE_OK;
}

/* ****************************************************************************
 函 数 名  : set_clear_shield_alarm_flag
 功能描述  : 按照告警屏蔽配置文件，配置告警屏蔽位。
 输入参数  : filename：告警屏蔽配置文件 alarm_option_type：类型
 输出参数  : 无
 返 回 值  : 成功:EDGE_OK 失败：EDGE_ERR
**************************************************************************** */
int set_clear_shield_alarm_flag(const char *filename, int alarm_option_type)
{
    FILE *fd = NULL;
    unsigned int level;
    unsigned int fault_id, sub_fault_id;
    ALARM_MSG_FAULT_INFO cur_fault_info = {0};
    char innerid[RESOURCE_MAX_LEN] = {0};
    char alarm_info_buf[ALARM_INFO_SIZE + 1] = {0};

    if (!(fd = safety_fopen(filename, "r"))) {
        ALARM_LOG_ERR("open path failed");
        return EDGE_ERR;
    }

    for (int i = 0; i < MAX_LINE_LENGTH; i++) {
        if (fgets(alarm_info_buf, ALARM_INFO_SIZE - 1, fd) == NULL) {
            break;
        }
        if (!strstr(alarm_info_buf, "@aabb")) {
            continue;
        }

        if ((memset_s(innerid, RESOURCE_MAX_LEN, 0, RESOURCE_MAX_LEN) != 0) ||
            (memset_s(&cur_fault_info, sizeof(cur_fault_info), 0, sizeof(cur_fault_info)) != 0)) {
            ALARM_LOG_ERR("memset_s error!");
            continue;
        }
        
        int ret = sscanf_s(alarm_info_buf, "%[^@]@%d@%d@%d@%[^@]@aabb", innerid, RESOURCE_MAX_LEN,
                           &fault_id, &sub_fault_id, &level, cur_fault_info.resource, RESOURCE_MAX_LEN);
        if (ret != ALARM_INFO_NUMS) {
            ALARM_LOG_ERR("sscanf failed %d", ret);
            continue;
        }

        cur_fault_info.fault_id = (unsigned short)fault_id;
        cur_fault_info.sub_fault_id = (unsigned short)sub_fault_id;
        cur_fault_info.fault_level = (unsigned short)level;
        cur_fault_info.resource[RESOURCE_MAX_LEN - 1] = '\0';

        // 根据fault_id，sub_fault_id，resource，设置告警清除标记。
        ret = set_clear_shield_by_fault_info(&cur_fault_info, alarm_option_type);
        if (ret != EDGE_OK) {
            ALARM_LOG_ERR("set shield flag failed,id(%d),sub_id(%d),resource(%s).",
                          cur_fault_info.fault_id, cur_fault_info.sub_fault_id, cur_fault_info.resource);
            continue;
        }
    }

    (void)fclose(fd);
    return EDGE_OK;
}

/* ****************************************************************************
 函 数 名  : unset_clear_shield_alarm_flag
 功能描述  : 清除所有检测项的告警屏蔽标记。
 输入参数  : 无
 输出参数  : 无
 返 回 值  : 无
**************************************************************************** */
static void unset_clear_shield_alarm_flag()
{
    time_t time_curr = {0};
    ALARM_LV1_STRU *fault_lv1_item = NULL;

    for (int index = 0; index < FAULT_LV1_MAX_SIZE; index++) {
        fault_lv1_item = g_alarm_global_info.alarm_info_stru[index];
        if (fault_lv1_item == NULL) {
            continue;
        }
        while (fault_lv1_item != NULL) {
            if (fault_lv1_item->shield_alarm_flag == SHILED_FLAG_SHIELDED) {
                (void)time(&time_curr);
                fault_lv1_item->fault_info.raise_time_stamp = time_curr;
            }
            fault_lv1_item->shield_alarm_flag = SHILED_FLAG_UNSHIELDED;
            fault_lv1_item = fault_lv1_item->next;
        }
    }
    return;
}

static int reset_shield_alarm_flag()
{
    int ret;
    unsigned int active_num = 0;
    int alarm_level = ALARM_LEVEL_DEFAULT;
    pthread_mutex_lock(&g_alarm_global_info.mutex);
    unset_clear_shield_alarm_flag();
    ret = set_clear_shield_alarm_flag(SHIELD_ALARM_FILE, SHIELD_ALARM_TYPE);
    (void)fault_record_to_file(&active_num, &alarm_level);
    publish_alarm_event(alarm_level);
    pthread_mutex_unlock(&g_alarm_global_info.mutex);
    return ret;
}

static int reset_shield_alarm_func(char *dstbuff, unsigned int buf_len, int *flag)
{
    int ret;
    char *next_event = NULL;
    struct inotify_event event = {0};
    unsigned int event_size = sizeof(struct inotify_event);
    unsigned int size = 0;
    if ((dstbuff == NULL) || (flag == NULL) || (event_size > buf_len)) {
        ALARM_LOG_ERR("invalid inputs.");
        return EDGE_ERR;
    }

    next_event = dstbuff;
    do {
        if (memcpy_s(&event, event_size, next_event, event_size) != 0) {
            ALARM_LOG_ERR("memcpy failed!");
            return EDGE_ERR;
        }

        if ((event.mask & IN_MODIFY) || (event.mask & IN_DELETE_SELF)) {
            ret = reset_shield_alarm_flag();
            if (ret != EDGE_OK) {
                ALARM_LOG_ERR("reset shield alarm flag failed!");
                return EDGE_ERR;
            }
            *flag = TRUE;
        }

        if ((event.mask & IN_DELETE_SELF) != 0) {
            return EDGE_OK;
        }

        size += event_size + event.len;
        if (size + event_size > buf_len) {
            ALARM_LOG_ERR("will out of buf_len!");
            return EDGE_ERR;
        }
        next_event += event_size + event.len;
    } while (event.wd > 0);

    return 1;
}

/* ****************************************************************************
 函 数 名  : inotify_alarm_file
 功能描述  : 监控告警清除或告警屏蔽的配置文件，如果发生变化或删除，则进行配置。
 输入参数  : filename：告警消除或者告警屏蔽配置文件
 输出参数  : flag：告警屏蔽文件更新标记
 返 回 值  : 成功：EDGE_OK 失败：EDGE_ERR
**************************************************************************** */
static int inotify_alarm_file(const char *filename, int *flag)
{
    int ret;
    ssize_t read_ret;
    int instance;
    int watchid;
    char dstbuff[INOTIFY_BUFF_MAX_LEN + 1] = {0};
    unsigned int event_mask = IN_MODIFY | IN_DELETE_SELF;

    if ((filename == NULL) || (flag == NULL)) {
        ALARM_LOG_ERR("invalid inputs.");
        return EDGE_ERR;
    }

    if (access(filename, F_OK) != 0) {
        ALARM_LOG_ERR("file %s not exist.", filename);
        return EDGE_ERR;
    }

    instance = inotify_init();
    if (instance < 0) {
        ALARM_LOG_ERR("inotify_init failed!");
        return EDGE_ERR;
    }

    watchid = inotify_add_watch(instance, filename, event_mask);
    if (watchid < 0) {
        close(instance);
        return EDGE_ERR;
    }

    while (TRUE) {
        if (memset_s(dstbuff, INOTIFY_BUFF_MAX_LEN + 1, 0, INOTIFY_BUFF_MAX_LEN + 1) != 0) {
            ALARM_LOG_ERR("memset_s error!");
            inotify_rm_watch(instance, watchid);
            close(instance);
            return EDGE_ERR;
        }
        read_ret = read(instance, (void *)dstbuff, INOTIFY_BUFF_MAX_LEN);
        if (read_ret == -1 && errno == EINTR) {
            sleep(TIME_INTERVAL);
            continue;
        }

        ret = reset_shield_alarm_func(dstbuff, INOTIFY_BUFF_MAX_LEN, flag);
        if (ret == EDGE_OK) { // Self was deleted
            inotify_rm_watch(instance, watchid);
            close(instance);
            return EDGE_OK;
        } else if (ret == EDGE_ERR) { // set_clear_shield_alarm_flag failed
            inotify_rm_watch(instance, watchid);
            close(instance);
            return EDGE_ERR;
        }
    }

    return EDGE_OK;
}

/* ****************************************************************************
 函 数 名  : fault_manager_shield_alarm_thread
 功能描述  : 告警屏蔽配置文件的监控线程
 输入参数  : 无
 输出参数  : 无
 返 回 值  : 无
**************************************************************************** */
void *fault_manager_shield_alarm_thread(void *p_void)
{
    int ret;
    (void)p_void;
    prctl(PR_SET_NAME, "fault_manager_shield_alarm_thread", 0, 0, 0);

    ret = pthread_detach(pthread_self());
    if (ret != EDGE_OK) {
        ALARM_LOG_ERR("thread detach fail,ret=%d.", ret);
    }

    // inotify监控告警屏蔽配置文件
    do {
        ret = inotify_alarm_file(SHIELD_ALARM_FILE, &g_shield_alarm_flag);
        if (ret != EDGE_OK) {
            ALARM_LOG_ERR("inotify alarm file failed!");
        }
        sleep(TIME_INTERVAL);
    } while (COMMON_TRUE);
}
