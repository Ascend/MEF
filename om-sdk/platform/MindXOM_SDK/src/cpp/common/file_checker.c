/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
 * Description: 文件/路径校验相关接口
 * Author: huawei
 * Create: 2022-02-12
 */

#include "file_checker.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <regex.h>
#include <errno.h>
#include <unistd.h>
#include <dlfcn.h>
#include <fcntl.h>
#include <limits.h>
#include <sys/types.h>
#include "securec.h"

/* ****************************************************************************
 函 数 名  : get_dir_by_file_path
 功能描述  : 文件路径分割
 输入参数  : const char *srcStr 文件的完整路径(入参为文件的绝对路径: 即"文件目录路径/文件名"格式)
            char *dir_path 切割后得到的文件目录的路径，可为空
            char *file_name 切割后得到的文件名，可为空
            unsigned int buf_len  为dir_path和file_name的缓存大小，需要调用者保证，不超过COMMON_VALUE_LEN_256
 使用注意  : 入参是否为标准的文件绝对路径，由调用者保证(如硬编码路径场景可直接使用)
 返回结果  ：处理成功返回COMMON_SUCCESS
**************************************************************************** */
static int get_dir_by_file_path(const char *srcStr, char *dir_path, char *file_name, unsigned int buf_len)
{
    if ((srcStr == NULL) || (strlen(srcStr) == 0) || (strlen(srcStr) >= COMMON_VALUE_LEN_256) ||
        (srcStr[0] != '/') || (srcStr[strlen(srcStr) - 1] == '/') ||
        (((dir_path != NULL) || (file_name != NULL)) && ((buf_len == 0) || (buf_len > COMMON_VALUE_LEN_256)))) {
        SYS_LOG_ERROR("invalid inputs");
        return COMMON_INVALID_PARAM;
    }

    /* 1、获取srcStr中最后一个路径分割符的位置，左边为路径名，右边为文件名 */
    const char *pathLastDelimiter = strrchr(srcStr, '/');
    if ((pathLastDelimiter == NULL) || (pathLastDelimiter < srcStr)) {
        SYS_LOG_ERROR("Src path error, not found '/'.");
        return COMMON_INVALID_PARAM;
    }

    /* 2、获取切割符合左侧的值 */
    if (dir_path) {
        if (memset_s(dir_path, buf_len, 0, buf_len) != 0) {
            SYS_LOG_ERROR("memset_s failed.");
        }

        if (strncpy_s(dir_path, buf_len, srcStr, (uint32_t)(pathLastDelimiter - srcStr)) != 0) {
            SYS_LOG_ERROR("strncpy_s failed.");
            return COMMON_INVALID_PARAM;
        }
    }

    /* 3、获取切割符合右侧的值 */
    if (file_name) {
        if (memset_s(file_name, buf_len, 0, buf_len) != 0) {
            SYS_LOG_ERROR("memset_s failed.");
        }

        if (strcpy_s(file_name, buf_len, pathLastDelimiter + 1) != 0) {
            SYS_LOG_ERROR("strcpy_s failed.");
            return COMMON_INVALID_PARAM;
        }
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : check_dir_link
 功能描述  : 纯目录软链接校验
 输入参数  : const char *dirPath 目录的完整路径(入参为目录的绝对路径)
 使用注意  : 目录路径入参最后不能为“/”符号，否则将返回入参非法
 返回结果  ：非软链接返回COMMON_SUCCESS
**************************************************************************** */
int check_dir_link(const char *dirPath)
{
    char stdPathName[PATH_MAX] = {0};

    if ((dirPath == NULL) || (strlen(dirPath) == 0) || (strlen(dirPath) >= COMMON_VALUE_LEN_256) ||
        (dirPath[0] != '/') || (dirPath[strlen(dirPath) - 1] == '/')) {
        SYS_LOG_ERROR("invalid inputs");
        return COMMON_INVALID_PARAM;
    }

    if (memset_s(stdPathName, PATH_MAX, 0, PATH_MAX) != 0) {
        SYS_LOG_ERROR("memset_s failed.");
    }

    if (realpath(dirPath, stdPathName) == NULL) {
        SYS_LOG_ERROR("realpath failed,errno:%d.", errno);
        return COMMON_INVALID_PARAM;
    }

    if (strlen(dirPath) != strlen(stdPathName)) {
        SYS_LOG_ERROR("check dir path include symbolic link failed.");
        return COMMON_ERROR;
    }

    if (strncmp(stdPathName, dirPath, strlen(dirPath)) != 0) {
        SYS_LOG_ERROR("dir path include symbolic link.");
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : check_file_link
 功能描述  : 对存在的文件进行软链接校验(仅支持文件存在的场景，本质上校验了文件和文件的目录路径)
 输入参数  : const char *path 文件的绝对路径
 使用注意  : path 指向的文件若不存在，则返回COMMON_INVALID_PARAM
 返回结果  ：非软链接, 则返回COMMON_SUCCESS
**************************************************************************** */
int check_file_link(const char *path)
{
    char stdPathName[PATH_MAX] = {0};

    if ((path == NULL) || (strlen(path) == 0) || (strlen(path) >= COMMON_VALUE_LEN_256)) {
        SYS_LOG_ERROR("invalid inputs");
        return COMMON_INVALID_PARAM;
    }

    if (memset_s(stdPathName, PATH_MAX, 0, PATH_MAX) != 0) {
        SYS_LOG_ERROR("memset_s failed.");
    }

    if (realpath(path, stdPathName) == NULL) {
        SYS_LOG_ERROR("realpath failed,errno:%d.", errno);
        return COMMON_INVALID_PARAM;
    }

    if (strlen(path) != strlen(stdPathName)) {
        SYS_LOG_ERROR("check file path include symbolic link failed.");
        return COMMON_ERROR;
    }

    if (strncmp(stdPathName, path, strlen(path)) != 0) {
        SYS_LOG_ERROR("file path include symbolic link.");
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : check_file_owner
 功能描述  : 文件的属主校验
 输入参数  : const char *path 文件的完整路径
            unsigned int expect_uid 期望的属主ID
 使用注意  : 无
 返回结果  ：是对应owner返回COMMON_SUCCESS
**************************************************************************** */
int check_file_owner(const char *path, unsigned int expect_uid)
{
    if (path == NULL) {
        return COMMON_INVALID_PARAM;
    }

    struct stat dirStat;
    if (stat(path, &dirStat) != 0) {
        return COMMON_INVALID_PARAM;
    }

    if ((dirStat.st_uid) != expect_uid) {
        SYS_LOG_ERROR("not expect uid(%u) path:%s", expect_uid, path);
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : check_path_format
 功能描述  : 路径格式是否合法
 输入参数  : const char *path 完整路径
 使用注意  : 无
 返回结果  ：合法返回COMMON_SUCCESS
**************************************************************************** */
static int check_path_format(const char *path)
{
    int cflags = REG_EXTENDED;
    regmatch_t pmatch;
    const size_t match_num = 1;
    regex_t reg;
    const char *reg_name = "\\/[a-z0-9A-Z_./-]+";
    char error_buf[COMMON_VALUE_LEN_256] = {'\0'};

    if (path == NULL) {
        SYS_LOG_ERROR("parameter is invalid");
        return COMMON_INVALID_PARAM;
    }

    if (memset_s(&pmatch, sizeof(pmatch), -1, sizeof(pmatch)) != 0) {
        SYS_LOG_ERROR("memset_s error!");
        return COMMON_ERROR;
    }

    int status = regcomp(&reg, reg_name, cflags);
    if (status != 0) {
        regerror(status, &reg, error_buf, sizeof(error_buf));
        SYS_LOG_ERROR("regcomp error_buf:%s", error_buf);
        return COMMON_ERROR;
    }

    status = regexec(&reg, path, match_num, &pmatch, 0);
    if (status != 0) {
        regerror(status, &reg, error_buf, sizeof(error_buf));
        regfree(&reg);
        SYS_LOG_ERROR("regexec error_buf:%s", error_buf);
        return COMMON_ERROR;
    }
    regfree(&reg);

    size_t len = strlen(path);
    if ((pmatch.rm_eo - pmatch.rm_so) != (int)len) {
        SYS_LOG_ERROR("reg cmp failed");
        return COMMON_ERROR;
    }
    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : check_path_relative
 功能描述  : 校验路径是否包含相对路径
 输入参数  : const char *path 完整路径
 使用注意  : 无
 返回结果  ：不包含相对路径返回COMMON_SUCCESS
**************************************************************************** */
static int check_path_relative(const char *path)
{
    char *feature = "..";

    if (path == NULL) {
        SYS_LOG_ERROR("parameter is invalid");
        return COMMON_INVALID_PARAM;
    }

    if (strstr(path, feature) != NULL) {
        SYS_LOG_ERROR("relative path");
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : check_file_path_valid
 功能描述  : 检查文件路径是否合法
 输入参数  : const char *path 完整路径
 使用注意  : 无
 返回结果  ：路径合法返回COMMON_SUCCESS
**************************************************************************** */
int check_file_path_valid(const char* path)
{
    if ((path == NULL) || (strlen(path) == 0) || (strlen(path) >= COMMON_VALUE_LEN_256) ||
        (path[0] != '/')) {
        SYS_LOG_ERROR("invalid inputs");
        return COMMON_INVALID_PARAM;
    }

    if ((check_path_format(path) != COMMON_SUCCESS) ||
        (check_path_relative(path) != COMMON_SUCCESS)) {
        SYS_LOG_ERROR("check path failed");
        return COMMON_ERROR;
    }

    /* 文件存在, 则校验文件及其目录是否为软链接; 文件不存在，则校验对应的目录是否为软链接 */
    if (access(path, F_OK) == 0) {
        if (check_file_link(path) == COMMON_SUCCESS) {
            return COMMON_SUCCESS;
        }
        SYS_LOG_ERROR("check_file_link failed,path:%s", path);
    } else {
        char dirPath[COMMON_VALUE_LEN_256] = {0};
        if (get_dir_by_file_path(path, dirPath, NULL, COMMON_VALUE_LEN_256) != COMMON_SUCCESS) {
            SYS_LOG_ERROR("get_dir_by_file_path failed,path:%s", path);
            return COMMON_ERROR;
        }

        if (check_dir_link(dirPath) == COMMON_SUCCESS) {
            return COMMON_SUCCESS;
        }
        SYS_LOG_ERROR("check_dir_link failed,dirPath:%s", dirPath);
    }

    return COMMON_ERROR;
}

/* ****************************************************************************
 函 数 名  : check_dir_path_valid
 功能描述  : 检查目录路径是否合法
 输入参数  : const char *path 完整目录路径
 使用注意  : 无
 返回结果  ：路径合法返回COMMON_SUCCESS
**************************************************************************** */
int check_dir_path_valid(const char* path)
{
    if ((path == NULL) || (strlen(path) == 0) || (strlen(path) >= COMMON_VALUE_LEN_256) ||
        (path[0] != '/')) {
        SYS_LOG_ERROR("invalid inputs");
        return COMMON_INVALID_PARAM;
    }

    if ((check_path_format(path) != COMMON_SUCCESS) ||
        (check_path_relative(path) != COMMON_SUCCESS)) {
        SYS_LOG_ERROR("check path failed");
        return COMMON_ERROR;
    }

    if (check_dir_link(path) != COMMON_SUCCESS) {
        SYS_LOG_ERROR("check_dir_link failed,path:%s", path);
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : get_env_var_dir
 功能描述  : 获取路径相关环境变量
 输入参数  : env_name  环境变量名
            env_buff 环境变量缓存
            env_buff_len 环境变量缓存大小
**************************************************************************** */
int get_env_var_dir(const char* env_name, char* env_buff, int env_buff_len)
{
    if ((env_name == NULL) || (env_buff == NULL) ||
        (env_buff_len <= 0) || (env_buff_len > COMMON_VALUE_LEN_256)) {
        SYS_LOG_ERROR("parameter is invalid, env_buff_len:%d", env_buff_len);
        return COMMON_INVALID_PARAM;
    }

    char *value = getenv(env_name);
    if (value == NULL) {
        SYS_LOG_ERROR("Failed to get env var");
        return COMMON_ERROR;
    }

    if (strcpy_s(env_buff, env_buff_len, value) != EOK) {
        SYS_LOG_ERROR("Failed to copy env var");
        return COMMON_ERROR;
    }

    int ret = check_dir_path_valid(env_buff);
    if (ret != COMMON_SUCCESS) {
        SYS_LOG_ERROR("check_dir_path_valid ret:%d!", ret);
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : get_full_valid_path
 功能描述  : 获取完整的有效路径
 输入参数  : full_path  完整路径
            full_path_len 完整路径缓存大小
            dir_path 目录路径参数
            dir_path_len 目录路径参数缓存大小
            part_path  部分路径
**************************************************************************** */
int get_full_valid_path(char* full_path, int full_path_len, const char* dir_path,
                        int dir_path_len, const char* part_path)
{
    if ((full_path == NULL) || (full_path_len <= 0) || (full_path_len > COMMON_VALUE_LEN_256) ||
        (dir_path == NULL) || (dir_path_len <= 0) || (dir_path_len > COMMON_VALUE_LEN_256) ||
        (part_path == NULL)) {
        SYS_LOG_ERROR("parameter is invalid!");
        return COMMON_INVALID_PARAM;
    }

    if (memset_s(full_path, full_path_len, 0, full_path_len) != 0) {
        SYS_LOG_ERROR("memset_s error!");
    }

    if (strcpy_s(full_path, full_path_len, dir_path) != EOK) {
        SYS_LOG_ERROR("strcpy_s error!");
        return COMMON_INVALID_PARAM;
    }

    if ((strlen(full_path) + strlen(part_path) + 1) > (size_t)full_path_len) {
        SYS_LOG_ERROR("buf len check failed! full_path_len:%d", full_path_len);
        return COMMON_INVALID_PARAM;
    }

    if (strcat_s(full_path, full_path_len, part_path) != EOK) {
        SYS_LOG_ERROR("strcat_s error!");
        return COMMON_INVALID_PARAM;
    }

    int ret = check_file_path_valid(full_path);
    if (ret != COMMON_SUCCESS) {
        SYS_LOG_ERROR("check_file_path_valid ret:%d!", ret);
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : check_file_regular
 功能描述  : 文件是否为普通文件,如如数据文件、可执行文件等
 输入参数  : const char *file_path 文件完整路径
 使用注意  : 无
 返回结果  ：是普通文件则返回COMMON_SUCCESS
**************************************************************************** */
static int check_file_regular(const char *file_path)
{
    if (file_path == NULL) {
        return COMMON_INVALID_PARAM;
    }

    /* 获取文件状态 */
    struct stat stBuf;
    if (lstat(file_path, &stBuf) < 0) {
        return COMMON_INVALID_PARAM;
    }

    /* 判断[入参]是否为普通文件 */
    if (!S_ISREG(stBuf.st_mode)) {
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : safety_dlopen
 功能描述  : 文件是否为普通文件,如如数据文件、可执行文件等
 输入参数  : const char* libfile 库文件绝对路径，对应原dlopen接口的libfile参数
            int flag 对应原dlopen接口的flag参数
            int is_check_owner 是否启用属主校验，取值参考 COMMON_TRUE
            unsigned int expect_owner_uid 期望的属主ID，与is_check_owner配合使用
 使用注意  : 无
 返回结果  ：成功返回对应的库文件句柄，失败返回NULL
**************************************************************************** */
void *safety_dlopen(const char* libfile, int flag, int is_check_owner, unsigned int expect_owner_uid)
{
    if (libfile == NULL) {
        SYS_LOG_ERROR("path null.");
        return NULL;
    }

    /* 校验文件路径 */
    int ret = check_file_path_valid(libfile);
    if (ret != COMMON_SUCCESS) {
        SYS_LOG_ERROR("check_file_path_valid ret:%d!", ret);
        return NULL;
    }

    /* 校验是否为普通文件 */
    ret = check_file_regular(libfile);
    if (ret != COMMON_SUCCESS) {
        SYS_LOG_ERROR("check_file_regular fail ret:%d.", ret);
        return NULL;
    }

    /* 校验属主 */
    if (is_check_owner == COMMON_TRUE) {
        ret = check_file_owner(libfile, expect_owner_uid);
        if (ret != COMMON_SUCCESS) {
            SYS_LOG_ERROR("check_file_owner fail ret:%d.", ret);
            return NULL;
        }
    }

    return dlopen(libfile, flag);
}

/* ****************************************************************************
 函 数 名  : safety_fopen
 功能描述  : 文件是否为普通文件,如如数据文件、可执行文件等
 输入参数  : const char *path 文件绝对路径
            const char *mode 文件的打开模式
 使用注意  : 无
 返回结果  ：成功返回对应的文件句柄，失败返回NULL
**************************************************************************** */
FILE *safety_fopen(const char *path, const char *mode)
{
    if ((path == NULL) || (mode == NULL)) {
        SYS_LOG_ERROR("Input path is NULL");
        return NULL;
    }

    if (check_file_path_valid(path) != COMMON_SUCCESS) {
        SYS_LOG_ERROR("check_file_path_valid failed");
        return NULL;
    }

    return fopen(path, mode);
}

/* ****************************************************************************
 函 数 名  : safety_chmod_by_fd
 功能描述  : 根据文件描述符修改文件的权限
 输入参数  : const FILE *fd 文件描述符
            mode_t mode 待修改的文件权限值
 使用注意  : 无
 返回结果  ：仅成功返回COMMON_SUCCESS
**************************************************************************** */
int safety_chmod_by_fd(FILE *fd, mode_t mode)
{
    if (fd == NULL) {
        return COMMON_INVALID_PARAM;
    }
    int dst_fd = fileno(fd);
    if ((dst_fd <= 0) || (fchmod(dst_fd, mode) != 0)) {
        return COMMON_ERROR;
    }

    return COMMON_SUCCESS;
}

/* ****************************************************************************
 函 数 名  : get_file_size
 功能描述  : 获取文件大小
 输入参数  : const char *file 文件名路径
 使用注意  : 无
 返回结果  ：成功返回文件大小，失败返回-1
**************************************************************************** */
long get_file_size(const char *file)
{
    if (file == NULL) {
        return -1;
    }

    struct stat st = {0};
    if (stat(file, &st) < 0) {
        SYS_LOG_ERROR("get file size failed: %d.", errno);
        return -1;
    }

    return st.st_size;
}