# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.


class ConfigPathConstants:
    """
    配置文件路径常量类
    """
    ALARM_INFO_EN_JSON = "/usr/local/mindx/MindXOM/software/RedfishServer/config/alarm_info_en.json"
    DEFAULT_CAPABILITY_FILE = "/usr/local/mindx/MindXOM/software/RedfishServer/config/default_capability.json"
    SYS_CONFIG_PATH = "/home/data/config/redfish/"
    ETC_RESOLV_PATH = "/etc/resolv.conf"


class BaseConfigPermissionConstants:
    # 配置文件目录权限要求
    CONFIG_DIR_MODE = "750"
    CONFIG_DIR_USER = "root"
    CONFIG_DIR_GROUP = "root"

    # 配置文件权限要求
    CONFIG_FILE_MODE = "640"
    CONFIG_FILE_USER = "root"
    CONFIG_FILE_GROUP = "root"
