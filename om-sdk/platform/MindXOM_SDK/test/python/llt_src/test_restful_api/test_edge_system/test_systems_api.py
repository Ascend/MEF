# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
import os
from collections import namedtuple
from unittest.mock import patch

from flask.testing import FlaskClient
from pytest_mock import MockerFixture

from common.file_utils import FileCheck, FileUtils, FileCreate
from common.utils.result_base import Result
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service import alarm_views
from system_service import nfs_views
from system_service import security_service_views
from ut_utils.create_client import get_client, TestUtils
from ut_utils.models import MockPrivilegeAuth

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from user_manager.user_manager import UserManager
    from system_service.systems_blueprint import system_bp

# Security Rules
GetSecurityRules = namedtuple("GetSecurityRules", "expect, lib_interface, code")
PatchSecurityRules = namedtuple("PatchSecurityRules", "expect, lock, data, check_pwd, lib_interface, code")
ExportSecurityRules = namedtuple("ExportSecurityRules", "code, lib_interface, check_path, send")
ImportSecurityRules = namedtuple("ImportSecurityRules", "expect, lock, data, check_pwd, lib_interface, code")
ImportPunyDict = namedtuple("ImportPunyDict", "expect, lock, check_pwd, lib_interface, code, data")
ExportPunyDict = namedtuple("ExportPunyDict", "expect, lock, lib_interface, check_path, code")
DeletePunyDict = namedtuple("DeletePunyDict", "expect, data, lock, get_all_info, lib_interface, code")
DownloadCSRFile = namedtuple("DeletePunyDict", "expect, lock, create_dir, lib_interface, send, code")

# Alarm
GetSystemAlarm = namedtuple("GetSystemAlarm", "get_resource, code")
GetSystemAlarmInfo = namedtuple("GetSystemAlarmInfo", "expect, lib_interface, code")
GetSystemAlarmShield = namedtuple("GetSystemAlarmShield", "expect, lib_interface, code")
IncreaseSystemAlarmShield = namedtuple("IncreaseSystemAlarmShield", "data, lock, expect, headers, lib_interface, code")
DecreaseSystemAlarmShield = namedtuple("DecreaseSystemAlarmShield", "data, lock, expect, headers, lib_interface, code")

# Nfs
GetNfs = namedtuple("GetNfs", "expect, lib_interface, code")
MountNfs = namedtuple("MountNfs", "headers, data, expect, lock, lib_interface, code")
UnmountNfs = namedtuple("UnmountNfs", "headers, data, expect, lock, lib_interface, code")

# Reset
SystemReset = namedtuple("SystemReset", "headers, data, expect, lock, lib_intf_exclusive_status, code")
RestoreDefaults = namedtuple("RestoreDefaults", "headers, data, expect, lock, lib_intf_exclusive_status, code")

client: FlaskClient = get_client(system_bp)


