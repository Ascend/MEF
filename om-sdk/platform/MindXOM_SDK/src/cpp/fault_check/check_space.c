/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: 磁盘空间故障检测功能实现
 * Author: huawei
 * Create: 2020-11-12
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

const char* OUTPUT_FILE_PATH = "/run/dir_space_info";

// 检查临时文件路径有效性
static int check_output_path(void)
{
    if (check_file_path_valid(OUTPUT_FILE_PATH) != EDGE_OK) {
        FAULT_LOG_ERR("dir_space_info path is not valid");
        return -1;
    }
    return 0;
}

// 构建df命令
static int build_df_command(char* buffer, size_t size, const char* path)
{
    if ((buffer == NULL) || (sprintf_s(buffer, size, "df -h %s", path) < 0)) {
        FAULT_LOG_ERR("build command fail for path %s", path);
        return -1;
    }
    return 0;
}

// 添加过滤命令链
static int append_filter_pipeline(char* buffer, size_t size)
{
    const char* filter = "|awk 'END{print}' |awk '{print $5}' |awk -F '%' '{print $1}' > /run/dir_space_info";
    if (strncat_s(buffer, size, filter, strlen(filter)) != 0) {
        FAULT_LOG_ERR("append filter fail");
        return -1;
    }
    return 0;
}

// 执行命令并保存输出
static int execute_and_save_command(const char* command)
{
    if (system(command) != 0) {
        FAULT_LOG_ERR("execute command fail: %s", command);
        return -1;
    }
    return 0;
}

// 读取并解析使用率
static int read_and_parse_usage(int* usage, const char* filePath)
{
    char file_buff[FILE_BUFF_SIZE] = {0};
    FILE *temp_fd = safety_fopen(filePath, "r");
    if (temp_fd == NULL) {
        FAULT_LOG_ERR("open file failed: %s", filePath);
        return -1;
    }

    (void)fread(file_buff, 1, sizeof(file_buff) - 1, temp_fd);
    file_buff[FILE_BUFF_SIZE - 1] = '\0';
    size_t result_len = strlen(file_buff);
    // 去掉换行
    if (result_len > 0) {
        *(file_buff + result_len - 1) = '\0';
        *usage = StrToInt(file_buff, FILE_BUFF_SIZE);
    } else {
        (void)fclose(temp_fd);
        return -1;
    }

    (void)fclose(temp_fd);
    return 0;
}

// 处理单个目录空间检查
static int process_directory_space(FAULT_SPACE_FULL_INFO* dir_info)
{
    char block_info[MAX_BUFF_SIZE] = {0};

    if ((build_df_command(block_info, MAX_BUFF_SIZE, dir_info->path_name) != 0) ||
        (append_filter_pipeline(block_info, MAX_BUFF_SIZE) != 0)) {
        return -1;
    }

    if (execute_and_save_command(block_info) != 0) {
        return -1;
    }

    int curr_usage = 0;
    if (read_and_parse_usage(&curr_usage, OUTPUT_FILE_PATH) != 0) {
        return -1;
    }

    if ((curr_usage <= MAX_PERCENT) && (curr_usage >= dir_info->occur_thresh)) {
        if (dir_info->status != 1) {
            FAULT_LOG_ERR("%s space full(%d).", dir_info->path_name, curr_usage);
        }
        dir_info->status = 1;
    } else if ((dir_info->status == 1) && (curr_usage < dir_info->resume_thresh)) {
        FAULT_LOG_ERR("%s space full resume(%d).", dir_info->path_name, curr_usage);
        dir_info->status = 0;
    }

    return 0;
}

static void fault_get_space_full_info(void)
{
    if (check_output_path() != 0) {
        return;
    }

    for (unsigned int index = 0; index < g_fault_space_full_num; index++) {
        if (process_directory_space(&g_fault_space_full_info[index]) != 0) {
            continue;
        }
    }

    (void)unlink(OUTPUT_FILE_PATH);
}

// 检测空间满
int fault_check_space_full(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    FAULT_SPACE_FULL_INFO *curr_dir_info = NULL;

    (void)fault_id;
    (void)sub_id;

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
