#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from typing import Type

from common.db.database import DataBase
from om_event_subscription.models import Subscription, SubscriptionCert, ActiveAlarm, AlarmReportTask, SubsPreCert


def register_extend_om_models(database: Type[DataBase]):
    database.register_models(Subscription, SubscriptionCert, ActiveAlarm, AlarmReportTask, SubsPreCert)
