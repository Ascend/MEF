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
    RF_SYSTEM1_ETHERNET_ETH_X_URL = "/redfish/v1/Systems/EthernetInterfaces/"

    def __init__(self, expect_ret, code: int, label: str, eth_id, patch_return):
        self.expect_ret = expect_ret
        self.patch = None
        self.patch_return = patch_return
        super().__init__(url=self.RF_SYSTEM1_ETHERNET_ETH_X_URL + str(eth_id),
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


def test_get_rf_system1_ethernet_eth_x():
    TestGetRfSystem1EthernetCollection(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
        code=400,
        label="Get ethernet interface info failed due to eth_id wrong.",
        patch_return=None,
        eth_id="1"
    )
    TestGetRfSystem1EthernetCollection(
        expect_ret=json.dumps(
            {"@odata.context": "/redfish/v1/$metadata#Systems/Members/EthernetInterfaces/Members/$entity",
             "@odata.id": "/redfish/v1/Systems/EthernetInterfaces/GMAC0",
             "@odata.type": "#EthernetInterface.v1_8_0.EthernetInterface",
             "Id": "GMAC0", "Name": None, "Description": "Ethernet Interface over Wired Network Adapter",
             "PermanentMACAddress": None, "MACAddress": None, "InterfaceEnabled": None,
             "LinkStatus": None, "IPv4Addresses": [], "IPv6Addresses": [],
             "IPv6DefaultGateway": None, "NameServers": [],
             "Oem": {
                 "WorkMode": None, "AdapterType": None, "LteDataSwitchOn": False, "Connections": [],
                 "Statistic": {
                     "SendPackages": None, "RecvPackages": None, "ErrorPackages": None, "DropPackages": None}}}
        ),
        code=200,
        label="Get ethernet interface info success.",
        patch_return={"status": 200, "message": {"eth_id": "GMAC0"}},
        eth_id="GMAC0"
    )
    TestGetRfSystem1EthernetCollection(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Internal server error",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100011}}]}},
        ),
        code=500,
        label="Get ethernet interface info exception.",
        patch_return={"status": 200, "message": "eth_id"},
        eth_id="GMAC0"
    )


test_get_rf_system1_ethernet_eth_x()
