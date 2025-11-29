/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 * Description: main
 * Author: dingbanglin wx1009413
 * Create: 2021-11-11
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