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

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestEventService(GetTest):
    BASE_URL = "/redfish/v1/EventService"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.BASE_URL, code=code, label=label)

    def call_back_assert(self, test_response: str):
        assert self.expect_ret in test_response


def init_test_get_event_service():
    TestEventService(
        expect_ret=json.dumps(
            {
                "@odata.context": "/redfish/v1/$metadata#EventService",
                "@odata.id": "/redfish/v1/EventService",
                "@odata.type": "#EventService.v1_8_0.EventService",
                "Id": "EventService",
                "Name": "EventService",
                "Subscriptions": {
                    "@odata.id": "/redfish/v1/EventService/Subscriptions"
                },
                "Oem": {
                    "ServiceCert": {
                        "@odata.id": "/redfish/v1/EventService/ServiceCert"
                    }
                }
            }
        ),
        label="test get event service success!",
        code=200,
    )


init_test_get_event_service()
