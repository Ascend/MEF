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
 * Description: test ensd
 */

#include <stdio.h>
#include <stdlib.h>
#include <libgen.h>
#include <cstdio>
#include <iostream>
#include "securec.h"
#include "test_ens.h"

#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif
#endif /* __cplusplus */

#include "ens.h"

#ifdef __cplusplus
#if __cplusplus
}
#endif
#endif /* __cplusplus */

using namespace testing;
using namespace std;

namespace ENS_TEST {
const int MAX_BUFF_SIZE = 512;

TEST_F(EnsTest, test_ens_work)
{
    /* 通过shell设置环境变量或者如下代码设置 */
    char *home_path = nullptr;
    char *rpath = nullptr;
    char *bin_dir = dirname(rpath);

    if (bin_dir) {
        home_path = dirname(bin_dir);
    }
    free(rpath);
    rpath = nullptr;
    EXPECT_EQ(0, (home_path == nullptr));
    
    char *arg[1];
    arg[0] = bin_dir;
    std::cout << "dt test ens_work start bin_dir:" << bin_dir << home_path<< std::endl;
    int ret = ens_work(1, arg);
    std::cout << "dt test ens_work end ret:" <<ret<< std::endl;
    EXPECT_EQ(0, ret);
}

} // namespace ENS_TEST