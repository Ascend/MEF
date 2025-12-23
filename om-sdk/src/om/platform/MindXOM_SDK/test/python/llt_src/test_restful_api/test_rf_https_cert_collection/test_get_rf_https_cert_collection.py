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


class TestGetRfHttpsCertCollection(GetTest):
    """查询SSL证书资源信息"""
    RF_HTTP_CERT_COLLECTION_URL = "/redfish/v1/Systems/SecurityService/HttpsCert"

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


def test_get_rf_https_cert_collection():
    TestGetRfHttpsCertCollection(
        expect_ret=json.dumps(
            {"@odata.context": "/redfish/v1/$metadata#Systems/SecurityService/HttpsCert/$entity",
             "@odata.id": "/redfish/v1/Systems/SecurityService/HttpsCert",
             "@odata.type": "#MindXEdgeHttpsCert.v1_0_0.MindXEdgeHttpsCert", "Id": "HttpsCert",
             "Name": "Https cert info",
             "X509CertificateInformation": {
                 "ServerCert": {"Subject": None, "Issuer": None, "ValidNotBefore": None, "ValidNotAfter": None,
                                "SerialNumber": None, "FingerPrint": None, "HttpsCertEable": None,
                                "ExpiredDayRemaining": None}},
             "CertificateSigningRequest": None,
             "Actions": {
                 "#HttpsCert.ImportServerCertificate":
                     {
                         "target": "/redfish/v1/Systems/SecurityService/HttpsCert"
                                   "/Actions/HttpsCert.ImportServerCertificate"},
                 }}
        ),
        code=200,
        label="Query security service success",
        patch_return={"status": 200, "message": {"test": "test"}}
    )
    TestGetRfHttpsCertCollection(
        expect_ret=json.dumps(
            {"error": {
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
                    }}]}},
        ),
        code=500,
        label="Query security service exception.",
        patch_return={"status": 200, "message": "test"}
    )


test_get_rf_https_cert_collection()
