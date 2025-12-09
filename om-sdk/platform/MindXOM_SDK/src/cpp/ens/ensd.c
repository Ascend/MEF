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
 * Description: ensd的主入口
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
