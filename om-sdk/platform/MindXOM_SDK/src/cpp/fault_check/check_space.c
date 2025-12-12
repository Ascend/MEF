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
 * Description: 磁盘空间故障检测功能实现
 */
#include <stdio.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdlib.h>
#include <sys/types.h>
#include "securec.h"
#include "fault_check.h"
#include "certproc.h"
#include "file_checker.h"

#define MAX_PERCENT 100
#define MAX_BUFF_SIZE 512
#define FILE_BUFF_SIZE 32

/* 空间满检测 */
typedef struct {
    char path_name[256];
    int status;        // 当前状态
    int occur_thresh;  // 内存占有率
    int resume_thresh; // 内存占有率
} FAULT_SPACE_FULL_INFO;

FAULT_SPACE_FULL_INFO g_fault_space_full_info[] = {
    {"/",               0, 85, 80},
    {"/run",            0, 85, 80},
    {"/home/data",      0, 85, 80},
    {"/home/log",       0, 85, 80},
    {"/opt",            0, 85, 80},
    {"/var/lib/docker", 0, 85, 80},
};

unsigned int g_fault_space_full_num = sizeof(g_fault_space_full_info) / sizeof(FAULT_SPACE_FULL_INFO);

static void fault_get_space_full_info(void)
{
    char file_buff[FILE_BUFF_SIZE] = {0};
    char block_info[MAX_BUFF_SIZE] = {0};
    FAULT_SPACE_FULL_INFO *curr_dir_info = NULL;
    char filt[] = "|awk 'END{print}' |awk '{print $5}' |awk -F '%' '{print $1}' > /run/dir_space_info";
    if (check_file_path_valid("/run/dir_space_info") != EDGE_OK) {
        FAULT_LOG_ERR("dir_space_info path is not valid");
        return;
    }

    for (unsigned int index = 0; index < g_fault_space_full_num; index++) {
        if ((memset_s(block_info, MAX_BUFF_SIZE, 0, MAX_BUFF_SIZE) != 0) ||
            (memset_s(file_buff, FILE_BUFF_SIZE, 0, FILE_BUFF_SIZE) != 0)) {
            FAULT_LOG_ERR("memset_s error!");
            break;
        }
        curr_dir_info = &g_fault_space_full_info[index];
        if (sprintf_s(&block_info[0], MAX_BUFF_SIZE, "df -h %s", &curr_dir_info->path_name[0]) < 0) {
            FAULT_LOG_ERR("call sprintf_s fail(%s).", &block_info[0]);
            continue;
        }

        if (strncat_s(&block_info[0], MAX_BUFF_SIZE, filt, strlen(filt) + 1) != 0) {
            FAULT_LOG_ERR("call strncat_s fail (%s).", &block_info[0]);
            continue;
        }

        (void)system(&block_info[0]);
        FILE *temp_fd = safety_fopen("/run/dir_space_info", "r");
        if (temp_fd == NULL) {
            FAULT_LOG_ERR("open file failed,index:%d", index);
            continue;
        }

        (void)fread(file_buff, 1, sizeof(file_buff) - 1, temp_fd);
        file_buff[FILE_BUFF_SIZE - 1] = '\0';
        size_t result_len = strlen(file_buff);
        if (result_len == 0) {
            (void)fclose(temp_fd);
            continue;
        }

        *(&file_buff[0] + result_len - 1) = '\0'; // 去掉换行
        int curr = StrToInt(file_buff, FILE_BUFF_SIZE);
        if ((curr <= MAX_PERCENT) && (curr >= curr_dir_info->occur_thresh)) {
            if (curr_dir_info->status != 1) {
                FAULT_LOG_ERR("curr_dir_info_path space full(%d).", curr);
            }
            curr_dir_info->status = 1;
        } else if ((curr_dir_info->status == 1) && (curr < curr_dir_info->resume_thresh)) {
            FAULT_LOG_ERR("curr_dir_info_path space full resume(%d).", curr);
            curr_dir_info->status = 0;
        }

        (void)fclose(temp_fd);
    }
    (void)unlink("/run/dir_space_info");
}

// 检测空间满
int fault_check_space_full(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    FAULT_SPACE_FULL_INFO *curr_dir_info = NULL;

    if (value == NULL) {
        FAULT_LOG_ERR("input null");
        return EDGE_ERR;
    }

    *value = FAULT_STATUS_OK;
    fault_get_space_full_info();

    for (unsigned int index = 0; index < g_fault_space_full_num; index++) {
        curr_dir_info = &g_fault_space_full_info[index];
        if (curr_dir_info->status == 1) {
            *value = FAULT_STATUS_ERR;
            break;
        }
    }

    return EDGE_OK;
}
