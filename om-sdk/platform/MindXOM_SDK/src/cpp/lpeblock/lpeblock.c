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
 */
// gcc lpeblock.c -ldl -shared -fPIE -fpic -o liblpeblock.so
#include "lpeblock.h"

#include <dlfcn.h>
#include <stdarg.h>
#include <string.h>
#include <errno.h>
#include <syslog.h>
#include <unistd.h>
#include <sys/file.h>
#include <sys/stat.h>
#include <gnu/lib-names.h>   /* defines LIBC_SO */

static int (*libc_openat)(int dirfd, const char *pathname, int flag, ...);
static int (*libc_execve)(const char *pathname, char *const argv[], char *const envp[]);
static int (*libc_chmod)(const char *pathname, mode_t mode);
static int (*libc_chown)(const char *pathname, uid_t owner, gid_t group);
static int (*libc_fchownat)(int dirfd, const char *pathname, uid_t owner, gid_t group, int flags);
static int (*libc_fchmodat)(int dirfd, const char *pathname, mode_t mode, int flags);


static void *libc_handle;
static char process_name[256] = {0};

#define TRUE 1
#define FALSE 0
#define E_OK  0
#define E_NOK (-1)

static void __attribute__ ((constructor)) setup(void)
{
    if (libc_handle != NULL) {
        return;
    }
    dlerror();                  /* clear any existing error */

    libc_handle = dlopen(LIBC_SO, RTLD_LAZY);
    if (!libc_handle) {
        syslog(LOG_INFO, "LPEBLOCK setup failed: can't find libc: %s\n", dlerror());
        return;
    }

    libc_openat = dlsym(libc_handle, "openat");
    if (libc_openat == NULL) {
        syslog(LOG_INFO, "LPEBLOCK setup openat failed");
        return;
    }
    libc_execve = dlsym(libc_handle, "execve");
    if (libc_execve == NULL) {
        syslog(LOG_INFO, "LPEBLOCK setup execve failed");
        return;
    }
    libc_chmod  = dlsym(libc_handle, "chmod");
    if (libc_chmod == NULL) {
        syslog(LOG_INFO, "LPEBLOCK setup chmod failed");
        return;
    }
    libc_chown  = dlsym(libc_handle, "chown");
    if (libc_chown == NULL) {
        syslog(LOG_INFO, "LPEBLOCK setup chown failed");
        return;
    }
    libc_fchownat = dlsym(libc_handle, "fchownat");
    if (libc_fchownat == NULL) {
        syslog(LOG_INFO, "LPEBLOCK setup fchownat failed");
        return;
    }
    libc_fchmodat = dlsym(libc_handle, "fchmodat");
    if (libc_fchmodat == NULL) {
        syslog(LOG_INFO, "LPEBLOCK setup fchmodat failed");
        return;
    }

    readlink("/proc/self/exe", process_name, sizeof(process_name));
}

static int end_with(const char *str, const char *suffix)
{
    if (str == NULL || suffix == NULL) {
        return FALSE;
    }
    size_t lenstr = strlen(str);
    size_t lensuffix = strlen(suffix);
    if (lensuffix > lenstr) {
        return FALSE;
    }
    if (strncmp(str + (lenstr - lensuffix), suffix, lensuffix) == 0) {
        return TRUE;
    }
    return FALSE;
}

static int lpe_check(const char* func, const char* path)
{
    if (strncmp(path, "/proc", strlen("/proc")) == 0) {
        return E_OK;
    }
    if (strncmp(path, "/dev", strlen("/dev")) == 0) {
        return E_OK;
    }

    struct stat st;
    if (stat(path, &st) != 0 || st.st_uid == 0) {
        // file not exist or owner is root
        return E_OK;
    }

    uid_t id = (uid_t)geteuid();
    if (id == st.st_uid) {
        return E_OK;
    }

    if (end_with(process_name, "python3") == TRUE) {
        if (end_with(path, "py") == TRUE || end_with(path, "pyc") == TRUE) {
            syslog(LOG_INFO, "LPEBLOCK: uid[%d] process[%s] call[%s] uid[%d] file[%s] blocked",
                   id, process_name, func, st.st_uid, path);
            errno = EPERM;
            return E_NOK;
        }
    }
    // root visit non root file
    syslog(LOG_INFO, "LPEBLOCK: uid[%d] process[%s] call[%s] uid[%d] file[%s] skipped",
           id, process_name, func, st.st_uid, path);
    return E_OK;
}

