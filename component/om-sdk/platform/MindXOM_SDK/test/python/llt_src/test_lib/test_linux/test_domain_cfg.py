from collections import namedtuple

from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.app_common_method import AppCommonMethod
from common.checkers import IPV4Checker
from common.utils.result_base import Result
from ut_utils.mock_utils import mock_cdll

with mock_cdll():
    from lib.Linux.EdgeSystem.domain_cfg import DomainCfg

TestHostConfig = namedtuple("TestHostConfig", "expect, exist_count, check_is_link, host_list")
TestGetHostList = namedtuple("TestGetHostList", "expect, f, check_ipv4_format, domain_check")
TestRequestDataCheck = namedtuple("TestRequestDataCheck",
                                  "expect, request_data_dict, is_forbidden_domain")


class TestDomainCfg:
    use_cases = {
        "test_request_data_check": {
            "normal": ([0, "ok"], {'static_host_list': [], "name_server": [{"ip_address": "51.38.69.171"}],
                                   "net_manager_domain": "fd.fusiondirector.huawei.com"}, False),
            "invalid_when_request_data_dict_is_none": ([1, "request parameter is empty"], None, False),
            "invalid_when_name_server_isnot_list": ([2, "invalid parameter"],
                                                    {'static_host_list': [], "name_server": 1,
                                                     "net_manager_domain": "fd.fusiondirector.huawei.com"}, False),
            "invalid_when_len_name_server_larger_than_3":
                ([3, "invalid list len"], {'static_host_list': [], "name_server": [1, 2, 3, 4],
                                           "net_manager_domain": "fd.fusiondirector.huawei.com"}, False),
            "invalid_when_first_check_ipv4_format_is_false":
                ([3, "invalid ip format"], {'static_host_list': [{"ip_address": "127.0.0.0"}],
                                            "name_server": [{"ip_address": "51.38.69.171"}],
                                            "net_manager_domain": "fd.fusiondirector.huawei.com"}, False),
            "invalid_when_domain_check_is_false":
                ([4, "invalid domain format"], {'static_host_list': [{"ip_address": "51.38.69.171", "name": 1}],
                                                "name_server": [{"ip_address": "51.38.69.171"}],
                                                "net_manager_domain": "fd.fusiondirector.huawei.com"}, False),
            "invalid_when_is_forbidden_domain_is_true":
                ([4, "invalid domain format"], {'static_host_list': [{"ip_address": "51.38.69.171", "name": "abc123"}],
                                                "name_server": [{"ip_address": "51.38.69.171"}],
                                                "net_manager_domain": "fd.fusiondirector.huawei.com"}, True),
            "invalid_when_second_check_ipv4_format_is_false":
                ([5, "invalid ip format"], {'static_host_list': [{"ip_address": "51.38.69.171", "name": "localhost"}],
                                            "name_server": [{"ip_address": "127.0.0.0"}],
                                            "net_manager_domain": "fd.fusiondirector.huawei.com"}, False),
        },
        "test_get_host_list": {
            "invalid_when_check_ipv4_format_is_false":
                ([],
                 ("127.0.0.1   Euler localhost localhost.localdomain localhost4 localhost4.localdomain4",
                  "::1         Euler localhost localhost.localdomain localhost6 localhost6.localdomain6",
                  "51.38.66.223 fd.fusiondirector.huawei.com"
                  ),
                 False, True),
            "invalid_when_domain_check_is_false":
                ([],
                 ("127.0.0.1   Euler localhost localhost.localdomain localhost4 localhost4.localdomain4",
                  "::1         Euler localhost localhost.localdomain localhost6 localhost6.localdomain6",
                  "51.38.66.223 fd.fusiondirector.huawei.com"
                  ),
                 True, False),
        },
        "test_host_config": {
            "normal": (None, 0, Result(True), ({"name": "a", "ip_address": "a"}, )),
            "invalid_when_exist_count_is_120": (None, 120, Result(True), ({"name": "a", "ip_address": "a"}, )),
            "invalid_when_check_is_link_is_false": (None, 0, Result(False), ({"name": "a", "ip_address": "a"}, )),
        },
    }

    def test_request_data_check(self, mocker: MockerFixture, model: TestRequestDataCheck):
        mocker.patch.object(DomainCfg, "is_forbidden_domain", return_value=model.is_forbidden_domain)
        domain_cfg = DomainCfg()
        assert domain_cfg.request_data_check(model.request_data_dict) == model.expect

    def test_get_host_list(self, mocker: MockerFixture, model: TestGetHostList):
        mocker.patch.object(AppCommonMethod, "check_ipv4_format", return_value=model.check_ipv4_format)
        mocker.patch.object(IPV4Checker, "domain_check", return_value=model.domain_check)
        domain_cfg = DomainCfg()
        assert domain_cfg.get_host_list() == model.expect

    def test_host_config(self, mocker: MockerFixture, model: TestHostConfig):
        mocker.patch.object(FileCheck, "check_is_link", return_value=model.check_is_link)
        domain_cfg = DomainCfg()
        assert domain_cfg.host_config(model.host_list, "fd.fusiondirector.huawei.com") == model.expect
