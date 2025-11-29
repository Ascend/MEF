// Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

#include <dlfcn.h>
#include <stdarg.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <sys/file.h>
#include <sys/stat.h>
#include <gnu/lib-names.h>
#include "test_lpe_block.h"

#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif
#endif /* __cplusplus */

#include "lpeblock.h"

#ifdef __cplusplus
#if __cplusplus
}
#endif
#endif /* __cplusplus */

using namespace testing;
using namespace std;

namespace LPE_BLOCK_TEST {

    TEST(LpeBlockTest, test_openat_check_failed)
    {
        /* openat */
        int dirfd = 1;
        const char *pathname;
        int flags = 1;
        std::cout << "dt test_openat_check_failed start: " << dirfd << pathname << flags;
        AMOCKER(strncmp).will(returnValue(1));
        int ret = openat(dirfd, pathname, flags);
        std::cout << "dt test_openat_check_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(LpeBlockTest, test_openat_libc_openat_null)
    {
        /* openat */
        int dirfd = 1;
        const char *pathname;
        int flags = 1;
        std::cout << "dt test_openat_libc_openat_null start: " << dirfd << pathname << flags;
        AMOCKER(strncmp).will(returnValue(0));
        int ret = openat(dirfd, pathname, flags);
        std::cout << "dt test_openat_libc_openat_null end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(LpeBlockTest, test_execve_libc_check_failed)
    {
        /* execve */
        const char *pathname = "/opt/buildtools/auto_fdisk.sh";
        char *const argv[] = {
                "PATH=/usr/bin",
                "HOME=/home/user",
                "LANG=en_US.UTF-8",
                NULL
        };
        char *const envp[] = {
                "PATH=/usr/bin",
                "HOME=/home/user",
                "LANG=en_US.UTF-8",
                NULL
        };
        std::cout << "dt test_execve_libc_check_failed start: " << pathname << argv << envp;
        AMOCKER(geteuid).will(returnValue(1));
        int ret = execve(pathname, argv, envp);
        std::cout << "dt test_execve_libc_check_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(LpeBlockTest, test_execve_libc_execve_null)
    {
        /* execve */
        const char *pathname = "/opt/auto_fdisk.sh";
        char *const argv[] = {
            "PATH=/usr/bin",
            "HOME=/home/user",
            "LANG=en_US.UTF-8",
            NULL
        };
        char *const envp[] = {
                "PATH=/usr/bin",
                "HOME=/home/user",
                "LANG=en_US.UTF-8",
                NULL
        };
        std::cout << "dt test_execve_libc_execve_null start: " << pathname << argv << envp;
        AMOCKER(geteuid).will(returnValue(0));
        int ret = execve(pathname, argv, envp);
        std::cout << "dt test_execve_libc_execve_null end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(LpeBlockTest, test_chmod_libc_chmod_null)
    {
        /* chmod */
        const char *pathname = "/opt/auto_fdisk.sh";
        mode_t mode;
        std::cout << "dt test_chmod_libc_chmod_null start: " << pathname << mode;
        int ret = chmod(pathname, mode);
        std::cout << "dt test_chmod_libc_chmod_null end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(LpeBlockTest, test_fchmodat_libc_fchmodat_null)
    {
        /* fchmodat */
        int dirfd = 0;
        const char *pathname = "/opt/auto_fdisk.sh";
        mode_t mode;
        int flags = 0;
        std::cout << "dt test_fchmodat_libc_fchmodat_null start: "<< dirfd << pathname << mode;
        int ret = fchmodat(dirfd, pathname, mode, flags);
        std::cout << "dt test_fchmodat_libc_fchmodat_null end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(LpeBlockTest, test_chown_libc_chown_null)
    {
        /* chown */
        const char *pathname = "/opt/auto_fdisk.sh";
        uid_t owner;
        gid_t group;
        std::cout << "dt test_chown_libc_chown_null start: " << pathname << owner << group;
        int ret = chown(pathname, owner, group);
        std::cout << "dt test_chown_libc_chown_null end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(LpeBlockTest, test_fchownat_libc_fchownat_null)
    {
        /* fchownat */
        int dirfd = 0;
        const char *pathname = "/opt/auto_fdisk.sh";
        uid_t owner;
        gid_t group;
        int flags = 0;
        std::cout << "dt test_fchownat_libc_fchownat_null start: " << dirfd << pathname << owner << group;
        int ret = fchownat(dirfd, pathname, owner, group, flags);
        std::cout << "dt test_fchownat_libc_fchownat_null end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }
}