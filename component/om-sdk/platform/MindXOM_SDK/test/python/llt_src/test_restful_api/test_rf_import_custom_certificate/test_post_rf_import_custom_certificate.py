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

from test_restful_api.test_z_main.restful_test_base import PostTest
from user_manager.user_manager import UserManager


class TestPostRfImportCustomCertificate(PostTest):
    """查询系统信息"""
    RF_HTTP_CERT_COLLECTION_URL = "/redfish/v1/Systems/SecurityService/HttpsCert/" \
                                  "Actions/HttpsCert.ImportServerCertificate"

    def __init__(self, expect_ret, code: int, label: str, data, patch_return):
        self.expect_ret = expect_ret
        self.patch_lib = None
        self.patch_user = None
        self.patch_return = patch_return
        super().__init__(url=self.RF_HTTP_CERT_COLLECTION_URL,
                         code=code,
                         label=label,
                         data=data
                         )

    def before(self):
        self.patch_user = patch.object(UserManager, "get_all_info", return_value=self.patch_return[0])
        self.patch_lib = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                               return_value=self.patch_return[1])
        self.patch_user.start()
        self.patch_lib.start()

    def after(self):
        if self.patch_lib:
            self.patch_lib.stop()

        if self.patch_user:
            self.patch_user.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def test_post_rf_import_custom_certificate():
    TestPostRfImportCustomCertificate(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.MalformedJSON",
                       "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [{
                           "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                           "Description": "Indicates that the request body was malformed JSON.  "
                                          "Could be duplicate, syntax error,etc.",
                           "Message": "The request body submitted was malformed JSON "
                                      "and could not be parsed by the receiving service.",
                           "Severity": "Critical", "NumberOfArgs": 0, "ParamTypes": None,
                           "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                           "Oem": {"status": None}}]}}
        ),
        code=400,
        label="Import custom certificate failed due to data is not json.",
        data="test",
        patch_return=[{}, {}]
    )
    TestPostRfImportCustomCertificate(
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
        label="Import custom certificate failed due to FileName has '..'!",
        data={"FileName": "/test/../test.txt", "Password": "Aa@123456"},
        patch_return=[{}, {}]
    ),
    TestPostRfImportCustomCertificate(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.Success",
                       "message": "Operation success. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that no error has occurred.",
                            "Message": "Import custom certificate successfully.",
                            "Severity": "OK", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            }]}}),
        code=202,
        label="Import custom certificate success!",
        data={"FileName": "test.crt", "Password": "Aa@123456"},
        patch_return=[{"status": 200, "message": "success"}, {"status": 200, "message": "test"}]
    ),
    TestPostRfImportCustomCertificate(
        expect_ret=json.dumps(
            {
                "error": {"code": "Base.1.0.Success",
                          "message": "Operation success. See ExtendedInfo for more information.",
                          "@Message.ExtendedInfo": [
                              {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                               "Description": "Indicates that no error has occurred.",
                               "Message": "Import custom certificate successfully, but legality is risky.",
                               "Severity": "OK", "NumberOfArgs": None, "ParamTypes": None,
                               "Resolution": "None",
                               }]}}),
        code=206,
        label="Import custom certificate ERR.006!",
        data={"FileName": "test.crt", "Password": "Aa@123456"},
        patch_return=[{"status": 200, "message": "success"}, {"status": 400,
                      "message": [110307, "Importing a custom certificate success but legality is risky"]}]
    )
    TestPostRfImportCustomCertificate(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Import custom certificate failed.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": None}}]}},
        ),
        code=400,
        label="Import custom certificate ERR.003!",
        data={"FileName": "test.crt", "Password": "Aa@123456"},
        patch_return=[{"status": 200, "message": "success"},
                      {"status": 400, "message": "Import custom certificate failed."}]
    )
    TestPostRfImportCustomCertificate(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Import custom certificate failed.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": None}}]}},
        ),
        code=400,
        label="Import custom certificate exception!",
        data={"FileName": "test.crt", "Password": "Aa@123456"},
        patch_return=[{"status": 200, "message": "success"}, "test"]
    )


test_post_rf_import_custom_certificate()
