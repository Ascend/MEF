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


def test_ethernet_get():
    """2.5.17 查询以太网接口资源信息"""

    payload1 = {
        "label": 'device_id is ""',
        "request_obj": {
            "method": "GET",
            "url": EdgeSystemUrls.ETHERNET_INTERFACES,
            "params": "",
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
        "label": 'device_id is "32"',
        "request_obj": {
            "method": "GET",
            "url": EdgeSystemUrls.ETHERNET_INTERFACES,
            "params": "32",
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


def test_ethernet_patch():
    """2.5.18 修改以太网接口资源信息"""

    payload1 = {
        "label": 'device_id is ""',
        "request_obj": {
            "method": "PATCH",
            "url": EdgeSystemUrls.ETHERNET_INTERFACES,
            "params": "6788",
            "data": {
                "IPv4Addresses": [{
                    "Address": "10.10.10.5",
                    "SubnetMask": "255.255.0.0",
                    "AddressOrigin": "Static",
                    "Gateway": "",
                    "VlanId": None,
                    "Tag": "net"
                }, {
                    "Address": "10.10.10.8",
                    "SubnetMask": "255.255.0.0",
                    "Gateway": "10.10.10.1",
                    "VlanId": None,
                    "Tag": "test",
                    "ConnectTest": True,
                    "RemoteTestIp": "10.10.20.10",
                    "AddressOrigin": "Static"
                }]
            },
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

    payload2 = {
        "label": 'device_id is ""',
        "request_obj": {
            "method": "PATCH",
            "url": EdgeSystemUrls.ETHERNET_INTERFACES,
            "params": "GMAC0",
            "data": {
                "IPv4Addresses": [{
                    "Address": "10.10.10.5",
                    "SubnetMask": "255.255.0.0",
                    "AddressOrigin": "Static",
                    "Gateway": "",
                    "VlanId": 11,
                    "Tag": "net"
                }, {
                    "Address": "10.10.10.268",
                    "SubnetMask": "255.255.0.0",
                    "Gateway": "10.10.10.1",
                    "VlanId": None,
                    "Tag": "test",
                    "ConnectTest": True,
                    "RemoteTestIp": "10.10.20.10",
                    "AddressOrigin": "Static"
                }]
            },
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


test_ethernet_get()
test_ethernet_patch()
