# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
import os
from collections import namedtuple
from unittest.mock import patch

from flask.testing import FlaskClient
from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.result_base import Result
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service import security_service_views
from ut_utils.create_client import TestUtils
from ut_utils.create_client import get_client
from ut_utils.models import MockPrivilegeAuth

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from user_manager.user_manager import UserManager
    from system_service.systems_blueprint import system_bp

GetSecurityRules = namedtuple("GetSecurityRules", "expect, lib_interface, code")
PatchSecurityRules = namedtuple("PatchSecurityRules", "expect, lock, data, check_pwd, lib_interface, code")
ExportSecurityRules = namedtuple("ExportSecurityRules", "code, lib_interface, check_path, send")
ImportSecurityRules = namedtuple("ImportSecurityRules", "expect, lock, data, check_pwd, lib_interface, code")


class TestSecurityRules:
    client: FlaskClient = get_client(system_bp)
    use_cases = {
        "test_get_security_rules": {
            "success": (
                {"@odata.context": "/redfish/v1/$metadata#Systems/SecurityService/SecurityLoad",
                 "@odata.id": "/redfish/v1/Systems/SecurityService/SecurityLoad",
                 "@odata.type": "#MindXEdgeSecurityService.v1_0_0.MindXEdgeSecurityService",
                 "Id": "SecurityLoad",
                 "Name": "Security Load",
                 "load_cfg": [],
                 "Actions": {
                     "#SecurityLoad.Import": {
                         "target": "/redfish/v1/Systems/SecurityService/SecurityLoad/Actions/SecurityLoad.Import"
                     },
                     "#SecurityLoad.Export": {
                         "target": "/redfish/v1/Systems/SecurityService/SecurityLoad/Actions/SecurityLoad.Export"
                     }
                 }},
                [{"status": 200, "message": {"load_cfg": []}}], 200
            ),
            "failed": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Send message failed.",
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
                [{"status": 400, "message": "Send message failed."}], 400
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
                                "Message": "Query SecurityLoad info failed.",
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
                Exception(), 400
            ),
        },
        "test_patch_security_rules": {
            "success": (
                {
                    "error": {
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Config security load successfully.",
                                "Severity": "OK",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                False, {"Password": "Edge@9000", "load_cfg": [{"enable": "false", "ip_addr": None}]},
                {"status": 200, "message": ""}, [{"status": 200, "message": ""}], 202
            ),
            "locked": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Config security load failed because SecurityServiceViews modify is busy.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "Oem": {"status": None},
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                True, {"Password": "Edge@9000", "load_cfg": [{"enable": "false", "ip_addr": None}]}, None, None, 400
            ),
            "not json": (
                {
                    "error": {
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON.  "
                                               "Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON "
                                           "and could not be parsed by the receiving service.",
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
                False, "123", None, None, 400
            ),
            "data wrong": (
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
                False, {"Password": None, "load_cfg": [{"enable": "false", "ip_addr": None}]}, None, None, 400
            ),
            "pwd wrong": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "The user name or password error.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 110207
                                }
                            }
                        ]
                    }
                },
                False, {"Password": "Edge@9001", "load_cfg": [{"enable": "false", "ip_addr": None}]},
                {"status": 400, "message": [110207, "The user name or password error."]}, None, 400
            ),
            "config failed": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Config security load failed.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100025
                                }
                            }
                        ]
                    }
                },
                False, {"Password": "Edge@9001", "load_cfg": [{"enable": "false", "ip_addr": None}]},
                {"status": 200, "message": ""},
                [{"status": 400, "message": [100025, "Config security load failed."]}],
                400
            ),
            "config exception": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Config security load failed.",
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
                False, {"Password": "Edge@9001", "load_cfg": [{"enable": "false", "ip_addr": None}]},
                {"status": 200, "message": ""}, Exception(), 400
            )
        },
        "test_export_security_rules": {
            "check pass failed": (400, {"status": 200, "message": ""}, Result(False, "not exist"), None),
            "export failed": (400, {"status": 400, "message": "err"}, Result(True), None)
        },
        "test_import_security_rules": {
            "success": (
                {
                    "error": {
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Import configuration of security load successfully.",
                                "Severity": "OK",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                False, {"file_name": "session_sec_cfg.ini", "Password": "Edge@9000"},
                {"status": 200, "message": ""}, {"status": 200, "message": ""}, 202
            ),
            "locked": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Import configuration of security load failed "
                                           "because SecurityServiceViews modify is busy.",
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
                True, {"file_name": "session_sec_cfg.ini", "Password": "Edge@9000"},
                None, None, 400
            ),
            "not json": (
                {
                    "error": {
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON.  "
                                               "Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON and "
                                           "could not be parsed by the receiving service.",
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
                False, [], None, None, 400
            ),
            "data wrong": (
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
                False, {"file_name": "session_sec_cfg.in", "Password": "Edge@9000"},
                None, None, 400
            ),
            "pwd wrong": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "The user name or password error.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 110207
                                }
                            }
                        ]
                    }
                },
                False, {"file_name": "session_sec_cfg.ini", "Password": "Edge@9001"},
                {"status": 400, "message": [110207, "The user name or password error."]}, None, 400
            ),
            "import failed": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Import configuration of security load failed.",
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
                False, {"file_name": "session_sec_cfg.ini", "Password": "Edge@9001"},
                {"status": 200, "message": ""},
                [{"status": 400, "message": "Import configuration of security load failed."}], 400
            ),
            "import exception": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Import configuration of security load failed.",
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
                False, {"file_name": "session_sec_cfg.ini", "Password": "Edge@9001"},
                {"status": 200, "message": ""}, Exception(), 400
            )
        }

    }

    def test_get_security_rules(self, mocker: MockerFixture, model: GetSecurityRules):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        response = TestSecurityRules.client.get("/redfish/v1/Systems/SecurityService/SecurityLoad")
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    def test_patch_security_rules(self, mocker: MockerFixture, model: PatchSecurityRules):
        mocker.patch.object(security_service_views, "SECURITY_LOAD_LOCK").locked.return_value = model.lock
        mocker.patch.object(UserManager, "get_all_info", return_value=model.check_pwd)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        security_service_views.g = TestUtils
        response = TestSecurityRules.client.patch("/redfish/v1/Systems/SecurityService/SecurityLoad",
                                                  data=json.dumps(model.data))
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    def test_export_security_rules(self, mocker: MockerFixture, model: ExportSecurityRules):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=model.check_path)
        mocker.patch.object(security_service_views, "send_from_directory", return_value=model.send)
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch("os.remove")
        response = TestSecurityRules.client.post(
            "/redfish/v1/Systems/SecurityService/SecurityLoad/Actions/SecurityLoad.Export")
        # 导出文件只比较code
        assert response.status_code == model.code

    def test_import_security_rules(self, mocker: MockerFixture, model: ImportSecurityRules):
        mocker.patch.object(security_service_views, "IMPORT_SECURITY_LOAD_LOCK").locked.return_value = model.lock
        mocker.patch.object(UserManager, "get_all_info", return_value=model.check_pwd)
        mocker.patch.object(os.path, "exists", return_value=True)
        mocker.patch("os.remove")
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        system_bp.g = TestUtils
        response = TestSecurityRules.client.post(
            "/redfish/v1/Systems/SecurityService/SecurityLoad/Actions/SecurityLoad.Import",
            data=json.dumps(model.data))
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect