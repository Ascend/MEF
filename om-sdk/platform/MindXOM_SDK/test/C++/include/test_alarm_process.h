// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
