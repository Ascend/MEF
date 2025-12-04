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
import time

from flask.testing import FlaskClient
from pytest_mock import MockerFixture

from common.constants.base_constants import CommonConstants
from common.constants.error_codes import LogErrorCodes
from common.file_utils import FileCreate
from lib_restful_adapter import LibRESTfulAdapter
from system_service import log_services_views
from system_service import partitions_views
from system_service.log_services_views import LoggerCollectUtils
from test_bp_api.create_client import get_client
from test_mqtt_api.get_log_info import GetLogInfo
from ut_utils.models import MockPrivilegeAuth

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from system_service.systems_blueprint import system_bp

getLog = GetLogInfo()

GetSystem = namedtuple("GetSystem", "expect, return_value")
PatchSystem = namedtuple("PatchSystem", "expect, data")
EthIpList = namedtuple("EthIpList", "get_resource, return_value, code")
GetProcessor = namedtuple("GetProcessor", "expect")
PostLogServices = namedtuple("PostLogServices", "expect, data")

# LogServices
GetLogServices = namedtuple("GetLogServices", "get_resource, code")
GetLogCollect = namedtuple("GetLogCollect", "expect, lib_interface, code")
SystemLogDownload = namedtuple("SystemLogDownload", "data, lock, expect, headers, lib_interface, code, create_dir")
LogErrorDict = namedtuple("LogErrorDict", "error_code, input_err_info")

# ExtendedDeviceServices
ExtendedDeviceInfo = namedtuple("ExtendedDeviceInfo", "get_resource, extend_id, return_value, code")

# SimpleStorages
SystemStorageCollection = namedtuple("SystemStorageCollection", "get_resource, return_value, code")
SystemStorageInfo = namedtuple("SystemStorageInfo", "get_resource, storage_id, return_value, code")

