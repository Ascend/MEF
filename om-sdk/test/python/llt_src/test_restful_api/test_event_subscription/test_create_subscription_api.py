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


import json

from unittest.mock import patch

from om_event_subscription.models import Subscription
from om_event_subscription.subscription_views import subs_mgr
from test_restful_api.test_z_main.restful_test_base import PostTest


class TestPostCreateSubscription(PostTest):
    BASE_URL = "/redfish/v1/EventService/Subscriptions"

    def __init__(self, expect_ret, code: int, label: str, data: dict):
        self.expect_ret = expect_ret
        self.data = data
        super().__init__(url=self.BASE_URL,
                         code=code,
                         data=data,
                         label=label)

    def before(self):
        self.patch1 = patch.object(subs_mgr, "check_destination_existed", return_value=False)
        self.patch1.start()
        self.patch2 = patch.object(subs_mgr, "get_subscription_num", return_value=0)
        self.patch2.start()
        self.patch3 = patch.object(subs_mgr, "get_available_min_id", return_value=1)
        self.patch3.start()
        self.patch4 = patch.object(subs_mgr, "add_subscription", return_value=None)
        self.patch4.start()
        self.patch5 = patch.object(Subscription, "from_dict", return_value={})
        self.patch5.start()

    def after(self):
        if self.patch1:
            self.patch1.stop()
        if self.patch2:
            self.patch2.stop()
        if self.patch3:
            self.patch3.stop()
        if self.patch4:
            self.patch4.stop()
        if self.patch5:
            self.patch5.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret in test_response


def init_test_post_create_subscription():
    TestPostCreateSubscription(
        expect_ret=json.dumps(
            {
                "@odata.context": "/redfish/v1/$metadata#EventService/Subscriptions/$entity",
                "@odata.id": "/redfish/v1/EventService/Subscriptions",
                "@odata.type": "#EventDestination.v1_1_1.EventDestination",
                "Id": 1,
                "Name": "EventSubscription 1",
                "Destination": "https://51.38.68.206:8081/alarm/event",
                "EventTypes": [
                    "Alert"
                ],
                "HttpHeaders": None,
                "Protocol": "Redfish",
                "Context": "event subscription context string"
            }
        ),
        data={
            "Destination": "https://51.38.68.206:8081/alarm/event",
            "EventTypes": ["Alert"],
            "HttpHeaders": {
                "X-Auth-Token": "this-is-token-eSight-test"
            },
            "Context": "event subscription context string",
            "Protocol": "Redfish"
        },
        label="test post create subscription successfully!",
        code=200,
    ),

    TestPostCreateSubscription(
        expect_ret=json.dumps(
            {
                "error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 100024
                            }
                        }
                    ]
                }
            }
        ),
        data={
            "Destination": "https://51.38.68.206:8081/alarm/event",
            "EventTypes": ["Alert"],
            "HttpHeaders": {
                "X-Auth-Token": "this-is-token-eSight-test"
            },
            "Context": "event subscription context string",
            "Protocol": "Redfish111"
        },
        label="test post create subscription failed due to parameter is invalid!",
        code=400,
    ),


init_test_post_create_subscription()
