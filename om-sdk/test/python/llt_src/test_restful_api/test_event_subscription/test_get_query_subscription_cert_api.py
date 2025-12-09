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

from mock import patch

from om_event_subscription.subscription_mgr import SubscriptionsCertMgr
from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetQuerySubscriptionCert(GetTest):
    BASE_URL = "/redfish/v1/EventService/ServiceCert"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.BASE_URL, code=code, label=label)

    def before(self):
        self.patch1 = patch.object(SubscriptionsCertMgr, "get_obj_by_id", return_value=1)
        self.patch1.start()

    def after(self):
        if self.patch1:
            self.patch1.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret in test_response


def init_test_get_query_subscription_cert():
    TestGetQuerySubscriptionCert(
        expect_ret=json.dumps(
            {
                "@odata.context": "/redfish/v1/$metadata#EventService/ServiceCert/$entity",
                "@odata.id": "/redfish/v1/EventService/ServiceCert",
                "@odata.type": "#MindXEdgeSecurityService.v1_0_0.MindXEdgeSecurityService",
                "Id": "ServiceCert",
                "Name": "ServiceCert",
                "RemoteHttpsServerCertChainInfo": [
                    {
                        "CertId": 1,
                        "Usage": "EventSubscription"
                    }
                ],
                "Actions": {
                    "#ServiceCert.ImportRemoteHttpsServiceRootCA": {
                        "target": "/redfish/v1/EventService/ServiceCert/Actions/ServiceCert.ImportRemoteHttpsServiceRootCA"
                    },
                    "#ServiceCert.DeleteRemoteHttpsServiceRootCA": {
                        "target": "/redfish/v1/EventService/ServiceCert/Actions/ServiceCert.DeleteRemoteHttpsServiceRootCA"
                    },
                    "#ServiceCert.ImportRemoteHttpsServiceCrl": {
                        "target": "/redfish/v1/EventService/ServiceCert/Actions/ServiceCert.ImportRemoteHttpsServiceCrl"
                    }
                }
            }
        ),
        label="test get query subscription cert successfully!",
        code=200,
    ),


init_test_get_query_subscription_cert()
