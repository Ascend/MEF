/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 *
 * 文 件 名     : test_file_checker.h
 * 日    期     : 2023/08/16
 * 功能描述     : file_checker 的测试用例
 */
#ifndef TEST_H
#define TEST_H

#include <cstdio>
#include <string>
#include "llt_AutoStarUT.h"

using namespace testing;
using namespace std;

namespace FILE_CHECKER_TEST {
    class FileCheckerTest : public testing::Test {
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
} // namespace CERT_TEST

#endif