# Partition
GetSystemPartition = namedtuple("GetSystemPartition", "get_resource, return_value, code")
CreateSystemPartition = namedtuple("CreateSystemPartition", "data, lock, expect, headers, lib_interface, code")
GetSystemPartitionInfo = namedtuple("GetSystemPartitionInfo", "partition_id, get_resource, return_value, code")
MountSystemPartition = namedtuple("MountSystemPartition", "data, lock, expect, headers, lib_interface, code")
UnmountSystemPartition = namedtuple("UnmountSystemPartition", "data, lock, expect, headers, lib_interface, code")
DeleteSystemPartition = namedtuple("DeleteSystemPartition", "partition_id, get_resource, return_value, headers, code")


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
        "test_eth_ip_list": {
            "success": (
                {
                  "@odata.context": "/redfish/v1/$metadata#Systems/SimpleStorages/$entity",
                  "@odata.id": "/redfish/v1/Systems/SimpleStorages",
                  "@odata.type": "#SimpleStorageCollection.SimpleStorageCollection",
                  "Name": "Simple Storage Collection",
                  "Members@odata.count": None,
                  "Members": [
                    {
                      "@odata.id": "/redfish/v1/Systems/SimpleStorages/entityID"
                    }
                  ]
                },
                {"status": 200, "message": dict()},
                200,
            ),
            "failed": (
                {
                  "@odata.context": "/redfish/v1/$metadata#Systems/SimpleStorages/$entity",
                  "@odata.id": "/redfish/v1/Systems/SimpleStorages",
                  "@odata.type": "#SimpleStorageCollection.SimpleStorageCollection",
                  "Name": "Simple Storage Collection",
                  "Members@odata.count": None,
                  "Members": [
                    {
                      "@odata.id": "/redfish/v1/Systems/SimpleStorages/entityID"
                    }
                  ]
                },
                {"status": 400, "message": dict()},
                400,
            ),
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

    def test_get_system_time(self):
        response = self.client.get("/redfish/v1/Systems/SystemTime")
        assert response.status_code == 200
        expect = {
              "@odata.context": "/redfish/v1/$metadata#Systems/SystemTime",
              "@odata.id": "/redfish/v1/Systems/SystemTime",
              "@odata.type": "#MindXEdgeSystemTime.MindXEdgeSystemTime",
              "Id": "SystemTime",
              "Name": "SystemTime",
              "Datetime": time.ctime()
        }
        assert response.get_json(force=True) == expect

    def test_eth_ip_list(self, mocker: MockerFixture, model: EthIpList):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get(f"/redfish/v1/Systems/EthIpList")
        assert response.status_code == model.code

    def test_get_processor(self, model: GetProcessor):
        response = self.client.get("/redfish/v1/Systems/Processors")
        assert response.status_code == 200
        assert response.get_json(force=True) == model.expect

    def test_post_log_services_download(self, model: PostLogServices):
        response = self.client.post("/redfish/v1/Systems/LogServices/Actions/download", data=json.dumps(model.data))
        assert response.status_code == 400
        assert response.get_json(force=True) == model.expect


class TestLogServicesViews:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_get_system_log_services": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/LogServices/$entity",
                    "@odata.id": "/redfish/v1/Systems/LogServices",
                    "@odata.type": "#LogServiceCollection.LogServiceCollection",
                    "Name": "LogService Collection",
                    "Members@odata.count": None,
                    "Members": [
                        {
                            "@odata.id": "/redfish/v1/Systems/LogServices/entityID"
                        }
                    ],
                    "Oem": {
                        "progress": {
                            "@odata.id": "redfish/v1/Systems/LogServices/progress"
                        },
                        "Actions": {
                            "#download": {
                                "target": "/redfish/v1/Systems/LogServices/Actions/download"
                            }
                        }
                    }
                },
                200,
            ),
            "failed": (
                Exception,
                500,
            )
        },
        "test_system_log_collect_progress": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/LogServices/progress",
                    "@odata.id": "/redfish/v1/Systems/LogServices/progress",
                    "@odata.type": "#Task.v1_6_0.Task",
                    "Description": "Get download logs progress.",
                    "Id": "Log Collection Task",
                    "Name": "Log Collection Task",
                    "PercentComplete": None,
                    "TaskState": None
                },
                [{"status": 200, "message": {"AlarMessages": []}}], 200
            ),
            "failed": (
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Send message failed.",
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
                [{"status": 400, "message": "Send message failed."}], 400
            ),
        },
        "test_system_log_download": {
            "locked": (
                None,
                True,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Collect log failed.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 110001
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400, None,
            ),
            "invalid_json": (
                None,
                False,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON. "
                                               " Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON and "
                                           "could not be parsed by the receiving service.",
                                "NumberOfArgs": 0,
                                "Oem": {
                                    "status": None
                                },
                                "ParamTypes": None,
                                "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                                "Severity": "Critical"
                            }
                        ],
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information."
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400, None,
            ),
            "invalid_param": (
                {"name": "MindXom"},
                False,
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400, None,
            ),
            "cannot_create_dir": (
                {"name": "MindXOM"},
                False,
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Collect log failed",
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400, False,
            ),
            "collect_failed": (
                {"name": "MindXOM"},
                False,
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Collect log failed",
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 400, "message": "log collect failed."}], 400, True,
            ),
        },
        "test_make_log_error_dict": {
            "empty_input_err_info": (
                LogErrorCodes.ERROR_LOG_COLLECT,
                ""
            )
        },
    }

    @staticmethod
    def test_get_system_log_services(mocker: MockerFixture, model: GetLogServices):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        response = TestLogServicesViews.client.get("/redfish/v1/Systems/LogServices")
        assert response.status_code == model.code

    @staticmethod
    def test_system_log_collect_progress(mocker: MockerFixture, model: GetLogCollect):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        response = TestLogServicesViews.client.get("/redfish/v1/Systems/LogServices/progress")
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_system_log_download(mocker: MockerFixture, model: SystemLogDownload):
        mocker.patch.object(log_services_views, "LOG_LOCK").locked.return_value = model.lock
        mocker.patch("os.path.exists")
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        url = "/redfish/v1/Systems/LogServices/Actions/download"
        response = TestLogServicesViews.client.post(url, data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_make_log_error_dict(model: LogErrorDict):
        ret = LoggerCollectUtils.make_log_error_dict(model.error_code, model.input_err_info)

        if not model.input_err_info:
            assert model.error_code.messageKey in getLog.get_log()
        else:
            assert model.input_err_info in getLog.get_log()

        assert ret["status"] == CommonConstants.ERR_CODE_400 and \
               ret["message"] == [model.error_code.code, model.error_code.messageKey]


class TestExtendedDevicesViews:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_get_system_extended_device_info": {
            "invalid_extend_id": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/ExtendedDevices/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/ExtendedDevices/oDataID",
                    "@odata.type": "#MindXEdgeExtendedDevice.v1_0_0.MindXEdgeExtendedDevice",
                    "Id": None,
                    "Name": None,
                    "DeviceClass": None,
                    "DeviceName": None,
                    "Manufacturer": None,
                    "Model": None,
                    "SerialNumber": None,
                    "Location": None,
                    "FirmwareVersion": None,
                    "Status": {
                        "State": None,
                        "Health": None
                    }
                },
                "a", None, 400
            ),
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/ExtendedDevices/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/ExtendedDevices/oDataID",
                    "@odata.type": "#MindXEdgeExtendedDevice.v1_0_0.MindXEdgeExtendedDevice",
                    "Id": None,
                    "Name": None,
                    "DeviceClass": None,
                    "DeviceName": None,
                    "Manufacturer": None,
                    "Model": None,
                    "SerialNumber": None,
                    "Location": None,
                    "FirmwareVersion": None,
                    "Status": {
                        "State": None,
                        "Health": None
                    }
                },
                "abc",
                {"status": 200, "message": dict()},
                200,
            ),
            "failed": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/ExtendedDevices/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/ExtendedDevices/oDataID",
                    "@odata.type": "#MindXEdgeExtendedDevice.v1_0_0.MindXEdgeExtendedDevice",
                    "Id": None,
                    "Name": None,
                    "DeviceClass": None,
                    "DeviceName": None,
                    "Manufacturer": None,
                    "Model": None,
                    "SerialNumber": None,
                    "Location": None,
                    "FirmwareVersion": None,
                    "Status": {
                        "State": None,
                        "Health": None
                    }
                },
                "abc",
                {"status": 400, "message": dict()},
                400,
            ),
        },
    }

    @staticmethod
    def test_get_system_extended_device_info(mocker: MockerFixture, model: ExtendedDeviceInfo):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = TestExtendedDevicesViews.client.get(f"/redfish/v1/Systems/ExtendedDevices/{model.extend_id}")
        assert response.status_code == model.code


