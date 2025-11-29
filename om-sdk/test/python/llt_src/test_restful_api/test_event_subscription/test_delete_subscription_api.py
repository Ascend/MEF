#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.


import json

from mock import patch

from om_event_subscription.subscription_views import subs_mgr
from test_restful_api.test_z_main.restful_test_base import DeleteTest


class TestDeleteSubscription(DeleteTest):
    BASE_URL = "/redfish/v1/EventService/Subscriptions/"

    def __init__(self, expect_ret, code: int, label: str, subscription_id: int):
        self.expect_ret = expect_ret
        self.patch1 = None
        send_url = self.BASE_URL + str(subscription_id)
        super().__init__(url=send_url, code=code, label=label, data={})

    def before(self):
        self.patch1 = patch.object(subs_mgr, "delete_subscription", return_value=1)
        self.patch1.start()

    def after(self):
        if self.patch1:
            self.patch1.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret in test_response


def init_test_delete_subscription():
    TestDeleteSubscription(
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
        label="test delete subscription failed due to subscription_id is invalid!",
        code=400,
        subscription_id=2,
    ),
    TestDeleteSubscription(
        expect_ret=json.dumps(
            {
                "error": {
                    "code": "Base.1.0.Success",
                    "message": "Operation success. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [{
                        "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                        "Description": "Indicates that no error has occurred.",
                        "Message": "Success",
                        "Severity": "OK",
                        "NumberOfArgs": None,
                        "ParamTypes": None,
                        "Resolution": "None"
                    }]
                }
            }
        ),
        label="test delete subscription successfully!",
        code=200,
        subscription_id=1,
    ),


init_test_delete_subscription()