class TestSecurityServiceViews:
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
                                "Description": "Indicates that the request body was malformed JSON."
                                               "  Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON"
                                           " and could not be parsed by the receiving service.",
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
        },
        "test_import_puny_dict": {
            "locked": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Import puny dict failed because SecurityServiceViews import is busy.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "Oem": {"status": None},
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                True, None, None, 400, {"FileName": "import.conf", "Password": "Huawei123"}
            ),
            "not_json": (
                {
                    "error": {
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON."
                                               "  Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON"
                                           " and could not be parsed by the receiving service.",
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
                False, None, None, 400, "asd"
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
                False, None, None, 400, {"FileName": "\\", "Password": "Huawei@456"}
            ),
            "pwd_invalid": (
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
                False, {"status": 400, "message": [110207, "The user name or password error."]}, None, 400,
                {"FileName": "test123.conf", "Password": "Huawei@456"}
            ),
            "admin_locked": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "User lock state locked.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 110206
                                }
                            }
                        ]
                    }
                },
                False, {"status": 400, "message": [110206, "User lock state locked."]}, None, 400,
                {"FileName": "test123.conf", "Password": "Huawei@456"}
            ),
            "normal_failed": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Import puny dict failed.",
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
                False, {"status": 200, "message": ""}, {"status": 400, "message": "Import puny dict failed."}, 400,
                {"FileName": "test123.conf", "Password": "Huawei@456"}
            ),
            "succeed": (
                {
                    "error": {
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Import puny dict successfully.",
                                "Severity": "OK",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                False, {"status": 200, "message": ""}, {"status": 200, "message": {""}},
                202, {"FileName": "test123.conf", "Password": "Huawei@456"}
            ),
        },
        "test_export_puny_dict": {
            "locked": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Export puny dict is busy.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "Oem": {"status": None},
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                }, True, [None, None], None, 400
            ),
            "invalid_ret": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Export puny dict failed.",
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
                }, False, [{"status": 400, "message": "Export puny dict failed."}, None], None, 400
            ),
            "ERR.008": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "File not exist.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 110105
                                }
                            }
                        ]
                    }
                }, False, [{"status": 400, "message": "ERR.008."}, None], None, 400
            ),
            "invalid_path": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "export puny dict failed",
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
                }, False, [{"status": 200, "message": ""}, None], False, 400
            ),
            "Exception": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "export puny dict failed",
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
                }, False, [{"status": 200, "message": ""}, Exception], False, 400
            )
        },
        "test_delete_puny_dict": {
            "locked": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Delete puny dict failed because SecurityServiceViews Delete is busy.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "Oem": {"status": None},
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                {"Password": "Huawei123"}, True, None, None, 400
            ),
            "not_json": (
                {
                    "error": {
                        "code": "Base.1.0.MalformedJSON",
                        "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that the request body was malformed JSON."
                                               "  Could be duplicate, syntax error,etc.",
                                "Message": "The request body submitted was malformed JSON"
                                           " and could not be parsed by the receiving service.",
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
                111, False, None, None, 400
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
                {"test": "Huawei123"}, False, None, None, 400
            ),
            "pwd_invalid": (
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
                {"Password": "Huawei123"}, False,
                {"status": 400, "message": [110207, "The user name or password error."]}, None, 400
            ),
            "admin_locked": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "User lock state locked.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 110206
                                }
                            }
                        ]
                    }
                },
                {"Password": "Huawei123"}, False, {"status": 400, "message": [110206, "User lock state locked."]},
                None, 400
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
                                "Message": "Delete puny dict failed.",
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
                {"Password": "Huawei123"}, False, {"status": 200, "message": ""},
                {"status": 400, "message": "Delete puny dict failed."}, 400
            ),
            "succeed": (
                {
                    "error": {
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Delete puny dict successfully.",
                                "Severity": "OK",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                {"Password": "Huawei123"}, False, {"status": 200, "message": ""},
                {"status": 200, "message": {""}}, 202
            ),
        },
        "test_download_csr_file": {
            "locked": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "SecurityServiceViews export is busy.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "Oem": {"status": None},
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                True, None, None, None, 400
            ),
            "create_failed": (
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Create path run_web_cert failed.",
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
                False, False, None, None, 400
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
                                "Message": "export csr failed",
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
                False, True, Exception, None, 400
            ),
        },
    }

    @staticmethod
    def test_get_security_rules(mocker: MockerFixture, model: GetSecurityRules):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        response = client.get("/redfish/v1/Systems/SecurityService/SecurityLoad")
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_patch_security_rules(mocker: MockerFixture, model: PatchSecurityRules):
        mocker.patch.object(security_service_views, "SECURITY_LOAD_LOCK").locked.return_value = model.lock
        mocker.patch.object(UserManager, "get_all_info", return_value=model.check_pwd)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        security_service_views.g = TestUtils
        response = client.patch("/redfish/v1/Systems/SecurityService/SecurityLoad", data=json.dumps(model.data))
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_export_security_rules(mocker: MockerFixture, model: ExportSecurityRules):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=model.check_path)
        mocker.patch.object(security_service_views, "send_from_directory", return_value=model.send)
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch("os.remove")
        response = client.post("/redfish/v1/Systems/SecurityService/SecurityLoad/Actions/SecurityLoad.Export")
        # 导出文件只比较code
        assert response.status_code == model.code

    @staticmethod
    def test_import_security_rules(mocker: MockerFixture, model: ImportSecurityRules):
        mocker.patch.object(security_service_views, "IMPORT_SECURITY_LOAD_LOCK").locked.return_value = model.lock
        mocker.patch.object(UserManager, "get_all_info", return_value=model.check_pwd)
        mocker.patch.object(os.path, "exists", return_value=True)
        mocker.patch("os.remove")
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        system_bp.g = TestUtils
        response = client.post("/redfish/v1/Systems/SecurityService/SecurityLoad/Actions/SecurityLoad.Import",
                               data=json.dumps(model.data))
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_import_puny_dict(mocker: MockerFixture, model: ImportPunyDict):
        mocker.patch.object(security_service_views, "PUNY_DICT_IMPORT_LOCK").locked.return_value = model.lock
        mocker.patch.object(UserManager, "get_all_info", return_value=model.check_pwd)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        mocker.patch.object(RedfishGlobals, "get_user_locked_state", return_value=True)
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch("os.path.join")
        mocker.patch("os.rename")
        security_service_views.g = TestUtils
        resp = client.post("/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictImport",
                           data=json.dumps(model.data))
        assert model.expect == resp.get_json(force=True)
        assert resp.status_code == model.code

    @staticmethod
    def test_export_puny_dict(mocker: MockerFixture, model: ExportPunyDict):
        mocker.patch.object(security_service_views, "PUNY_DICT_EXPORT_LOCK").locked.return_value = model.lock
        mocker.patch.object(FileUtils, "delete_full_dir")
        mocker.patch.object(FileCreate, "create_dir")
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("shutil.copyfile")
        resp = client.post("/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictExport")
        assert resp.status_code == model.code
        assert resp.get_json(force=True) == model.expect

    @staticmethod
    def test_delete_puny_dict(mocker: MockerFixture, model: DeletePunyDict):
        mocker.patch.object(security_service_views, "PUNY_DICT_DELETE_LOCK").locked.return_value = model.lock
        mocker.patch.object(UserManager, "get_all_info", return_value=model.get_all_info)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        mocker.patch.object(RedfishGlobals, "get_user_locked_state", return_value=True)
        security_service_views.g = TestUtils
        resp = client.post("/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictDelete",
                           data=json.dumps(model.data))
        assert model.expect == resp.get_json(force=True)
        assert resp.status_code == model.code

    @staticmethod
    def test_download_csr_file(mocker: MockerFixture, model: DownloadCSRFile):
        mocker.patch.object(security_service_views, "DOWNLOAD_CSR_LOCK").locked.return_value = model.lock
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        mocker.patch("os.path.exists", return_value=False)
        mocker.patch("os.open")
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = "abc"
        mocker.patch.object(security_service_views, "send_from_directory", return_value=model.send)
        resp = client.post("/redfish/v1/Systems/SecurityService/downloadCSRFile")
        assert resp.status_code == model.code
        assert model.expect == resp.get_json(force=True)


