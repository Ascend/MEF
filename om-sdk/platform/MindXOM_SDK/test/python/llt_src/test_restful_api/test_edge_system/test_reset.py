# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

from config_urls import EdgeSystemUrls
from test_class import RequestTest


def test_reset():
    """2.5.46 远程恢复出厂设置操作"""

    payload1 = {
        "label": "test_reset: error password",
        "request_obj": {
            "method": "POST",
            "url": EdgeSystemUrls.RESET,
            "params": "",
            "data": {
                "ethernet": "eth0",
                "root_pwd": "password"
            }
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
        "label": "test_reset: error param",
        "request_obj": {
            "method": "POST",
            "url": EdgeSystemUrls.RESET,
            "params": "",
            "data": {
                "ethernet": "ethddd0",
                "root_pwd": "password"
            }
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

    RequestTest(**payload1)
    RequestTest(**payload2)


test_reset()
