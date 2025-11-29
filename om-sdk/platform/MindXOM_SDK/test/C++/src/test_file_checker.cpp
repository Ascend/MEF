// Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <regex.h>
#include <errno.h>
#include <unistd.h>
#include <dlfcn.h>
#include <fcntl.h>
#include <limits.h>
#include <libgen.h>
#include <sys/types.h>
#include <sys/time.h>
#include "securec.h"
#include "test_file_checker.h"

#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif
#endif /* __cplusplus */

#include "file_checker.h"

#ifdef __cplusplus
#if __cplusplus
}
#endif
#endif /* __cplusplus */

using namespace testing;
using namespace std;

namespace FILE_CHECKER_TEST {
    TEST(FileCheckerTest, test_check_dir_link_dirPath_is_empty)
    {
        /* check_dir_link */
        char const *dirPath = "";
        std::cout << "dt test_check_dir_link_dirPath_is_empty start: " << dirPath;
        int ret = check_dir_link(dirPath);
        std::cout << "dt test_check_dir_link_dirPath_is_empty end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_check_dir_link_realpath_failed)
    {
        /* check_dir_link */
        char const *dirPath = "/home/test";
        std::cout << "dt test test_check_dir_link_realpath_failed start: " << dirPath;
        AMOCKER(memset_s).will(returnValue(1));
        AMOCKER(realpath).will(returnValue(NULL));
        int ret = check_dir_link(dirPath);
        std::cout << "dt test test_check_dir_link_realpath_failed end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_check_dir_link_strncmp_failed)
    {
        /* check_dir_link */
        char const *dirPath = "/home/test";
        std::cout << "dt test test_check_dir_link_strncmp_failed start: " << dirPath;
        AMOCKER(memset_s).will(returnValue(0));
        AMOCKER(realpath).will(returnValue("/home/test"));
        AMOCKER(strlen).will(returnValue(10));
        AMOCKER(strncmp).will(returnValue(1));
        int ret = check_dir_link(dirPath);
        std::cout << "dt test test_check_dir_link_strncmp_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_dir_link_success)
    {
        /* check_dir_link */
        char const *dirPath = "/home/test";
        std::cout << "dt test test_check_dir_link_success start: " << dirPath;
        AMOCKER(memset_s).will(returnValue(0));
        AMOCKER(realpath).will(returnValue("/home/test"));
        AMOCKER(strlen).will(returnValue(10));
        AMOCKER(strncmp).will(returnValue(0));
        int ret = check_dir_link(dirPath);
        std::cout << "dt test test_check_dir_link_success end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(FileCheckerTest, test_check_file_link_path_is_empty)
    {
        /* check_file_link */
        char const *path = "";
        std::cout << "dt test_check_file_link_path_is_empty start: " << path;
        int ret = check_file_link(path);
        std::cout << "dt test_check_file_link_path_is_empty end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_check_file_link_realpath_failed)
    {
        /* check_file_link */
        char const *path = "/home/test.conf";
        std::cout << "dt test test_check_file_link_realpath_failed start: " << path;
        AMOCKER(memset_s).will(returnValue(1));
        AMOCKER(realpath).will(returnValue(NULL));
        int ret = check_file_link(path);
        std::cout << "dt test test_check_file_link_realpath_failed end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_check_file_link_strncmp_failed)
    {
        /* check_file_link */
        char const *path = "/home/test.conf";
        std::cout << "dt test test_check_file_link_strncmp_failed start: " << path;
        AMOCKER(memset_s).will(returnValue(0));
        AMOCKER(realpath).will(returnValue("/home/test.conf"));
        AMOCKER(strlen).will(returnValue(10));
        AMOCKER(strncmp).will(returnValue(1));
        int ret = check_file_link(path);
        std::cout << "dt test test_check_file_link_strncmp_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_file_link_success)
    {
        /* check_file_link */
        char const *path = "/home/test.conf";
        std::cout << "dt test test_check_file_link_success start: " << path;
        AMOCKER(memset_s).will(returnValue(0));
        AMOCKER(realpath).will(returnValue("/home/test"));
        AMOCKER(strlen).will(returnValue(10));
        AMOCKER(strncmp).will(returnValue(0));
        int ret = check_file_link(path);
        std::cout << "dt test test_check_file_link_success end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(FileCheckerTest, test_check_file_owner_path_null)
    {
        /* check_file_owner */
        char const *path = NULL;
        unsigned int expectUid = 0;
        std::cout << "dt test test_check_file_owner_path_null start: " << path << expectUid;
        AMOCKER(stat).will(returnValue(1));
        int ret = check_file_owner(path, expectUid);
        std::cout << "dt test test_check_file_owner_path_null end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_check_file_owner_stat_failed)
    {
        /* check_file_owner */
        char const *path = "/home/data/config";
        unsigned int expectUid = 10000;
        std::cout << "dt test test_check_file_owner_stat_failed start: " << path << expectUid;
        AMOCKER(stat).will(returnValue(1));
        int ret = check_file_owner(path, expectUid);
        std::cout << "dt test test_check_file_owner_stat_failed end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_check_file_path_valid_path_null)
    {
        /* check_file_path_valid */
        char const *path = "";
        std::cout << "dt test test_check_file_path_valid_path_null start: " << path;
        int ret = check_file_path_valid(path);
        std::cout << "dt test test_check_file_path_valid_path_null end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_check_file_path_valid_path_invalid)
    {
        /* check_file_path_valid */
        char const *path = "/abc*";
        std::cout << "dt test test_check_file_path_valid_path_invalid start: " << path;
        int ret = check_file_path_valid(path);
        std::cout << "dt test test_check_file_path_valid_path_invalid end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_file_path_valid_path_invalid2)
    {
        /* check_file_path_valid */
        char const *path = "/abc..";
        std::cout << "dt test test_check_file_path_valid_path_invalid2 start: " << path;
        int ret = check_file_path_valid(path);
        std::cout << "dt test test_check_file_path_valid_path_invalid2 end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_file_path_valid_file_exist)
    {
        /* check_file_path_valid */
        char const *path = "/etc/vconsole.conf";
        std::cout << "dt test test_check_file_path_valid_file_exist start: " << path;
        int ret = check_file_path_valid(path);
        std::cout << "dt test test_check_file_path_valid_file_exist end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(FileCheckerTest, test_check_file_path_valid_file_get_dir_failed)
    {
        /* check_file_path_valid */
        char const *path = "/abc/abc";
        std::cout << "dt test test_check_file_path_valid_file_get_dir_failed start: " << path;
        int ret = check_file_path_valid(path);
        std::cout << "dt test test_check_file_path_valid_file_get_dir_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_file_path_valid_file_check_dir_failed)
    {
        /* check_file_path_valid */
        char const *path = "/etc/abc";
        std::cout << "dt test test_check_file_path_valid_file_check_dir_failed start: " << path;
        AMOCKER(check_dir_link).will(returnValue(1));
        int ret = check_file_path_valid(path);
        std::cout << "dt test test_check_file_path_valid_file_check_dir_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_file_path_valid_file_check_dir_success)
    {
        /* check_file_path_valid */
        char const *path = "/etc/abc";
        std::cout << "dt test test_check_file_path_valid_file_check_dir_success start: " << path;
        AMOCKER(check_dir_link).will(returnValue(0));
        int ret = check_file_path_valid(path);
        std::cout << "dt test test_check_file_path_valid_file_check_dir_success end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(FileCheckerTest, test_check_dir_path_valid_path_null)
    {
        /* check_dir_path_valid */
        char const *path = "";
        std::cout << "dt test test_check_dir_path_valid_path_null start: " << path;
        int ret = check_dir_path_valid(path);
        std::cout << "dt test test_check_dir_path_valid_path_null end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_check_dir_path_valid_path_invalid)
    {
        /* check_dir_path_valid */
        char const *path = "/abc*";
        std::cout << "dt test test_check_dir_path_valid_path_invalid start: " << path;
        AMOCKER(check_dir_link).will(returnValue(0));
        int ret = check_dir_path_valid(path);
        std::cout << "dt test test_check_dir_path_valid_path_invalid end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_dir_path_valid_path_invalid2)
    {
        /* check_dir_path_valid */
        char const *path = "/abc..";
        std::cout << "dt test test_check_dir_path_valid_path_invalid2 start: " << path;
        AMOCKER(check_dir_link).will(returnValue(0));
        int ret = check_dir_path_valid(path);
        std::cout << "dt test test_check_dir_path_valid_path_invalid2 end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_dir_path_valid_path_link)
    {
        /* check_dir_path_valid */
        char const *path = "/abc";
        std::cout << "dt test test_check_dir_path_valid_path_link start: " << path;
        AMOCKER(check_dir_link).will(returnValue(1));
        int ret = check_dir_path_valid(path);
        std::cout << "dt test test_check_dir_path_valid_path_link end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_check_dir_path_valid_success)
    {
        /* check_dir_path_valid */
        char const *path = "/abc";
        std::cout << "dt test test_check_dir_path_valid_success start: " << path;
        AMOCKER(check_dir_link).will(returnValue(0));
        int ret = check_dir_path_valid(path);
        std::cout << "dt test test_check_dir_path_valid_success end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(FileCheckerTest, test_get_env_var_dir_param_wrong)
    {
        /* get_env_var_dir */
        char const *envName = "abc";
        char *envBuff = "abc";
        int envBuffLen = 257;
        std::cout << "dt test test_get_env_var_dir_param_wrong start: " << envName << envBuff << envBuffLen;
        int ret = get_env_var_dir(envName, envBuff, envBuffLen);
        std::cout << "dt test test_get_env_var_dir_param_wrong end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_get_env_var_dir_getenv_failed)
    {
        /* get_env_var_dir */
        char const *envName = "abc";
        char *envBuff = "abc";
        int envBuffLen = 4;
        std::cout << "dt test test_get_env_var_dir_getenv_failed start: " << envName << envBuff << envBuffLen;
        AMOCKER(getenv).will(returnValue(NULL));
        int ret = get_env_var_dir(envName, envBuff, envBuffLen);
        std::cout << "dt test test_get_env_var_dir_getenv_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_get_env_var_dir_strcpy_s_failed)
    {
        /* get_env_var_dir */
        char const *envName = "abc";
        char *envBuff = "abc";
        int envBuffLen = 4;
        std::cout << "dt test test_get_env_var_dir_strcpy_s_failed start: " << envName << envBuff << envBuffLen;
        AMOCKER(getenv).will(returnValue("abc"));
        AMOCKER(strcpy_s).will(returnValue(1));
        int ret = get_env_var_dir(envName, envBuff, envBuffLen);
        std::cout << "dt test test_get_env_var_dir_strcpy_s_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_get_env_var_dir_check_path_failed)
    {
        /* get_env_var_dir */
        char const *envName = "abc";
        char *envBuff = "abc";
        int envBuffLen = 4;
        std::cout << "dt test test_get_env_var_dir_check_path_failed start: " << envName << envBuff << envBuffLen;
        AMOCKER(getenv).will(returnValue("abc"));
        AMOCKER(strcpy_s).will(returnValue(0));
        AMOCKER(check_dir_path_valid).will(returnValue(1));
        int ret = get_env_var_dir(envName, envBuff, envBuffLen);
        std::cout << "dt test test_get_env_var_dir_check_path_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_get_env_var_dir_check_success)
    {
        /* get_env_var_dir */
        char const *envName = "abc";
        char *envBuff = "abc";
        int envBuffLen = 4;
        std::cout << "dt test test_get_env_var_dir_check_success start: " << envName << envBuff << envBuffLen;
        AMOCKER(getenv).will(returnValue("abc"));
        AMOCKER(strcpy_s).will(returnValue(0));
        AMOCKER(check_dir_path_valid).will(returnValue(0));
        int ret = get_env_var_dir(envName, envBuff, envBuffLen);
        std::cout << "dt test test_get_env_var_dir_check_success end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(FileCheckerTest, test_get_full_valid_path_param_invalid)
    {
        /* get_full_valid_path */
        char *fullPath = "/abc";
        int fullPathLen = 4;
        const char *dirPath = "/abc";
        int dirPathLen = 4;
        const char *partPath = NULL;
        std::cout << "dt test test_get_full_valid_path_param_invalid start: " \
        << fullPath << fullPathLen << dirPath << dirPathLen << partPath;
        int ret = get_full_valid_path(fullPath, fullPathLen, dirPath, dirPathLen, partPath);
        std::cout << "dt test test_get_full_valid_path_param_invalid end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_get_full_valid_path_strcpy_s_failed)
    {
        /* get_full_valid_path */
        char *fullPath = "/abc";
        int fullPathLen = 4;
        const char *dirPath = "/abc";
        int dirPathLen = 4;
        const char *partPath = "/abc";
        std::cout << "dt test test_get_full_valid_path_strcpy_s_failed start: " \
        << fullPath << fullPathLen << dirPath << dirPathLen << partPath;
        AMOCKER(memset_s).will(returnValue(1));
        AMOCKER(strcpy_s).will(returnValue(1));
        int ret = get_full_valid_path(fullPath, fullPathLen, dirPath, dirPathLen, partPath);
        std::cout << "dt test test_get_full_valid_path_strcpy_s_failed end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_get_full_valid_path_strlen_failed)
    {
        /* get_full_valid_path */
        char *fullPath = "/abc";
        int fullPathLen = 4;
        const char *dirPath = "/abc";
        int dirPathLen = 4;
        const char *partPath = "/abc";
        std::cout << "dt test test_get_full_valid_path_strlen_failed start: " \
        << fullPath << fullPathLen << dirPath << dirPathLen << partPath;
        AMOCKER(memset_s).will(returnValue(0));
        AMOCKER(strcpy_s).will(returnValue(0));
        int ret = get_full_valid_path(fullPath, fullPathLen, dirPath, dirPathLen, partPath);
        std::cout << "dt test test_get_full_valid_path_strlen_failed end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_get_full_valid_path_strcat_s_failed)
    {
        /* get_full_valid_path */
        char *fullPath = "/abc";
        int fullPathLen = 10;
        const char *dirPath = "/abc";
        int dirPathLen = 4;
        const char* partPath = "/abc";
        std::cout << "dt test test_get_full_valid_path_strcat_s_failed start: " \
        << fullPath << fullPathLen << dirPath << dirPathLen << partPath;
        AMOCKER(memset_s).will(returnValue(0));
        AMOCKER(strcpy_s).will(returnValue(0));
        AMOCKER(strcat_s).will(returnValue(1));
        int ret = get_full_valid_path(fullPath, fullPathLen, dirPath, dirPathLen, partPath);
        std::cout << "dt test test_get_full_valid_path_strcat_s_failed end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_get_full_valid_path_check_failed)
    {
        /* get_full_valid_path */
        char *fullPath = "/abc";
        int fullPathLen = 10;
        const char *dirPath = "/abc";
        int dirPathLen = 4;
        const char* partPath = "/abc";
        std::cout << "dt test test_get_full_valid_path_check_failed start: " \
        << fullPath << fullPathLen << dirPath << dirPathLen << partPath;
        AMOCKER(memset_s).will(returnValue(0));
        AMOCKER(strcpy_s).will(returnValue(0));
        AMOCKER(strcat_s).will(returnValue(0));
        AMOCKER(strcat_s).will(returnValue(1));
        int ret = get_full_valid_path(fullPath, fullPathLen, dirPath, dirPathLen, partPath);
        std::cout << "dt test test_get_full_valid_path_check_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_get_full_valid_path_success)
    {
        /* get_full_valid_path */
        char *fullPath = "/abc";
        int fullPathLen = 10;
        const char *dirPath = "/abc";
        int dirPathLen = 4;
        const char* partPath = "/abc";
        std::cout << "dt test test_get_full_valid_path_success start: " \
        << fullPath << fullPathLen << dirPath << dirPathLen << partPath;
        AMOCKER(memset_s).will(returnValue(0));
        AMOCKER(strcpy_s).will(returnValue(0));
        AMOCKER(strcat_s).will(returnValue(0));
        AMOCKER(check_file_path_valid).will(returnValue(0));
        int ret = get_full_valid_path(fullPath, fullPathLen, dirPath, dirPathLen, partPath);
        std::cout << "dt test test_get_full_valid_path_success end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(FileCheckerTest, test_safety_dlopen_param_invalid)
    {
        /* safety_dlopen */
        const char *libfile = NULL;
        int flag = 10;
        int isCheckOwner = 1;
        unsigned int expectOwnerUid = 4;
        std::cout << "dt test test_safety_dlopen_param_invalid start: " \
        << libfile << flag << isCheckOwner << expectOwnerUid;
        void * ret = safety_dlopen(libfile, flag, isCheckOwner, expectOwnerUid);
        std::cout << "dt test test_safety_dlopen_param_invalid end:" << ret << std::endl;
        EXPECT_EQ(NULL, ret);
    }

    TEST(FileCheckerTest, test_safety_dlopen_check_path_failed)
    {
        /* safety_dlopen */
        const char *libfile = NULL;
        int flag = 10;
        int isCheckOwner = 1;
        unsigned int expectOwnerUid = 4;
        std::cout << "dt test test_safety_dlopen_check_path_failed start: " \
        << libfile << flag << isCheckOwner << expectOwnerUid;
        AMOCKER(check_file_path_valid).will(returnValue(1));
        void * ret = safety_dlopen(libfile, flag, isCheckOwner, expectOwnerUid);
        std::cout << "dt test test_safety_dlopen_check_path_failed end:" << ret << std::endl;
        EXPECT_EQ(NULL, ret);
    }

    TEST(FileCheckerTest, test_safety_dlopen_check_regular_failed)
    {
        /* safety_dlopen */
        const char *libfile = "/etc/vconsole.conf";
        int flag = 10;
        int isCheckOwner = 1;
        unsigned int expectOwnerUid = 4;
        std::cout << "dt test test_safety_dlopen_check_regular_failed start: " \
        << libfile << flag << isCheckOwner << expectOwnerUid;
        AMOCKER(check_file_path_valid).will(returnValue(1));
        void * ret = safety_dlopen(libfile, flag, isCheckOwner, expectOwnerUid);
        std::cout << "dt test test_safety_dlopen_check_regular_failed end:" << ret << std::endl;
        EXPECT_EQ(NULL, ret);
    }

    TEST(FileCheckerTest, test_safety_dlopen_check_owner_failed)
    {
        /* safety_dlopen */
        const char *libfile = "/etc/vconsole.conf";
        int flag = 1;
        int isCheckOwner = 1;
        unsigned int expectOwnerUid = 4;
        std::cout << "dt test test_safety_dlopen_check_owner_failed start: " \
        << libfile << flag << isCheckOwner << expectOwnerUid;
        AMOCKER(check_file_path_valid).will(returnValue(0));
        AMOCKER(check_file_owner).will(returnValue(1));
        void *ret = safety_dlopen(libfile, flag, isCheckOwner, expectOwnerUid);
        std::cout << "dt test test_safety_dlopen_check_owner_failed end:" << ret << std::endl;
        EXPECT_EQ(NULL, ret);
    }

    TEST(FileCheckerTest, test_safety_fopen_param_invalid)
    {
        /* safety_fopen */
        const char *path = "/etc/vconsole.conf";
        const char *mode = NULL;
        std::cout << "dt test test_safety_fopen_param_invalid start: " << path << mode;
        FILE *ret = safety_fopen(path, mode);
        std::cout << "dt test test_safety_fopen_param_invalid end:" << ret << std::endl;
        EXPECT_EQ(NULL, ret);
    }

    TEST(FileCheckerTest, test_safety_fopen_check_path_failed)
    {
        /* safety_fopen */
        const char *path = "/etc/vconsole.conf";
        const char *mode = "r";
        std::cout << "dt test test_safety_fopen_check_path_failed start: " << path << mode;
        AMOCKER(check_file_path_valid).will(returnValue(1));
        FILE *ret = safety_fopen(path, mode);
        std::cout << "dt test test_safety_fopen_check_path_failed end:" << ret << std::endl;
        EXPECT_EQ(NULL, ret);
    }

    TEST(FileCheckerTest, test_safety_chmod_by_fd_param_invalid)
    {
        /* safety_chmod_by_fd */
        FILE *fd = NULL;
        mode_t mode = S_IRUSR | S_IWUSR;
        std::cout << "dt test test_safety_chmod_by_fd_param_invalid start: " << mode;
        int ret = safety_chmod_by_fd(fd, mode);
        std::cout << "dt test test_safety_chmod_by_fd_param_invalid end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
    }

    TEST(FileCheckerTest, test_safety_chmod_by_fd_fileno_failed)
    {
        /* safety_chmod_by_fd */
        FILE *fd;
        mode_t mode = S_IRUSR | S_IWUSR;
        std::cout << "dt test test_safety_chmod_by_fd_fileno_failed start: " << mode;
        AMOCKER(fileno).will(returnValue(-1));
        int ret = safety_chmod_by_fd(fd, mode);
        std::cout << "dt test test_safety_chmod_by_fd_fileno_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(FileCheckerTest, test_safety_chmod_by_fd_success)
    {
        /* safety_chmod_by_fd */
        FILE *fd;
        mode_t mode = S_IRUSR | S_IWUSR;
        std::cout << "dt test test_safety_chmod_by_fd_success start: ";
        AMOCKER(fileno).will(returnValue(1));
        AMOCKER(fchmod).will(returnValue(0));
        int ret = safety_chmod_by_fd(fd, mode);
        std::cout << "dt test test_safety_chmod_by_fd_success end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(FileCheckerTest, test_get_file_size_param_invalid)
    {
        /* get_file_size */
        const char *file = NULL;
        std::cout << "dt test test_get_file_size_param_invalid start: " << file;
        long ret = get_file_size(file);
        std::cout << "dt test test_get_file_size_param_invalid end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }
}