class TestAlarmViews:
    use_cases = {
        "test_get_system_alarm": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Alarm",
                    "@odata.id": "/redfish/v1/Systems/Alarm",
                    "@odata.type": "MindXEdgeAlarm.v1_0_0.MindXEdgeAlarm",
                    "Id" : "Alarm",
                    "Name": "Alarm",
                    "AlarmInfo": {
                        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmInfo"
                    },
                    "AlarmShield": {
                        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield"
                    }
                },
                200,
            ),
            "failed": (
                Exception,
                500,
            )
        },
        "test_get_system_alarm_info": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Alarm/AlarmInfo",
                    "@odata.id": "/redfish/v1/Systems/Alarm/AlarmInfo",
                    "@odata.type": "MindXEdgeAlarm.v1_0_0.MindXEdgeAlarm",
                    "Id": "Alarm Info",
                    "Name": "Alarm Info",
                    "AlarMessages": []
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
                        "Message": "Internal server error",
                        "NumberOfArgs": None,
                        "Oem": {
                          "status": 100011
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
                [{"status": 400, "message": "Send message failed."}], 500
            )
        },
        "test_get_system_alarm_shield": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Alarm/AlarmShield",
                    "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield",
                    "@odata.type": "MindXEdgeAlarm.v1_0_0.MindXEdgeAlarm",
                    "Id": "Alarm Shield",
                    "Name": "Alarm Shield",
                    "AlarmShieldMessages": [],
                    "Increase": {
                        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield/Increase"
                    },
                    "Decrease": {
                        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield/Decrease"
                    }
                },
                [{"status": 200, "message": {"AlarMessages": []}}], 200,
            ),
            "failed": (
                {
                  "error": {
                    "@Message.ExtendedInfo": [
                      {
                        "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                        "Description": "Indicates that a general error has occurred.",
                        "Message": "Internal server error",
                        "NumberOfArgs": None,
                        "Oem": {
                          "status": 100011
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
                [{"status": 400, "message": "Send message failed."}], 500
            )
        },
        "test_increase_system_alarm_shield": {
            "success": (
                {
                    "AlarmShieldMessages":
                        [
                            {
                                "UniquelyIdentifies": "asdas",
                                "AlarmId": "00000000",
                                "PerceivedSeverity": "1",
                                "AlarmInstance": "MCU"
                            }
                        ]
                },
                False,
                {
                    "error": {
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Increase alarm shield successfully.",
                                "Severity": "OK",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 200, "message": ""}], 200,
            ),
            "failed": (
                {
                    "AlarmShieldMessages":
                        [
                            {
                                "UniquelyIdentifies": "asdas",
                                "AlarmId": "00000000",
                                "PerceivedSeverity": "1",
                                "AlarmInstance": "MCU"
                            }
                        ]
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
                                "Message": "",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": None},
                            }
                        ]
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 400, "message": ""}], 400,
            ),
            "locked": (
                {
                    "AlarmShieldMessages":
                        [
                            {
                                "UniquelyIdentifies": "asdas",
                                "AlarmId": "00000000",
                                "PerceivedSeverity": "1",
                                "AlarmInstance": "MCU"
                            }
                        ]
                },
                True,
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Increase alarm shield failed because increasing alarm shield is busy.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": None},
                            }
                        ]
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 400, "message": ""}], 400,
            ),
        },
        "test_decrease_system_alarm_shield": {
            "success": (
                {
                    "AlarmShieldMessages":
                        [
                            {
                                "UniquelyIdentifies": "asdas",
                                "AlarmId": "00000000",
                                "PerceivedSeverity": "1",
                                "AlarmInstance": "MCU"
                            }
                        ]
                },
                False,
                {
                    "error": {
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Decrease alarm shield successfully.",
                                "Severity": "OK",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 200, "message": ""}], 200,
            ),
            "failed": (
                {
                    "AlarmShieldMessages":
                        [
                            {
                                "UniquelyIdentifies": "asdas",
                                "AlarmId": "00000000",
                                "PerceivedSeverity": "1",
                                "AlarmInstance": "MCU"
                            }
                        ]
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
                                "Message": "",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": None},
                            }
                        ]
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 400, "message": ""}], 400,
            ),
            "locked": (
                {
                    "AlarmShieldMessages":
                        [
                            {
                                "UniquelyIdentifies": "asdas",
                                "AlarmId": "00000000",
                                "PerceivedSeverity": "1",
                                "AlarmInstance": "MCU"
                            }
                        ]
                },
                True,
                {
                    "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Decrease alarm shield failed because increasing alarm shield is busy.",
                                "Severity": "Critical",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": None},
                            }
                        ]
                    }
                },
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                [{"status": 400, "message": ""}], 400,
            ),
        },
    }

    @staticmethod
    def test_get_system_alarm(mocker: MockerFixture, model: GetSystemAlarm):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        response = client.get("/redfish/v1/Systems/Alarm")
        assert response.status_code == model.code

    @staticmethod
    def test_get_system_alarm_info(mocker: MockerFixture, model: GetSystemAlarmInfo):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        response = client.get("/redfish/v1/Systems/Alarm/AlarmInfo")
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_get_system_alarm_shield(mocker: MockerFixture, model: GetSystemAlarmShield):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        response = client.get("/redfish/v1/Systems/Alarm/AlarmShield")
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_increase_system_alarm_shield(mocker: MockerFixture, model: IncreaseSystemAlarmShield):
        mocker.patch.object(alarm_views, "INCREASE_ALARM_SHIELD").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        response = client.patch("/redfish/v1/Systems/Alarm/AlarmShield/Increase",
                                data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_decrease_system_alarm_shield(mocker: MockerFixture, model: DecreaseSystemAlarmShield):
        mocker.patch.object(alarm_views, "DECREASE_ALARM_SHIELD").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        response = client.patch("/redfish/v1/Systems/Alarm/AlarmShield/Decrease",
                                data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect


class TestNfsViews:
    use_cases = {
        "test_get_nfs": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Members/NfsManage",
                    "@odata.type": "#MindXEdgeNfsManage.v1_0_0.MindXEdgeNfsManage",
                    "@odata.id": "/redfish/v1/Systems/NfsManage",
                    "Id": "1",
                    "Name": "Nfs Manage",
                    "nfsList": [],
                    "Actions": {
                        "#NfsManage.Mount": {
                          "target": "/redfish/v1/Systems/NfsManage/Actions/NfsManage.Mount"
                        },
                        "#NfsManage.Unmount": {
                          "target": "/redfish/v1/Systems/NfsManage/Actions/NfsManage.Unmount"
                        }
                    }
                },
                [{"status": 200, "message": {"AlarMessages": []}}],
                200,
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
            )
        },
        "test_mount_nfs": {
            "locked": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Mount NFS request failed because lock is locked.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100028
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
                True, None, 400
            ),
            "invalid_param": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "ServerIP": "serverIp",
                    "ServerDir": "serverPath",
                    "FileSystem": "version",
                    "MountPath": "mountPath",
                },
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100024
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
                False, None, 400
            ),
            "modify_failed": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "ServerIP": "192.168.2.108",
                    "ServerDir": "/home",
                    "FileSystem": "nfs4",
                    "MountPath": "/home/test",
                },
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
                False, [{"status": 400, "message": "Send message failed."}], 400
            ),
            "modify_success": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "ServerIP": "192.168.2.108",
                    "ServerDir": "/home",
                    "FileSystem": "nfs4",
                    "MountPath": "/home/test",
                },
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Mount NFS successfully.",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Severity": "OK"
                            }
                        ],
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information."
                    }
                },
                False, [{"status": 200, "message": "OK."}], 200
            ),
        },
        "test_unmount_nfs": {
            "locked": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Unmount NFS request failed because lock is locked.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100028
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
                True, None, 400
            ),
            "invalid_param": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "ServerIP": "serverIp",
                    "ServerDir": "serverPath",
                    "FileSystem": "version",
                    "MountPath": "mountPath",
                },
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100024
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
                False, None, 400
            ),
            "modify_failed": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "ServerIP": "192.168.2.108",
                    "ServerDir": "/home",
                    "FileSystem": "nfs4",
                    "MountPath": "/home/test",
                },
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
                False, [{"status": 400, "message": "Send message failed."}], 400
            ),
            "modify_success": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {
                    "ServerIP": "192.168.2.108",
                    "ServerDir": "/home",
                    "FileSystem": "nfs4",
                    "MountPath": "/home/test",
                },
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Unmount NFS successfully.",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Severity": "OK"
                            }
                        ],
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information."
                    }
                },
                False, [{"status": 200, "message": "OK."}], 200
            ),
        },
    }

    @staticmethod
    def test_get_nfs(mocker: MockerFixture, model: GetNfs):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        response = client.get("/redfish/v1/Systems/NfsManage")
        assert response.status_code == model.code and response.get_json(force=True) == model.expect

    @staticmethod
    def test_mount_nfs(mocker: MockerFixture, model: MountNfs):
        mocker.patch.object(nfs_views, "MOUNT_NFS_LOCK").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        url = "/redfish/v1/Systems/NfsManage/Actions/NfsManage.Mount"
        response = client.post(url, data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_unmount_nfs(mocker: MockerFixture, model: UnmountNfs):
        mocker.patch.object(nfs_views, "UNMOUNT_NFS_LOCK").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_interface)
        url = "/redfish/v1/Systems/NfsManage/Actions/NfsManage.Unmount"
        response = client.post(url, data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect


class TestActionViews:
    use_cases = {
        "test_system_reset": {
            "locked": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "The operation is busy.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100028
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
                True, None, 400
            ),
            "get_exclusive_status_failed": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "get ExclusiveStatus failed.",
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
                False, {"status": 400, "message": "get ExclusiveStatus failed."}, 400
            ),
            "get_message_failed": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "The operation is busy.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100028
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
                False, {"status": 200, "message": "get ExclusiveStatus success."}, 400
            ),
            "invalid_param": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {"ResetType": None},
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100024
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
                False, {"status": 200, "message": {"system_busy": False}}, 400
            ),
            "actions_success": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {"ResetType": "GracefulRestart"},
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Restart system (GracefulRestart) successfully.",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Severity": "OK"
                            }
                        ],
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information."
                    }
                },
                False,
                {"status": 200, "message": {"system_busy": False}}, 200
            ),
        },
        "test_restore_defaults": {
            "locked": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "The operation is busy.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100028
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
                True, None, 400
            ),
            "get_exclusive_status_failed": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "get ExclusiveStatus failed.",
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
                False, {"status": 400, "message": "get ExclusiveStatus failed."}, 400
            ),
            "get_message_failed": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                None,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "The operation is busy.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100028
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
                False, {"status": 200, "message": "get ExclusiveStatus success."}, 400
            ),
            "invalid_param": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {"ethernet": ":eth00", "root_pwd": "password"},
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100024
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
                False, {"status": 200, "message": {"system_busy": False}}, 400
            ),
            "restore_success": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                {"ethernet": "eth0", "root_pwd": "password"},
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Restore defaults system successfully.",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Severity": "OK"
                            }
                        ],
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information."
                    }
                },
                False,
                {"status": 200, "message": {"system_busy": False}}, 200
            ),
        },
    }

    @staticmethod
    def test_system_reset(mocker: MockerFixture, model: SystemReset):
        mocker.patch.object(RedfishGlobals, "high_risk_exclusive_lock").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_intf_exclusive_status)
        mocker.patch("sys.exit")
        url = "/redfish/v1/Systems/Actions/ComputerSystem.Reset"
        response = client.post(url, data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    @staticmethod
    def test_restore_defaults(mocker: MockerFixture, model: RestoreDefaults):
        mocker.patch.object(RedfishGlobals, "high_risk_exclusive_lock").locked.return_value = model.lock
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_intf_exclusive_status)
        url = "/redfish/v1/Systems/Actions/RestoreDefaults.Reset"
        response = client.post(url, data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

