// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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