class TestSimpleStoragesViews:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_get_system_simple_storages_collection": {
            "success": (
                {
                  "@odata.context": "/redfish/v1/$metadata#Systems/SimpleStorages/$entity",
                  "@odata.id": "/redfish/v1/Systems/SimpleStorages",
                  "@odata.type": "#SimpleStorageCollection.SimpleStorageCollection",
                  "Name": "Simple Storage Collection",
                  "Members@odata.count": None,
                  "Members": [
                    {
                      "@odata.id": "/redfish/v1/Systems/SimpleStorages/entityID"
                    }
                  ]
                },
                {"status": 200, "message": dict()},
                200,
            ),
            "failed": (
                {
                  "@odata.context": "/redfish/v1/$metadata#Systems/SimpleStorages/$entity",
                  "@odata.id": "/redfish/v1/Systems/SimpleStorages",
                  "@odata.type": "#SimpleStorageCollection.SimpleStorageCollection",
                  "Name": "Simple Storage Collection",
                  "Members@odata.count": None,
                  "Members": [
                    {
                      "@odata.id": "/redfish/v1/Systems/SimpleStorages/entityID"
                    }
                  ]
                },
                {"status": 400, "message": dict()},
                400,
            ),
        },
        "test_get_system_storage_info": {
            "invalid_storage_id": (
                {
                  "@odata.context": "/redfish/v1/$metadata#Systems/SimpleStorages/$entity",
                  "@odata.id": "/redfish/v1/Systems/SimpleStorages",
                  "@odata.type": "#SimpleStorageCollection.SimpleStorageCollection",
                  "Name": "Simple Storage Collection",
                  "Members@odata.count": None,
                  "Members": [
                    {
                      "@odata.id": "/redfish/v1/Systems/SimpleStorages/entityID"
                    }
                  ]
                },
                "a", None, 400
            ),
            "success": (
                {
                  "@odata.context": "/redfish/v1/$metadata#Systems/SimpleStorages/$entity",
                  "@odata.id": "/redfish/v1/Systems/SimpleStorages",
                  "@odata.type": "#SimpleStorageCollection.SimpleStorageCollection",
                  "Name": "Simple Storage Collection",
                  "Members@odata.count": None,
                  "Members": [
                    {
                      "@odata.id": "/redfish/v1/Systems/SimpleStorages/entityID"
                    }
                  ]
                },
                "1",
                {"status": 200, "message": dict()},
                200,
            ),
            "failed": (
                {
                  "@odata.context": "/redfish/v1/$metadata#Systems/SimpleStorages/$entity",
                  "@odata.id": "/redfish/v1/Systems/SimpleStorages",
                  "@odata.type": "#SimpleStorageCollection.SimpleStorageCollection",
                  "Name": "Simple Storage Collection",
                  "Members@odata.count": None,
                  "Members": [
                    {
                      "@odata.id": "/redfish/v1/Systems/SimpleStorages/entityID"
                    }
                  ]
                },
                "1",
                {"status": 400, "message": dict()},
                400,
            ),
        },
    }

    @staticmethod
    def test_get_system_simple_storages_collection(mocker: MockerFixture, model: SystemStorageCollection):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = TestSimpleStoragesViews.client.get(f"/redfish/v1/Systems/SimpleStorages")
        assert response.status_code == model.code

    @staticmethod
    def test_get_system_storage_info(mocker: MockerFixture, model: SystemStorageInfo):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = TestSimpleStoragesViews.client.get(f"/redfish/v1/Systems/SimpleStorages/{model.storage_id}")
        assert response.status_code == model.code


