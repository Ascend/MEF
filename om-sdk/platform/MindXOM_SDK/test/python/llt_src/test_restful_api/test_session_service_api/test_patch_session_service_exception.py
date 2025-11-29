# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import PatchTest


class TestPatchSessionService(PatchTest):
    SESSION_SERVICE_URL = "/redfish/v1/SessionService"

    def __init__(self, expect_ret, code: int, label: str, data):
        self.expect_ret = expect_ret
        self.data = data
        self.patch = None
        super().__init__(url=self.SESSION_SERVICE_URL, code=code, label=label, data=self.data)

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                           side_effect=Exception())
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_get_session_service_instances_exception():
    TestPatchSessionService(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Modify session info failed.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": None}}]}}),
                            label="test patch sessions failed due to data is not json!",
                            code=400,
                            data={"SessionTimeout": 5, "Password": "test"},
                            )


init_get_session_service_instances_exception()