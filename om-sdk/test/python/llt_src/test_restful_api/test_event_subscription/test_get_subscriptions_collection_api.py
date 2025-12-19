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


import json

from mock import patch

from om_event_subscription.subscription_views import subs_mgr
from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetSubscriptionCollection(GetTest):
    BASE_URL = "/redfish/v1/EventService/Subscriptions"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.BASE_URL, code=code, label=label)

    def before(self):
        self.patch = patch.object(subs_mgr, "get_first_subscription", return_value=[])
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret in test_response


def init_test_get_subscription_collection():
    TestGetSubscriptionCollection(
        expect_ret=json.dumps(
            {
                "@odata.context": "/redfish/v1/$metadata#EventService/Subscriptions/$entity",
                "@odata.id": "/redfish/v1/EventService/Subscriptions",
                "@odata.type": "#EventDestinationCollection.EventDestinationCollection",
                "Name": "Event Subscriptions Collection",
                "Members@odata.count": 0,
                "Members": []
            }
        ),
        label="test get subscription collection!",
        code=200,
    )


init_test_get_subscription_collection()
