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
 * Description: 文件安全校验相关接口
 */
#ifndef _FILE_CHECKER_H_
#define _FILE_CHECKER_H_

#include <stdint.h>
#include <stdio.h>
#include <sys/stat.h>
#include "base_type.h"

int check_file_owner(const char *path, unsigned int expect_uid);
int check_dir_link(const char *dirPath);
int check_file_link(const char *path);
int check_file_path_valid(const char* path);
int check_dir_path_valid(const char* path);
int get_env_var_dir(const char* env_name, char* env_buff, int env_buff_len);
int get_full_valid_path(char* full_path, int full_path_len, const char* dir_path,
                        int dir_path_len, const char* part_path);
void *safety_dlopen(const char* libfile, int flag, int is_check_owner, unsigned int expect_owner_uid);
FILE *safety_fopen(const char *path, const char *mode);
int safety_chmod_by_fd(FILE *fd, mode_t mode);
long get_file_size(const char *file);

#endif