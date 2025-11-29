/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
 *
 * 文 件 名     : test_cms.h
 * 日    期     : 2022/08/16
 * 功能描述     : cms 的测试用例
 */
#ifndef TEST_H
#define TEST_H

#include <cstdio>
#include <string>
#include "llt_AutoStarUT.h"

using namespace testing;
using namespace std;

namespace CMS_TEST {
class CmsTest : public testing::Test {
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
} // namespace CMS_TEST

#endif