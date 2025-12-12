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

from flask.testing import FlaskClient
from pytest_mock import MockerFixture

from lib_restful_adapter import LibRESTfulAdapter
from test_bp_api.create_client import get_client
from ut_utils.models import MockPrivilegeAuth

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from system_service.systems_blueprint import system_bp

GetSystem = namedtuple("GetSystem", "expect, return_value")
PatchSystem = namedtuple("PatchSystem", "expect, data")
GetProcessor = namedtuple("GetProcessor", "expect")
PostLogServices = namedtuple("PostLogServices", "expect, data")


class TestSystem:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_get_system": {
            "success": ({"@odata.context": "/redfish/v1/$metadata#Systems",
                         "@odata.id": "/redfish/v1/Systems",
                         "@odata.type": "#ComputerSystem.v1_18_0.ComputerSystem",
                         "Id": "1", "Name": "Computer System", "HostName": "Atlas200", "UUID": "",
                         "Model": "", "SerialNumber": "", "AssetTag": "", "SupportModel": "",
                         "Status": {
                             "Health": None
                         },
                         "Processors": {
                             "@odata.id": "/redfish/v1/Systems/Processors"
                         },
                         "Memory": {
                             "@odata.id": "/redfish/v1/Systems/Memory"
                         },
                         "EthernetInterfaces": {
                             "@odata.id": "/redfish/v1/Systems/EthernetInterfaces"
                         },
                         "LogServices": {
                             "@odata.id": "/redfish/v1/Systems/LogServices"
                         },
                         "SimpleStorages": {
                             "@odata.id": "/redfish/v1/Systems/SimpleStorages"
                         },
                         "Oem": {
                             "PCBVersion": "Ver.C", "Temperature": None, "Power": None,
                             "Voltage": None, "CpuHeating": None, "DiskHeating": None,
                             "AiTemperature": None, "UsbHubHeating": None,
                             "KernelVersion": None, "Uptime": None,
                             "Datetime": None, "DateTimeLocalOffset": None,
                             "CpuUsage": None, "MemoryUsage": None, "ProcessorArchitecture": None,
                             "OSVersion": None,
                             "Firmware": [{
                                 "BoardId": None,
                                 "InactiveVersion": None,
                                 "Module": None,
                                 "UpgradeResult": None,
                                 "Version": None,
                                 "UpgradeProcess": None,
                             }],
                             "InactiveConfiguration": None,
                             "NTPService": {
                                 "@odata.id": "/redfish/v1/Systems/NTPService"
                             },
                             "ExtendedDevices": {
                                 "@odata.id": "/redfish/v1/Systems/ExtendedDevices"
                             },
                             "LTE": {
                                 "@odata.id": "/redfish/v1/Systems/LTE"
                             },
                             "Partitions": {
                                 "@odata.id": "/redfish/v1/Systems/Partitions"
                             },
                             "NfsManage": {
                                 "@odata.id": "/redfish/v1/Systems/NfsManage"
                             },
                             "SecurityService": {
                                 "@odata.id": "/redfish/v1/Systems/SecurityService"
                             },
                             "Alarm": {
                                 "@odata.id": "/redfish/v1/Systems/Alarm"
                             },
                             "SystemTime": {
                                 "@odata.id": "/redfish/v1/Systems/SystemTime"
                             },
                             "EthIpList": {
                                 "@odata.id": "/redfish/v1/Systems/EthIpList"
                             },
                             "Modules": {
                                 "@odata.id": "/redfish/v1/Systems/Modules"
                             }
                         },
                         "Actions": {
                             "#ComputerSystem.Reset": {
                                 "target": "/redfish/v1/Systems/Actions/ComputerSystem.Reset"
                             },
                             "Oem": {
                                 "#RestoreDefaults.Reset": {
                                     "target": "/redfish/v1/Systems/Actions/RestoreDefaults.Reset"
                                 }
                             }
                         }}, {"status": 200, "message": {"test": "test"}})
        },
        "test_patch_system": {
            "failed-Datetime-invalid": (
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": 100024}}]}}, {"DateTime": "Wed"}),
            "failed-DateTimeLocalOffset-invalid": (
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": 100024}}]}},
                {"DateTimeLocalOffset": "UTC..(UTC, +0000)"}),
            "failed-AssetTag-invalid": (
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": 100024}}]}},
                {"AssetTag": "111111111111111111111111111111111111111111111\
                                                 111111111111111111111111111111111111111111111\
                                                 111111111111111111111111111111111111111111111\
                                                 111111111111111111111111111111111111111111111\
                                                 111111111111111111111111111111111111111111111\
                                                 111111111111111111111111111111111"}),
            "failed-HostName-invalid": (
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": 100024}}]}},
                {"HostName": "-Euler00"})
        },
        "test_get_processor": {
            "success": ({
                            "@odata.context": "/redfish/v1/$metadata#Systems/Processors/#entity",
                            "@odata.id": "/redfish/v1/Systems/Processors",
                            "@odata.type": "#ProcessorCollection.ProcessorCollection",
                            "Name": "Processors Collection",
                            "Members@odata.count": 2,
                            "Members": [
                                {
                                    "@odata.id": "/redfish/v1/Systems/Processors/CPU"
                                },
                                {
                                    "@odata.id": "/redfish/v1/Systems/Processors/AiProcessor"
                                }
                            ]
                        },)
        },
        "test_post_log_services_download": {
            "failed-log-service-download-name-invalid": ({
                         "error": {"code": "Base.1.0.GeneralError",
                                   "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                                   "@Message.ExtendedInfo": [
                                       {
                                           "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                           "Description": "Indicates that a general error has occurred.",
                                           "Message": "Parameter is invalid.",
                                           "Severity": "Critical",
                                           "NumberOfArgs": None, "ParamTypes": None,
                                           "Resolution": "None",
                                           "Oem": {"status": 100024}}]}}, {"name": "XXX"},),
            "failed-log-service-download-name-none": ({
                    "error": {"code": "Base.1.0.GeneralError",
                              "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                              "@Message.ExtendedInfo": [
                                  {
                                      "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                      "Description": "Indicates that a general error has occurred.",
                                      "Message": "Parameter is invalid.",
                                      "Severity": "Critical",
                                      "NumberOfArgs": None, "ParamTypes": None,
                                      "Resolution": "None",
                                      "Oem": {"status": 100024}}]}}, {"name": ""},)
        },

    }

    def test_get_system(self, mocker: MockerFixture, model: GetSystem):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get("/redfish/v1/Systems")
        assert response.status_code == 200
        assert response.get_json(force=True) == model.expect

    def test_patch_system(self, model: PatchSystem):
        response = self.client.patch("/redfish/v1/Systems", data=json.dumps(model.data))
        assert response.status_code == 400
        assert response.get_json(force=True) == model.expect

    def test_get_processor(self, model: GetProcessor):
        response = self.client.get("/redfish/v1/Systems/Processors")
        assert response.status_code == 200
        assert response.get_json(force=True) == model.expect

    def test_post_log_services_download(self, model: PostLogServices):
        response = self.client.post("/redfish/v1/Systems/LogServices/Actions/download", data=json.dumps(model.data))
        assert response.status_code == 400
        assert response.get_json(force=True) == model.expect
