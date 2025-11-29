/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
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
