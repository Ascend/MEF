import json

from config_urls import EdgeSystemUrls
from my_test import MyTest


def test_lte_patch():
    """2.5.15 配置LTE接口资源信息"""

    payload = {
        "label": 'patch data auth_type is "6"',
        "request_obj": {
            "method": "PATCH",
            "url": EdgeSystemUrls.LTE,
            "params": "",
            "data": {
                "apn_name": "lte",
                "apn_user": "admin",
                "apn_passwd": "MyStrongPassword",
                "auth_type": "6",
            },
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
        "label": 'patch data auth_type is "0"',
        "request_obj": {
            "method": "PATCH",
            "url": EdgeSystemUrls.LTE,
            "params": "",
            "data": {
                "apn_name": "lte",
                "apn_user": "admin",
                "apn_passwd": "MyStrongPassword",
                "auth_type": "0",
            },
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

    MyTest(**payload)
    MyTest(**payload2)


test_lte_patch()
