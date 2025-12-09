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

from common.kmc_lib.kmc_adapter import TableAdapter
from om_event_subscription.models import Subscription
from redfish_db.session import session_maker


class SubscriptionPsdAdapter(TableAdapter):
    """订阅destination及credential加密字段的Kmc密钥更新适配器"""

    session = session_maker
    model = Subscription
    filter_by = "id"
    cols = ("destination", "credential")
