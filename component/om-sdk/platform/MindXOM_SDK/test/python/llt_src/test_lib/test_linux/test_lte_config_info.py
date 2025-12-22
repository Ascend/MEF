import configparser
from collections import namedtuple

from common.file_utils import FileCheck
from common.utils.result_base import Result
from pytest_mock import MockerFixture
from ut_utils.mock_utils import mock_cdll

from common.constants.error_codes import CommonErrorCodes
from common.common_methods import CommonMethods

with mock_cdll():
    from lib.Linux.systems.lte_config_info import LteConfigInfo
    from devm.device_mgr import DEVM
    from devm.exception import DeviceManagerError
    from devm.device_mgr import Device

PACKAGE_NAME = "lib.Linux.systems.lte_config_info"

CheckApnName = namedtuple("CheckApnName", "expect, mode, name_len")
PatchRequest = namedtuple("PatchRequest", "expect, get_attribute, set_attribute, lock, path_valid, sections, "
                                          "bool, request, write")


class TestLteConfigInfo:
    use_cases = {
        "test_check_apn_name": {
            "0-99": ([CommonMethods.ERROR, CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey], 0, 100),
            "1-40": ([CommonMethods.ERROR, CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey], 1, 40),
            "2-39": ([CommonMethods.ERROR, CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey], 2, 40),
            "3-39": ([CommonMethods.ERROR, CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey], 3, 40),
            "OK": ([CommonMethods.OK, ""], 3, 39),
        },
        "test_patch_request": {
            "null lte": (
                [CommonMethods.ERROR, "Lte does not exist."], [False, None, None], None, None, None, None, None, None,
                None
            ),
            "busy": (
                [CommonMethods.ERROR, "Lte config is busy"], [True, None, None], None, True, None, None, None, None,
                None
            ),
            "path valid": (
                [CommonMethods.ERROR, "Lte config reads failed"], [True, None, None], None, False, Result(False),
                None, None, None, None
            ),
            "can't operate APN": (
                [CommonMethods.ERROR, "ERR.007,Lte data is open, can not operate APN!"], [True, True, None], None,
                False, Result(True), ["lte_apn"], [True, True], None, None
            ),
            "parameter is invalid": (
                [CommonMethods.ERROR, "Parameter is invalid."], [True, False, None], None, False, Result(True),
                ["lte_apn"], [True, True],
                {"apn_name": "", "apn_user": "apnuser@#!", "apn_passwd": "123456", "auth_type": "4"}, None
            ),
            "invalid apn name": (
                [CommonMethods.ERROR, "Parameter is invalid."], [True, False, 1], None, False, Result(True),
                ["lte_apn"], [True, True],
                {"apn_name": "apn_name111111111111111111111111111111111111111111111",
                 "apn_user": "", "apn_passwd": "", "auth_type": "3"}, None
            ),
            "set apn info exception": (
                [CommonMethods.ERROR, "Parameter is invalid."], [True, False, 1], DeviceManagerError(), False,
                Result(True), ["lte_apn"], [True, True],
                {"apn_name": "apn_name", "apn_user": "apnuser", "apn_passwd": "123456", "auth_type": "3"}, None
            ),
            "ok": (
                [CommonMethods.OK, ], [True, False, 3], None, False, Result(True), ["lte_apn"], [True, True],
                {"apn_name": "apn_name", "apn_user": "", "apn_passwd": "", "auth_type": "3"}, None

            ),
            "write err": (
                [CommonMethods.ERROR, "Parameter is invalid."], [True, False, 1], None, False, Result(True),
                ["lte_apn"], [True, True],
                {"apn_name": "apn_name", "apn_user": "", "apn_passwd": "", "auth_type": "3"}, Exception(),

            )
        }
    }

    @staticmethod
    def test_check_apn_name(model: CheckApnName):
        assert LteConfigInfo.check_apn_name(model.mode, model.name_len) == model.expect

    @staticmethod
    def test_patch_request(mocker: MockerFixture, model: PatchRequest):
        mocker.patch.object(DEVM, "get_device").return_value = Device
        mocker.patch.object(Device, "get_attribute", side_effect=model.get_attribute)
        mocker.patch.object(Device, "set_attribute", side_effect=model.set_attribute)
        mocker.patch.object(LteConfigInfo, "LTE_CONFIG_LOCK").locked.return_value = model.lock
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.path_valid)
        config = mocker.patch.object(configparser, "RawConfigParser").return_value
        config.sections.return_value = model.sections
        config.getboolean.side_effect = model.bool
        mocker.patch.object(FileCheck, "check_is_link_exception")
        mocker.patch("os.open")
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = model.write
        assert LteConfigInfo().patch_request(model.request) == model.expect
