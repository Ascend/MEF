/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 * Description: test ensd
 * Create: 2021-11-11
 */
#ifndef TEST_ENS_H
#define TEST_ENS_H

#include <cstdio>
#include <string>
#include "llt_AutoStarUT.h"

using namespace testing;
using namespace std;

namespace ENS_TEST {
class EnsTest : public testing::Test {
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
} // namespace ENS_TEST

#endif