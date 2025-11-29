// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
#ifndef TEST_H
#define TEST_H

#include <cstdio>
#include <string>
#include "llt_AutoStarUT.h"

using namespace testing;
using namespace std;

namespace LPE_BLOCK_TEST {
    class LpeBlockTest : public testing::Test {
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
} // namespace LPE_BLOCK_TEST

#endif