class TestPartitionViews:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_get_system_partitions": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Partitions/$entity",
                    "@odata.id": "/redfish/v1/Systems/Partitions",
                    "@odata.type": "#MindXEdgePartitionCollection.MindXEdgePartitionCollection",
                    "Name": "Partition Collection",
                    "Members@odata.count": None,
                    "Members": [
                        {
                            "@odata.id": "/redfish/v1/Systems/Partitions/entityID"
                        }
                    ],
                    "Mount": {
                        "@odata.id": "/redfish/v1/Systems/Partitions/Mount"
                    },
                    "Unmount": {
                        "@odata.id": "/redfish/v1/Systems/Partitions/Unmount"
                    }
                },
                {"status": 200, "message": dict()},
                200,
            ),
            "failed": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Partitions/$entity",
                    "@odata.id": "/redfish/v1/Systems/Partitions",
                    "@odata.type": "#MindXEdgePartitionCollection.MindXEdgePartitionCollection",
                    "Name": "Partition Collection",
                    "Members@odata.count": None,
                    "Members": [
                        {
                            "@odata.id": "/redfish/v1/Systems/Partitions/entityID"
                        }
                    ],
                    "Mount": {
                        "@odata.id": "/redfish/v1/Systems/Partitions/Mount"
                    },
                    "Unmount": {
                        "@odata.id": "/redfish/v1/Systems/Partitions/Unmount"
                    }
                },
                {"status": 400, "message": dict()},
                400,
            ),
        },
        "test_create_system_partitions": {
            "locked": (
                None,
                True,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Create system partition failed because PartitionView modify is busy.",
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400
            ),
            "invalid_json": (
                None,
                False,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON. "
                                               " Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON and "
                                           "could not be parsed by the receiving service.",
                                "NumberOfArgs": 0,
                                "Oem": {
                                    "status": None
                                },
                                "ParamTypes": None,
                                "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                                "Severity": "Critical"
                            }
                        ],
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information."
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400,
            ),
            "invalid_param": (
                {
                    "Number": 100,
                    "CapacityBytes": "0.5",
                    "Links": [{
                        "Device": {
                            "@odata.id": "/dev/mdisk0"
                        }
                    }],
                    "FileSystem": "ext4",
                },
                False,
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400,
            ),
            "failed": (
                {
                    "Number": 1,
                    "CapacityBytes": "0.5",
                    "Links": [{
                        "Device": {
                            "@odata.id": "/dev/mdisk0"
                        }
                    }],
                    "FileSystem": "ext4",
                },
                False,
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "create_system_partitions failed.",
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 400, "message": "create_system_partitions failed."}], 400,
            )
        },
        "test_get_system_partition_info": {
            "invalid_partition_id": (
                "!@#",
                None,
                None,
                400,
            ),
            "success": (
                "mdisk0p100",
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Partitions/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/Partitions/oDataID",
                    "@odata.type": "#MindXEdgePartition.v1_0_0.MindXEdgePartition",
                    "Id": None,
                    "Name": None,
                    "CapacityBytes": None,
                    "FreeBytes": None,
                    "Links": [
                        {
                            "Device": {
                                "@odata.id": None
                            },
                            "DeviceName": None,
                            "Location": None
                        }
                    ],
                    "MountPath": None,
                    "Primary": False,
                    "FileSystem": None,
                    "Status": {
                        "State": None,
                        "Health": None
                    }
                },
                {"status": 200, "message": dict()},
                200,
            ),
            "failed": (
                "mdisk0p100",
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Partitions/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/Partitions/oDataID",
                    "@odata.type": "#MindXEdgePartition.v1_0_0.MindXEdgePartition",
                    "Id": None,
                    "Name": None,
                    "CapacityBytes": None,
                    "FreeBytes": None,
                    "Links": [
                        {
                            "Device": {
                                "@odata.id": None
                            },
                            "DeviceName": None,
                            "Location": None
                        }
                    ],
                    "MountPath": None,
                    "Primary": False,
                    "FileSystem": None,
                    "Status": {
                        "State": None,
                        "Health": None
                    }
                },
                {"status": 400, "message": dict()},
                400,
            ),
        },
        "test_mount_system_partitions": {
            "locked": (
                None,
                True,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Mount partition failed because Partition View mount is busy.",
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400
            ),
            "invalid_json": (
                None,
                False,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON. "
                                               " Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON and "
                                           "could not be parsed by the receiving service.",
                                "NumberOfArgs": 0,
                                "Oem": {
                                    "status": None
                                },
                                "ParamTypes": None,
                                "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                                "Severity": "Critical"
                            }
                        ],
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information."
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400,
            ),
            "invalid_param": (
                {
                    "MountPath": "/opt/mount",
                },
                False,
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400,
            ),
            "failed": (
                {
                    "MountPath": "/opt/mount",
                    "PartitionID": "mdisk0p11",
                },
                False,
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "create_system_partitions failed.",
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 400, "message": "create_system_partitions failed."}], 400,
            )
        },
        "test_unmount_system_partitions": {
            "locked": (
                None,
                True,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Unmount partition failed because Partition View unmount is busy.",
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400
            ),
            "invalid_json": (
                None,
                False,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON. "
                                               " Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON and "
                                           "could not be parsed by the receiving service.",
                                "NumberOfArgs": 0,
                                "Oem": {
                                    "status": None
                                },
                                "ParamTypes": None,
                                "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                                "Severity": "Critical"
                            }
                        ],
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information."
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400,
            ),
            "invalid_param": (
                {
                    "PartitionID": "",
                },
                False,
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None, 400,
            ),
            "failed": (
                {
                    "PartitionID": "mdisk0p11",
                },
                False,
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "create_system_partitions failed.",
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
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 400, "message": "create_system_partitions failed."}], 400,
            )
        },
        "test_delete_system_partitions": {
            "invalid_partition_id": (
                "!@#",
                None,
                None,
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                400,
            ),
            "success": (
                "mdisk0p100",
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Partitions/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/Partitions/oDataID",
                    "@odata.type": "#MindXEdgePartition.v1_0_0.MindXEdgePartition",
                    "Id": None,
                    "Name": None,
                    "CapacityBytes": None,
                    "FreeBytes": None,
                    "Links": [
                        {
                            "Device": {
                                "@odata.id": None
                            },
                            "DeviceName": None,
                            "Location": None
                        }
                    ],
                    "MountPath": None,
                    "Primary": False,
                    "FileSystem": None,
                    "Status": {
                        "State": None,
                        "Health": None
                    }
                },
                {"status": 200, "message": dict()},
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                202,
            ),
            "failed": (
                "mdisk0p100",
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Partitions/Members/$entity",
                    "@odata.id": "/redfish/v1/Systems/Partitions/oDataID",
                    "@odata.type": "#MindXEdgePartition.v1_0_0.MindXEdgePartition",
                    "Id": None,
                    "Name": None,
                    "CapacityBytes": None,
                    "FreeBytes": None,
                    "Links": [
                        {
                            "Device": {
                                "@odata.id": None
                            },
                            "DeviceName": None,
                            "Location": None
                        }
                    ],
                    "MountPath": None,
                    "Primary": False,
                    "FileSystem": None,
                    "Status": {
                        "State": None,
                        "Health": None
                    }
                },
                {"status": 400, "message": dict()},
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                400,
            ),
        },
    }

    def test_get_system_partitions(self, mocker: MockerFixture, model: GetSystemPartition):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get(f"/redfish/v1/Systems/Partitions")
        assert response.status_code == model.code

    def test_create_system_partitions(self, mocker: MockerFixture, model: CreateSystemPartition):
        mocker.patch.object(partitions_views, "CREATE_PARTITION_LOCK").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        url = "/redfish/v1/Systems/Partitions"
        response = self.client.post(url, data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    def test_get_system_partition_info(self, mocker: MockerFixture, model: GetSystemPartitionInfo):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.get(f"/redfish/v1/Systems/Partitions/{model.partition_id}")
        assert response.status_code == model.code

    def test_mount_system_partitions(self, mocker: MockerFixture, model: MountSystemPartition):
        mocker.patch.object(partitions_views, "MOUNT_PARTITION_LOCK").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        url = "/redfish/v1/Systems/Partitions/Mount"
        response = self.client.patch(url, data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    def test_unmount_system_partitions(self, mocker: MockerFixture, model: UnmountSystemPartition):
        mocker.patch.object(partitions_views, "UNMOUNT_PARTITION_LOCK").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        url = "/redfish/v1/Systems/Partitions/Unmount"
        response = self.client.patch(url, data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    def test_delete_system_partitions(self, mocker: MockerFixture, model: DeleteSystemPartition):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.return_value)
        response = self.client.delete(f"/redfish/v1/Systems/Partitions/{model.partition_id}", headers=model.headers)
        assert response.status_code == model.code
