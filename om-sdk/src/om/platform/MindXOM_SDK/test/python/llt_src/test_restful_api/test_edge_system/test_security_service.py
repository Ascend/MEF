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

from config_urls import EdgeSystemUrls
from test_class import RequestTest


def test_security_service_patch():
    """2.5.42 修改证书过期提醒时间"""

    payload1 = {
        "label": "test_security_service_patch: error password",
        "request_obj": {
            "method": "PATCH",
            "url": EdgeSystemUrls.SECURITY_SERVICE + 'HttpsCertAlarmTime',
            "params": "",
            "data": {
                "CertAlarmTime": 100,
                "Password": "password"
            }
        },
        "response_obj": {
            "code":
            400,
            "expect_response":
            json.dumps(
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Internal server error",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": 100011}}]}}
            ),
        },
    }

    payload2 = {
        "label": "test_security_service_patch: error param",
        "request_obj": {
            "method": "PATCH",
            "url": EdgeSystemUrls.SECURITY_SERVICE + 'HttpsCertAlarmTime',
            "params": "",
            "data": {
                "CertAlarmTime": "wq",
                "Password": "password"
            }
        },
        "response_obj": {
            "code":
            400,
            "expect_response":
            json.dumps(
                {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
        },
    }

    RequestTest(**payload1)
    RequestTest(**payload2)


test_security_service_patch()
