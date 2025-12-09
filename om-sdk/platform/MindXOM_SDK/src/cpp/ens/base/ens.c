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

#include "ens.h"
#include "file_checker.h"
#include "ens_log.h"
#include "ens_base.h"
#include "ens_context.h"

static char ens_home[COMMON_VALUE_LEN_256];

static int ens_proc_args(int argc, char *argv[])
{
    int ch;
    char *p = NULL;

    while ((ch = getopt(argc, argv, "c:V")) != -1) {
        switch (ch) {
            case 'c':
                p = optarg;
                break;
            case 'V':
                break;
            case '?':
                break;
            default:
                ENS_LOG_ERR("Unexpected options [%c].", ch);
                break;
        }
    }

    if ((p == NULL) || (strlen(p) >= COMMON_VALUE_LEN_256)) {
        ENS_LOG_WARN("fail to get path.");
        return ENS_ERR;
    }

    char *path = malloc(COMMON_VALUE_LEN_256);
    if (path == NULL) {
        ENS_LOG_ERR("malloc fail.");
        return ENS_ERR;
    }

    if ((sprintf_s(path, COMMON_VALUE_LEN_256 - 1, "%s", p) < 0) ||
        (check_file_path_valid(path) != COMMON_SUCCESS)) {
        free(path);
        path = NULL;
        ENS_LOG_ERR("sprintf_s fail. path.");
        return ENS_ERR;
    }

    ens_ctx.config.file_name = path;
    return ENS_OK;
}

static int ens_get_home_path_by_file(char *path)
{
    int ret = ENS_ERR;

    ens_ctx.home_path = NULL;
    if (check_dir_path_valid(path) != COMMON_SUCCESS) {
        ENS_LOG_FATAL("invalid path.");
        return ENS_ERR;
    }

    char *bin_dir = dirname(path);
    if (bin_dir != NULL) {
        char *home_dir = dirname(bin_dir);
        if (home_dir == NULL) {
            ENS_LOG_FATAL("Failed get HOME_PATH by executable.");
            return ENS_ERR;
        }

        size_t homepath_len = strlen(home_dir) + 1;
        if (homepath_len > COMMON_VALUE_LEN_256) {
            ENS_LOG_FATAL("invalid homepath_len:%lu.", homepath_len);
            return ENS_ERR;
        }

        if (memset_s(ens_home, COMMON_VALUE_LEN_256, 0, COMMON_VALUE_LEN_256) != 0) {
            ENS_LOG_ERR("memset_s failed.");
            return ENS_ERR;
        }
        if (strcpy_s(ens_home, homepath_len, home_dir) != 0) {
            ENS_LOG_FATAL("strcpy fail");
        } else {
            ret = ENS_OK;
            ens_ctx.home_path = ens_home;
        }
    } else {
        ENS_LOG_FATAL("Failed get executable's directory.");
    }
    return ret;
}

static int get_default_ens_config_file(void)
{
    if (ens_ctx.config.file_name != NULL) {
        ENS_LOG_WARN("config file has been got,not need to get default config.");
        return ENS_OK;
    }

    char *file_name = malloc(COMMON_VALUE_LEN_256);
    if (file_name == NULL) {
        ENS_LOG_ERR("malloc fail.");
        return ENS_ERR;
    }

    if (memset_s(file_name, COMMON_VALUE_LEN_256, 0, COMMON_VALUE_LEN_256) != 0) {
        ENS_LOG_ERR("memset_s fail, exit.");
        free(file_name);
        file_name = NULL;
        return ENS_ERR;
    }

    if ((sprintf_s(file_name, COMMON_VALUE_LEN_256 - 1, "%s/conf/ensd.conf", ens_ctx.home_path) < 0) ||
        (check_file_path_valid(file_name) != COMMON_SUCCESS)) {
        ENS_LOG_ERR("invalid path: ensd.conf.");
        free(file_name);
        file_name = NULL;
        return ENS_ERR;
    }

    ens_ctx.config.file_name = file_name;
    return ENS_OK;
}

int ens_work(int argc, char *argv[])
{
    if (memset_s(&ens_ctx, sizeof(ens_ctx), 0, sizeof(ens_ctx)) != 0) {
        ENS_LOG_ERR("memset fail");
        return ENS_ERR;
    }

    if (get_env_var_dir("ENS_HOME", ens_home, COMMON_VALUE_LEN_256) == COMMON_SUCCESS) {
        ens_ctx.home_path = ens_home;
    } else {
        if (ens_get_home_path_by_file(argv[0]) != ENS_OK) {
            ENS_LOG_FATAL("Failed get ENS_HOME, exit.");
            return ENS_ERR;
        }
    }

    if (ens_proc_args(argc, argv) != ENS_OK) {
        ENS_LOG_WARN("proc args failed.");
    }

    if (get_default_ens_config_file() != ENS_OK) {
        ENS_LOG_WARN("get default_ens_config_file failed.");
        return ENS_ERR;
    }

    int ret = ens_ctx_initialize();
    free(ens_ctx.config.file_name);
    ens_ctx.config.file_name = NULL;
    if (ret != 0) {
        ENS_LOG_FATAL("Initilization failed.");
        return ENS_ERR;
    }

    ENS_LOG_WARN("ensd init ok!");
    return ENS_OK;
}
