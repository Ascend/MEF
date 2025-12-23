from collections import namedtuple

from pytest_mock import MockerFixture

from common.checkers import IPV4Checker, CheckResult
from common.checkers import NotExistsChecker, PasswordComplexityChecker, \
    ExtensionChecker, GatewayChecker, LogNameChecker, TimeHourMinChecker, Ipv4WithMaskChecker, MacAddrChecker, \
    DateChecker
from common.net_check import NetCheck
from conftest import TestBase


class TestChecker(TestBase):
    CheckDictCase = namedtuple("CheckDictCase",
                               "tested_class, excepted_success, excepted_reason, param_key, input_param")
    CheckDictLengthCase = namedtuple("CheckDictCase", "tested_class, excepted_success, "
                                                      "excepted_reason, param_key, input_param, min_len, max_len")
    CheckDictChoiceCase = namedtuple("CheckDictCase", "tested_class, excepted_success, "
                                                      "excepted_reason, param_key, input_param, choice")
    CheckPasswordOkCase = namedtuple("CheckPasswordOkCase", "excepted, param")

    CheckGateWayCase = namedtuple("CheckDictCase", "excepted_success, excepted_reason, "
                                                   "is_same_network_segment, get_num_for_mask, net_work_address")
    CheckTimeCase = namedtuple("CheckTimeCase", "excepted, time")
    CheckIPV4Case = namedtuple("CheckIPV4Case", "excepted_success, excepted_reason, param_key, input_param, mock_check")
    CheckMacAddrCase = namedtuple("CheckMacAddrCase",
                                  "excepted_success, excepted_reason, param_key, "
                                  "input_param, mac_addr_single_cast_check")
    CheckFormatTimeCase = namedtuple("CheckFormatTimeCase", "excepted, input_time")

    use_cases = {

        "test_check_dict": {
            "not_exists_checker_is_existed": (NotExistsChecker, False, "NotExistsChecker: 1 exists", "1", {"1": "1"}),
            "not_exists_checker_not_existed": (NotExistsChecker, True, "", "2", {"1": "1"}),
            "password_complexity_checker_complexity_illegal": (PasswordComplexityChecker, False,
                                                               "Insufficient password complexity!", "password",
                                                               {"password": "12345abcd"}),
            "password_complexity_checker_complexity_legal": (PasswordComplexityChecker, True, "", "password",
                                                             {"password": "W12345abcd"}),
            "log_name_checker_empty": (LogNameChecker, False, "Log names no values", "log_name", {"log_name": ""}),
            "log_name_checker_repeated_name": (LogNameChecker, False, "Log names have duplicate values",
                                               "log_name", {"log_name": "a b c a"}),
            "log_name_checker_not_issubset": (LogNameChecker, False, "Log names wrong",
                                              "log_name", {"log_name": "a b c"}),
            "log_name_checker_success": (LogNameChecker, True, "", "log_name", {"log_name": "NPU"}),
            "time_hour_min_checker_value_illegal": (TimeHourMinChecker, False,
                                                    "Invalid time format", "time", {"time": "12.12"}),
            "time_hour_min_checker_value_legal": (TimeHourMinChecker, True, "", "time", {"time": "12:12"}),
            "mac_addr_checker_len_wrong": (MacAddrChecker, False, "Parameter mac_addr len invalid",
                                           "mac_addr", {"mac_addr": "fc:1b:d1:6d:be:8"}),
            "date_checker_len_wrong": (DateChecker, False, "", "date", {"date": "2022"}),
            "date_checker_format_wrong": (DateChecker, False, "", "date", {"date": "2022:11:11"}),
            "date_checker_success": (DateChecker, True, "", "date", {"date": "2022-11-11"}),

        },
        "test_check_length_dict": {
            "password_complexity_checker_length_wrong": (
                PasswordComplexityChecker, False, "String length checker: invalid length of password", "password",
                {"password": "1"}, 8, 20)
        },
        "test_check_choice_dict": {
            "extension_checker_choices_failed": (ExtensionChecker, False,
                                                 "ExtensionChecker: Invalid extension of file_name",
                                                 "file_name", {"file_name": "abcdefg.cfg"}, ("txt",)),
            "extension_checker_choices_success": (ExtensionChecker, True, "",
                                                  "file_name", {"file_name": "abcdefg.txt"}, ("txt",)),
        },
        "test_check_dict_gate_way": {
            "same_network_segment_fail": (False, "The gateway does not match the ip and subnet mask",
                                          False, "32", None),
            "net_work_address_fail": (False, "The gateway is the same as network address",
                                      True, "32", "90.90.11.1"),
            "net_work_address_success": (True, "", True, "32", "90.90.11.111"),
        },
        "test_check_time_hour_min": {
            "without_colon": (False, "test"),
            "item_num_wrong": (False, "12:12:12"),
            "item_len_wrong": (False, "1" * 129 + ":12"),
            "hh_type_wrong": (False, "aa:12"),
            "hh_value_wrong": (False, "-1:12"),
            "mm_value_wrong": (False, "12:61"),
            "hh_24_mm_not_zero": (False, "24:10"),
            "success": (True, "24:00"),
        },
        "test_ipv4_check_dict": {
            "ipv4_check_failed": (False, "Parameter ip_addr with mask invalid",
                                  "ip", {"ip": "1/123"}, CheckResult.make_failed("failed")),
            "mask_num_len_wrong": (False, "Parameter mask invalid",
                                   "ip", {"ip": "123/" + "123" * 50}, CheckResult.make_success()),
            "mask_num_type_wrong": (False, "extract ip/mask info fail",
                                    "ip", {"ip": "123/a"}, CheckResult.make_success()),
            "mask_num_value_wrong": (False, "Parameter mask invalid",
                                     "ip", {"ip": "123/33"}, CheckResult.make_success()),
            "without_slash_ipv4_check_failed": (False, "Parameter single ip_addr invalid",
                                                "ip", {"ip": "123"}, CheckResult.make_failed("failed")),
            "start_127": (False, "Parameter single ip_addr reserved",
                          "ip", {"ip": "127"}, CheckResult.make_success()),
            "success": (True, "", "ip", {"ip": "51.38.69.69"}, CheckResult.make_success()),
        },
        "test_mac_addr_check_dict": {
            "single_check_failed": (False, "Parameter mac_addr error", "mac_addr",
                                    {"mac_addr": "fc:1b:d1:6d:be:89"}, False),
            "single_check_success": (True, "", "mac_addr", {"mac_addr": "fc:1b:d1:6d:be:89"}, True),
        },
        "test_check_format_time": {
            "wrong": (False, "2022:12:12"),
            "normal": (True, "2022-12-12")
        }
    }

    def test_check_dict(self, model: CheckDictCase):
        ret = model.tested_class(model.param_key).check_dict(model.input_param)
        assert model.excepted_success == ret.success
        if model.excepted_reason:
            assert model.excepted_reason == ret.reason

    def test_check_length_dict(self, model: CheckDictLengthCase):
        ret = model.tested_class(model.param_key, model.min_len, model.max_len).check_dict(model.input_param)
        assert model.excepted_success == ret.success
        assert model.excepted_reason == ret.reason

    def test_check_choice_dict(self, model: CheckDictChoiceCase):
        ret = model.tested_class(model.param_key, model.choice).check_dict(model.input_param)
        assert model.excepted_success == ret.success
        assert model.excepted_reason == ret.reason

    def test_check_dict_gate_way(self, mocker: MockerFixture, model: CheckGateWayCase):
        mocker.patch.object(NetCheck, "is_same_network_segment", return_value=model.is_same_network_segment)
        mocker.patch.object(NetCheck, "get_num_for_mask", return_value=model.get_num_for_mask)
        mocker.patch.object(NetCheck, "net_work_address", return_value=model.net_work_address)
        gateway_data = {"Address": "90.90.11.11",
                        "Gateway": "90.90.11.1",
                        "SubnetMask": "1"}
        ret = GatewayChecker().check_dict(gateway_data)
        assert model.excepted_success == ret.success
        assert model.excepted_reason == ret.reason

    def test_check_time_hour_min(self, model: CheckTimeCase):
        ret = TimeHourMinChecker.check_time_hour_min(model.time)
        assert model.excepted == ret

    def test_ipv4_check_dict(self, mocker: MockerFixture, model: CheckIPV4Case):
        mocker.patch.object(IPV4Checker, "check", return_value=model.mock_check)
        ret = Ipv4WithMaskChecker(model.param_key).check_dict(model.input_param)
        assert model.excepted_success == ret.success
        assert model.excepted_reason == ret.reason

    def test_mac_addr_check_dict(self, mocker: MockerFixture, model: CheckMacAddrCase):
        mocker.patch.object(NetCheck, "mac_addr_single_cast_check", return_value=model.mac_addr_single_cast_check)
        ret = MacAddrChecker(model.param_key).check_dict(model.input_param)
        assert model.excepted_success == ret.success
        assert model.excepted_reason == ret.reason

    def test_check_format_time(self, model: CheckFormatTimeCase):
        ret = DateChecker()._check_format_time(model.input_time)
        assert model.excepted == ret[0]
