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

#include <cstdio>
#include <iostream>
#include "llt_AutoStarUT.h"

using namespace testing;

extern "C" {
    unsigned long __stack_chk_guard;
    void __stack_chk_guard_setup(void)
    {
        __stack_chk_guard = 0xBAAAAAAD;//provide some magic numbers
    }

    void __stack_chk_fail(void)
    {
        /* Error message */
    }// will be called when guard variable is corrupted 
}

int main(int argc, char *argv[])
{
    std::cout << "dt test MindXOM start" << std::endl;
    int ret = Init_UT(argc, (char **)argv, true);
    std::cout << "dt test MindXOM finished" << ret << std::endl;
    return ret;
}