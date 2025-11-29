# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetRfEdgeSystemAiCollection(GetTest):
    """查询系统信息"""
    RF_EDGE_SYSTEM_AI_COLLECTION_URL = "/redfish/v1/Systems/Processors/AiProcessor"

    def __init__(self, expect_ret, code: int, label: str, patch_return):
        self.expect_ret = expect_ret
        self.patch = None
        self.patch_return = patch_return
        super().__init__(url=self.RF_EDGE_SYSTEM_AI_COLLECTION_URL,
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


def test_get_rf_edge_system_ai_collection():
    TestGetRfEdgeSystemAiCollection(
        expect_ret=json.dumps(
            {
                "@odata.context": "/redfish/v1/$metadata#Systems/Processors/Members/$entity",
                "@odata.id": "/redfish/v1/Systems/Processors/AiProcessor",
                "@odata.type": "#Processor.v1_15_0.Processor",
                "Name": "AiProcessor",
                "Id": "AiProcessor",
                "Manufacturer": None,
                "Model": None,
                "NpuVersion": None,
                "Status": {
                    "Health": None,
                    "State": None
                },
                "Oem": {
                    "Count": None,
                    "Capability": {
                        "Calc": None,
                        "Ddr": None
                    },
                    "OccupancyRate": {
                        "AiCore": None,
                        "AiCpu": None,
                        "CtrlCpu": None,
                        "DdrUsage": None,
                        "DdrBandWidth": None
                    }
                }
            }
        ),
        code=200,
        label="get system ai success",
        patch_return={"status": 200, "message": {"test": "test"}}
    )
    TestGetRfEdgeSystemAiCollection(
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
        label="get system ai exception",
        patch_return={"status": 200, "message": "test"}
    )


test_get_rf_edge_system_ai_collection()
