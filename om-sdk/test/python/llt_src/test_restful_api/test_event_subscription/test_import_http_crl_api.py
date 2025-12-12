#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.


import json

from mock import patch

from om_event_subscription.subscription_mgr import SubscriptionsCertMgr
from test_restful_api.test_z_main.restful_test_base import PostTest


class TestImportHttpCrl(PostTest):
    BASE_URL = "/redfish/v1/EventService/ServiceCert/Actions/ServiceCert.ImportRemoteHttpsServiceCrl"

    def __init__(self, expect_ret, code: int, label: str, data: dict):
        self.expect_ret = expect_ret
        self.data = data
        super().__init__(url=self.BASE_URL,
                         code=code,
                         data=data,
                         label=label)

    def before(self):
        self.patch1 = patch.object(SubscriptionsCertMgr, "get_first",
                                   return_value={"cert_contents": "-----BEGIN CERTIFICATE-----\n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\nr3KeMilgYMY/qac8iw==\n-----END CERTIFICATE-----\n"})
        self.patch1.start()
        self.patch2 = patch.object(SubscriptionsCertMgr, "update_crt_with_crl", return_value=None)
        self.patch2.start()

    def after(self):
        if self.patch1:
            self.patch1.stop()
        if self.patch2:
            self.patch2.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret in test_response


def init_test_import_http_crl():
    TestImportHttpCrl(
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
                                "status": None
                            }
                        }
                    ]
                }
            }
        ),
        label="test import http crl failed due to invalid cert!",
        code=400,
        data={
                "Type": "text",
                "Content": "1111",
                "RootCertId": 1
        }
    )


init_test_import_http_crl()
