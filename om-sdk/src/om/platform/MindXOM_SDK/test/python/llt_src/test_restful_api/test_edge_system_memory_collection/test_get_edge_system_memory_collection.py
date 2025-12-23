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
from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetRfEdgeSystemMemoryCollection(GetTest):
    """查询内存概要信息"""
    RF_EDGE_SYSTEM_COLLECTION_URL = "/redfish/v1/Systems/Memory"

    def __init__(self, expect_ret, code: int, label: str, patch_return):
        self.expect_ret = expect_ret
        self.patch = None
        self.patch_return = patch_return
        super().__init__(url=self.RF_EDGE_SYSTEM_COLLECTION_URL,
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


def test_get_rf_edge_system_memory_collection():
    TestGetRfEdgeSystemMemoryCollection(
        expect_ret=json.dumps({
                    "@odata.type": "#Memory.v1_15_0.Memory",
                    "@odata.context": "/redfish/v1/$metadata#Systems/Memory",
                    "@odata.id": "/redfish/v1/Systems/Memory",
                    "Id": "Memory",
                    "Name": "Memory",
                    "Oem": {
                        "TotalSystemMemoryGiB": None
                    }
                }),
        code=200,
        label="get system memory success",
        patch_return={"status": 200, "message": {"test": "test"}}
    )
    TestGetRfEdgeSystemMemoryCollection(
        expect_ret=json.dumps({
                "error":{
                "code": "Base.1.0.GeneralError",
                "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                "@Message.ExtendedInfo": [{
                    "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                    "Description": "Indicates that a general error has occurred.",
                    "Message": "Internal server error",
                    "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                    "Resolution": "None",
                    "Oem": {
                        "status": 100011
                    }}]}
            }),
        code=500,
        label="get system memory exception",
        patch_return={"status": 200, "message": "test"}
    )


test_get_rf_edge_system_memory_collection()