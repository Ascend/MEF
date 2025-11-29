# coding: utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
import json
from collections import namedtuple
from unittest.mock import patch

from flask import Blueprint
from flask import Flask
from flask.testing import FlaskClient
from pytest_mock import MockerFixture

from ibma_redfish_serializer import SuccessMessageResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter
from system_service import lte_views
from system_service.systems_serializer import LteConfigInfoResourceSerializer
from system_service.systems_serializer import LteResourceSerializer
from system_service.systems_serializer import LteStatusInfoResourceSerializer
from ut_utils.models import MockPrivilegeAuth

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from system_service.systems_blueprint import system_bp

app = Flask(__name__)


def get_client(blueprint: Blueprint):
    app.register_blueprint(blueprint)
    app.testing = True
    return app.test_client()


class TestLteViews:
    client: FlaskClient = get_client(system_bp)
    GetLteInfoCase = namedtuple("GetLteInfoCase", ["expect_status_code", "expect_data", "resource"])
    GetLteStatusCase = namedtuple("GetLteStatusCase",
                                  ["expect_status_code", "expect_data", "resource", "lib_interface"])
    PatchLteStatusCase = namedtuple("PatchLteStatusCase",
                                    ["expect_status_code", "expect_data", "headers", "body", "locked", "resource",
                                     "lib_interface"])
    GetLteConfigCase = namedtuple("GetLteConfigCase",
                                  ["expect_status_code", "expect_data", "resource", "lib_interface"])
    PatchLteConfigCase = namedtuple("PatchLteConfigCase",
                                    ["expect_status_code", "expect_data", "headers", "body", "locked", "resource",
                                     "lib_interface"])
    use_cases = {
        "test_rf_get_system_lte": {
            "success": GetLteInfoCase(
                expect_status_code=200,
                expect_data={
                    "@odata.type": "#MindXEdgeLTE.v1_0_0.MindXEdgeLTE",
                    "@odata.context": "/redfish/v1/$metadata#Systems/LTE",
                    "@odata.id": "/redfish/v1/Systems/LTE",
                    "Id": "LTE",
                    "Name": "LTE",
                    "StatusInfo": {
                        "@odata.id": "/redfish/v1/Systems/LTE/StatusInfo"
                    },
                    "ConfigInfo": {
                        "@odata.id": "/redfish/v1/Systems/LTE/ConfigInfo"
                    }
                },
                resource=None,
            ),
            "exception_with_invalid_resource": GetLteInfoCase(
                expect_status_code=404,
                expect_data={
                    "error": {
                        "code": "Base.1.0.ResourceDoesNotExists",
                        "message": "A ResourceDoesNotExists has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [{
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a resource change or creation was attempted but that the "
                                           "operation cannot proceed because the resource does not exists.",
                            "Message": "The requested resource does not exists.",
                            "Severity": "Critical",
                            "NumberOfArgs": 0,
                            "ParamTypes": None,
                            "Resolution": "Resource does not exists",
                            "Oem": {
                                "status": None
                            }
                        }]
                    }
                },
                resource="invalid json format resource data",
            ),
        },
        "test_rf_get_system_lte_status_info": {
            "success": GetLteStatusCase(
                expect_status_code=200,
                expect_data={
                    "@odata.type": "#MindXEdgeLTE.v1_0_0.MindXEdgeLTE",
                    "@odata.context": "/redfish/v1/$metadata#Systems/LTE/StatusInfo",
                    "@odata.id": "/redfish/v1/Systems/LTE/StatusInfo",
                    "Id": "LTE StatusInfo",
                    "Name": "LTE StatusInfo",
                    "default_gateway": True,
                    "lte_enable": True,
                    "sim_exist": False,
                    "state_lte": None,
                    "state_data": None,
                    "network_signal_level": None,
                    "network_type": None,
                    "ip_addr": None
                },
                resource=None,
                lib_interface={
                    "status": 200,
                    "message": {
                        "default_gateway": True,
                        "lte_enable": True,
                        "sim_exist": False,
                        "state_data": None,
                        "state_lte": None,
                        "network_signal_level": None,
                        "network_type": None,
                        "ip_addr": None
                    }
                },
            ),
            "socket_path_not_exist": GetLteStatusCase(
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
            "exception_with_invalid_resource": GetLteStatusCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Query LTE status info failed.",
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
                resource="invalid json format resource data",
                lib_interface={
                    "status": 200,
                    "message": {
                        "default_gateway": True,
                        "lte_enable": True,
                        "sim_exist": False,
                        "state_data": None,
                        "state_lte": None,
                        "network_signal_level": None,
                        "network_type": None,
                        "ip_addr": None
                    }
                },
            ),
        },
        "test_rf_patch_system_lte_status_info": {
            "success": PatchLteStatusCase(
                expect_status_code=200,
                expect_data={
                    "@odata.type": "#MindXEdgeLTE.v1_0_0.MindXEdgeLTE",
                    "@odata.context": "/redfish/v1/$metadata#Systems/LTE/StatusInfo",
                    "@odata.id": "/redfish/v1/Systems/LTE/StatusInfo",
                    "Id": "LTE StatusInfo",
                    "Name": "LTE StatusInfo",
                    "default_gateway": False,
                    "lte_enable": True,
                    "sim_exist": True,
                    "state_lte": True,
                    "state_data": True,
                    "network_signal_level": 4,
                    "network_type": "4G",
                    "ip_addr": "10.10.10.10"
                },
                headers={
                    "X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"
                },
                body={
                    "state_lte": True,
                    "state_data": True
                },
                locked=False,
                resource=None,
                lib_interface={
                    "status": 200,
                    "message": {
                        "default_gateway": False,
                        "lte_enable": True,
                        "sim_exist": True,
                        "state_lte": True,
                        "state_data": True,
                        "network_signal_level": 4,
                        "network_type": "4G",
                        "ip_addr": "10.10.10.10"
                    }
                },
            ),
            "locked": PatchLteStatusCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Modify LTE status failed because LteView modify is busy.",
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
                headers={
                    "X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"
                },
                body={
                    "state_lte": True,
                    "state_data": True
                },
                locked=True,
                resource=None,
                lib_interface={
                    "status": 200,
                    "message": {
                        "default_gateway": False,
                        "lte_enable": True,
                        "sim_exist": True,
                        "state_lte": True,
                        "state_data": True,
                        "network_signal_level": 4,
                        "network_type": "4G",
                        "ip_addr": "10.10.10.10"
                    }
                },
            ),
            "exception_with_invalid_resource_data": PatchLteStatusCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Modify LTE status failed",
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
                headers={
                    "X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"
                },
                body={
                    "state_lte": True,
                    "state_data": True
                },
                locked=False,
                resource="invalid json format resource data",
                lib_interface={
                    "status": 200,
                    "message": {
                        "default_gateway": False,
                        "lte_enable": True,
                        "sim_exist": True,
                        "state_lte": True,
                        "state_data": True,
                        "network_signal_level": 4,
                        "network_type": "4G",
                        "ip_addr": "10.10.10.10"
                    }
                },
            ),
        },
        "test_rf_get_system_lte_config_info": {
            "success": GetLteConfigCase(
                expect_status_code=200,
                expect_data={
                    "@odata.type": "#MindXEdgeLTE.v1_0_0.MindXEdgeLTE",
                    "@odata.context": "/redfish/v1/$metadata#Systems/LTE/ConfigInfo",
                    "@odata.id": "/redfish/v1/Systems/LTE/ConfigInfo",
                    "Id": "LTE ConfigInfo",
                    "Name": "LTE ConfigInfo",
                    "apn_name": "ctnet",
                    "apn_user": "user",
                    "auth_type": "2",
                    "mode_type": 3
                },
                resource=None,
                lib_interface={
                    "status": 200,
                    "message": {
                        "apn_name": "ctnet",
                        "apn_user": "user",
                        "auth_type": "2",
                        "mode_type": 3
                    }
                }
            ),
            "exception_with_invalid_resource": GetLteConfigCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Query LTE config info failed.",
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
                resource="invalid json format resource data",
                lib_interface={
                    "status": 200,
                    "message": {
                        "apn_name": "ctnet",
                        "apn_user": "user",
                        "auth_type": "2",
                        "mode_type": 3
                    }
                }
            ),
        },
        "test_rf_patch_system_lte_config_info": {
            "success": PatchLteConfigCase(
                expect_status_code=202,
                expect_data={
                    "error": {
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Config LTE APN successfully.",
                                "Severity": "OK",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                headers={
                    "X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"
                },
                body={
                    "apn_name": "ctnet",
                    "apn_user": "user",
                    "apn_passwd": "1234",
                    "auth_type": "2"
                },
                locked=False,
                resource=None,
                lib_interface={
                    "status": 200,
                    "message": {
                        "apn_name": None,
                        "apn_user": None,
                        "auth_type": "0",
                        "mode_type": None
                    }
                }
            ),
            "locked": PatchLteConfigCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Set LTE config info failed because LteView modify is busy.",
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
                headers={
                    "X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"
                },
                body={
                    "apn_name": "ctnet",
                    "apn_user": "user",
                    "apn_passwd": "1234",
                    "auth_type": "2"
                },
                locked=True,
                resource=None,
                lib_interface={
                    "status": 200,
                    "message": {
                        "apn_name": None,
                        "apn_user": None,
                        "auth_type": "0",
                        "mode_type": None
                    }
                }
            ),
            "exception_with_invalid_resource_data": PatchLteConfigCase(
                expect_status_code=400,
                expect_data={
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Config LTE APN failed.", "Severity": "Critical",
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
                headers={
                    "X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"
                },
                body={
                    "apn_name": "ctnet",
                    "apn_user": "user",
                    "apn_passwd": "1234",
                    "auth_type": "2"
                },
                locked=False,
                resource="invalid json format resource data",
                lib_interface={
                    "status": 200,
                    "message": {
                        "apn_name": None,
                        "apn_user": None,
                        "auth_type": "0",
                        "mode_type": None
                    }
                }
            ),
            "invalid_lib_interface_result": PatchLteConfigCase(
                expect_status_code=400,
                expect_data={
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
                                    "status": None
                                }
                            }
                        ]
                    }
                },
                headers={
                    "X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"
                },
                body={
                    "apn_name": "ctnet",
                    "apn_user": "user",
                    "apn_passwd": "1234",
                    "auth_type": "2"
                },
                locked=False,
                resource=None,
                lib_interface={
                    "status": 400,
                    "message": "Parameter is invalid."
                }
            ),
        },
    }

    @classmethod
    def test_rf_get_system_lte(cls, mocker: MockerFixture, model: GetLteInfoCase):
        if model.resource:
            mocker.patch.object(LteResourceSerializer.service, "get_resource", return_value=model.resource)
        resp = cls.client.get(path="/redfish/v1/Systems/LTE")
        assert model.expect_data == resp.get_json(force=True)
        assert model.expect_status_code == resp.status_code

    @classmethod
    def test_rf_get_system_lte_status_info(cls, mocker: MockerFixture, model: GetLteStatusCase):
        if model.resource:
            mocker.patch.object(LteStatusInfoResourceSerializer.service, "get_resource", return_value=model.resource)
        if model.lib_interface:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        resp = cls.client.get(path="/redfish/v1/Systems/LTE/StatusInfo")
        assert model.expect_data == resp.get_json(force=True)
        assert model.expect_status_code == resp.status_code

    @classmethod
    def test_rf_patch_system_lte_status_info(cls, mocker: MockerFixture, model: PatchLteStatusCase):
        if model.resource:
            mocker.patch.object(LteStatusInfoResourceSerializer.service, "get_resource", return_value=model.resource)
        if model.lib_interface:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)

        mocker.patch.object(lte_views, "LTE_STATUS_INFO_LOCK").locked.return_value = model.locked
        uri = "/redfish/v1/Systems/LTE/StatusInfo"
        resp = cls.client.patch(path=uri, headers=model.headers, data=json.dumps(model.body))
        assert model.expect_data == resp.get_json(force=True)
        assert model.expect_status_code == resp.status_code

    @classmethod
    def test_rf_get_system_lte_config_info(cls, mocker: MockerFixture, model: GetLteConfigCase):
        if model.resource:
            mocker.patch.object(LteConfigInfoResourceSerializer.service, "get_resource", return_value=model.resource)
        if model.lib_interface:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        resp = cls.client.get(path="/redfish/v1/Systems/LTE/ConfigInfo")
        assert model.expect_data == resp.get_json(force=True)
        assert model.expect_status_code == resp.status_code

    @classmethod
    def test_rf_patch_system_lte_config_info(cls, mocker: MockerFixture, model: PatchLteConfigCase):
        if model.resource:
            mocker.patch.object(SuccessMessageResourceSerializer.service, "get_resource", return_value=model.resource)
        if model.lib_interface:
            mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        mocker.patch.object(lte_views, "LTE_CONFIG_INFO_LOCK").locked.return_value = model.locked
        uri = "/redfish/v1/Systems/LTE/ConfigInfo"
        resp = cls.client.patch(path=uri, headers=model.headers, data=json.dumps(model.body))
        assert model.expect_data == resp.get_json(force=True)
        assert model.expect_status_code == resp.status_code
