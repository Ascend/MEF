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

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from system_service.systems_blueprint import system_bp

GetSystemModuleCollection = namedtuple("GetSystemModuleCollection", "expect, return_value")
GetSystemModuleInfo = namedtuple("GetSystemModuleInfo", "expect, return_value")
GetSystemDeviceInfo = namedtuple("GetSystemDeviceInfo", "expect, return_value")
PatchSystemDeviceInfo = namedtuple("PatchSystemDeviceInfo", "expect, return_value, data")


class TestModuleViews:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_rf_get_system_module_collection": {
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
                }, {"status": 200, "message": ["npu", "eth", "wifi", "disk", "lte", "A200"]})},
        "test_rf_get_system_module_info": {
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
                {"status": 200, "message": {"devices": ["npu1", "npu2"],
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
                                               }}
            ),
            "failed-invalid-module_id": ({
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
                }}, {"status": 404, "message": "tv is fail"}), },
        "test_rf_get_system_device_info": {
            "success": ({
                            "@odata.context": "/redfish/v1/$metadata#Systems/Modules/Members/npu/Members/$entity",
                            "@odata.id": "/redfish/v1/Systems/Modules/npu/npu1",
                            "@odata.type": "#MindXEdgeDeviceInfo.MindXEdgeDeviceInfo",
                            "Id": "npu1",
                            "Name": "Device Info",
                            "Attributes": None
                        }, {"status": 200, "message": {"Attributes": None}}),
            "failed-invalid-device_id": ({
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
                }}, {"status": 404, "message": "npu404 is fail"})},
        "test_rf_patch_system_device_info": {
            "success": ({
                            "@odata.context": "/redfish/v1/$metadata#Systems/Modules/Members/npu/Members/$entity",
                            "@odata.id": "/redfish/v1/Systems/Modules/npu/npu1",
                            "@odata.type": "#MindXEdgeDeviceInfo.MindXEdgeDeviceInfo",
                            "Id": "npu1",
                            "Name": "Device Info",
                            "Attributes": None
            }, {"status": 200, "message": {"Attributes": None}},
             {"Attributes": None}),
            "failed-invalid-device_id": ({
                "error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "npu402 is fail",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": None
                            }
                        }
                    ]
                }}, {"status": 404, "message": "npu402 is fail"},
                    {"Attributes": None}),
            "failed-invalid-parameter": ({
                "error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "invalid parameter",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": None
                            }
                        }
                    ]
                }}, {"status": 404, "message": "invalid parameter"},
                {"Attributes": None})
        }
    }

    def test_rf_get_system_module_collection(self, mocker: MockerFixture, model: GetSystemModuleCollection):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get("/redfish/v1/Systems/Modules")
        assert response.status_code == 200
        assert response.get_json(force=True) == model.expect

    def test_rf_get_system_module_info(self, mocker: MockerFixture, model: GetSystemModuleInfo):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get("/redfish/v1/Systems/Modules/npu")
        assert response.get_json(force=True) == model.expect

    def test_rf_get_system_device_info(self, mocker: MockerFixture, model: GetSystemDeviceInfo):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get("/redfish/v1/Systems/Modules/npu/npu1")
        assert response.get_json(force=True) == model.expect

    def rf_patch_system_device_info(self, mocker: MockerFixture, model: PatchSystemDeviceInfo):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get("/redfish/v1/Systems/Modules/npu/npu1", data=json.dumps(model.data))
        assert response.get_json(force=True) == model.expect

