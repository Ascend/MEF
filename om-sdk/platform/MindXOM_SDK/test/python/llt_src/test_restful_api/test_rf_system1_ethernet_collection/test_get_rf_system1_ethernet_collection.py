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

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetRfSystem1EthernetCollection(GetTest):
    """获取以太网接口集合信息"""
    RF_HTTP_CERT_COLLECTION_URL = "/redfish/v1/Systems/EthernetInterfaces"

    def __init__(self, expect_ret, code: int, label: str, patch_return):
        self.expect_ret = expect_ret
        self.patch = None
        self.patch_return = patch_return
        super().__init__(url=self.RF_HTTP_CERT_COLLECTION_URL,
                         code=code,
                         label=label,
                         )

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                           return_value=self.patch_return)
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def test_get_rf_system1_ethernet_collection():
    TestGetRfSystem1EthernetCollection(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Query ethernet interface resources failed.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": None}}]}}
        ),
        code=400,
        label="Query ethernet interface resources exception.",
        patch_return={"status": 400, "message": "Query ethernet interface resources failed."}
    )
    TestGetRfSystem1EthernetCollection(
        expect_ret=json.dumps(
            {"@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
             "@odata.context": "/redfish/v1/$metadata#Systems/Members/EthernetInterfaces/$entity",
             "@odata.id": "/redfish/v1/Systems/EthernetInterfaces", "Name": "Ethernet Interface Collection",
             "Description": "System NICs on Servers", "Members@odata.count": 4,
             "Members": [{"@odata.id": "/redfish/v1/Systems/EthernetInterfaces/t"},
                         {"@odata.id": "/redfish/v1/Systems/EthernetInterfaces/e"},
                         {"@odata.id": "/redfish/v1/Systems/EthernetInterfaces/s"},
                         {"@odata.id": "/redfish/v1/Systems/EthernetInterfaces/t"}]},
        ),
        code=200,
        label="Query ethernet interface resources success.",
        patch_return={"status": 200, "message": ["t", "e", "s", "t"]}
    )


test_get_rf_system1_ethernet_collection()
