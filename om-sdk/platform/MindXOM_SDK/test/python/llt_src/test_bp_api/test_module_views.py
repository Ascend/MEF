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
from collections import namedtuple
from unittest.mock import patch
from pytest_mock import MockerFixture
from flask.testing import FlaskClient

from lib_restful_adapter import LibRESTfulAdapter
from ut_utils.models import MockPrivilegeAuth
from test_bp_api.create_client import get_client
from system_service.module_views import module_replace_odata_id

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from system_service.systems_blueprint import system_bp

GetSystemModuleCollection = namedtuple("GetSystemModuleCollection", "expect, return_value, code")
GetSystemModuleInfo = namedtuple("GetSystemModuleInfo", "expect, return_value, module_id")
GetSystemDeviceInfo = namedtuple("GetSystemDeviceInfo", "expect, return_value, module_id, device_id")
PatchSystemDeviceInfo = namedtuple("PatchSystemDeviceInfo", "headers, expect, return_value, data, module_id, device_id")


class TestModuleViews:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_get_system_module_collection": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Modules/$entity",
                    "@odata.id": "/redfish/v1/Systems/Modules",
                    "@odata.type": "#MindXEdgeModuleCollection.MindXEdgeModuleCollection",
                    "Name": "Device Module Collection",
                    "Members@odata.count": 6,
                    "Members": [
                        {
                            "@odata.id": "/redfish/v1/Systems/Modules/npu"
                        },
                        {
                            "@odata.id": "/redfish/v1/Systems/Modules/eth"
                        },
                        {
                            "@odata.id": "/redfish/v1/Systems/Modules/wifi"
                        },
                        {
                            "@odata.id": "/redfish/v1/Systems/Modules/disk"
                        },
                        {
                            "@odata.id": "/redfish/v1/Systems/Modules/lte"
                        },
                        {
                            "@odata.id": "/redfish/v1/Systems/Modules/A200"
                        }
                    ]
                }, {"status": 200, "message": ["npu", "eth", "wifi", "disk", "lte", "A200"]}, 200),
            "failed-invalid-module_id": (
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "fail",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": None
                                },
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Severity": "Critical"
                            }
                        ],
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information."
                    }
                },
                {"status": 400, "message": "fail"}, 400
            ),
            "exception": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Internal server error",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100011
                                }
                            }
                        ]
                    }
                },
                Exception, 500
            ),
        },
        "test_get_system_module_info": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Modules/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/Modules/npu",
                    "@odata.type": "#MindXEdgeModuleInfoCollection.MindXEdgeModuleInfoCollection",
                    "Id": "npu",
                    "Name": "Device Module Info",
                    "Members@odata.count": 2,
                    "Members": [
                        {
                            "@odata.id": "/redfish/v1/Systems/Modules/npu/npu1"
                        },
                        {
                            "@odata.id": "/redfish/v1/Systems/Modules/npu/npu2"
                        },
                    ],
                    "ModuleInfo": {
                        "temperature": {
                            "description": "npu temperature",
                            "type": "int",
                            "id": 65537,
                            "accessMode": "Read"
                        },
                        "health": {
                            "description": "npu health",
                            "type": "int",
                            "id": 65538,
                            "accessMode": "Read"
                        },
                        "memory": {
                            "description": "npu memory",
                            "type": "long long",
                            "id": 65539,
                            "accessMode": "Read"
                        }
                    }
                },
                {
                    "status": 200,
                    "message": {
                        "devices": [
                            "npu1",
                            "npu2"
                        ],
                        "ModuleInfo": {
                            "temperature": {
                                "description": "npu temperature",
                                "type": "int",
                                "id": 65537,
                                "accessMode": "Read"
                            },
                            "health": {
                                "description": "npu health",
                                "type": "int",
                                "id": 65538,
                                "accessMode": "Read"
                            },
                            "memory": {
                                "description": "npu memory",
                                "type": "long long",
                                "id": 65539,
                                "accessMode": "Read"
                            }
                        }
                    }
                },
                "npu"
            ),
            "failed-invalid-module_id": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "tv is fail",
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
                },
                {"status": 404, "message": "tv is fail"}, "npu"
            ),
            "invalid_param": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100024
                                }
                            }
                        ]
                    }
                },
                None, "npu@#"
            ),
            "exception": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Internal server error",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100011
                                }
                            }
                        ]
                    }
                },
                Exception, "npu",
            ),
        },
        "test_get_system_device_info": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Modules/Members/npu/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/Modules/npu/npu1",
                    "@odata.type": "#MindXEdgeDeviceInfo.MindXEdgeDeviceInfo",
                    "Id": "npu1",
                    "Name": "Device Info",
                    "Attributes": None
                },
                {"status": 200, "message": {"Attributes": None}},
                "npu", "npu1",
            ),
            "failed-invalid-device_id": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "npu404 is fail",
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
                },
                {"status": 404, "message": "npu404 is fail"},
                "npu", "npu1",
            ),
            "invalid_param": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100024
                                }
                            }
                        ]
                    }
                },
                None, "npu+++", "npu+"
            ),
            "exception": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Internal server error",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100011
                                }
                            }
                        ]
                    }
                },
                Exception, "npu", "npu1"
            ),
        },
        "test_patch_system_device_info": {
            "success": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Modules/Members/npu/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/Modules/npu/npu1",
                    "@odata.type": "#MindXEdgeDeviceInfo.MindXEdgeDeviceInfo",
                    "Id": "npu1",
                    "Name": "Device Info",
                    "Attributes": None
                },
                {"status": 200, "message": {"Attributes": None}},
                {"Attributes": dict()}, "npu", "npu1",
            ),
            "failed-invalid-parameter": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100024
                                }
                            }
                        ]
                    }
                },
                {"status": 404, "message": "invalid parameter"},
                {"Attributes": None}, "npu+++", "npu"
            ),
            "invalid_json": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "error": {
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON.  "
                                               "Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON and could not be "
                                           "parsed by the receiving service.",
                                "Severity": "Critical",
                                "NumberOfArgs": 0,
                                "ParamTypes": None,
                                "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                                "Oem": {
                                    "status": None
                                }
                            }
                        ]
                    }
                },
                {"status": 404, "message": "invalid parameter"},
                "Attributes", "npu", "npu1"
            ),
            "invalid_param_type": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100024
                                }
                            }
                        ]
                    }
                },
                {"status": 404, "message": "invalid parameter"},
                {"Attributes": ""}, "npu", "npu1"
            ),
        }
    }

    @staticmethod
    def test_module_replace_odata_id():
        resp_json = {
            "@odata.context": "/redfish/v1/$metadata#Systems/Modules/Members/oDataID1/Members/$entity",
            "@odata.id": "/redfish/v1/Systems/Modules/oDataID1/oDataID2",
            "@odata.type": "#MindXEdgeDeviceInfo.MindXEdgeDeviceInfo",
            "Id": None,
            "Name": "Device Info",
            "Attributes": None
        }
        module_id = "module_a"
        device_id = "device_1"
        module_replace_odata_id(resp_json, module_id, device_id)
        assert resp_json["Id"] == device_id and \
               resp_json["@odata.id"] == f"/redfish/v1/Systems/Modules/{module_id}/{device_id}" and \
               resp_json["@odata.context"] == f"/redfish/v1/$metadata#Systems/Modules/Members/{module_id}" \
                                              f"/Members/$entity"

    def test_get_system_module_collection(self, mocker: MockerFixture, model: GetSystemModuleCollection):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get("/redfish/v1/Systems/Modules")
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    def test_get_system_module_info(self, mocker: MockerFixture, model: GetSystemModuleInfo):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get(f"/redfish/v1/Systems/Modules/{model.module_id}")
        assert response.get_json(force=True) == model.expect

    def test_get_system_device_info(self, mocker: MockerFixture, model: GetSystemDeviceInfo):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get(f"/redfish/v1/Systems/Modules/{model.module_id}/{model.device_id}")
        assert response.get_json(force=True) == model.expect

    def test_patch_system_device_info(self, mocker: MockerFixture, model: PatchSystemDeviceInfo):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        url = f"/redfish/v1/Systems/Modules/{model.module_id}/{model.device_id}"
        response = self.client.patch(url, data=json.dumps(model.data), headers=model.headers)
        assert response.get_json(force=True) == model.expect
