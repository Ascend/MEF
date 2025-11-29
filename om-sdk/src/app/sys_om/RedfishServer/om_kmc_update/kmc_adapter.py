#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from common.kmc_lib.kmc_adapter import TableAdapter
from om_event_subscription.models import Subscription
from redfish_db.session import session_maker


class SubscriptionPsdAdapter(TableAdapter):
    """订阅destination及credential加密字段的Kmc密钥更新适配器"""

    session = session_maker
    model = Subscription
    filter_by = "id"
    cols = ("destination", "credential")
