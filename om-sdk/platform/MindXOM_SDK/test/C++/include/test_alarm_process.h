// Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

/*
 * 文 件 名     : test_alarm_process.h
 * 日    期     : 2023/06/12
 * 功能描述     : alarm 的测试用例
 */
#ifndef TEST_H
#define TEST_H

#include <cstdio>
#include <string>
#include "llt_AutoStarUT.h"

using namespace testing;
using namespace std;

namespace ALARM_PROCESS_TEST {
    class AlarmProcessTest : public testing::Test {
    public:
        static void SetUpTestCase()
        {}

        static void TearDownTestCase()
        {}

        virtual void SetUp()
        {}

        virtual void TearDown()
        {}
    };
} // namespace AlarmProcessTest

#endif
