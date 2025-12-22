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

from om_event_subscription.subscription_views import subs_cert_mgr
from test_restful_api.test_z_main.restful_test_base import PostTest


class TestImportRootCert(PostTest):
    BASE_URL = "/redfish/v1/EventService/ServiceCert/Actions/ServiceCert.ImportRemoteHttpsServiceRootCA"

    def __init__(self, expect_ret, code: int, label: str, data: dict):
        self.expect_ret = expect_ret
        self.data = data
        super().__init__(url=self.BASE_URL,
                         code=code,
                         data=data,
                         label=label)

    def before(self):
        self.patch1 = patch.object(subs_cert_mgr, "overwrite_subs_cert", return_value=None)
        self.patch1.start()

    def after(self):
        if self.patch1:
            self.patch1.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret in test_response


def init_test_import_root_cert():
    TestImportRootCert(
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
        label="test import root cert failed due to invalid cert!",
        code=400,
        data={
                "Type": "text",
                "Content": "1111",
                "RootCertId": 1
        }
    )


init_test_import_root_cert()
