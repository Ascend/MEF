#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from om_event_subscription.subscription_mgr import BaseManager
from om_event_subscription.models import ActiveAlarm


class ActiveAlarmManager(BaseManager):
    model = ActiveAlarm
