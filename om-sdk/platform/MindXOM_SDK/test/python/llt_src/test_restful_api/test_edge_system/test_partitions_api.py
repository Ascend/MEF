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


def test_partition_get():
    """2.5.22 查询磁盘分区资源信息"""

    payload1 = {
        "label": 'partition_id is "?sda1"',
        "request_obj": {
            "method": "GET",
            "url": EdgeSystemUrls.PARTITIONS,
            "params": "?sda1",
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
        "label": 'partition_id is "&?partition_id=sda1"',
        "request_obj": {
            "method": "GET",
            "url": EdgeSystemUrls.PARTITIONS,
            "params": "&?partition_id=sda1",
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

    RequestTest(**payload1)
    RequestTest(**payload2)


def test_partition_post():
    """2.5.23 创建磁盘分区"""

    payload1 = {
        "label": 'post data is error: number',
        "request_obj": {
            "method": "POST",
            "url": EdgeSystemUrls.PARTITIONS,
            "params": "",
            "data": {
                "Number": 2,
                "CapacityBytes": 1,
                "Links": [{
                    "Device": {
                        "@odata.id": "/dev/sda"
                    }
                }],
                "FileSystem": "ext4"
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
        "label": 'post data is error: @odata.id',
        "request_obj": {
            "method": "POST",
            "url": EdgeSystemUrls.PARTITIONS,
            "params": "",
            "data": {
                "Number": 2,
                "CapacityBytes": 1,
                "Links": [{
                    "Device": {
                        "@odata.id": "/sda"
                    }
                }],
                "FileSystem": "ext4"
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

    RequestTest(**payload1)
    RequestTest(**payload2)


def test_partition_delete():
    """2.5.24 删除磁盘分区"""

    payload1 = {
        "label": 'error param is: partition_id = "sda.."',
        "request_obj": {
            "method": "DELETE",
            "url": EdgeSystemUrls.PARTITIONS,
            "params": "sda..",
            "data": ""
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


def test_partition_patch():
    """2.5.25 挂载/解挂磁盘分区"""

    payload2 = {
        "label": 'error param is: body MountPath = "/home/software" ',
        "request_obj": {
            "method": "PATCH",
            "url": EdgeSystemUrls.PARTITIONS,
            "params": "",
            "data": {
                "MountPath": "/home/software"
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

    RequestTest(**payload2)


test_partition_get()
test_partition_post()
test_partition_delete()
test_partition_patch()
