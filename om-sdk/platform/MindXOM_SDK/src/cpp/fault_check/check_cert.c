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
 * Description: 证书故障检测功能实现
 */
#include <stdio.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/wait.h>
#include "certproc.h"
#include "fault_check.h"
#include "file_checker.h"

#define DIR_NGINX_CERT                          "/home/data/config/default/server_kmc.cert"
#define DIR_REDFISH_CERT                        "/home/data/config/redfish/server_kmc.cert"
#define CERT_UPDATE_FILE                        "/run/certupdate"
#define CERT_FAULTTIME_UPDATE_FILE              "/run/certWarnTimeUpdate"
#define CERT_FAULTTIME_CONFIG_FILE              "/home/data/ies/certWarnTime.ini"
#define ERR_CERT_VALIDITY_WARN                  (0x20140000 + 0x09)
#define CHECK_FLAG_VALUE                        0x55
#define MAX_CERT_WARN_DAY                       180
#define MIN_CERT_WARN_DAY                       7

int g_warnDay = 10;

static void fault_update_cert_warnTime()
{
    FILE *fp = NULL;
    size_t read_len;
    char  time_info[MAX_BUF_LEN_128] = {0};
    int day;

    if (access(CERT_FAULTTIME_CONFIG_FILE, F_OK) == 0) {
        fp = safety_fopen(CERT_FAULTTIME_CONFIG_FILE, "r");
        if (fp == NULL) {
            return;
        }

        read_len = fread(time_info, 1, sizeof(time_info) - 1, fp);
        (void)fclose(fp);

        if (read_len <= 0) {
            FAULT_LOG_ERR("certWartTime`s content is null.");
            return;
        }

        day = StrToInt(time_info, MAX_BUF_LEN_128);
        if ((day >= MIN_CERT_WARN_DAY) && (day <= MAX_CERT_WARN_DAY)) {
            g_warnDay = day;
        }
    }
}

static unsigned int get_cert_status(const char *CERT_FILE)
{
    unsigned int state_root = 0;
    int day_root = 0;
    signed long ret_root;

    if (access(CERT_FILE, F_OK) == 0) {
        if (check_cert_security(CERT_FILE) != 0) {
            return state_root;
        }
        ret_root = certPeriodicalCheck(CERT_FILE, &day_root);
        if (((ret_root == 0) || (ret_root == ERR_CERT_VALIDITY_WARN)) && (day_root < g_warnDay)) {
            state_root = 1;
        }
    }
    return state_root;
}

static unsigned int fault_cert_status(void)
{
    unsigned int state_nginx;
    unsigned int state_redfish;

    fault_update_cert_warnTime();
    state_nginx = get_cert_status(DIR_NGINX_CERT);
    state_redfish = get_cert_status(DIR_REDFISH_CERT);
    OM_LOG_INFO("state_nginx flag(%d), state_redfish flag(%d).",
                state_nginx, state_redfish);
    return (state_nginx || state_redfish);
}

int fault_check_cert_warn(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    unsigned int state;
    static unsigned short alarm_state = FAULT_STATUS_OK;
    if (value == NULL) {
        FAULT_LOG_ERR("input value is NULL!");
        return EDGE_ERR;
    }
    state = fault_cert_status();
    if (state != 0) {
        *value = FAULT_STATUS_ERR;
    } else {
        *value = FAULT_STATUS_OK;
    }
    if (alarm_state != *value) {
        alarm_state = *value;
    }
    return EDGE_OK;
}

int check_cert_update_flag(void)
{
    int flag = 0;
    /* 存在证书告警阈值更新文件标示 */
    if (access(CERT_FAULTTIME_UPDATE_FILE, F_OK) == 0) {
        (void)unlink(CERT_FAULTTIME_UPDATE_FILE);
        flag = 1;
    }

    /* 存在证书更新文件标示 */
    if (access(CERT_UPDATE_FILE, F_OK) == 0) {
        (void)unlink(CERT_UPDATE_FILE);
        flag = 1;
    }
    return flag;
}
