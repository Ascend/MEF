/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 * Description: ensd的主入口
 * Create：2021-11-22
 * Modify: Clean Code 2021-11-22
 */

#include <unistd.h>
#include "ens.h"
#include "ens_log.h"

int main(int argc, char *argv[])
{
    int ret = ens_work(argc, argv);
    if (ret != 0) {
        ENS_LOG_FATAL("ensd init failed.ret:%d", ret);
        return ret;
    }

    while (1) {
        sleep(1);
    }
}
