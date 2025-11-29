from collections import namedtuple

import pytest

from common.file_utils import FileCheck
from common.utils.result_base import Result
from pytest_mock import MockerFixture

from ut_utils.mock_utils import mock_cdll

with mock_cdll():
    from lib.Linux.systems.security_service.security_service_clib \
        import get_cert_info, get_http_cert_effective

PACKAGE_PATH = "lib.Linux.systems.security_service.security_service_clib"
DEFAULT_CERT_PATH = "/home/data/config/default/server_kmc.cert"

GetCertInfo = namedtuple("GetCertInfo", "expect, available, path_valid, info")
GetCertInfoExc = namedtuple("GetCertInfoExc", "available, path_valid, info")
GetHttpCertEffective = namedtuple("GetHttpCertEffective", "expect, path_valid, effective")
GetHttpCertEffectiveExc = namedtuple("GetHttpCertEffectiveExc", "path_valid, effective")
SetCustomCertificate = namedtuple("SetCustomCertificate", "expect, init, ret_set, verify")


class TestSecurityServiceClib:
    use_cases = {
        "test_get_cert_info": {
            "success": ([b"", b"", b"", b"", b""], [True, True], Result(True), 0),
        },
        "test_get_cert_info_exception": {
            "file not exists": ([False, False], None, None),
            "check path failed": ([True, True], Result(False), None),
            "get info failed": ([True, True], Result(True), -1),
        },
        "test_get_http_cert_effective": {
            "expire": ([0, 0], Result(True), 0x20140009),
            "normal": ([0, 0], Result(True), 0),
        },
        "test_get_http_cert_effective_exception": {
            "check path failed": (Result(False), None),
            "func err": (Result(True), -1),
        }
    }

    @staticmethod
    def test_get_cert_info(mocker: MockerFixture, model: GetCertInfo):
        mocker.patch(f"{PACKAGE_PATH}.check_lib_available", side_effect=model.available)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.path_valid)
        mocker.patch(f"{PACKAGE_PATH}.getcertinfo", return_value=model.info)
        assert get_cert_info(DEFAULT_CERT_PATH) == model.expect

    @staticmethod
    def test_get_cert_info_exception(mocker: MockerFixture, model: GetCertInfoExc):
        mocker.patch(f"{PACKAGE_PATH}.check_lib_available", side_effect=model.available)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.path_valid)
        mocker.patch(f"{PACKAGE_PATH}.getcertinfo", return_value=model.info)
        with pytest.raises(Exception):
            get_cert_info(DEFAULT_CERT_PATH)

    @staticmethod
    def test_get_http_cert_effective(mocker: MockerFixture, model: GetHttpCertEffective):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.path_valid)
        mocker.patch(f"{PACKAGE_PATH}.get_cert_effective", return_value=model.effective)
        assert get_http_cert_effective(DEFAULT_CERT_PATH) == model.expect

    @staticmethod
    def test_get_http_cert_effective_exception(mocker: MockerFixture, model: GetHttpCertEffectiveExc):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.path_valid)
        mocker.patch(f"{PACKAGE_PATH}.get_cert_effective", return_value=model.effective)
        with pytest.raises(Exception):
            get_http_cert_effective(DEFAULT_CERT_PATH)
