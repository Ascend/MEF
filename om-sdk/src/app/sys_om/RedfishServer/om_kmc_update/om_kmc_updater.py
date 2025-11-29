#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from typing import Type

from common.kmc_lib.kmc_updater import MultiKmcUpdater
from om_kmc_update.kmc_adapter import SubscriptionPsdAdapter


def extend_om_updater_and_adapters(OMKmcUpdater: Type[MultiKmcUpdater]):
    OMKmcUpdater.extend_adapters("redfish", SubscriptionPsdAdapter)
