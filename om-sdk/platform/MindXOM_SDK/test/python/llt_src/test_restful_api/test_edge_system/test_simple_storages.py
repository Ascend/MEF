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

from config_urls import EdgeSystemUrls
from test_class import RequestTest


def test_simple_storage_get():
    """2.5.20 查询简单存储资源信息"""

    payload = {
        "label": 'test simple storage api when storage_id is "?"',
        "request_obj": {
            "method": "GET",
            "url": EdgeSystemUrls.SIMPLE_STORAGES,
            "params": "?",
            "data": "",
        },
        "response_obj": {
            "code":
            404,
            "expect_response":
            json.dumps({
                "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "The requested URL was not found on the server",
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
            }),
        },
    }

    payload2 = {
        "label": 'test simple storage api when storage_id is "ddd"',
        "request_obj": {
            "method": "GET",
            "url": EdgeSystemUrls.SIMPLE_STORAGES,
            "params": "ddd",
            "data": "",
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

    payload3 = {
        "label": 'test simple storage api when storage_id is "&?id=12"',
        "request_obj": {
            "method": "GET",
            "url": EdgeSystemUrls.SIMPLE_STORAGES,
            "params": "&?id=12",
            "data": "",
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

    RequestTest(**payload)
    RequestTest(**payload2)
    RequestTest(**payload3)


test_simple_storage_get()
