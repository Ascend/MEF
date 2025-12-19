/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   OMSDK is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 * Description: CleanCode增加devm.c描述
 */
#include <stdlib.h>
#include <dlfcn.h>
#include <securec.h>
#include "base_type.h"
#include "ens_base.h"
#include "ens_api.h"
#include "ens_log.h"
#include "file_checker.h"

#define DEVM_DRIVER_LIB_NAME "/usr/local/lib/libsystem_adapter.so"
#define INT_SIZE 4
#define ALARM_ID 65540

typedef int (*PROCESS_ALARM_EVENT_FUNC)(int alarm_num);
int (*g_pfn_fm_subscribe_alarm_event)(PROCESS_ALARM_EVENT_FUNC p_func);

static int process_alarm_event_func(int alarm_level)
{
    int ret;
    int fd;
    int (*dev_load)(void);
    int (*dev_open)(char *, int *);
    int (*set_attribute)(int, unsigned int, unsigned char *);
    ENS_LOG_INFO("alarm_level: %d", alarm_level);
    void *p_devm_driver_lib = safety_dlopen(DEVM_DRIVER_LIB_NAME, RTLD_GLOBAL | RTLD_LAZY, COMMON_TRUE, ROOT_UID);
    if (p_devm_driver_lib == NULL) {
        ENS_LOG_INFO("failed to open %s, devm driver APIs will not be available", DEVM_DRIVER_LIB_NAME);
        return -1;
    }

    dev_load = dlsym(p_devm_driver_lib, "dev_load");
    dev_open = dlsym(p_devm_driver_lib, "dev_open");
    set_attribute = dlsym(p_devm_driver_lib, "set_attribute");
    if (dev_load == NULL || dev_open == NULL || set_attribute == NULL) {
        ENS_LOG_INFO("failed to load set alarm level function.");
        dlclose(p_devm_driver_lib);
        return -1;
    }

    ret = dev_load();
    if (ret != 0) {
        ENS_LOG_INFO("failed to call dev load function.");
        dlclose(p_devm_driver_lib);
        return -1;
    }

    ret = dev_open("system0", &fd);
    if (ret != 0) {
        ENS_LOG_INFO("failed to call dev open function.");
        dlclose(p_devm_driver_lib);
        return -1;
    }

    unsigned int buffSize = INT_SIZE * 3;
    unsigned char *buff = (unsigned char *)malloc(buffSize);
    if (buff == NULL) {
        ENS_LOG_INFO("call malloc fail.");
        dlclose(p_devm_driver_lib);
        return -1;
    }
    *(int *)(buff) = ALARM_ID;
    *(int *)(buff + INT_SIZE) = INT_SIZE;
    *(int *)(buff + INT_SIZE * 2) = alarm_level;
    ret = set_attribute(fd, buffSize, buff);
    if (ret != 0) {
        ENS_LOG_INFO("failed to call set alarm level function.");
        dlclose(p_devm_driver_lib);
        free(buff);
        return -1;
    }

    free(buff);
    return ret;
}

int devm_load(void)
{
    int ret = ens_intf_import("subscribe_alarm_event", (void **)&g_pfn_fm_subscribe_alarm_event);
    ENS_LOG_INFO("devm load ret:%d", ret);
    return ENS_OK;
}

int devm_unload(void)
{
    ENS_LOG_WARN("devm unload,but do nothing now!");
    return ENS_OK;
}

int devm_start(void)
{
    if (g_pfn_fm_subscribe_alarm_event == NULL) {
        ENS_LOG_ERR("g_pfn_fm_subscribe_alarm_event null");
        return ENS_ERR;
    }

    int ret = g_pfn_fm_subscribe_alarm_event(process_alarm_event_func);
    ENS_LOG_INFO("subscribe_alarm_event ret:%d", ret);
    return ENS_OK;
}

int devm_stop(void)
{
    ENS_LOG_WARN("devm stop,but do nothing now!");
    return ENS_OK;
}
