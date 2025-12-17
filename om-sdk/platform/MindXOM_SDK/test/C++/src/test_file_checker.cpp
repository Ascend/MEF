// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
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

    TEST(FileCheckerTest, test_check_file_link_path_is_empty)
    {
        /* check_file_link */
        char const *path = "";
        std::cout << "dt test_check_file_link_path_is_empty start: " << path;
        int ret = check_file_link(path);
        std::cout << "dt test_check_file_link_path_is_empty end:" << ret << std::endl;
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

    TEST(FileCheckerTest, test_check_dir_path_valid_path_null)
    {
        /* check_dir_path_valid */
        char const *path = "";
        std::cout << "dt test test_check_dir_path_valid_path_null start: " << path;
        int ret = check_dir_path_valid(path);
        std::cout << "dt test test_check_dir_path_valid_path_null end:" << ret << std::endl;
        EXPECT_EQ(-2, ret);
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