static int lpe_exe_check(const char* func, const char* path)
{
    struct stat st;
    if (stat(path, &st) != 0) {
        // file not exist
        return E_OK;
    }

    uid_t id = (uid_t)geteuid();
    if (id == 0 && st.st_uid != 0) {
        // process euid is root  but file owner is not
        syslog(LOG_INFO, "LPEBLOCK: uid[%d] process[%s] call[%s] uid[%d] file[%s] blocked",
               id, process_name, func, st.st_uid, path);
        errno = EPERM;
        return E_NOK;
    }
    return E_OK;
}

/**
 * @brief 检查chown/chmod提权问题
 *        如果path是软连接，且软连接本身的owner和软连接target的owner不一致时返回错误
 */
static int lpe_change_permission_check(const char* func, const char* path, const int flags)
{
    struct stat st;
    if (lstat(path, &st) != 0 || st.st_uid == 0) {
        // (link or file) not exist or owner is root
        return E_OK;
    }

    uid_t id = (uid_t)getuid();
    if (!S_ISLNK(st.st_mode)) {
        // path is not softlink
        if (id != st.st_uid) {
            // root visit non root file
            syslog(LOG_INFO, "LPEBLOCK: uid[%d] process[%s] call[%s] uid[%d] file[%s] skipped",
                   id, process_name, func, st.st_uid, path);
        }
        return E_OK;
    }

    // chown -h
    if ((flags & AT_SYMLINK_NOFOLLOW) != 0) {
        return E_OK;
    }

    // path is softlink, get real file st
    struct stat real_st;
    if (stat(path, &real_st) != 0) {
        // link target is not exist
        syslog(LOG_INFO, "LPEBLOCK: uid[%d] process[%s] call[%s] uid[%d] link file[%s] errno[%d], blocked",
               id, process_name, func, st.st_uid, path, errno);
        return E_NOK;
    }
    if (st.st_uid != real_st.st_uid) {
        // link and target not same owner
        syslog(LOG_INFO, "LPEBLOCK: uid[%d] process[%s] call[%s] uid[%d/%d] file[%s] blocked",
               id, process_name, func, st.st_uid, real_st.st_uid, path);
        errno = EPERM;
        return E_NOK;
    }
    return E_OK;
}

int openat(int dirfd, const char *pathname, int flags, ...)
{
    if (lpe_check(__FUNCTION__, pathname) != E_OK) {
        return -1;
    }
    va_list argp;
    va_start(argp, flags);
    mode_t mode = va_arg(argp, mode_t);
    va_end(argp);
    if (libc_openat == NULL) {
        syslog(LOG_INFO, "LPEBLOCK can't find openat");
        return -1;
    }
    return libc_openat(dirfd, pathname, flags, mode);
}

int execve(const char *pathname, char *const argv[], char *const envp[])
{
    if (lpe_exe_check(__FUNCTION__, pathname) != E_OK) {
        return -1;
    }
    if (libc_execve == NULL) {
        syslog(LOG_INFO, "LPEBLOCK can't find execve");
        return -1;
    }
    return libc_execve(pathname, argv, envp);
}

int chmod(const char *pathname, mode_t mode)
{
    if (lpe_change_permission_check(__FUNCTION__, pathname, 0) != E_OK) {
        return -1;
    }
    if (libc_chmod == NULL) {
        syslog(LOG_INFO, "LPEBLOCK can't find chmod");
        return -1;
    }
    return libc_chmod(pathname, mode);
}

int fchmodat(int dirfd, const char *pathname, mode_t mode, int flags)
{
    if (flags < 0 || lpe_change_permission_check(__FUNCTION__, pathname, flags) != E_OK) {
        return -1;
    }
    if (libc_fchmodat == NULL) {
        syslog(LOG_INFO, "LPEBLOCK can't find fchmodat");
        return -1;
    }
    return libc_fchmodat(dirfd, pathname, mode, flags);
}

int chown(const char *pathname, uid_t owner, gid_t group)
{
    if (lpe_change_permission_check(__FUNCTION__, pathname, 0) != E_OK) {
        return -1;
    }
    if (libc_chown == NULL) {
        syslog(LOG_INFO, "LPEBLOCK can't find chown");
        return -1;
    }
    return libc_chown(pathname, owner, group);
}

int fchownat(int dirfd, const char *pathname, uid_t owner, gid_t group, int flags)
{
    if (flags < 0 || lpe_change_permission_check(__FUNCTION__, pathname, flags) != E_OK) {
        return -1;
    }
    if (libc_fchownat == NULL) {
        syslog(LOG_INFO, "LPEBLOCK can't find fchownat");
        return -1;
    }
    return libc_fchownat(dirfd, pathname, owner, group, flags);
}

