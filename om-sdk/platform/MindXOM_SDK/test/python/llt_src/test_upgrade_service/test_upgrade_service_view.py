# coding: utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
import json
import shutil
from collections import namedtuple
from unittest.mock import patch

from flask.testing import FlaskClient
from pytest_mock import MockerFixture

from ut_utils.create_client import get_client
from ut_utils.models import MockPrivilegeAuth

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from upgrade_service.upgrade_blueprint import https_upgrade_service_bp

from common.file_utils import FileUtils
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from upgrade_service.upgrade_serializer import GetUpgradeServiceActionSerializer
from upgrade_service.upgrade_serializer import GetUpgradeServiceResourceSerializer
from upload_mark_file import UploadMarkFile
from ut_utils.mock_utils import mock_path_exists


class TestUpgradeServiceView:
    client: FlaskClient = get_client(https_upgrade_service_bp)
    GetResourceCase = namedtuple("GetResourceCase", ["expect_status_code", "expect_data", "resource"])
    GetActionsCase = namedtuple("GetActionsCase", ["expect_status_code", "expect_data", "resource", "lib_interface"])
    UpgradeActionsCase = namedtuple("UpgradeActionsCase",
                                    ["expect_status_code", "expect_data", "headers", "body", "locked", "lib_interface",
                                     "exists"])
    ResetActionsCase = namedtuple("ResetActionsCase",
                                  ["expect_status_code", "expect_data", "headers", "body", "locked", "lib_interface",
                                   "resource"])
    use_cases = {
        "test_get_upgrade_service_resource": {
            "success": GetResourceCase(
                expect_status_code=200,
                expect_data={
                    "@odata.context": "/redfish/v1/$metadata#UpdateService",
                    "@odata.id": "/redfish/v1/UpdateService",
                    "@odata.type": "#UpdateService.v1_0_0.UpdateService",
                    "Id": "UpdateService",
                    "Name": "Update Service",
                    "Status": {
                        "Status": None, "Health": None
                    },
                    "ServiceEnabled": None,
                    "Actions": {
                        "#UpdateService.SimpleUpdate": {
                            "target": "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"
                        },
                        "#UpdateService.Reset": {
                            "target": "/redfish/v1/UpdateService/Actions/UpdateService.Reset"
                        }
                    },
                    "FirmwareInventory": {
                        "@odata.id": "/redfish/v1/UpdateService/FirmwareInventory"
                    }
                },
                resource=None,
            ),
            "invalid_resource": GetResourceCase(
                expect_status_code=404,
                expect_data={
                    "error": {
                        "code": "Base.1.0.ResourceDoesNotExists",
                        "message": "A ResourceDoesNotExists has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a resource change or creation was attempted but that the"
                                               " operation cannot proceed because the resource does not exists.",
                                "Message": "The requested resource does not exists.",
                                "Severity": "Critical",
                                "NumberOfArgs": 0,
                                "ParamTypes": None,
                                "Resolution": "Resource does not exists",
                                "Oem": {"status": None}
                            }
                        ]
                    }
                },
                resource="invalid resource data",
            ),
        },
        "test_get_upgrade_service_actions": {
            "success": GetActionsCase(
                expect_status_code=200,
                expect_data={
                    "@odata.context": "/redfish/v1/$metadata#TaskService/Tasks/Members/$entity",
                    "@odata.type": "#Task.v1_6_0.Task",
                    "@odata.id": "/redfish/v1/TaskService/Tasks/1",
                    "Id": "1",
                    "Name": "Upgrade Task",
                    "TaskState": "New",
                    "StartTime": "",
                    "Messages": {
                        "upgradeState": "ERR.0-1-Not upgraded"
                    },
                    "PercentComplete": 0,
                    "Module": "",
                    "Version": ""
                },
                resource=None,
                lib_interface={
                    "status": 200,
                    "message": {
                        "StartTime": "",
                        "TaskState": "New",
                        "Messages": {
                            "upgradeState": "ERR.0-1-Not upgraded"
                        },
                        "Id": "1",
                        "Name": "Upgrade Task",
                        "Version": "",
                        "PercentComplete": 0,
                        "Module": ""
                    }
                }
            ),
            "failed_with_invalid_resource": GetActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Socket path is not exist.",
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
                resource="invalid resource data",
                lib_interface=None,
            ),
            "failed_with_invalid_ret_dict": GetActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Socket path is not exist.",
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
                resource=None,
                lib_interface=None,
            ),
        },
        "test_rf_upgrade_service_actions": {
            "success": UpgradeActionsCase(
                expect_status_code=200,
                expect_data={
                    "@odata.context": "/redfish/v1/$metadata#TaskService/Tasks/Members/$entity",
                    "@odata.type": "#Task.v1_6_0.Task",
                    "@odata.id": "/redfish/v1/TaskService/Tasks/1",
                    "Id": "1",
                    "Name": "Upgrade Task",
                    "TaskState": "Running",
                    "StartTime": "2024-02-18 15:31:52",
                    "Messages": {
                        "upgradeState": "Running"
                    },
                    "PercentComplete": 0,
                    "Module": "",
                    "Version": ""
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ImageURI": "A500-A2-firmware_1.0.23.zip",
                    "TransferProtocol": "https"
                },
                locked=False,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": False}
                    },
                    {
                        "status": 200,
                        "message": {
                            "StartTime": "2024-02-18 15:31:52",
                            "TaskState": "Running",
                            "Messages": {
                                "upgradeState": "Running"
                            },
                            "Id": "1",
                            "Name": "Upgrade Task",
                            "Version": "",
                            "PercentComplete": 0,
                            "Module": ""
                        }
                    }
                ],
                exists=True,
            ),
            "high_risk_op_locked": UpgradeActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The operation is busy.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 100028
                            }
                        }]
                    }
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ImageURI": "A500-A2-firmware_1.0.23.zip",
                    "TransferProtocol": "https"
                },
                locked=True,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": False}
                    },
                    {
                        "status": 200,
                        "message": {
                            "StartTime": "2024-02-18 15:31:52",
                            "TaskState": "Running",
                            "Messages": {
                                "upgradeState": "Running"
                            },
                            "Id": "1",
                            "Name": "Upgrade Task",
                            "Version": "",
                            "PercentComplete": 0,
                            "Module": ""
                        }
                    }
                ],
                exists=True,
            ),
            "exclusive_status_busy": UpgradeActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The operation is busy.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 100028
                            }
                        }]
                    }
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ImageURI": "A500-A2-firmware_1.0.23.zip",
                    "TransferProtocol": "https"
                },
                locked=False,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": True}
                    },
                    {
                        "status": 200,
                        "message": {
                            "StartTime": "2024-02-18 15:31:52",
                            "TaskState": "Running",
                            "Messages": {
                                "upgradeState": "Running"
                            },
                            "Id": "1",
                            "Name": "Upgrade Task",
                            "Version": "",
                            "PercentComplete": 0,
                            "Module": ""
                        }
                    }
                ],
                exists=True,
            ),
            "zipfile_not_exist": UpgradeActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "A500-A2-firmware_1.0.23.zip not exist",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110105
                            }
                        }]
                    }
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ImageURI": "A500-A2-firmware_1.0.23.zip",
                    "TransferProtocol": "https"
                },
                locked=False,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": False}
                    },
                    {
                        "status": 200,
                        "message": {
                            "StartTime": "2024-02-18 15:31:52",
                            "TaskState": "Running",
                            "Messages": {
                                "upgradeState": "Running"
                            },
                            "Id": "1",
                            "Name": "Upgrade Task",
                            "Version": "",
                            "PercentComplete": 0,
                            "Module": ""
                        }
                    }
                ],
                exists=False,
            ),
            "exception_without_x_real_ip_in_headers": UpgradeActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Internal server error",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": None
                            }
                        }]
                    }
                },
                headers={"X-Auth-Token": "abc"},
                body={
                    "ImageURI": "A500-A2-firmware_1.0.23.zip",
                    "TransferProtocol": "https"
                },
                locked=False,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": False}
                    },
                    {
                        "status": 200,
                        "message": {
                            "StartTime": "2024-02-18 15:31:52",
                            "TaskState": "Running",
                            "Messages": {
                                "upgradeState": "Running"
                            },
                            "Id": "1",
                            "Name": "Upgrade Task",
                            "Version": "",
                            "PercentComplete": 0,
                            "Module": ""
                        }
                    }
                ],
                exists=True,
            ),
        },
        "test_rf_upgrade_reset_actions": {
            "success": ResetActionsCase(
                expect_status_code=200,
                expect_data={
                    "@odata.context": "/redfish/v1/$metadata#TaskService/Tasks/Members/$entity",
                    "@odata.type": "#Task.v1_6_0.Task",
                    "@odata.id": "/redfish/v1/TaskService/Tasks/1",
                    "Id": None,
                    "Name": "Upgrade Task",
                    "TaskState": None,
                    "StartTime": None,
                    "Messages": None,
                    "PercentComplete": None,
                    "Module": None,
                    "Version": None
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ResetType": "GracefulRestart"
                },
                locked=False,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": False}
                    },
                    {'status': 200, 'message': {}}
                ],
                resource=None,
            ),
            "high_risk_op_locked": ResetActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The operation is busy.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 100028
                            }
                        }]
                    }
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ResetType": "GracefulRestart"
                },
                locked=True,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": False}
                    },
                    {'status': 200, 'message': {}}
                ],
                resource=None,
            ),
            "exclusive_status_busy": ResetActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The operation is busy.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 100028
                            }
                        }]
                    }
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ResetType": "GracefulRestart"
                },
                locked=False,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": True}
                    },
                    {
                        "status": 200, "message": {}
                    }
                ],
                resource=None,
            ),
            "upgrade_effect_failed": ResetActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "ERR.02-Request data is invalid.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": None
                            }
                        }]
                    }
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ResetType": "GracefulRestart"
                },
                locked=False,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": False}
                    },
                    {
                        "status": 400, "message": "ERR.02-Request data is invalid."
                    }
                ],
                resource=None,
            ),
            "exception_with_invalid_resource_data": ResetActionsCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Internal server error",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": None
                            }
                        }]
                    }
                },
                headers={"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                body={
                    "ResetType": "GracefulRestart"
                },
                locked=False,
                lib_interface=[
                    {
                        "status": 200, "message": {"system_busy": False}
                    },
                    {
                        "status": 200, "message": {}
                    }
                ],
                resource="invalid format resource data",
            ),
        },
    }

    @classmethod
    def test_get_upgrade_service_resource(cls, mocker: MockerFixture, model: GetResourceCase):
        if model.resource:
            mocker.patch.object(GetUpgradeServiceResourceSerializer.service, "get_resource",
                                resturn_value=model.resource)

        resp = cls.client.get("/redfish/v1/UpdateService")
        assert json.loads(resp.data) == model.expect_data
        assert resp.status_code == model.expect_status_code

    @classmethod
    def test_get_upgrade_service_actions(cls, mocker: MockerFixture, model: GetActionsCase):
        if model.resource:
            mocker.patch.object(GetUpgradeServiceResourceSerializer.service, "get_resource",
                                resturn_value=model.resource)
        if model.lib_interface:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)

        resp = cls.client.get("/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate")
        assert json.loads(resp.data) == model.expect_data
        assert resp.status_code == model.expect_status_code

    @classmethod
    def test_rf_upgrade_service_actions(cls, mocker: MockerFixture, model: UpgradeActionsCase):
        mocker.patch.object(RedfishGlobals, "high_risk_exclusive_lock").locked.return_value = model.locked
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        mock_path_exists(mocker, return_value=model.exists)
        mocker.patch.object(shutil, "move")
        mocker.patch.object(UploadMarkFile, "clear_upload_mark_file")
        mocker.patch.object(FileUtils, "delete_dir_content")
        uri = "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"
        resp = cls.client.post(path=uri, headers=model.headers, data=json.dumps(model.body))
        assert json.loads(resp.data) == model.expect_data
        assert resp.status_code == model.expect_status_code

    @classmethod
    def test_rf_upgrade_reset_actions(cls, mocker: MockerFixture, model: ResetActionsCase):
        if model.resource:
            mocker.patch.object(GetUpgradeServiceActionSerializer.service, "get_resource",
                                resturn_value=model.resource)
        mocker.patch.object(RedfishGlobals, "high_risk_exclusive_lock").locked.return_value = model.locked
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        uri = "/redfish/v1/UpdateService/Actions/UpdateService.Reset"
        resp = cls.client.post(path=uri, headers=model.headers, data=json.dumps(model.body))
        assert json.loads(resp.data) == model.expect_data
        assert resp.status_code == model.expect_status_code
