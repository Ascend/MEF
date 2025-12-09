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


def test_nfs_post():
    """2.5.27 挂载/解挂NFS分区"""

    payload = {
        "label": 'error is server_dir = ".."',
        "request_obj": {
            "method": "POST",
            "url": EdgeSystemUrls.NFS,
            "params": "",
            "data": {
                "type": "mount",
                "ServerIP": "192.168.2.108",
                "ServerDir": "..",
                "FileSystem": "nfs3",
                "MountPath": "/home/test"
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
        "label": 'error is MountPath = "../home/test"',
        "request_obj": {
            "method": "POST",
            "url": EdgeSystemUrls.NFS,
            "params": "",
            "data": {
                "type": "mount",
                "ServerIP": "192.168.2.108",
                "ServerDir": "/home",
                "FileSystem": "nfs3",
                "MountPath": "../home/test"
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

    RequestTest(**payload)
    RequestTest(**payload2)


test_nfs_post()
