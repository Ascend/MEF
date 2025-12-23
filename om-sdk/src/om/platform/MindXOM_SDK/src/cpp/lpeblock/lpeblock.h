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
 * 功能描述     : 禁止提权操作
 */
#ifndef LPEBLOCK_H
#define LPEBLOCK_H

#include <sys/types.h>

int openat(int dirfd, const char *pathname, int flags, ...);

int execve(const char *pathname, char *const argv[], char *const envp[]);

int chmod(const char *pathname, mode_t mode);

int fchmodat(int dirfd, const char *pathname, mode_t mode, int flags);

int chown(const char *pathname, uid_t owner, gid_t group);

int fchownat(int dirfd, const char *pathname, uid_t owner, gid_t group, int flags);

#endif
