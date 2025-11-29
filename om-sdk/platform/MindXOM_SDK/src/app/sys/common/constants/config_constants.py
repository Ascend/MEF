# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.


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
