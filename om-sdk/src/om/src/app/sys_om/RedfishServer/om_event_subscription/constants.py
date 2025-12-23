#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from enum import Enum

# 支持的订阅ID
MIN_SUBSCRIPTION_ID = 1
MAX_SUBSCRIPTION_ID = 1

# 告警ID字段数量
ALARM_ID_NUM = 3

# 告警分隔符
ALARM_SEPARATOR = "&"

# kmc 加解密的文件路径
ESIGHT_KSF = "/home/data/config/redfish/redfish_encrypt.keystore"
ESIGHT_BAK_KSF = "/home/data/config/redfish/redfish_encrypt_backup.keystore"
ESIGHT_ALG_CONFIG = "/home/data/config/redfish/om_alg.json"

# 证书以及吊销列表最大size
CERT_CRL_FILE_MAX_SIZE_BYTES = 10 * 1024


class EventTypes(Enum):
    """
    功能描述：支持的订阅类型
    """
    STATUS_CHANGE = "StatusChange"
    RESOURCE_UPDATED = "ResourceUpdated"
    RESOURCE_ADDED = "ResourceAdded"
    RESOURCE_REMOVED = "ResourceRemoved"
    ALERT = "Alert"

    @classmethod
    def values(cls):
        return [elem.value for elem in cls]

    @classmethod
    def get_supported_types(cls):
        return [cls.ALERT.value]


class AlarmTypes(Enum):
    """支持的告警类型"""
    ALERT = "Alert"

    @classmethod
    def values(cls):
        return [elem.value for elem in cls]


class TaskStatus(Enum):
    """告警任务状态"""
    WAIT_ADD = 1
    ADDED = 2
    WAIT_DELETE = 3
    DELETED = 4

    @classmethod
    def values(cls):
        return [elem.value for elem in cls]

    @classmethod
    def wait_add_and_delete(cls):
        return cls.WAIT_ADD.value, cls.WAIT_DELETE.value

    @classmethod
    def added_and_deleted(cls):
        return cls.ADDED.value, cls.DELETED.value


class AlarmStatus(Enum):
    """告警状态"""
    DELETED = 0
    ACTIVE = 1

    @classmethod
    def values(cls):
        return [elem.value for elem in cls]


TASK_ALARM_STATUS_RELATIONSHIP = {
    TaskStatus.WAIT_DELETE.value: AlarmStatus.DELETED.value,
    TaskStatus.WAIT_ADD.value: AlarmStatus.ACTIVE.value,
}


class ESightSeverity(Enum):
    CRITICAL = "Critical"
    MAJOR = "Major"
    MINOR = "Minor"

    @classmethod
    def values(cls):
        return [elem.value for elem in cls]


class OmSeverity(Enum):
    CRITICAL = 0
    MAJOR = 1
    MINOR = 2

    @classmethod
    def values(cls):
        return [elem.value for elem in cls]


ALARM_SEVERITY_MAPPING = {
    "eSight": {
        OmSeverity.CRITICAL.value: ESightSeverity.CRITICAL.value,
        OmSeverity.MAJOR.value: ESightSeverity.MAJOR.value,
        OmSeverity.MINOR.value: ESightSeverity.MINOR.value,
    }
}

RESP_STATUS_LIST = (200, 201)


class Protocol(Enum):
    """
    功能描述：支持使用的协议类型
    """
    REDFISH = "Redfish"
