import ssl
from collections import namedtuple
from unittest.mock import mock_open
from unittest.mock import patch

import pytest
from pytest_mock import MockerFixture

from common.checkers import CheckResult
from common.checkers.param_checker import NTPServiceChecker
from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.file_utils import FileUtils
from common.kmc_lib.kmc import Kmc
from common.kmc_lib.tlsconfig import TlsConfig
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result
from ibma_redfish_globals import RedfishGlobals
from ut_utils.models import MockPrivilegeAuth

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from ibma_redfish_urls import RedfishURIs

from system_service.systems_common import lib_rest_check_status, modify_null_tag_to_default

GetUserLockedStateCase = namedtuple("GetUserLockedStateCase", "expected, ret_dict")
SendOperationalLogCase = namedtuple("SendOperationalLogCase", "print_log")
CheckExternalParameterCase = namedtuple("CheckExternalParameterCase", "expected, check")
CreateDirCase = namedtuple("CreateDirCase", "expected, create_dir")
CheckUploadFileSizeCase = namedtuple("CheckUploadFileSizeCase", "expected, filename, file_size")
SaveFileCase = namedtuple("SaveFileCase", "expected, delete_dir, create_file, read, disk_usage")
SaveFileExceptionCase = namedtuple("SaveFileExceptionCase", "read, disk_usage")
FileRequestErrorCase = namedtuple("FileRequestErrorCase", "expected, error_msg")
ModifyNullTagToDefaultCase = namedtuple("FileRequestErrorCase", "expected, request_dict")
GetSslContextCase = namedtuple("GetSslContextCase", "check_path, read_data, get_ssl_context")


class TestRedfishURIs:
    use_cases = {
        "test_get_user_locked_state": {
            "state_110206": (True, {"status": 400, "message": [110206, "test"]},),
            "state_not_110206": (False, {"status": 400, "message": [110207, "test"]},)
        },
        "test_send_operational_log": {
            "print_log": (True,),
            "print_log_not": (False,)
        },
        "test_check_external_parameter": {
            "success": (None, CheckResult(True, "success")),
            "failed": ({"status": 400, "message": [100024, "Parameter is invalid."]}, CheckResult(False, "failed"))
        },
        "test_create_all_required_upload_dir": {
            "success": (True, Result(True)),
            "failed": (False, Result(False))
        },
        "test_save_file": {
            "delete_failed": (False, False, None, None, None),
            "create_file_failed": (False, True, Result(False), None, None),
            "normal": (True, True, Result(True), b'testtesttesttest', 1024)
        },
        "test_file_request_error": {
            "errno_28": ("The size of uploaded file is larger than the remaining capacity.",
                         "[Errno 28] No space left on device"),
            "errno_not_28": ("request imgfile failed:failed", "failed"),

        },
        "test_modify_null_tag_to_default": {
            "addresses_not_list": ({"IPv4Addresses": "test"}, {"IPv4Addresses": "test"}),
            "address_not_dict": ({"IPv4Addresses": ["test"]}, {"IPv4Addresses": ["test"]}),
            "normal": ({"IPv4Addresses": [{'Tag': '123'}]}, {"IPv4Addresses": [{"Tag": "123"}]}),
        },
        "test_lib_rest_check_status": {
            "success": (True, {"status": 200}),
            "failed": (False, {"status": 300})
        },
        "test_get_ssl_context_exception": {
            "check_path_failed": (Result(False), None, None),
            "get_ssl_context_failed": (Result(True), "test", [False, ssl.SSLContext()]),
        }
    }

    request = namedtuple("request", "method, headers")

    def test_get_user_locked_state(self, model: GetUserLockedStateCase):
        ret = RedfishGlobals.get_user_locked_state(model.ret_dict)
        assert model.expected == ret

    def test_send_operational_log(self, mocker: MockerFixture, model: SendOperationalLogCase):
        mocker.patch.object(RedfishGlobals, "set_operational_log")
        TestRedfishURIs.request.method = "PATCH"
        ret = RedfishURIs.send_operational_log(TestRedfishURIs.request, "test", model.print_log, "admin")
        assert ret is None

    def test_check_external_parameter(self, mocker: MockerFixture, model: CheckExternalParameterCase):
        mocker.patch.object(RedfishGlobals, "make_error_response", return_value="test")
        mocker.patch.object(NTPServiceChecker, "check", return_value=model.check)
        ret = RedfishGlobals.check_external_parameter(NTPServiceChecker, "test")
        assert model.expected == ret

    def test_check_external_parameter_exception(self, mocker: MockerFixture):
        mocker.patch.object(RedfishGlobals, "make_error_response", return_value="exception")
        mocker.patch.object(NTPServiceChecker, "check", side_effect=Exception())
        ret = RedfishGlobals.check_external_parameter(NTPServiceChecker, "exception")
        assert ret == {"status": 400, "message": [100024, "Parameter is invalid."]}

    def test_create_all_required_upload_dir(self, mocker: MockerFixture, model: CreateDirCase):
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        ret = RedfishGlobals.create_all_required_upload_dir()
        assert model.expected == ret

    def test_save_file(self, mocker: MockerFixture, model: SaveFileCase):
        flask_file = namedtuple("flask_file", "filename, read")
        flask_file.filename = "test.conf"
        mocker.patch.object(flask_file, "read", side_effect=[model.read, b''])
        usage = namedtuple('usage', 'free')
        usage.free = model.disk_usage
        mocker.patch.object(FileUtils, "delete_dir_content", return_value=model.delete_dir)
        mocker.patch.object(FileCreate, "create_file", return_value=model.create_file)
        mocker.patch("os.fdopen")
        mocker.patch("os.open")
        mocker.patch("shutil.disk_usage", return_value=usage)
        mocker.patch.object(AppCommonMethod, "force_remove_file", return_value=model.create_file)
        ret = RedfishGlobals.save_file(flask_file)
        assert model.expected == ret

    def test_file_request_error(self, model: FileRequestErrorCase):
        ret = RedfishURIs.file_request_error(model.error_msg)
        assert model.expected == ret

    def test_modify_null_tag_to_default(self, model: ModifyNullTagToDefaultCase):
        ret = modify_null_tag_to_default(model.request_dict)
        assert model.expected == ret

    def test_lib_rest_check_status(self, model: GetUserLockedStateCase):
        request_data_dict = {"DateTimeLocalOffset": True}
        ret = lib_rest_check_status(request_data_dict, model.ret_dict)
        assert model.expected == ret

    def test_start_flask(self, mocker: MockerFixture):
        mocker.patch.object(RedfishGlobals, "get_http_port", return_value=8001)
        app = namedtuple("app", "run")
        mocker.patch.object(app, "run")
        ret = RedfishURIs._start_flask(app, "test")
        assert ret is None

    def test_get_ssl_context_exception(self, mocker: MockerFixture, model: GetSslContextCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("builtins.open", mock_open(read_data=model.read_data))
        mocker.patch.object(TlsConfig, "get_ssl_context", return_value=model.get_ssl_context)
        mocker.patch.object(Kmc, "decrypt", return_value="test")
        with pytest.raises(Exception):
            RedfishURIs._get_ssl_context()

    def test_get_ssl_context(self, mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(True))
        mocker.patch("builtins.open", mock_open(read_data="TEST"))
        mocker.patch.object(TlsConfig, "get_ssl_context", return_value=[True, ssl.SSLContext()])
        mocker.patch.object(Kmc, "decrypt", return_value="test")
        ret = RedfishURIs._get_ssl_context()
        assert ssl.OP_NO_TLSv1_2 not in ret.options
