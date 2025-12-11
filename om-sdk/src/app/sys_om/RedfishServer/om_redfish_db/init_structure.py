#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

"""
首建模型字段持久化(需要平滑的表才记录)
    当首次新增了一个sqlite数据库ORM模型，必须将模型对应的字段名写入到此文件中的INIT_COLUMNS中；
    当版本升级修改了sqlite数据库ORM模型，不能更改INIT_COLUMNS对应表的初始字段；
"""
from typing import Dict, Iterable

INIT_COLUMNS: Dict[str, Iterable[str]] = {
    "subscriptions": ("id", "destination", "types", "protocol", "credential", "create_time", "update_time"),
    "subscriptions_cert": ("id", "root_cert_id", "type", "cert_contents", "crl_contents"),
    "active_alarm": ("id", "alarm_id", "alarm_instance", "alarm_name", "type", "severity", "create_time"),
    "alarm_report_task": ("id", "subscriber_id", "event_type", "event_id", "event_name", "severity", "event_detail",
                          "reason", "extra_column", "event_timestamp", "task_status"),
}
